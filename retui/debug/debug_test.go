package debug

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestEnableDisable(t *testing.T) {
	Disable()
	if Enabled() {
		t.Fatal("expected disabled by default after Disable()")
	}
	Enable()
	defer Disable()
	if !Enabled() {
		t.Fatal("expected enabled after Enable()")
	}
}

func TestLogLevelGating(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	Enable()
	defer Disable()
	SetLevel(LevelWarn)

	Infof("should not appear")
	Warnf("should appear")

	out := buf.String()
	if strings.Contains(out, "should not appear") {
		t.Errorf("Infof output leaked at LevelWarn: %q", out)
	}
	if !strings.Contains(out, "should appear") {
		t.Errorf("Warnf output missing: %q", out)
	}
}

func TestDisabledProducesNoOutput(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	Disable()
	SetLevel(LevelTrace)
	Errorf("should not appear while disabled")

	if buf.Len() != 0 {
		t.Errorf("expected no output while disabled, got %q", buf.String())
	}
}

func TestEntriesRingBuffer(t *testing.T) {
	ClearEntries()
	SetHistorySize(3)
	Enable()
	SetLevel(LevelTrace)
	defer Disable()

	Infof("one")
	Infof("two")
	Infof("three")
	Infof("four")

	es := Entries()
	if len(es) != 3 {
		t.Fatalf("expected 3 retained entries, got %d", len(es))
	}
	if es[0].Message != "two" || es[2].Message != "four" {
		t.Errorf("unexpected ring contents: %+v", es)
	}

	SetHistorySize(512) // restore default for other tests
	ClearEntries()
}

func TestReadMemStats(t *testing.T) {
	m := ReadMemStats()
	if m.Sys == 0 {
		t.Error("expected non-zero Sys memory")
	}
}

func TestPrintMemoryBackwardCompatible(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	Disable()

	PrintMemory()

	out := buf.String()
	for _, want := range []string{
		"Alloc:",
		"TotalAlloc:",
		"Sys:",
		"GC:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("PrintMemory output missing %q: %q", want, out)
		}
	}
}

func TestLogMemoryRespectsGating(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	Disable()
	LogMemory()
	if buf.Len() != 0 {
		t.Errorf("expected LogMemory to produce nothing while disabled, got %q", buf.String())
	}

	Enable()
	defer Disable()
	SetLevel(LevelInfo)
	LogMemory()
	if !strings.Contains(buf.String(), "Alloc:") {
		t.Errorf("expected LogMemory output once enabled, got %q", buf.String())
	}
}

func TestFrameStats(t *testing.T) {
	Enable()
	defer Disable()
	ResetFrameStats()

	for i := 0; i < 10; i++ {
		BeginFrame()
		time.Sleep(2 * time.Millisecond)
		EndFrame()
	}

	fs := CurrentFrameStats()

	if fs.FrameCount != 10 {
		t.Fatalf("expected 10 frames recorded, got %d", fs.FrameCount)
	}

	if fs.FrameTime <= 0 {
		t.Errorf("expected positive frame time, got %v", fs.FrameTime)
	}

	if fs.FPS <= 0 {
		t.Errorf("expected positive FPS, got %v", fs.FPS)
	}
}

func TestFrameStatsOffWhenDisabled(t *testing.T) {
	Disable()
	ResetFrameStats()

	BeginFrame()
	time.Sleep(time.Millisecond)
	EndFrame()

	fs := CurrentFrameStats()
	if fs.FrameCount != 0 {
		t.Errorf("expected no frames recorded while disabled, got %d", fs.FrameCount)
	}
}

func TestRecoverAndLog(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	func() {
		defer RecoverAndLog("test")()
		panic("boom")
	}()

	if !strings.Contains(buf.String(), "panic in test: boom") {
		t.Errorf("expected panic log, got %q", buf.String())
	}
}

func TestRecoverAndReraisePropagatesPanic(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	defer func() {
		r := recover()
		if r != "boom" {
			t.Errorf("expected panic to propagate with value 'boom', got %v", r)
		}
		if !strings.Contains(buf.String(), "panic in test-reraise: boom") {
			t.Errorf("expected panic log before reraise, got %q", buf.String())
		}
	}()

	func() {
		defer RecoverAndReraise("test-reraise")()
		panic("boom")
	}()
}
