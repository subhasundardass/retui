package retui

import (
	"testing"
)

// ─── Helper ──────────────────────────────────────────────────────────────────

// emptyElement returns an empty Element for testing
func emptyElement() Element {
	return Element{}
}

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
		t.Error("CreateContext returned nil")
	}
	if ctx.defaultValue != "default" {
		t.Errorf("Expected defaultValue 'default', got %v", ctx.defaultValue)
	}
}

func TestContextProvideAndUse(t *testing.T) {
	ctx := CreateContext("default")

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
