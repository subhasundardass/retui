package retui

import (
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/subhasundardass/retui/retui/debug"
)

var debugPanelVisible atomic.Bool

// ToggleDebugPanel toggles the built-in diagnostics panel.
func ToggleDebugPanel() {
	debugPanelVisible.Store(!debugPanelVisible.Load())
}

// DebugPanelVisible reports whether the diagnostics panel is visible.
func DebugPanelVisible() bool {
	return debugPanelVisible.Load()
}

// CloseDebugPanel closes the diagnostics panel.
func CloseDebugPanel() {
	debugPanelVisible.Store(false)
}

// DebugPanel renders RetUI's built-in full-screen diagnostics dashboard.
//
// The panel lives in the retui package so applications do not need to
// import or register a diagnostics component.
//
// Layout is a 2x2 grid of bordered sections beneath a header bar, with a
// keybinding footer pinned to the bottom:
//
//	┌──────── RETUI / DEVELOPER DIAGNOSTICS ────────┐
//	│ ┌──── RENDERING ────┐ ┌──── MEMORY ──────────┐ │
//	│ └────────────────────┘ └───────────────────────┘ │
//	│ ┌── GARBAGE COLLECTION ──┐ ┌──── RUNTIME ─────┐ │
//	│ └─────────────────────────┘ └───────────────────┘ │
//	│ RetUI Diagnostics       F12 Close  R Reset  C Clear │
//	└─────────────────────────────────────────────────┘
func DebugPanel() Element {
	stats := debug.CurrentFrameStats()
	mem := debug.CurrentMemStats()

	return Box(
		Props{
			Width:     Grow(1),
			Height:    Grow(1),
			Padding:   [4]int{1, 1, 1, 1},
			Gap:       1,
			Direction: Column,
		},
		NewStyle().
			Border(retuiDebugBorder()),

		// Header
		debugPanelHeader(),

		// Main dashboard: a 2x2 grid of sections.
		Box(
			Props{
				Width:  Grow(1),
				Height: Grow(1),
				Gap:    0,
			},
			NewStyle(),

			// Row 1 — Rendering / Memory
			debugPanelRow(
				debugPanelSection(
					"RENDERING",
					debugRow("FPS", fmt.Sprintf("%.1f", stats.FPS)),
					debugRow("Render Rate", fmt.Sprintf("%.1f / sec", stats.RenderRate)),
					debugRow("Frame Time", formatDebugDuration(stats.FrameTime)),
					debugRow("Average Frame", formatDebugDuration(stats.AvgFrameTime)),
					debugRow("Frame Count", fmt.Sprintf("%d", stats.FrameCount)),
				),
				debugPanelSection(
					"MEMORY",
					debugRow("Allocated", formatDebugBytes(mem.Alloc)),
					debugRow("Total Allocated", formatDebugBytes(mem.TotalAlloc)),
					debugRow("System", formatDebugBytes(mem.Sys)),
					debugRow("Heap Objects", fmt.Sprintf("%d", mem.HeapObjects)),
					debugRow("Heap In Use", formatDebugBytes(mem.HeapInUse)),
					debugRow("Stack In Use", formatDebugBytes(mem.StackInUse)),
					debugRow("Mallocs", fmt.Sprintf("%d", mem.Mallocs)),
					debugRow("Frees", fmt.Sprintf("%d", mem.Frees)),
				),
			),

			// Row 2 — Garbage Collection / Runtime
			debugPanelRow(
				debugPanelSection(
					"GARBAGE COLLECTION",
					debugRow("GC Cycles", fmt.Sprintf("%d", mem.NumGC)),
					debugRow("GC / sec", fmt.Sprintf("%.2f", mem.GCPerSec)),
					debugRow("GC Pause", formatDebugDuration(mem.GCPause)),
					debugRow("GC Pause / sec", formatDebugDuration(mem.GCPausePerSec)),
					debugRow("Last GC", formatDebugGCTime(mem.LastGC)),
					debugRow("GC CPU", fmt.Sprintf("%.2f%%", mem.GCCPUFraction*100)),
				),
				debugPanelSection(
					"RUNTIME",
					debugRow("Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine())),
					debugRow("Go Version", runtime.Version()),
					debugRow("OS", runtime.GOOS),
					debugRow("Architecture", runtime.GOARCH),
					debugRow("CPU", fmt.Sprintf("%d", runtime.NumCPU())),
					debugRow("PID", fmt.Sprintf("%d", os.Getpid())),
				),
			),
		),

		// Footer
		// debugPanelFooter(),
	)
}

// -----------------------------------------------------------------------------
// Header
// -----------------------------------------------------------------------------
func formatDebugGCTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	return formatDebugUptime(time.Since(t)) + " ago"
}
func debugPanelHeader() Element {
	return Box(
		Props{
			Width:   Grow(1),
			Justify: JustifySpaceBetween,
			Align:   AlignCenter,
			Padding: [4]int{0, 1, 0, 1},
		},
		NewStyle(),

		Text(
			" RETUI  /  DEVELOPER DIAGNOSTICS ",
			NewStyle().
				Foreground(Gold).Bold(true),
		),

		Text(
			"F12  Close",
			NewStyle(),
		),
	)
}
func formatDebugUptime(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	d = d.Truncate(time.Second)

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute
	d -= minutes * time.Minute

	seconds := d / time.Second

	switch {
	case days > 0:
		return fmt.Sprintf(
			"%dd%02dh%02dm%02ds",
			days,
			hours,
			minutes,
			seconds,
		)

	case hours > 0:
		return fmt.Sprintf(
			"%dh%02dm%02ds",
			hours,
			minutes,
			seconds,
		)

	case minutes > 0:
		return fmt.Sprintf(
			"%dm%02ds",
			minutes,
			seconds,
		)

	default:
		return fmt.Sprintf(
			"%ds",
			seconds,
		)
	}
}

