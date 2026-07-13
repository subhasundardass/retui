package window

import (
	"github.com/subhasundardass/retui/retui"
)

// OverlayRenderer handles rendering windows as overlays on top of main content.
// It uses retui.Overlay() for absolute positioning of windows.
type OverlayRenderer struct{}

// NewOverlayRenderer creates a new overlay renderer with the given screen dimensions.
func NewOverlayRenderer() *OverlayRenderer {
	return &OverlayRenderer{}
}

// Render returns a full-screen element with windows properly overlaid on main content.
// The render order is: Main Content → Windows (in Z-order)
// Windows are rendered on top of main content using absolute positioning.
func (or *OverlayRenderer) Render(mainContent retui.Element) retui.Element {
	mgr := GetManager()

	// If no windows, just return the main content
	if mgr.Count() == 0 {
		return mainContent
	}

	// Start with main content as the base layer
	elements := []retui.Element{mainContent}

	// Add all visible windows as overlays in Z-order (bottom to top)
	for _, winID := range mgr.GetZOrder() {
		win := mgr.GetWindow(winID)
		if win == nil || !win.visible {
			continue
		}
		elements = append(elements, or.renderWindowAsOverlay(win))
	}

	// Stack everything in a full-screen container
	// Children are rendered in order: earlier elements are behind later ones
	return retui.Box(
		retui.Props{
			Width:  retui.Grow(1),
			Height: retui.Grow(1),
		},
		retui.NewStyle(),
		elements...,
	)
}

// renderWindowAsOverlay renders a single window as an absolute-positioned overlay.
// Uses retui.Overlay() for true absolute positioning at (w.X, w.Y).
func (or *OverlayRenderer) renderWindowAsOverlay(w *Window) retui.Element {
	// Build the complete window UI
	windowContent := or.buildWindowContent(w)

	// Place the window at its exact screen coordinates using Overlay
	return retui.Overlay(w.X, w.Y, windowContent)
}

// buildTitleBarWithColor creates title bar with custom color
func (or *OverlayRenderer) buildWindowContent(w *Window) retui.Element {
	content := w.StaticContent
	if w.RenderFn != nil {
		content = w.RenderFn() // rebuilt fresh, inside a.Render()'s bracket — hooks resolve correctly here
	}

	return retui.Box(
		retui.Props{Direction: retui.Column, Width: retui.Fit(), Height: retui.Fit()},
		retui.NewStyle(),
		or.buildTitleBar(w),
		retui.Box(
			retui.Props{Padding: [4]int{1, 1, 1, 1}},
			retui.NewStyle().Background(retui.Color{Type: retui.ColorRGB, R: 40, G: 40, B: 40}),
			content,
		),
	)
}

// buildTitleBar creates the window title bar with title text and close indicator.
func (or *OverlayRenderer) buildTitleBar(w *Window) retui.Element {
	// Set default title if empty
	title := w.Title
	if title == "" {
		title = "Window"
	}

	// Add [MODAL] indicator for modal windows
	if w.Modal {
		title = title + " [MODAL]"
	}

	// Title bar with blue background and white text
	return retui.Box(
		retui.Props{
			Direction: retui.Row,
			Width:     retui.Fixed(w.Width),
			Height:    retui.Fixed(1),
		},
		retui.NewStyle().
			Background(retui.Blue).
			Foreground(retui.White),
		// Title with a space prefix for padding
		retui.Text(" "+title, retui.NewStyle().Bold(true)),
		// Flexible spacer to push close button to the right
		retui.Box(
			retui.Props{Width: retui.Grow(1)},
			retui.NewStyle(),
		),
		// Close button indicator (visual only for now)
		retui.Text("[Esc]", retui.NewStyle().Foreground(retui.BrightRed)),
	)
}

// RenderWindowsOverlay is a convenience function that renders windows as overlays.
// This is the recommended way to integrate window management into your app.
//
// Usage:
//
//	func App(props retui.Props) retui.Element {
//	    mainContent := buildYourMainUI()
//	    return window.RenderWindowsOverlay(140, 40, mainContent)
//	}
func RenderWindowsOverlay(mainContent retui.Element) retui.Element {
	renderer := NewOverlayRenderer()

	return renderer.Render(mainContent)
}
