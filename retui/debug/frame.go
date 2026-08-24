package debug

import (
	"sync"
	"time"
)

// FrameStats contains performance statistics for completed render frames.
//
// FPS is the observed number of completed frames per second over the
// recent FPS sampling window.
//
// FrameTime is the duration of the most recently completed frame.
//
// AvgFrameTime is the average duration of frames in the frame-duration
// sampling window.
//
// FrameCount is the total number of completed frames since the tracker
// was started or last reset.
type FrameStats struct {
	FPS          float64
	RenderRate   float64
	FrameTime    time.Duration
	AvgFrameTime time.Duration
	FrameCount   uint64
}

const defaultFrameWindowSize = 60

// frameTracker tracks completed render frames.
//
// Two separate rolling windows are maintained:
//
//   - window contains recent frame durations.
//   - fpsTimes contains timestamps of completed frames.
//
// FPS is calculated from actual completed-frame timestamps:
//
//	(N - 1) / elapsed time
//
// It is intentionally not calculated as:
//
//	time.Second / AvgFrameTime
//
// because that represents theoretical frame throughput rather than
// observed FPS.
type frameTracker struct {
	mu sync.Mutex

	// Current frame start time.
	start time.Time

	// ---------------------------------------------------------------------
	// Frame duration samples
	// ---------------------------------------------------------------------

	window     []time.Duration
	windowPos  int
	windowSize int
	windowFull bool

	// Total completed frames since reset.
	count uint64

	// ---------------------------------------------------------------------
	// FPS samples
	// ---------------------------------------------------------------------

	// Completion timestamps of recent frames.
	fpsTimes []time.Time

	// Next position to write in fpsTimes.
	fpsTimePos int

	// Number of valid timestamps currently stored.
	//
	// This is deliberately tracked separately from fpsTimePos because
	// fpsTimePos is the next write position, not the number of samples.
	fpsTimeCount int

	// Maximum number of timestamps retained for FPS calculation.
	fpsWindowSize int
}

var frames = frameTracker{
	window:        make([]time.Duration, defaultFrameWindowSize),
	windowSize:    defaultFrameWindowSize,
	fpsTimes:      make([]time.Time, defaultFrameWindowSize),
	fpsWindowSize: defaultFrameWindowSize,
}

// BeginFrame marks the beginning of a complete render frame.
//
// It should be called once immediately before the complete render cycle:
//
//	BeginFrame()
//	    build UI
//	    layout
//	    paint
//	    flush
//	EndFrame()
//
// Do not call this around only the root component function. That would
// measure UI building time rather than a complete render frame.
func BeginFrame() {
	if !Enabled() {
		return
	}

	frames.mu.Lock()
	defer frames.mu.Unlock()

	// Protect against accidental nested BeginFrame calls.
	if !frames.start.IsZero() {
		return
	}

	frames.start = timeNow()
}

// EndFrame marks the completion of a render frame.
//
// It records:
//
//   - the duration of the completed frame,
//   - the total frame count,
//   - the completion timestamp used for FPS calculation.
func EndFrame() {
	if !Enabled() {
		return
	}

	now := timeNow()

	frames.mu.Lock()
	defer frames.mu.Unlock()

	// EndFrame without BeginFrame is ignored.
	if frames.start.IsZero() {
		return
	}

	duration := now.Sub(frames.start)

	// ---------------------------------------------------------------------
	// Record frame duration
	// ---------------------------------------------------------------------

	frames.count++

	frames.window[frames.windowPos] = duration

	frames.windowPos++
	if frames.windowPos >= frames.windowSize {
		frames.windowPos = 0
		frames.windowFull = true
	}

	// ---------------------------------------------------------------------
	// Record FPS timestamp
	// ---------------------------------------------------------------------

	frames.fpsTimes[frames.fpsTimePos] = now

	frames.fpsTimePos++
	if frames.fpsTimePos >= frames.fpsWindowSize {
		frames.fpsTimePos = 0
	}

	if frames.fpsTimeCount < frames.fpsWindowSize {
		frames.fpsTimeCount++
	}

	// Mark frame as completed.
	frames.start = time.Time{}
}

// CurrentFrameStats returns the current render-frame statistics.
//
// FPS is calculated from actual completed-frame timestamps:
//
//	frames completed / elapsed wall-clock time
//
// More precisely, for N timestamps:
//
//	(N - 1) / elapsed time between the oldest and newest timestamp
//
// This avoids confusing average render time with actual observed
// application frame rate.
func CurrentFrameStats() FrameStats {

	renderRate := 0.0

	if !Enabled() {
		return FrameStats{}
	}

	frames.mu.Lock()
	defer frames.mu.Unlock()

	if frames.count == 0 {
		return FrameStats{}
	}

	// ---------------------------------------------------------------------
	// Frame duration statistics
	// ---------------------------------------------------------------------

	n := frames.windowSize

	if !frames.windowFull {
		n = frames.windowPos
	}

	var avg time.Duration

	if n > 0 {
		var total time.Duration

		for i := 0; i < n; i++ {
			total += frames.window[i]
		}

		avg = total / time.Duration(n)
	}

	// The most recently completed frame is immediately before windowPos.
	lastFrameIndex := frames.windowPos - 1

	if lastFrameIndex < 0 {
		lastFrameIndex = frames.windowSize - 1
	}

	frameTime := frames.window[lastFrameIndex]

	// ---------------------------------------------------------------------
	// FPS
	// ---------------------------------------------------------------------

	fps := calculateFPSLocked()

	if avg > 0 {
		renderRate = float64(time.Second) / float64(avg)
	}

	return FrameStats{
		FPS:          fps,
		RenderRate:   renderRate,
		FrameTime:    frameTime,
		AvgFrameTime: avg,
		FrameCount:   frames.count,
	}
}

// calculateFPSLocked calculates observed FPS from completed-frame
// timestamps.
//
// frames.mu must be held by the caller.
func calculateFPSLocked() float64 {
	n := frames.fpsTimeCount

	// One timestamp does not define a frame rate because there is no
	// elapsed interval yet.
	if n < 2 {
		return 0
	}

	var oldest time.Time
	var newest time.Time

	// The timestamp ring may have wrapped, so find the oldest and newest
	// valid samples rather than assuming their physical order.
	for i := 0; i < n; i++ {
		t := frames.fpsTimes[i]

		if t.IsZero() {
			continue
		}

		if oldest.IsZero() || t.Before(oldest) {
			oldest = t
		}

		if newest.IsZero() || t.After(newest) {
			newest = t
		}
	}

	if oldest.IsZero() || newest.IsZero() {
		return 0
	}

	elapsed := newest.Sub(oldest)

	if elapsed <= 0 {
		return 0
	}

	// N completed frames contain N-1 intervals between timestamps.
	return float64(n-1) / elapsed.Seconds()
}

// ResetFrameStats resets all frame statistics and discards all existing
// duration and FPS samples.
func ResetFrameStats() {
	frames.mu.Lock()
	defer frames.mu.Unlock()

	frames.start = time.Time{}

	// Reset frame duration window.
	frames.windowPos = 0
	frames.windowFull = false

	for i := range frames.window {
		frames.window[i] = 0
	}

	// Reset total frame count.
	frames.count = 0

	// Reset FPS timestamp window.
	frames.fpsTimePos = 0
	frames.fpsTimeCount = 0

	for i := range frames.fpsTimes {
		frames.fpsTimes[i] = time.Time{}
	}
}
