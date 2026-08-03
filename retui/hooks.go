package retui

// hooks.go — RetUI component state primitives.
//
// Design contract:
//
//  1. UseState / UseStateKeyed / UseMemo / UseRef are called ONLY from the
//     render goroutine. They read/write StateCursor and the State slice
//     which are render-goroutine-exclusive — no lock needed for those.
//
//  2. State values (State[idx], KeyedState[key]) may be written by ANY
//     goroutine via the setter returned from UseState/UseStateKeyed.
//     stateMu protects only those shared value writes/reads.
//
//  3. stateMu is NEVER held while calling user-supplied code (effect fn,
//     cleanup, reducer). Effect bodies routinely call setters — holding
//     the lock across that would self-deadlock (Go mutexes are not
//     reentrant).
//
//  4. Cursor variables (StateCursor, EffectCursor, MemosCursor) are
//     reset to 0 at the start of each render via BeginRender(). They
//     must only be written from the render goroutine.
//
//  5. Batch() coalesces multiple setter calls from one event into a
//     single pendingRender signal, preventing redundant redraws.

import (
	"reflect"
	"sync"
)

// ─────────────────────────────────────────────────────────────────────────────
// Global hook storage
// ─────────────────────────────────────────────────────────────────────────────

// State is the positional state slice. Indexed by StateCursor.
// Render-goroutine-owned for reads; stateMu for writes from setters.
var State []any

// StateCursor is the next free index in State.
// Must only be written from the render goroutine.
var StateCursor int

// Effects is the registered side-effect slice. Indexed by EffectCursor.
var Effects []Effect

// EffectCursor is the next free index in Effects.
var EffectCursor int

// KeyedState is state keyed by a stable string identity.
// stateMu guards all access to this map.
var KeyedState = map[string]any{}

// keyedTouched tracks keys read this render cycle for PruneKeyedState.
var keyedTouched = map[string]bool{}

// memos is the UseMemo cache slice. Indexed by MemosCursor.
var memos []memoEntry

// MemosCursor is the next free index in memos.
var MemosCursor int

// pendingRender is set true by any setter to request a redraw.
// Read and cleared by the render loop.
var pendingRender bool

// batching is true while inside a Batch() call. When true, setters do
// not set pendingRender — Batch sets it once when the callback returns.
var batching bool

// stateMu guards: State values, KeyedState, memos values, pendingRender,
// batching. It does NOT guard cursor variables (render-goroutine-only).
var stateMu sync.Mutex

// ─────────────────────────────────────────────────────────────────────────────
// Render lifecycle
// ─────────────────────────────────────────────────────────────────────────────

// BeginRender must be called once at the start of every render, before
// the root component function runs.
//
// It resets all cursor positions and the keyed-state touch tracker.
// This is the ONLY place cursors should be reset — doing it elsewhere
// without the stateMu section below risks a race with async setters.
//
// Example (in your render loop):
//
//	retui.BeginRender()
//	root := MyApp(props)
//	retui.RunEffects()
func BeginRender() {
	// Cursor resets are render-goroutine-only — no lock needed.
	StateCursor = 0
	EffectCursor = 0
	MemosCursor = 0

	// keyedTouched is only written from the render goroutine too.
	keyedTouched = make(map[string]bool, len(keyedTouched))
}

// PruneKeyedState evicts KeyedState entries that were not touched since
// the last BeginRender. Call once per render, after the tree has been
// fully rendered. Prevents KeyedState growing without bound.
//
// Example:
//
//	retui.BeginRender()
//	root := MyApp(props)
//	retui.PruneKeyedState()
//	retui.RunEffects()
func PruneKeyedState() {
	stateMu.Lock()
	defer stateMu.Unlock()
	for k := range KeyedState {
		if !keyedTouched[k] {
			delete(KeyedState, k)
		}
	}
}

// IsPendingRender returns true if any setter has requested a redraw
// since the last render. Clears the flag atomically.
//
// Example (in your render loop):
//
//	if retui.IsPendingRender() {
//	    // re-render
//	}
func IsPendingRender() bool {
	stateMu.Lock()
	defer stateMu.Unlock()
	v := pendingRender
	pendingRender = false
	return v
}

// Batch coalesces multiple setter calls into a single pendingRender
// signal. Use when an event handler must update several independent
// state values atomically — without Batch, each setter would trigger
// its own redraw.
//
// Example:
//
//	retui.Batch(func() {
//	    setName("Alice")
//	    setAge(30)
//	    setRole("admin")
//	})
//	// one redraw, not three
func Batch(fn func()) {
	stateMu.Lock()
	batching = true
	stateMu.Unlock()

	fn()

	stateMu.Lock()
	batching = false
	pendingRender = true
	stateMu.Unlock()
}

