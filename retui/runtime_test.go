package retui

import (
	"runtime"
	"testing"
)

// BenchmarkRender measures the overhead of the multi-pass render cycle.
// Run with: go test -bench=BenchmarkRender -benchmem
func BenchmarkRender(b *testing.B) {
	app := NewApp(80, 24)
	// We don't call app.screen.Start() to avoid opening /dev/tty during bench

	fn := func(props Props) Element {
		return Text("benchmark", NewStyle())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Render(fn, Props{})
	}
}

// TestRenderMemoryLeak ensures that repeated renders do not leak memory.
func TestRenderMemoryLeak(t *testing.T) {
	app := NewApp(80, 24)
	fn := func(props Props) Element {
		// UseState/UseEffect to trigger hook logic
		UseState(0)
		return Text("leak-test", NewStyle())
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	initialAlloc := ms.Alloc

	for i := 0; i < 1000; i++ {
		app.Render(fn, Props{})
	}

	runtime.GC()
	runtime.ReadMemStats(&ms)

	// We allow for some minor growth due to internal Go runtime overhead,
	// but it should not be proportional to the 1000 iterations.
	if ms.Alloc > initialAlloc+(1024*1024) { // 1MB threshold
		t.Errorf("Potential memory leak detected: %d -> %d", initialAlloc, ms.Alloc)
	}
}