// -----------------------------------------------------------------------------
// Grid row
// -----------------------------------------------------------------------------

// debugPanelRow lays two dashboard sections side by side, each taking an
// equal share of the available width. Used to build the 2x2 metrics grid.
func debugPanelRow(left, right Element) Element {
	return Box(
		Props{
			Width: Grow(1),
			Gap:   0,
		},
		NewStyle(),

		left,
		right,
	)
}

// -----------------------------------------------------------------------------
// Sections
// -----------------------------------------------------------------------------

func debugPanelSection(title string, rows ...Element) Element {
	children := make([]Element, 0, len(rows)+1)

	children = append(
		children,
		Box(
			Props{
				Width: Grow(1),
				Padding: [4]int{
					0,
					1,
					0,
					1,
				},
			},
			NewStyle().
				Background(Gray(2)),

			Text(
				title,
				NewStyle().
					Foreground(Cyan),
			),
		),
	)

	children = append(children, rows...)

	return Box(
		Props{
			Width:  Grow(1),
			Height: Grow(1),
			Gap:    0,
			Padding: [4]int{
				1,
				1,
				1,
				1,
			},
			Direction: Column,
		},
		NewStyle().
			Border(retuiDebugSectionBorder()),

		children...,
	)
}

// -----------------------------------------------------------------------------
// Metric row
// -----------------------------------------------------------------------------

func debugRow(label, value string) Element {
	return Box(
		Props{
			Width:   Grow(1),
			Align:   AlignCenter,
			Justify: JustifySpaceBetween,
		},
		NewStyle(),

		Text(
			label,
			NewStyle(),
		),

		Text(
			value,
			NewStyle().
				Foreground(Cyan).Bold(true),
		),
	)
}

// -----------------------------------------------------------------------------
// Footer
// -----------------------------------------------------------------------------

// func debugPanelFooter() Element {
// 	return Box(
// 		Props{
// 			Width:   Grow(1),
// 			Justify: JustifySpaceBetween,
// 			Align:   AlignCenter,
// 			Padding: [4]int{0, 1, 0, 1},
// 		},
// 		NewStyle().
// 			Border(retuiDebugFooterBorder()),

// 		Text(
// 			"RetUI Diagnostics",
// 			NewStyle().
// 				Foreground(Gray(5)),
// 		),

// 		Box(
// 			Props{
// 				Gap: 0,
// 			},
// 			NewStyle(),

// 			Text(
// 				"F12",
// 				NewStyle().
// 					Foreground(Cyan),
// 			),
// 			Text("Close", Style{}),

// 			Text(
// 				"R",
// 				NewStyle().
// 					Foreground(Cyan),
// 			),
// 			Text("Reset", Style{}),

// 			Text(
// 				"C",
// 				NewStyle().
// 					Foreground(Cyan),
// 			),
// 			Text("Clear Logs", Style{}),
// 		),
// 	)
// }

// -----------------------------------------------------------------------------
// Borders
// -----------------------------------------------------------------------------

func retuiDebugBorder() Border {
	return Border{
		Top:    true,
		Right:  true,
		Bottom: true,
		Left:   true,
		Color:  Cyan,
	}
}

func retuiDebugSectionBorder() Border {
	return Border{
		Top:    true,
		Right:  true,
		Bottom: true,
		Left:   true,
		Color:  Gray(4),
	}
}

func retuiDebugFooterBorder() Border {
	return Border{
		Top:   true,
		Color: Gray(4),
	}
}

// -----------------------------------------------------------------------------
// Formatting
// -----------------------------------------------------------------------------

func formatDebugBytes(v uint64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)

	switch {
	case v >= gb:
		return fmt.Sprintf("%.2f GB", float64(v)/float64(gb))

	case v >= mb:
		return fmt.Sprintf("%.2f MB", float64(v)/float64(mb))

	case v >= kb:
		return fmt.Sprintf("%.2f KB", float64(v)/float64(kb))

	default:
		return fmt.Sprintf("%d B", v)
	}
}

func formatDebugDuration(d time.Duration) string {
	switch {
	case d <= 0:
		return "0ns"

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
