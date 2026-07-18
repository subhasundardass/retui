package retui

import (
	"testing"
	"time"
)

// ─── TestNewApp ─────────────────────────────────────────────────────────────

func TestNewApp(t *testing.T) {
	app := NewApp(80, 24)
	if app == nil {
		t.Error("NewApp returned nil")
	}
	if app.screen == nil {
		t.Error("App.screen is nil")
	}
	if app.focus == nil {
		t.Error("App.focus is nil")
	}
	if app.renderer == nil {
		t.Error("App.renderer is nil")
	}
}

func TestNewAppDimensions(t *testing.T) {
	// Test with fallback dimensions
	app := NewApp(80, 24)
	if app.screen.Width() != 80 && app.screen.termCols == 0 {
		t.Errorf("Expected width 80, got %d", app.screen.Width())
	}
	if app.screen.Height() != 24 && app.screen.termRows == 0 {
		t.Errorf("Expected height 24, got %d", app.screen.Height())
	}
}

// ─── TestAppRender ──────────────────────────────────────────────────────────

func TestAppRender(t *testing.T) {
	//Don't use defer screen.Stop() - Run handles it
	app := NewApp(80, 24)

	// Reset state
	State = nil
	StateCursor = 0
	Effects = nil
	EffectCursor = 0
	pendingRender = false

	renderCount := 0
	testFn := func(props Props) Element {
		renderCount++
		return Text("test", NewStyle())
	}

	// Initial render
	app.Render(testFn, Props{})

	//Render should call the function:
	// - Pass 1 (process keys, mutate state)
	// - Pass 2 (render with updated state)
	// - If pendingRender is true, a third pass is done
	// So renderCount should be 2 or 3 depending on pendingRender

	if renderCount < 2 {
		t.Errorf("Expected render count at least 2, got %d", renderCount)
	}
}

func TestAppRenderWithState(t *testing.T) {
	app := NewApp(80, 24)
	defer app.screen.Stop()

	State = nil
	StateCursor = 0
	Effects = nil
	EffectCursor = 0
	pendingRender = false

	var stateVal int
	testFn := func(props Props) Element {
		val, setVal := UseState(0)
		stateVal = val
		if CurrentKey.Code == KeyEnter {
			setVal(val + 1)
		}
		return Text("test", NewStyle())
	}

	// Initial render
	app.Render(testFn, Props{})
	if stateVal != 0 {
		t.Errorf("Expected state 0, got %d", stateVal)
	}
}

// ─── TestExit ──────────────────────────────────────────────────────────────

func TestExit(t *testing.T) {
	// Exit should not panic
	Exit()
}

// ─── TestAppRun ────────────────────────────────────────────────────────────

func TestAppRun(t *testing.T) {
	// This test runs the app briefly to ensure no panics
	app := NewApp(80, 24)
	defer app.screen.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(100 * time.Millisecond)
		Exit()
	}()

	go func() {
		app.Run(func(props Props) Element {
			return Text("test", NewStyle())
		}, Props{})
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("App.Run did not exit within 2 seconds")
	}
}

// ─── TestAppRunWithPanic ──────────────────────────────────────────────────

func TestAppRunWithPanic(t *testing.T) {
	app := NewApp(80, 24)
	defer app.screen.Stop()

	// Panic should be recovered
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected
			}
		}()
		app.Run(func(props Props) Element {
			panic("test panic")
		}, Props{})
	}()

	time.Sleep(100 * time.Millisecond)
}

// ─── TestKeyHandling ──────────────────────────────────────────────────────

func TestKeyHandlingInApp(t *testing.T) {
	app := NewApp(80, 24)
	defer app.screen.Stop()

	State = nil
	StateCursor = 0
	Effects = nil
	EffectCursor = 0
	pendingRender = false

	keyReceived := false
	testFn := func(props Props) Element {
		if CurrentKey.Code == KeyEnter {
			keyReceived = true
		}
		return Text("test", NewStyle())
	}

	// Set a key
	CurrentKey = Key{Code: KeyEnter}
	app.Render(testFn, Props{})
	CurrentKey = Key{} // Reset

	if !keyReceived {
		t.Error("Key was not received in render function")
	}
}

// ─── TestTicker ───────────────────────────────────────────────────────────

func TestTicker(t *testing.T) {
	app := NewApp(80, 24)
	defer app.screen.Stop()

	State = nil
	StateCursor = 0
	Effects = nil
	EffectCursor = 0
	pendingRender = false

	tickerReceived := false
	testFn := func(props Props) Element {
		if CurrentTick {
			tickerReceived = true
		}
		return Text("test", NewStyle())
	}

	// Simulate tick
	CurrentTick = true
	app.Render(testFn, Props{})
	CurrentTick = false

	if !tickerReceived {
		t.Error("Tick was not received in render function")
	}
}

// ─── BenchmarkAppRender ──────────────────────────────────────────────────

func BenchmarkAppRender(b *testing.B) {
	app := NewApp(80, 24)
	defer app.screen.Stop()

	State = nil
	StateCursor = 0
	Effects = nil
	EffectCursor = 0
	pendingRender = false

	testFn := func(props Props) Element {
		return Text("benchmark", NewStyle())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Render(testFn, Props{})
	}
}

// ─── BenchmarkAppRun ─────────────────────────────────────────────────────

func BenchmarkAppRun(b *testing.B) {
	app := NewApp(80, 24)
	defer app.screen.Stop()

	for i := 0; i < b.N; i++ {
		done := make(chan bool)
		go func() {
			time.Sleep(10 * time.Millisecond)
			Exit()
		}()
		go func() {
			app.Run(func(props Props) Element {
				return Text("benchmark", NewStyle())
			}, Props{})
			done <- true
		}()
		<-done
	}
}
