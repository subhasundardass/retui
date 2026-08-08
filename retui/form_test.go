package retui

import (
	"errors"
	"sync"
	"testing"
)

// personForm is the struct used across most tests below.
type personForm struct {
	Name  string
	Email string
	Age   int
}

// renderFrame simulates one render pass: BeginRender() resets hook cursors
// exactly like the production render loop, so repeated calls to UseForm (or
// any other hook) in the SAME order return the SAME stored value.
func renderFrame(t *testing.T, fn func()) {
	t.Helper()
	BeginRender()
	fn()
}

// setupForm resets all global hook state before a test and schedules the
// same reset after it, so tests don't leak State/Effects/pendingRender into
// each other. Not safe to run with t.Parallel() — hook state is a package
// global, same as production.
func setupForm(t *testing.T) {
	t.Helper()
	ResetComponentState()
	t.Cleanup(ResetComponentState)
}

// ============================================================================
// UseForm: creation & reuse
// ============================================================================

func TestUseForm_PersistsAcrossRenders(t *testing.T) {
	setupForm(t)

	var first, second *Form[personForm]

	renderFrame(t, func() {
		first = UseForm(personForm{Name: "Alice"})
	})
	renderFrame(t, func() {
		second = UseForm(personForm{Name: "ignored on reuse"})
	})

	if first != second {
		t.Fatalf("expected UseForm to return the same *Form[T] pointer across renders")
	}
	if got := second.Values().Name; got != "Alice" {
		t.Fatalf("expected value from first render to persist, got %q", got)
	}
}

func TestUseForm_MultipleFormsAtDifferentCallSites(t *testing.T) {
	setupForm(t)

	var formA *Form[personForm]
	var formB *Form[struct{ Query string }]

	renderFrame(t, func() {
		formA = UseForm(personForm{Name: "A"})
		formB = UseForm(struct{ Query string }{Query: "B"})
	})

	if formA.Values().Name != "A" {
		t.Fatalf("formA got wrong value: %+v", formA.Values())
	}
	if formB.Values().Query != "B" {
		t.Fatalf("formB got wrong value: %+v", formB.Values())
	}
}

func TestUseForm_TypeMismatchAtSlotResets(t *testing.T) {
	setupForm(t)

	renderFrame(t, func() {
		UseForm(personForm{Name: "Alice"})
	})

	// Simulate a hook-order bug: a different type shows up at the same
	// call-site index. UseRef's type-mismatch branch should reset the slot
	// rather than panic, and UseForm should build a fresh Form on top of it.
	var reset *Form[struct{ X int }]
	renderFrame(t, func() {
		reset = UseForm(struct{ X int }{X: 42})
	})

	if reset.Values().X != 42 {
		t.Fatalf("expected fresh form after type mismatch, got %+v", reset.Values())
	}
}

// ============================================================================
// Values / SetValues
// ============================================================================

func TestForm_Values_ReturnsCopy(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{Name: "Alice"}) })

	v := f.Values()
	v.Name = "mutated locally"

	if f.Values().Name != "Alice" {
		t.Fatalf("Values() should return a copy; internal state was mutated")
	}
}

func TestForm_SetValues_NoopWhenEqual(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{Name: "Alice"}) })

	IsPendingRender() // clear any pending flag from setup

	f.SetValues(personForm{Name: "Alice"})

	if f.IsDirty() {
		t.Fatalf("SetValues with an equal value should not mark the form dirty")
	}
	if IsPendingRender() {
		t.Fatalf("SetValues with an equal value should not request a render")
	}
}

func TestForm_SetValues_MarksDirtyAndRequestsRender(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{Name: "Alice"}) })

	IsPendingRender() // clear

	f.SetValues(personForm{Name: "Bob"})

	if !f.IsDirty() {
		t.Fatalf("expected form to be dirty after SetValues with a new value")
	}
	if f.Values().Name != "Bob" {
		t.Fatalf("expected updated value, got %q", f.Values().Name)
	}
	if !IsPendingRender() {
		t.Fatalf("expected SetValues to request a render")
	}
}