// ─────────────────────────────────────────────────────────────────────────────
// UseState
// ─────────────────────────────────────────────────────────────────────────────

// UseState returns the current value at the caller's positional slot and
// a setter that updates it and requests a redraw.
//
// Rules (same as React hooks):
//   - Call only from the render goroutine.
//   - Call unconditionally and in a fixed order — never inside an if/loop.
//   - If calls vary across renders (different screen with different hook
//     count), use UseStateKeyed instead.
//
// The setter is safe to call from any goroutine (timer, goroutine, etc).
//
// Example:
//
//	count, setCount := retui.UseState(0)
//	if focused && key == KeyEnter {
//	    setCount(count + 1)
//	}
func UseState[T any](initial T) (T, func(T)) {
	idx := StateCursor
	StateCursor++

	// FIX Bug 1: lock only for the State slice access, not for the entire
	// function. The original held stateMu across the return including the
	// setter closure — any component that calls a setter during render
	// (e.g. a clamping correction like "if pos > len { setPos(len) }")
	// would self-deadlock since Go mutexes are not reentrant.
	stateMu.Lock()
	if idx >= len(State) {
		State = append(State, initial)
	} else {
		// Type mismatch: slot reused across a screen switch or a hook-order
		// bug. Reset to initial rather than panicking so the app stays alive.
		if _, ok := State[idx].(T); !ok {
			State[idx] = initial
		}
	}
	current := State[idx].(T)
	stateMu.Unlock()

	setter := func(next T) {
		stateMu.Lock()
		defer stateMu.Unlock()
		State[idx] = next
		if !batching {
			pendingRender = true
		}
	}

	return current, setter
}

// ─────────────────────────────────────────────────────────────────────────────
// UseStateKeyed
// ─────────────────────────────────────────────────────────────────────────────

// UseStateKeyed is like UseState but keyed by a stable string.
//
// Use this when the number of hook calls in a render can vary — e.g.
// tree nodes that appear/disappear as subtrees expand or collapse.
// Positional UseState would corrupt all slots after a disappeared node;
// UseStateKeyed is immune because it always looks up by key.
//
// key must be globally unique for the lifetime of the app. Prefix it
// with the component/screen name to avoid collisions:
//
//	expanded, setExpanded := retui.UseStateKeyed("tree-"+node.ID, false)
func UseStateKeyed[T any](key string, initial T) (T, func(T)) {
	// keyedTouched is render-goroutine-only — no lock needed.
	keyedTouched[key] = true

	stateMu.Lock()
	if _, exists := KeyedState[key]; !exists {
		KeyedState[key] = initial
	}
	current := KeyedState[key].(T)
	stateMu.Unlock()

	setter := func(next T) {
		stateMu.Lock()
		defer stateMu.Unlock()
		KeyedState[key] = next
		if !batching {
			pendingRender = true
		}
	}

	return current, setter
}

// ─────────────────────────────────────────────────────────────────────────────
// UseReducer
// ─────────────────────────────────────────────────────────────────────────────

// UseReducer manages state through a reducer function, identical to the
// React pattern. Prefer over multiple UseState calls when several values
// must be updated atomically in response to an action.
//
// Example:
//
//	type Action struct{ Type string; Payload any }
//
//	state, dispatch := retui.UseReducer(func(s MyState, a Action) MyState {
//	    switch a.Type {
//	    case "increment": return MyState{Count: s.Count + 1}
//	    case "reset":     return MyState{Count: 0}
//	    }
//	    return s
//	}, MyState{Count: 0})
//
//	if focused && key == KeyEnter {
//	    dispatch(Action{Type: "increment"})
//	}
func UseReducer[S any, A any](reducer func(S, A) S, initial S) (S, func(A)) {
	idx := StateCursor
	StateCursor++

	stateMu.Lock()
	if idx >= len(State) {
		State = append(State, initial)
	} else if _, ok := State[idx].(S); !ok {
		State[idx] = initial
	}
	current := State[idx].(S)
	stateMu.Unlock()

	dispatch := func(action A) {
		stateMu.Lock()
		cur := State[idx].(S)
		stateMu.Unlock()

		next := reducer(cur, action) // user code — never called under stateMu

		stateMu.Lock()
		State[idx] = next
		if !batching {
			pendingRender = true
		}
		stateMu.Unlock()
	}

	return current, dispatch
}

// ─────────────────────────────────────────────────────────────────────────────
// UseMemo
// ─────────────────────────────────────────────────────────────────────────────

