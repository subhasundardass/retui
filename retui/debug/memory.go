package debug

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"
)

// MemStats is a stable snapshot of Go runtime memory statistics.
//
// Raw memory values come directly from runtime.MemStats.
// Derived rate metrics such as GCPerSec and GCPausePerSec are calculated
// by the memory sampler.
type MemStats struct {
	// -----------------------------------------------------------------
	// Memory
	// -----------------------------------------------------------------

	Alloc       uint64
	TotalAlloc  uint64
	Sys         uint64
	HeapObjects uint64

	HeapInUse  uint64
	StackInUse uint64
	Mallocs    uint64
	Frees      uint64

	// -----------------------------------------------------------------
	// Garbage collection
	// -----------------------------------------------------------------

	NumGC         uint32
	GCPerSec      float64
	GCPause       time.Duration
	GCPausePerSec time.Duration
	LastGC        time.Time
	GCCPUFraction float64

	// -----------------------------------------------------------------
	// Runtime
	// -----------------------------------------------------------------

	Goroutines uint64
}

// memorySampler stores the previous runtime statistics needed to calculate
// rate metrics.
type memorySampler struct {
	mu sync.RWMutex

	stats MemStats

	lastSampleTime time.Time
	lastNumGC      uint32
	lastPauseNs    uint64

	started bool
}

var memSampler memorySampler

// ReadMemStats captures a fresh snapshot of Go runtime statistics.
//
// This function reads runtime statistics directly and does not maintain
// sampling state.
//
// runtime.ReadMemStats may briefly stop the world, so avoid calling this
// on every render frame. For frequently displayed debug information,
// use CurrentMemStats().
func ReadMemStats() MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return memStatsFromRuntime(m)
}

// SampleMemStats reads the current runtime memory statistics and updates
// the internal sampler.
//
// Call this periodically, for example every 500ms or 1 second, rather
// than once per render frame.
func SampleMemStats() MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	now := time.Now()
	current := memStatsFromRuntime(m)

	memSampler.mu.Lock()
	defer memSampler.mu.Unlock()

	// First sample establishes the baseline.
	if !memSampler.started {
		memSampler.stats = current
		memSampler.lastSampleTime = now
		memSampler.lastNumGC = m.NumGC
		memSampler.lastPauseNs = m.PauseTotalNs
		memSampler.started = true

		return current
	}

	elapsed := now.Sub(memSampler.lastSampleTime)

	if elapsed > 0 {
		seconds := elapsed.Seconds()

		// -------------------------------------------------------------
		// GC rate
		// -------------------------------------------------------------

		gcDelta := m.NumGC - memSampler.lastNumGC

		current.GCPerSec = float64(gcDelta) / seconds

		// -------------------------------------------------------------
		// GC pause rate
		// -------------------------------------------------------------

		pauseDelta := m.PauseTotalNs - memSampler.lastPauseNs

		current.GCPausePerSec = durationFromNanoseconds(
			float64(pauseDelta) / seconds,
		)
	}

	// Keep the latest values.
	memSampler.stats = current
	memSampler.lastSampleTime = now
	memSampler.lastNumGC = m.NumGC
	memSampler.lastPauseNs = m.PauseTotalNs

	return current
}

// CurrentMemStats returns the most recently sampled memory statistics.
//
// Unlike ReadMemStats(), this does not call runtime.ReadMemStats() and is
// therefore safe to use from the render/footer path.
func CurrentMemStats() MemStats {
	memSampler.mu.RLock()
	defer memSampler.mu.RUnlock()

	return memSampler.stats
}

// ResetMemStatsSampler resets the memory sampler.
//
// The next call to SampleMemStats() establishes a new baseline for GC/s
// and GC pause rate calculations.
func ResetMemStatsSampler() {
	memSampler.mu.Lock()
	defer memSampler.mu.Unlock()

	memSampler.stats = MemStats{}
	memSampler.lastSampleTime = time.Time{}
	memSampler.lastNumGC = 0
	memSampler.lastPauseNs = 0
	memSampler.started = false
}