func TestForm_SetValues_CallsOnChange(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{Name: "Alice"}) })

	var got personForm
	calls := 0
	f.OnChange(func(v personForm) {
		calls++
		got = v
	})

	f.SetValues(personForm{Name: "Bob"})

	if calls != 1 {
		t.Fatalf("expected OnChange to be called once, got %d", calls)
	}
	if got.Name != "Bob" {
		t.Fatalf("expected OnChange to receive the new value, got %q", got.Name)
	}
}

// ============================================================================
// SetField / Field
// ============================================================================

func TestForm_SetField_UpdatesValueAndMarksTouched(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{}) })

	if err := f.SetField("Name", "Alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if f.Values().Name != "Alice" {
		t.Fatalf("expected Name to be set, got %q", f.Values().Name)
	}
	if !f.Touched("Name") {
		t.Fatalf("expected Name to be marked touched")
	}
	if f.Touched("Email") {
		t.Fatalf("expected Email to NOT be touched")
	}
	if !f.IsDirty() {
		t.Fatalf("expected form to be dirty after SetField")
	}
}

func TestForm_SetField_UnknownField(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{}) })

	if err := f.SetField("Nickname", "Al"); err == nil {
		t.Fatalf("expected error for unknown field")
	}
}

func TestForm_SetField_TypeMismatch(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{}) })

	if err := f.SetField("Age", "not an int"); err == nil {
		t.Fatalf("expected error assigning string to int field")
	}
}

func TestForm_SetField_NonStructT(t *testing.T) {
	setupForm(t)
	var f *Form[string]
	renderFrame(t, func() { f = UseForm("hello") })

	if err := f.SetField("Anything", "value"); err == nil {
		t.Fatalf("expected error calling SetField when T is not a struct")
	}
}

func TestForm_Field_Get(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{Name: "Alice", Age: 30}) })

	v, ok := f.Field("Age")
	if !ok {
		t.Fatalf("expected Field to find Age")
	}
	if v.(int) != 30 {
		t.Fatalf("expected Age 30, got %v", v)
	}

	if _, ok := f.Field("Nonexistent"); ok {
		t.Fatalf("expected Field to return false for an unknown field")
	}
}

// ============================================================================
// Touched / Reset / SetInitial
// ============================================================================

func TestForm_MarkTouched_WithoutChangingValue(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{Name: "Alice"}) })

	f.MarkTouched("Email")

	if !f.Touched("Email") {
		t.Fatalf("expected Email to be touched")
	}
	if f.Values().Email != "" {
		t.Fatalf("MarkTouched should not change the value")
	}
}

func TestForm_Reset(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{Name: "Alice"}) })

	_ = f.SetField("Name", "Bob")
	f.SetError("Name", errors.New("bad name"))
	IsPendingRender() // clear

	f.Reset()

	if f.Values().Name != "Alice" {
		t.Fatalf("expected Reset to restore initial value, got %q", f.Values().Name)
	}
	if f.IsDirty() {
		t.Fatalf("expected Reset to clear the dirty flag")
	}
	if f.Touched("Name") {
		t.Fatalf("expected Reset to clear touched state")
	}
	if f.Error("Name") != nil {
		t.Fatalf("expected Reset to clear errors")
	}
	if !IsPendingRender() {
		t.Fatalf("expected Reset to request a render")
	}
}

func TestForm_SetInitial(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{Name: "Alice"}) })

	_ = f.SetField("Name", "Bob")
	if !f.IsDirty() {
		t.Fatalf("expected form dirty before SetInitial")
	}

	f.SetInitial(f.Values())

	if f.IsDirty() {
		t.Fatalf("expected SetInitial to clear the dirty flag")
	}

	f.Reset()
	if f.Values().Name != "Bob" {
		t.Fatalf("expected Reset to revert to the new initial (Bob), got %q", f.Values().Name)
	}
}

// ============================================================================
// Errors / Validate
// ============================================================================

