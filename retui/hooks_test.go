package retui

import (
	"sync"
	"testing"
)

// ─── TestUseState ────────────────────────────────────────────────────────────

func TestUseState(t *testing.T) {
	// Reset state before test
	State = nil
	StateCursor = 0
	pendingRender = false

	// Test initial value
	val, setVal := UseState(42)
	if val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}

	// Test setting value
	setVal(100)
	if State[0] != 100 {
		t.Errorf("Expected state[0] = 100, got %v", State[0])
	}
	if !pendingRender {
		t.Error("Expected pendingRender to be true after setter call")
	}
}

func TestUseStateMultiple(t *testing.T) {
	// Reset state before test
	State = nil
	StateCursor = 0
	pendingRender = false

	s1, _ := UseState("hello")
	s2, _ := UseState(42)
	s3, _ := UseState(true)

	if s1 != "hello" {
		t.Errorf("Expected 'hello', got %v", s1)
	}
	if s2 != 42 {
		t.Errorf("Expected 42, got %v", s2)
	}
	if s3 != true {
		t.Errorf("Expected true, got %v", s3)
	}

	if len(State) != 3 {
		t.Errorf("Expected state length 3, got %d", len(State))
	}
}

func TestUseStateSetOrder(t *testing.T) {
	// Reset state before test
	State = nil
	StateCursor = 0
	pendingRender = false

	val1, set1 := UseState("a")
	val2, set2 := UseState("b")

	if val1 != "a" || val2 != "b" {
		t.Errorf("Initial values wrong: val1=%v, val2=%v", val1, val2)
	}

	set1("A")
	set2("B")

	if State[0] != "A" || State[1] != "B" {
		t.Errorf("State after set wrong: State[0]=%v, State[1]=%v", State[0], State[1])
	}
}

// ─── TestUseStateKeyed ──────────────────────────────────────────────────────

func TestUseStateKeyed(t *testing.T) {
	// Reset keyed state
	KeyedState = make(map[string]any)
	pendingRender = false

	val1, set1 := UseStateKeyed("key1", "initial1")
	val2, set2 := UseStateKeyed("key2", "initial2")

	if val1 != "initial1" {
		t.Errorf("Expected 'initial1', got %v", val1)
	}
	if val2 != "initial2" {
		t.Errorf("Expected 'initial2', got %v", val2)
	}

	set1("updated1")
	if KeyedState["key1"] != "updated1" {
		t.Errorf("Expected KeyedState['key1'] = 'updated1', got %v", KeyedState["key1"])
	}
	if !pendingRender {
		t.Error("Expected pendingRender to be true after setter call")
	}

	// Test second setter
	set2("updated2")
	if KeyedState["key2"] != "updated2" {
		t.Errorf("Expected KeyedState['key2'] = 'updated2', got %v", KeyedState["key2"])
	}
}

func TestUseStateKeyedSameKey(t *testing.T) {
	// Reset keyed state
	KeyedState = make(map[string]any)
	pendingRender = false

	val1, set1 := UseStateKeyed("sameKey", "first")
	if val1 != "first" {
		t.Errorf("Expected 'first', got %v", val1)
	}

	val2, _ := UseStateKeyed("sameKey", "shouldNotOverride")
	if val2 != "first" {
		t.Errorf("Expected 'first' (not overridden), got %v", val2)
	}

	set1("second")
	val3, _ := UseStateKeyed("sameKey", "ignored")
	if val3 != "second" {
		t.Errorf("Expected 'second', got %v", val3)
	}
}

// ─── TestUseEffect ──────────────────────────────────────────────────────────

func TestUseEffect(t *testing.T) {
	// Reset effects
	Effects = nil
	EffectCursor = 0

	callCount := 0

	effect := func() func() {
		callCount++
		return func() {
			// Cleanup called
		}
	}

	UseEffect(effect, []any{"dep1", 42})
	RunEffects()

	if callCount != 1 {
		t.Errorf("Expected effect called once, got %d", callCount)
	}
	//Check that the effect was executed (dirty should be false after RunEffects)
	if Effects[0].dirty {
		t.Error("Expected effect to be clean after RunEffects")
	}
}

// ─── TestContext ────────────────────────────────────────────────────────────

func TestCreateContext(t *testing.T) {
	ctx := CreateContext("default")
	if ctx == nil {
		t.Fatal("CreateContext returned nil")
	}
	// Now safe to access defaultValue
	if ctx.defaultValue != "default" {
		t.Errorf("expected defaultValue 'default', got %q", ctx.defaultValue)
	}
	// Now safe to access defaultValue
	if ctx.defaultValue != "default" {
		t.Errorf("expected defaultValue 'default', got %q", ctx.defaultValue)
	}

	if ctx.defaultValue != "default" {
		t.Errorf("Expected defaultValue 'default', got %v", ctx.defaultValue)
	}
}