// String renders memory statistics in a human-readable format.
func (m MemStats) String() string {
	return fmt.Sprintf(
		"Alloc: %.2f MB | TotalAlloc: %.2f MB | Sys: %.2f MB | HeapObjects: %d | HeapInUse: %.2f MB | StackInUse: %.2f MB | Mallocs: %d | Frees: %d | GC: %d | GC/s: %.2f | GC Pause: %s | GC Pause/s: %s | GC CPU: %.2f%% | Goroutines: %d",
		mb(m.Alloc),
		mb(m.TotalAlloc),
		mb(m.Sys),
		m.HeapObjects,
		mb(m.HeapInUse),
		mb(m.StackInUse),
		m.Mallocs,
		m.Frees,
		m.NumGC,
		m.GCPerSec,
		formatDuration(m.GCPause),
		formatDuration(m.GCPausePerSec),
		m.GCCPUFraction*100,
		m.Goroutines,
	)
}

// PrintMemory prints the most recently sampled memory statistics.
//
// Deprecated: retained for backward compatibility. Use CurrentMemStats()
// or SampleMemStats() for new code.
func PrintMemory() {
	fmt.Fprintln(currentOutput(), CurrentMemStats().String())
}

// LogMemory reports the most recently sampled memory statistics through
// the leveled debug logger.
func LogMemory() {
	if !shouldLog(LevelInfo) {
		return
	}

	writeEntry(LevelInfo, CurrentMemStats().String())
}

// memStatsFromRuntime converts runtime.MemStats into the application's
// stable MemStats representation.
func memStatsFromRuntime(m runtime.MemStats) MemStats {
	var lastGC time.Time

	if m.LastGC != 0 {
		lastGC = time.Unix(0, int64(m.LastGC))
	}

	return MemStats{
		// -------------------------------------------------------------
		// Memory
		// -------------------------------------------------------------

		Alloc:       m.Alloc,
		TotalAlloc:  m.TotalAlloc,
		Sys:         m.Sys,
		HeapObjects: m.HeapObjects,

		HeapInUse:  m.HeapInuse,
		StackInUse: m.StackInuse,
		Mallocs:    m.Mallocs,
		Frees:      m.Frees,

		// -------------------------------------------------------------
		// Garbage collection
		// -------------------------------------------------------------

		NumGC:         m.NumGC,
		GCPause:       latestGCPause(m),
		LastGC:        lastGC,
		GCCPUFraction: m.GCCPUFraction,

		// These are calculated by SampleMemStats().
		GCPerSec:      0,
		GCPausePerSec: 0,

		// -------------------------------------------------------------
		// Runtime
		// -------------------------------------------------------------

		Goroutines: uint64(runtime.NumGoroutine()),
	}
}

// latestGCPause returns the duration of the most recently completed
// garbage collection.
//
// runtime.MemStats.PauseNs is a circular buffer containing the pause
// duration for recent GC cycles. PauseEnd is used to identify the
// corresponding entry, but NumGC is sufficient to locate the latest
// completed cycle.
func latestGCPause(m runtime.MemStats) time.Duration {
	if m.NumGC == 0 {
		return 0
	}

	// PauseNs is a circular buffer with 256 entries.
	index := (m.NumGC - 1) % uint32(len(m.PauseNs))

	return time.Duration(m.PauseNs[index])
}

// durationFromNanoseconds safely converts nanoseconds represented as a
// float64 into time.Duration.
//
// time.Duration is an int64, while runtime.MemStats pause counters are
// uint64. Guarding the conversion avoids integer overflow and protects
// against impossible/out-of-range values.
func durationFromNanoseconds(ns float64) time.Duration {
	if ns <= 0 {
		return 0
	}

	if ns >= float64(math.MaxInt64) {
		return time.Duration(math.MaxInt64)
	}

	return time.Duration(ns)
}

// mb converts bytes to megabytes.
func mb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

// formatDuration produces compact durations suitable for the debug footer.
//
// Examples:
//
//	250µs
//	1.24ms
//	2.10s
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0ns"
	}

	switch {
	case d < time.Microsecond:
		return fmt.Sprintf(
			"%dns",
			d.Nanoseconds(),
		)

	case d < time.Millisecond:
		return fmt.Sprintf(
			"%.0fµs",
			float64(d)/float64(time.Microsecond),
		)

	case d < time.Second:
		return fmt.Sprintf(
			"%.2fms",
			float64(d)/float64(time.Millisecond),
		)

	default:
		return fmt.Sprintf(
			"%.2fs",
			d.Seconds(),
		)
	}
}