func TestForm_SetError_Error_Errors_IsValid(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{}) })

	if !f.IsValid() {
		t.Fatalf("expected a new form to be valid")
	}

	f.SetError("Email", errors.New("required"))

	if f.IsValid() {
		t.Fatalf("expected form to be invalid after SetError")
	}
	if f.Error("Email") == nil {
		t.Fatalf("expected Error to return the set error")
	}
	if len(f.Errors()) != 1 {
		t.Fatalf("expected exactly one error, got %d", len(f.Errors()))
	}

	f.SetError("Email", nil)

	if !f.IsValid() {
		t.Fatalf("expected form to be valid again after clearing its only error")
	}
}

func TestForm_Errors_ReturnsCopy(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{}) })

	f.SetError("Email", errors.New("required"))
	errs := f.Errors()
	delete(errs, "Email")

	if f.Error("Email") == nil {
		t.Fatalf("mutating the map returned by Errors() should not affect internal state")
	}
}

func TestForm_Validate(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{}) })

	requireName := func(v personForm) map[string]error {
		errs := map[string]error{}
		if v.Name == "" {
			errs["Name"] = errors.New("required")
		}
		return errs
	}

	if valid := f.Validate(requireName); valid {
		t.Fatalf("expected form to be invalid when Name is empty")
	}
	if f.Error("Name") == nil {
		t.Fatalf("expected Name error to be set")
	}

	_ = f.SetField("Name", "Alice")
	if valid := f.Validate(requireName); !valid {
		t.Fatalf("expected form to be valid once Name is set")
	}
	if f.Error("Name") != nil {
		t.Fatalf("expected Name error to be cleared, got %v", f.Error("Name"))
	}
}

// ============================================================================
// Submit
// ============================================================================

func TestForm_Submit_BlockedWhenInvalid(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{}) })

	f.SetError("Name", errors.New("required"))

	called := false
	err := f.Submit(func(v personForm) error {
		called = true
		return nil
	})

	if err == nil {
		t.Fatalf("expected Submit to return an error when the form is invalid")
	}
	if called {
		t.Fatalf("expected Submit not to call fn when the form is invalid")
	}
}

func TestForm_Submit_Success(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{}) })

	_ = f.SetField("Name", "Alice")

	var received personForm
	err := f.Submit(func(v personForm) error {
		received = v
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Name != "Alice" {
		t.Fatalf("expected fn to receive current values, got %q", received.Name)
	}
	if f.IsDirty() {
		t.Fatalf("expected a successful Submit to clear the dirty flag via SetInitial")
	}
}

func TestForm_Submit_PropagatesFnError(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{Name: "Alice"}) })

	_ = f.SetField("Name", "Bob")

	sentinel := errors.New("save failed")
	err := f.Submit(func(v personForm) error {
		return sentinel
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected Submit to propagate fn's error, got %v", err)
	}
	if !f.IsDirty() {
		t.Fatalf("expected form to remain dirty when Submit's fn fails")
	}
}

// ============================================================================
// Batch interaction
// ============================================================================

func TestForm_Batch_CoalescesPendingRender(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{}) })

	IsPendingRender() // clear

	Batch(func() {
		_ = f.SetField("Name", "Alice")
		_ = f.SetField("Email", "alice@example.com")
	})

	if !IsPendingRender() {
		t.Fatalf("expected Batch to request a render once the callback returns")
	}
	if IsPendingRender() {
		t.Fatalf("expected IsPendingRender to clear the flag once read")
	}
	if f.Values().Name != "Alice" || f.Values().Email != "alice@example.com" {
		t.Fatalf("expected both field updates to apply, got %+v", f.Values())
	}
}

// ============================================================================
// Concurrency (run with -race)
// ============================================================================

func TestForm_ConcurrentSetFieldAndValues(t *testing.T) {
	setupForm(t)
	var f *Form[personForm]
	renderFrame(t, func() { f = UseForm(personForm{}) })

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			_ = f.SetField("Age", n)
		}(i)
		go func() {
			defer wg.Done()
			_ = f.Values()
		}()
	}
	wg.Wait()
	// No assertion on the final value — this test exists to be run with
	// `go test -race` to catch data races in SetField/Values, not to check
	// a specific outcome from concurrent writers.
}