func TestContextProvideAndUse(t *testing.T) {
	ctx := CreateContext("default")
	if ctx == nil {
		t.Fatal("CreateContext returned nil")
	}
	// Now safe to access defaultValue
	if ctx.defaultValue != "default" {
		t.Errorf("expected defaultValue 'default', got %q", ctx.defaultValue)
	}
	// Now safe to access defaultValue
	if ctx.defaultValue != "default" {
		t.Errorf("expected defaultValue 'default', got %q", ctx.defaultValue)
	}
	if ctx == nil {
		t.Fatal("CreateContext returned nil")
	}
	if ctx.defaultValue != "default" {
		t.Errorf("expected defaultValue 'default', got %q", ctx.defaultValue)
	}

	// Test without Provide
	value := UseContext(ctx)
	if value != "default" {
		t.Errorf("Expected 'default', got %v", value)
	}

	// Test with Provide
	result := ctx.Provide("provided", func() Element {
		inner := UseContext(ctx)
		if inner != "provided" {
			t.Errorf("Expected 'provided', got %v", inner)
		}
		return Element{}
	})

	// result should be an Element (not nil)
	_ = result
}

func TestContextNested(t *testing.T) {
	ctx := CreateContext("default")

	ctx.Provide("outer", func() Element {
		if UseContext(ctx) != "outer" {
			t.Error("Expected 'outer' in outer Provide")
		}

		ctx.Provide("inner", func() Element {
			if UseContext(ctx) != "inner" {
				t.Error("Expected 'inner' in inner Provide")
			}
			return Element{}
		})

		// After inner Provide pops, should be back to outer
		if UseContext(ctx) != "outer" {
			t.Error("Expected 'outer' after inner pop")
		}
		return Element{}
	})

	// After all Provides pop, should be back to default
	if UseContext(ctx) != "default" {
		t.Error("Expected 'default' after all Provides pop")
	}
}

func TestContextWithPanic(t *testing.T) {
	ctx := CreateContext("default")

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic but got none")
		}
		// After panic, stack should be popped
		if UseContext(ctx) != "default" {
			t.Error("Expected 'default' after panic recovery")
		}
	}()

	ctx.Provide("shouldBePopped", func() Element {
		if UseContext(ctx) != "shouldBePopped" {
			t.Error("Expected 'shouldBePopped'")
		}
		panic("test panic")
	})
}

// ─── TestResetComponentState ───────────────────────────────────────────────

func TestResetComponentState(t *testing.T) {
	// Set up some state
	State = nil
	StateCursor = 0
	Effects = nil
	EffectCursor = 0

	UseState(42)
	UseState("hello")

	UseEffect(func() func() { return nil }, []any{1})

	if len(State) != 2 {
		t.Errorf("Expected State length 2, got %d", len(State))
	}
	if len(Effects) != 1 {
		t.Errorf("Expected Effects length 1, got %d", len(Effects))
	}

	// Reset
	ResetComponentState()

	if State != nil {
		t.Errorf("Expected State to be nil, got %v", State)
	}
	if StateCursor != 0 {
		t.Errorf("Expected StateCursor 0, got %d", StateCursor)
	}
	if Effects != nil {
		t.Errorf("Expected Effects to be nil, got %v", Effects)
	}
	if EffectCursor != 0 {
		t.Errorf("Expected EffectCursor 0, got %d", EffectCursor)
	}
}

// =======================================================================
// UseMemo — type-safety across slot reuse (e.g. screen switches)
// =======================================================================

// TestUseMemo_PanicsOnTypeMismatchWithoutGuard simulates the exact
// failure case: BeginRender() resets MemosCursor to 0 but never clears
// the underlying memos slice, so slot 0 can be reused across screens
// with a different result type. Before the fix, this panics on the
// unchecked entry.result.(T) assertion.
func TestUseMemo_SurvivesTypeChangeAcrossScreens(t *testing.T) {
	defer ResetComponentState()
	ResetComponentState()

	// "Screen A": memoizes a []int at slot 0.
	BeginRender()
	got := UseMemo(func() []int { return []int{1, 2, 3} }, []any{})
	if len(got) != 3 {
		t.Fatalf("screen A: got %v, want len 3", got)
	}
	RunEffects()

	// "Screen B": same slot 0, same (empty) deps, but a completely
	// different result type. Must recompute instead of panicking on
	// the stale []int stored in memos[0].
	BeginRender()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("UseMemo panicked on type mismatch across slot reuse: %v", r)
		}
	}()
	gotStr := UseMemo(func() string { return "hello" }, []any{})
	if gotStr != "hello" {
		t.Errorf("screen B: got %q, want %q", gotStr, "hello")
	}
}

func TestUseMemo_RecomputesWhenDepsChange(t *testing.T) {
	defer ResetComponentState()
	ResetComponentState()

	calls := 0
	compute := func() int {
		calls++
		return calls
	}

	BeginRender()
	first := UseMemo(compute, []any{1})
	RunEffects()

	BeginRender()
	second := UseMemo(compute, []any{1}) // same deps — should NOT recompute
	RunEffects()

	if first != second {
		t.Errorf("expected memoized value to be stable across renders with same deps: first=%d second=%d", first, second)
	}
	if calls != 1 {
		t.Errorf("expected compute() called once, got %d calls", calls)
	}

	BeginRender()
	third := UseMemo(compute, []any{2}) // different deps — should recompute
	RunEffects()

	if third == second {
		t.Error("expected recompute when deps changed, got same value")
	}
	if calls != 2 {
		t.Errorf("expected compute() called twice total, got %d calls", calls)
	}
}

