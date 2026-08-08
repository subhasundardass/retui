package retui

import (
	"fmt"
	"reflect"
	"sync"
)

// ============================================================================
// Form Hook
// ============================================================================
//
// UseForm is built on top of UseRef (see hooks.go), so it inherits the exact
// same contract as every other hook in this package:
//
//  1. Call only from the render goroutine.
//  2. Call unconditionally, in a fixed order — never inside an if/loop whose
//     condition can differ between renders.
//  3. Cursor management is handled entirely by BeginRender() (hooks.go) —
//     there is nothing form-specific to reset. Do NOT reset any cursor
//     yourself; UseForm shares the same positional slot list as UseState/
//     UseRef/UseReducer, via UseRef's built-in create-once-reuse logic.
//
// Form itself holds its own values/dirty/touched/errors under a private
// mutex, separate from stateMu, since a *Form[T] is a fairly heavyweight
// object compared to a plain state value — only the *pointer* to it lives
// in the shared State slice (via UseRef); everything Form-specific is
// managed here.

// Form holds a form's current values, plus dirty/touched/error tracking.
// All exported methods are safe for concurrent use from any goroutine.
type Form[T any] struct {
	mu       sync.RWMutex
	values   T
	initial  T
	dirty    bool
	touched  map[string]bool
	errors   map[string]error
	onChange func(T)
}

// UseForm returns a persistent Form[T] for this call site: created with
// `initial` on first render, reused (ignoring `initial`) on every
// subsequent render — the same semantics as UseState(initial).
func UseForm[T any](initial T) *Form[T] {
	ref := UseRef[*Form[T]](nil)

	if ref.Get() == nil {
		ref.Set(&Form[T]{
			values:  initial,
			initial: initial,
			touched: make(map[string]bool),
			errors:  make(map[string]error),
		})
	}

	return ref.Get()
}

// requestRender mirrors the exact pattern every setter in hooks.go uses:
// set pendingRender under stateMu, but only if not inside a Batch(). This
// means Form participates correctly in retui.Batch() — several SetField
// calls inside one Batch() coalesce into a single redraw, same as UseState.
func requestRender() {
	stateMu.Lock()
	if !batching {
		pendingRender = true
	}
	stateMu.Unlock()
}

// ============================================================================
// Values
// ============================================================================

// Values returns a copy of the form's current values.
func (f *Form[T]) Values() T {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.values
}

// SetValues replaces the entire value struct. No-ops (no re-render, no
// dirty flag) if v is deeply equal to the current values.
func (f *Form[T]) SetValues(v T) {
	f.mu.Lock()
	if reflect.DeepEqual(f.values, v) {
		f.mu.Unlock()
		return
	}
	f.values = v
	f.dirty = true
	cb := f.onChange
	f.mu.Unlock()

	if cb != nil {
		cb(v)
	}
	requestRender()
}

// SetValuesSilent replaces the entire value struct WITHOUT marking the form
// dirty, updating touched state, or firing OnChange. Still requests a
// render (the UI needs to reflect the change), and still no-ops if v is
// deeply equal to current values.
//
// Use this for UI-only state that happens to live in T alongside real form
// data — most commonly a FocusIndex field used for index-based Tab
// navigation. "Which field has keyboard focus" isn't a user edit, so it
// shouldn't trip IsDirty()/Touched()/gate Submit() the way actual data
// changes should. If FocusIndex (or similar) can live outside T instead
// (e.g. its own UseState), that's simpler than this — reach for
// SetValuesSilent specifically when the two are stuck in the same struct.
func (f *Form[T]) SetValuesSilent(v T) {
	f.mu.Lock()
	if reflect.DeepEqual(f.values, v) {
		f.mu.Unlock()
		return
	}
	f.values = v
	f.mu.Unlock()

	requestRender()
}

// fieldRef returns a settable reflect.Value for the named field on f.values.
// Caller MUST hold f.mu (write lock) for the duration of use — the returned
// Value aliases f.values directly.
func (f *Form[T]) fieldRef(name string) (reflect.Value, error) {
	v := reflect.ValueOf(&f.values).Elem()
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("retui: field access requires T to be a struct, got %s", v.Kind())
	}
	field := v.FieldByName(name)
	if !field.IsValid() {
		return reflect.Value{}, fmt.Errorf("retui: no field %q on %T", name, f.values)
	}
	if !field.CanSet() {
		return reflect.Value{}, fmt.Errorf("retui: field %q is unexported and cannot be set", name)
	}
	return field, nil
}

// SetField sets a single named field on T by reflection and marks the form
// dirty and that field touched. T must be a struct. Intended for wiring
// individual inputs' OnChange handlers without manually copying the whole
// struct out and back in for every field:
//
//	input.OnChange(func(id, v string) { form.SetField("Name", v) })
//
// Returns an error (not a panic) for bad field names/types, since this is
// typically called from UI event handlers where a hard crash is undesirable.
func (f *Form[T]) SetField(name string, value any) error {
	f.mu.Lock()

	field, err := f.fieldRef(name)
	if err != nil {
		f.mu.Unlock()
		return err
	}

	rv := reflect.ValueOf(value)
	if !rv.Type().AssignableTo(field.Type()) {
		f.mu.Unlock()
		return fmt.Errorf("retui: cannot assign %T to field %q (type %s)", value, name, field.Type())
	}

	field.Set(rv)
	f.dirty = true
	f.touched[name] = true
	cb := f.onChange
	vals := f.values
	f.mu.Unlock()

	if cb != nil {
		cb(vals)
	}
	requestRender()
	return nil
}