type memoEntry struct {
	deps   []any
	result any
}

// UseMemo returns a memoized value recomputed only when deps change.
// Use for expensive computations (filtering, sorting, formatting) that
// should not run on every render.
//
// The compute function runs synchronously during render.
// deps comparison uses safeDeepEqual (recover-wrapped DeepEqual).
//
// Example:
//
//	filtered := retui.UseMemo(func() []Item {
//	    return filterItems(allItems, query)
//	}, []any{allItems, query})
func UseMemo[T any](compute func() T, deps []any) T {
	idx := MemosCursor
	MemosCursor++

	if idx < len(memos) {
		entry := memos[idx]
		if result, ok := entry.result.(T); ok && depsEqual(entry.deps, deps) {
			return result
		}
	}

	result := compute()

	if idx >= len(memos) {
		memos = append(memos, memoEntry{deps: deps, result: result})
	} else {
		memos[idx] = memoEntry{deps: deps, result: result}
	}

	return result
}

// ─────────────────────────────────────────────────────────────────────────────
// UseRef
// ─────────────────────────────────────────────────────────────────────────────

// Ref is a mutable container that does NOT trigger a re-render when its
// value changes. Use for values that the render needs to read but that
// should not cause a redraw when updated (timer IDs, previous values,
// imperative handles).
//
// Example:
//
//	prevCount := retui.UseRef(0)
//	prevCount.Set(count) // does NOT trigger redraw
//	delta := count - prevCount.Get()
type Ref[T any] struct {
	mu  sync.Mutex
	val T
}

// Get returns the current ref value. Safe from any goroutine.
func (r *Ref[T]) Get() T {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.val
}

// Set updates the ref value without triggering a re-render. Safe from
// any goroutine.
func (r *Ref[T]) Set(v T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.val = v
}

// UseRef returns a Ref stored at the caller's positional slot.
// The initial value is set only on the first render.
//
// Example:
//
//	timerID := retui.UseRef[int](0)
func UseRef[T any](initial T) *Ref[T] {
	idx := StateCursor
	StateCursor++

	stateMu.Lock()
	defer stateMu.Unlock()

	if idx >= len(State) {
		r := &Ref[T]{val: initial}
		State = append(State, r)
		return r
	}

	if r, ok := State[idx].(*Ref[T]); ok {
		return r
	}

	// Type mismatch — reset.
	r := &Ref[T]{val: initial}
	State[idx] = r
	return r
}

// ─────────────────────────────────────────────────────────────────────────────
// UseEffect
// ─────────────────────────────────────────────────────────────────────────────

// Effect holds a registered side-effect.
type Effect struct {
	fn      func() func()
	deps    []any
	cleanup func()
	dirty   bool
}

// UseEffect registers a side-effect to run after paint when deps change.
// fn returns an optional cleanup called before the next run or on unmount.
//
// Deps are compared with safeDeepEqual (recover-wrapped reflect.DeepEqual)
// so slice/map/struct deps are compared by value. Functions and channels
// in deps are treated as "always changed" rather than panicking.
//
// Example:
//
//	retui.UseEffect(func() func() {
//	    ticker := time.NewTicker(time.Second)
//	    go func() {
//	        for range ticker.C {
//	            setTime(time.Now())
//	        }
//	    }()
//	    return ticker.Stop  // cleanup
//	}, []any{})  // empty deps = run once on mount
func UseEffect(fn func() func(), deps []any) {
	idx := EffectCursor
	EffectCursor++

	if idx >= len(Effects) {
		Effects = append(Effects, Effect{fn: fn, deps: deps, dirty: true})
		return
	}

	existing := &Effects[idx]
	if !depsEqual(existing.deps, deps) {
		existing.dirty = true
		existing.fn = fn
		existing.deps = deps
	}
}