// =======================================================================
// UseReducer — correctness after the DeepEqual-scan removal
// =======================================================================

type counterState struct {
	Count int
}

type counterAction struct {
	Delta int
}

func counterReducer(s counterState, a counterAction) counterState {
	return counterState{Count: s.Count + a.Delta}
}

func TestUseReducer_DispatchUpdatesState(t *testing.T) {
	defer ResetComponentState()
	ResetComponentState()

	BeginRender()
	state, dispatch := UseReducer(counterReducer, counterState{Count: 0})
	RunEffects()

	if state.Count != 0 {
		t.Fatalf("initial Count = %d, want 0", state.Count)
	}

	dispatch(counterAction{Delta: 5})

	// Re-render to observe the committed state at the same slot.
	BeginRender()
	state2, _ := UseReducer(counterReducer, counterState{Count: 0})
	RunEffects()

	if state2.Count != 5 {
		t.Errorf("after dispatch, Count = %d, want 5", state2.Count)
	}
}

// TestUseReducer_DoesNotCollideWithEqualValueElsewhere reproduces the
// wrong-slot risk from the old DeepEqual-scan implementation: two
// independent reducers of the same type, both currently holding an
// equal value, must never let a dispatch on one mutate the other.
func TestUseReducer_DoesNotCollideWithEqualValueElsewhere(t *testing.T) {
	defer ResetComponentState()
	ResetComponentState()

	BeginRender()
	_, dispatchA := UseReducer(counterReducer, counterState{Count: 0}) // slot 0
	_, dispatchB := UseReducer(counterReducer, counterState{Count: 0}) // slot 1 — equal value, same type
	RunEffects()

	dispatchA(counterAction{Delta: 1})

	BeginRender()
	stateA, dispatchA2 := UseReducer(counterReducer, counterState{Count: 0}) // slot 0
	stateB, _ := UseReducer(counterReducer, counterState{Count: 0})          // slot 1
	RunEffects()
	_ = dispatchA2

	if stateA.Count != 1 {
		t.Errorf("slot 0 (dispatched) Count = %d, want 1", stateA.Count)
	}
	if stateB.Count != 0 {
		t.Errorf("slot 1 (untouched) Count = %d, want 0 — dispatch on slot 0 leaked into slot 1", stateB.Count)
	}

	_ = dispatchB
}

func TestUseReducer_MultipleDispatchesBeforeRerenderAllApply(t *testing.T) {
	defer ResetComponentState()
	ResetComponentState()

	BeginRender()
	_, dispatch := UseReducer(counterReducer, counterState{Count: 0})
	RunEffects()

	// Fire several dispatches back to back, before any re-render observes
	// the intermediate state — each must build on the previous, not on
	// the stale value captured at initial render.
	dispatch(counterAction{Delta: 1})
	dispatch(counterAction{Delta: 1})
	dispatch(counterAction{Delta: 1})

	BeginRender()
	final, _ := UseReducer(counterReducer, counterState{Count: 0})
	RunEffects()

	if final.Count != 3 {
		t.Errorf("Count after 3 sequential dispatches = %d, want 3", final.Count)
	}
}

func TestUseReducer_ConcurrentDispatchIsRaceFree(t *testing.T) {
	// This test only asserts no data race / no panic under `go test -race`;
	// it does NOT assert every dispatched delta is preserved. As noted in
	// review, concurrent dispatch on the same reducer slot can lose updates
	// (last-write-wins) since read-modify-write isn't atomic across the
	// two lock sections. If that guarantee is later required, UseReducer
	// needs a per-slot mutex or action queue.
	defer ResetComponentState()
	ResetComponentState()

	BeginRender()
	_, dispatch := UseReducer(counterReducer, counterState{Count: 0})
	RunEffects()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dispatch(counterAction{Delta: 1})
		}()
	}
	wg.Wait()

	BeginRender()
	_, _ = UseReducer(counterReducer, counterState{Count: 0})
	RunEffects()
	// No assertion on final Count — see comment above.
}

// ─── Benchmarks ──────────────────────────────────────────────────────────────

func BenchmarkUseState(b *testing.B) {
	for i := 0; i < b.N; i++ {
		State = nil
		StateCursor = 0
		val, set := UseState(0)
		set(val + 1)
	}
}

func BenchmarkUseStateKeyed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		KeyedState = make(map[string]any)
		val, set := UseStateKeyed("key", 0)
		set(val + 1)
	}
}

func BenchmarkUseEffect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Effects = nil
		EffectCursor = 0
		UseEffect(func() func() { return nil }, []any{1, 2, 3})
		RunEffects()
	}
}

func BenchmarkContextProvide(b *testing.B) {
	ctx := CreateContext("default")
	for i := 0; i < b.N; i++ {
		ctx.Provide("value", func() Element {
			return Element{}
		})
	}
}
