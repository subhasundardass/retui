package retui

import (
	"testing"
)

// TestEffectCleanupOnUnmount verifies that when a component stops calling UseEffect
// (simulating an unmount or a conditional branch change), the cleanup function
// of the previously registered effect is executed and the effect is removed
// from the internal slice.
func TestEffectCleanupOnUnmount(t *testing.T) {
	app := NewApp(80, 24)
	cleanupCalled := false

	// Ensure clean slate for global hook state
	ResetComponentState()

	// 1. First Render: Register an effect
	app.Render(func(props Props) Element {
		UseEffect(func() func() {
			return func() {
				cleanupCalled = true
			}
		}, nil)
		return Element{}
	}, Props{})

	if cleanupCalled {
		t.Fatal("Effect cleanup was called prematurely during mount")
	}
	if len(Effects) != 1 {
		t.Fatalf("Expected 1 effect in internal slice, got %d", len(Effects))
	}

	// 2. Second Render: Simulate "unmounting" by not calling UseEffect
	app.Render(func(props Props) Element {
		// No UseEffect call here
		return Element{}
	}, Props{})

	if !cleanupCalled {
		t.Error("Effect cleanup was NOT called after the component stopped using the effect")
	}
	if len(Effects) != 0 {
		t.Errorf("Effect slice was NOT truncated; still has %d elements", len(Effects))
	}
}

// TestStateTruncation ensures that the global State slice is truncated
// when the component tree shrinks, preventing memory leaks from holding
// references to data no longer in use.
func TestStateTruncation(t *testing.T) {
	app := NewApp(80, 24)
	ResetComponentState()

	// 1. First Render: Populate state
	app.Render(func(props Props) Element {
		UseState("some persistent data")
		UseState(12345)
		return Element{}
	}, Props{})

	if len(State) != 2 {
		t.Fatalf("Expected 2 state slots, got %d", len(State))
	}

	// 2. Second Render: Shrink the tree (only 1 state call)
	app.Render(func(props Props) Element {
		UseState("only one")
		return Element{}
	}, Props{})

	if len(State) != 1 {
		t.Errorf("State slice was NOT truncated; expected 1, got %d", len(State))
	}
}

// TestEffectDependencyChangeCleanup verifies that changing effect dependencies
// correctly triggers the cleanup of the previous effect before running the new one.
func TestEffectDependencyChangeCleanup(t *testing.T) {
	app := NewApp(80, 24)
	ResetComponentState()

	cleanupCount := 0
	renderWithDep := func(dep int) {
		app.Render(func(props Props) Element {
			UseEffect(func() func() {
				return func() {
					cleanupCount++
				}
			}, []any{dep})
			return Element{}
		}, Props{})
	}

	// Initial mount
	renderWithDep(1)
	if cleanupCount != 0 {
		t.Error("Cleanup should not have run on mount")
	}

	// Change dependency
	renderWithDep(2)
	if cleanupCount != 1 {
		t.Errorf("Cleanup should have run once due to dependency change, got %d", cleanupCount)
	}

	// Same dependency - no cleanup
	renderWithDep(2)
	if cleanupCount != 1 {
		t.Errorf("Cleanup should not have run for identical dependency, got %d", cleanupCount)
	}
}