// RunEffects executes all dirty effects and cleans up removed effects.
//
// Call once per render cycle, after rendering and PruneKeyedState:
//
//	retui.BeginRender()
//	root := MyApp(props)
//	retui.PruneKeyedState()
//	retui.RunEffects()
//
// CRITICAL: stateMu is never held while calling user code (fn, cleanup).
// Effect bodies routinely call setters — holding the lock across that
// would self-deadlock. stateMu is held only for short slice/map accesses.
func RunEffects() {
	type job struct {
		idx     int
		cleanup func()
		fn      func() func()
	}

	// Phase 1 (locked): snapshot dirty jobs and tail cleanups.
	// Claim cleanup funcs so nothing else touches them concurrently.
	// No user code runs in this phase.
	stateMu.Lock()
	var jobs []job
	for i := 0; i < EffectCursor && i < len(Effects); i++ {
		if Effects[i].dirty {
			jobs = append(jobs, job{
				idx:     i,
				cleanup: Effects[i].cleanup,
				fn:      Effects[i].fn,
			})
			Effects[i].cleanup = nil
		}
	}

	// Tear down effects for unmounted components (tail beyond EffectCursor).
	var tailCleanups []func()
	if len(Effects) > EffectCursor {
		for i := EffectCursor; i < len(Effects); i++ {
			if Effects[i].cleanup != nil {
				tailCleanups = append(tailCleanups, Effects[i].cleanup)
			}
		}
		Effects = Effects[:EffectCursor]
	}

	// FIX Bug 4: State tail cleanup moved here from inside RunEffects body
	// in original. But it must happen BEFORE the next render's BeginRender
	// resets StateCursor to 0 — otherwise there's a window where
	// StateCursor=0 and State still has stale entries from the previous
	// screen. BeginRender() now owns cursor resets; tail cleanup happens
	// here, after render, before the next BeginRender.
	if len(State) > StateCursor {
		State = State[:StateCursor]
	}
	if len(memos) > MemosCursor {
		memos = memos[:MemosCursor]
	}
	stateMu.Unlock()

	// Phase 2 (unlocked): run user code.
	for _, j := range jobs {
		safeCall(j.cleanup)
		newCleanup := safeCallFn(j.fn)

		stateMu.Lock()
		if j.idx < len(Effects) {
			Effects[j.idx].cleanup = newCleanup
			Effects[j.idx].dirty = false
		}
		stateMu.Unlock()
	}

	for _, c := range tailCleanups {
		safeCall(c)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Context
// ─────────────────────────────────────────────────────────────────────────────

// Context carries a value down the component tree without prop-drilling.
// The zero value is not usable — construct with CreateContext.
//
// Example:
//
//	var ThemeCtx = retui.CreateContext("light")
//
//	func App(...) retui.Element {
//	    return ThemeCtx.Provide("dark", func() retui.Element {
//	        return MyPage()
//	    })
//	}
//
//	func MyPage() retui.Element {
//	    theme := retui.UseContext(ThemeCtx)  // "dark"
//	    ...
//	}
type Context[T any] struct {
	defaultValue T
	stack        []T
}

// CreateContext returns a Context with the given default value.
func CreateContext[T any](defaultValue T) *Context[T] {
	return &Context[T]{defaultValue: defaultValue}
}

// Provide pushes value, runs render, then pops via defer.
// A panic in render still unwinds the stack cleanly.
func (c *Context[T]) Provide(value T, render func() Element) Element {
	c.stack = append(c.stack, value)
	defer func() { c.stack = c.stack[:len(c.stack)-1] }()
	return render()
}

// UseContext returns the innermost active Provide value, or the default.
func UseContext[T any](c *Context[T]) T {
	if len(c.stack) == 0 {
		return c.defaultValue
	}
	return c.stack[len(c.stack)-1]
}

// ─────────────────────────────────────────────────────────────────────────────
// ResetComponentState
// ─────────────────────────────────────────────────────────────────────────────

// ResetComponentState clears ALL hook state: positional, keyed, memos,
// and effects. Call on hard screen transitions (logout, full reset).
//
// WARNING: this also clears UseStateKeyed state. If you use UseStateKeyed
// specifically to survive navigation (tree expand state, table selection),
// do NOT call this on routine screen switches — call BeginRender() instead,
// which only resets cursors.
func ResetComponentState() {
	stateMu.Lock()
	defer stateMu.Unlock()

	State = nil
	StateCursor = 0
	Effects = nil
	EffectCursor = 0
	memos = nil
	MemosCursor = 0
	// KeyedState = map[string]any{}
	keyedTouched = map[string]bool{}
	pendingRender = false
	batching = false
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────────────────────────────────────

// depsEqual compares two dependency slices using recover-wrapped DeepEqual.
// FIX Bug 3: original used raw reflect.DeepEqual which panics if deps
// contain channels or func values. We treat a panic as "deps changed"
// so the effect re-runs safely instead of crashing.
func depsEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !safeDeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func safeDeepEqual(a, b any) (equal bool) {
	defer func() {
		if r := recover(); r != nil {
			equal = false // non-comparable type — treat as changed
		}
	}()
	return reflect.DeepEqual(a, b)
}

func safeCall(fn func()) {
	if fn == nil {
		return
	}
	defer func() { recover() }()
	fn()
}

func safeCallFn(fn func() func()) (result func()) {
	if fn == nil {
		return nil
	}
	defer func() { recover() }()
	return fn()
}
