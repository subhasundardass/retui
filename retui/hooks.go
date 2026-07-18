package retui

import (
	"reflect"
)

type Effect struct {
	fn      func() func()
	deps    []any
	cleanup func()
	dirty   bool
}

var State []any
var StateCursor int = 0

var Effects []Effect
var EffectCursor int = 0

// KeyedState holds state for UseStateKeyed — a map keyed by a stable
// string identity rather than cursor position. This lets components like
// Tree maintain independent expand/collapse state per node across renders
// where the number of visible nodes changes (collapsed subtrees disappear,
// which would shift the positional StateCursor for every node after them).
var KeyedState = map[string]any{}

var pendingRender bool

func UseState[T any](initial T) (T, func(T)) {
	idx := StateCursor
	StateCursor++

	if idx >= len(State) {
		State = append(State, initial)
	} else {
		// State exists, but may belong to another screen.
		if _, ok := State[idx].(T); !ok {
			State[idx] = initial
		}
	}

	current := State[idx].(T)

	setter := func(next T) {
		State[idx] = next
		pendingRender = true
	}

	return current, setter
}

// UseStateKeyed is like UseState but keyed by a stable string instead of
// cursor position. Use this whenever the number of hook calls in a render
// can vary — e.g. a tree node that may or may not render children depending
// on expand state. Positional UseState would corrupt all subsequent state
// slots when nodes are toggled; UseStateKeyed is immune because it always
// looks up by key regardless of render order.
//
// key must be globally unique across your entire component tree for the
// lifetime of the app. For tree nodes, combine the node's ID with its
// depth or path: "tree-node-/src/main.go" or "tree-node-0-1-2".
func UseStateKeyed[T any](key string, initial T) (T, func(T)) {
	if _, exists := KeyedState[key]; !exists {
		KeyedState[key] = initial
	}

	current := KeyedState[key].(T)

	setter := func(next T) {
		KeyedState[key] = next
		pendingRender = true
	}

	return current, setter
}

// UseScreenReset resets all component hook state whenever currentID changes
// from the previous render. Call once at the top of your root component
// to safely switch between screens/routes without hook-order panics.
// func UseScreenReset(currentID string) {
// 	prevScreen, setPrevScreen := UseStateKeyed("root:prevScreen", "")
// 	if prevScreen != currentID {
// 		setPrevScreen(currentID)
// 		ResetComponentState()
// 	}
// }

// UseEffect registers a side-effect to run after paint whenever its
// deps change. fn returns an optional cleanup called before the next
// run or when the component unmounts.
//
// Deps comparison uses reflect.DeepEqual so slice/map/struct deps are
// compared by value, not pointer — a plain != on interface{} would
// panic at runtime for any non-comparable dep type (slice, map, etc).
func UseEffect(fn func() func(), deps []any) {
	idx := EffectCursor
	EffectCursor++

	newEffect := Effect{fn: fn, deps: deps, dirty: true}

	if idx >= len(Effects) {
		//New effect - append
		Effects = append(Effects, newEffect)
		return
	}

	//Existing effect - check if deps changed
	existing := &Effects[idx]
	changed := len(existing.deps) != len(newEffect.deps)
	if !changed {
		for i, dep := range newEffect.deps {
			if !reflect.DeepEqual(existing.deps[i], dep) {
				changed = true
				break
			}
		}
	}

	if changed {
		// Only mark dirty if it changed. Do NOT set to false if it was already
		// dirty from a previous pass in the same render cycle.
		existing.dirty = true
		existing.fn = newEffect.fn
		existing.deps = newEffect.deps
	}
}

// RunEffects runs all effects marked dirty up to the current EffectCursor.
// It also cleans up any effects that were present in previous renders but
// are no longer reachable (component unmounted).
func RunEffects() {
	// 1. Run active effects
	for i := 0; i < EffectCursor; i++ {
		if !Effects[i].dirty {
			continue
		}
		if Effects[i].cleanup != nil {
			Effects[i].cleanup()
		}
		Effects[i].cleanup = Effects[i].fn()
		Effects[i].dirty = false
	}

	// 2. Tail Cleanup: If the tree shrunk, clean up remaining effects
	if len(Effects) > EffectCursor {
		for i := EffectCursor; i < len(Effects); i++ {
			if Effects[i].cleanup != nil {
				Effects[i].cleanup()
			}
		}
		// Truncate to prevent memory leak
		Effects = Effects[:EffectCursor]
	}

	// 3. State Tail Cleanup: Prevent memory leaks from large objects in state
	if len(State) > StateCursor {
		State = State[:StateCursor]
	}
}

// Context carries a value down the component tree without prop-drilling.
// The zero value is not usable — construct with CreateContext.
type Context[T any] struct {
	defaultValue T
	stack        []T
}

// CreateContext returns a new Context whose UseContext readers see
// defaultValue when no enclosing Provide is active.
func CreateContext[T any](defaultValue T) *Context[T] {
	return &Context[T]{defaultValue: defaultValue}
}

// Provide pushes value onto the context's stack, runs render (during
// which any descendant calling UseContext observes value), then pops
// via defer so a panic in render still unwinds the stack cleanly.
func (c *Context[T]) Provide(value T, render func() Element) Element {
	c.stack = append(c.stack, value)
	defer func() { c.stack = c.stack[:len(c.stack)-1] }()
	return render()
}

// UseContext returns the innermost active Provide value, or
// defaultValue if no Provide is currently on the stack.
func UseContext[T any](c *Context[T]) T {
	if len(c.stack) == 0 {
		return c.defaultValue
	}
	return c.stack[len(c.stack)-1]
}

// ResetComponentState clears all positional state slots.
// Call this whenever the active screen changes so stale state
// from the previous screen does not corrupt the new screen's slots.
func ResetComponentState() {
	State = nil
	StateCursor = 0
	Effects = nil
	EffectCursor = 0
}