// SetFieldSilent sets a single named field on T by reflection WITHOUT
// marking the form dirty, updating touched state, or firing OnChange —
// the SetField equivalent of SetValuesSilent. See SetValuesSilent's doc
// for when this is the right call (UI-only state like FocusIndex living
// inside T).
func (f *Form[T]) SetFieldSilent(name string, value any) error {
	f.mu.Lock()

	field, err := f.fieldRef(name)
	if err != nil {
		f.mu.Unlock()
		return err
	}

	rv := reflect.ValueOf(value)
	if !rv.Type().AssignableTo(field.Type()) {
		f.mu.Unlock()
		return fmt.Errorf("retui: cannot assign %T to field %q (type %s)", value, name, field.Type())
	}

	field.Set(rv)
	f.mu.Unlock()

	requestRender()
	return nil
}

// Field returns the current value of a named field on T by reflection.
// The second return value is false if T isn't a struct or has no such field.
func (f *Form[T]) Field(name string) (any, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	v := reflect.ValueOf(f.values)
	if v.Kind() != reflect.Struct {
		return nil, false
	}
	field := v.FieldByName(name)
	if !field.IsValid() {
		return nil, false
	}
	return field.Interface(), true
}

// ============================================================================
// Dirty / Touched
// ============================================================================

// IsDirty reports whether values have changed since creation or the last
// Reset()/SetInitial().
func (f *Form[T]) IsDirty() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.dirty
}

// Touched reports whether a named field has ever been changed via SetField
// (SetValues does not update per-field touched state, since it doesn't know
// which fields changed).
func (f *Form[T]) Touched(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.touched[name]
}

// MarkTouched marks a field touched without changing its value — useful for
// "touch on blur" validation patterns (show an error only after the user
// has left the field at least once).
func (f *Form[T]) MarkTouched(name string) {
	f.mu.Lock()
	f.touched[name] = true
	f.mu.Unlock()
}

// Reset restores values to the initial value passed to UseForm (or the last
// SetInitial), and clears dirty/touched/error state.
func (f *Form[T]) Reset() {
	f.mu.Lock()
	f.values = f.initial
	f.dirty = false
	f.touched = make(map[string]bool)
	f.errors = make(map[string]error)
	f.mu.Unlock()
	requestRender()
}

// SetInitial updates what Reset() reverts to and clears the dirty flag,
// without changing current values. Typically called after a successful
// submit, so the just-saved values become the new "clean" baseline.
//
// Requests a render if the dirty flag actually changed (no-op if the form
// was already clean) — UI typically shows dirty state (e.g. an "unsaved
// changes" indicator), so this needs to repaint like any other flag flip.
func (f *Form[T]) SetInitial(v T) {
	f.mu.Lock()
	changed := f.dirty
	f.initial = v
	f.dirty = false
	f.mu.Unlock()

	if changed {
		requestRender()
	}
}

// OnChange registers a callback invoked after every successful SetValues or
// SetField call, with the new values. Only one callback is stored; calling
// OnChange again replaces the previous one.
func (f *Form[T]) OnChange(cb func(T)) {
	f.mu.Lock()
	f.onChange = cb
	f.mu.Unlock()
}

// ============================================================================
// Validation
// ============================================================================

// SetError attaches a validation error to a named field. Pass nil to clear
// the error for that field.
func (f *Form[T]) SetError(name string, err error) {
	f.mu.Lock()
	if err == nil {
		delete(f.errors, name)
	} else {
		f.errors[name] = err
	}
	f.mu.Unlock()
	requestRender()
}

// Error returns the current validation error for a field, or nil.
func (f *Form[T]) Error(name string) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.errors[name]
}

// Errors returns a copy of all current field errors, keyed by field name.
func (f *Form[T]) Errors() map[string]error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make(map[string]error, len(f.errors))
	for k, v := range f.errors {
		out[k] = v
	}
	return out
}

// IsValid reports whether the form currently has zero field errors.
func (f *Form[T]) IsValid() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.errors) == 0
}

// Validate runs fn against the current values and replaces the entire error
// set with whatever fn returns (field name -> error; omit valid fields).
// Returns true if the resulting error set is empty.
func (f *Form[T]) Validate(fn func(T) map[string]error) bool {
	f.mu.Lock()
	f.errors = fn(f.values)
	valid := len(f.errors) == 0
	f.mu.Unlock()
	requestRender()
	return valid
}

// ============================================================================
// Submit
// ============================================================================

// Submit calls fn with the current values only if the form has no
// outstanding validation errors, and treats those values as the new clean
// baseline (via SetInitial) if fn succeeds. Returns an error without calling
// fn if the form is currently invalid, or whatever fn itself returns.
func (f *Form[T]) Submit(fn func(T) error) error {
	f.mu.RLock()
	if len(f.errors) > 0 {
		n := len(f.errors)
		f.mu.RUnlock()
		return fmt.Errorf("retui: form has %d validation error(s)", n)
	}
	vals := f.values
	f.mu.RUnlock()

	if err := fn(vals); err != nil {
		return err
	}

	f.SetInitial(vals) // requests its own render if dirty changed
	return nil
}
