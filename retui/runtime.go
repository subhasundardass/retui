package retui

import (
	"os"
	"sync"
	"time"

	"github.com/subhasundardass/retui/retui/debug"
)

var CurrentScreenWidth int
var CurrentScreenHeight int

// ============================================================================
// App — Terminal Application Runtime
// ============================================================================

// App is the root controller for a terminal-based application.
//
// App manages:
//   - Screen lifecycle (initialization, cleanup)
//   - Event loop (keyboard, timer, resize)
//   - Rendering pipeline (three-pass system with state/effect hooks)
//   - Modal/window layer integration
//
// App is not thread-safe; all calls must come from the same goroutine
// (typically the main thread). Use channels or queues for inter-goroutine
// communication.
type App struct {
	screen   *Screen
	renderer *ComponentRenderer
	focus    *FocusManager
}

// ============================================================================
// App Initialization
// ============================================================================

// NewApp creates and initializes a new terminal application.
//
// Bootstrap Process:
//
//  1. Creates a Screen that writes to os.Stdout
//  2. Puts the terminal into raw mode (no echo, line buffering disabled)
//  3. Hides the cursor and enables bracketed paste mode
//  4. Detects the actual terminal dimensions
//  5. Creates the root Renderer
//  6. Returns a fully initialized App instance
//
// Parameters:
//
//	width:  Fallback width (in columns) if terminal size cannot be detected.
//	        Used when stdout is redirected or in CI environments where
//	        terminal.GetSize() fails (e.g., piped output, Docker).
//
//	height: Fallback height (in rows) if terminal size cannot be detected.
//
// Terminal Size Detection:
//
// In normal interactive terminals, the actual terminal viewport dimensions
// are detected automatically and used, regardless of width/height arguments.
// Fallback dimensions are only used if detection fails.
//
// Returns: A fully initialized App ready for Run(). NewApp panics if Screen
// creation fails.
//
// Example:
//
//	app := NewApp(80, 24)  // 80x24 fallback for piped/CI environments
//	app.Run(renderFunc, props)
//
// Note: NewApp starts the Screen to detect terminal size. The Screen is
// started again in Run() for proper cleanup (see Run() for details).
func NewApp(width, height int) *App {

	screen := NewScreen(width, height, os.Stdout)
	screen.Start()

	// Prefer real terminal dimensions over fallback args so layout fills
	// the actual viewport. Args remain a fallback for piped/CI where
	// terminal.GetSize fails.
	if screen.termCols > 0 && screen.termRows > 0 {
		screen.Resize(screen.termCols, screen.termRows)
	} else {
		screen.Resize(width, height)
	}

	renderer := NewRenderer(screen)

	return &App{
		screen:   screen,
		renderer: renderer,
		focus:    globalFocus,
	}
}

// ============================================================================
// Global Hooks & Event Channels
// ============================================================================

// RootRenderWrap, if set, wraps the root element before rendering.
// Used by overlay/modal packages (like window) to composite overlays without
// creating import cycles.
//
// Called twice per render frame:
//  1. Pass 1 (hooks phase): With realKey for modal/window processing
//  2. Pass 2 (render phase): With cleared key (already consumed)
//
// Return value from Pass 1 is discarded (side effects only).
// Return value from Pass 2 is used for rendering.
//
// Nil by default. Set by modal/window layer at init time.
var RootRenderWrap func(Element) Element

// WindowKeyDispatch, if set, intercepts keyboard input for modal/window layer.
// Called when a key is pressed to allow modal stack to handle it before
// the root application.
//
// Returns true if the key was handled by the modal layer (consumed);
// false if the root application should handle it.
//
// Nil by default. Set by window package.
var WindowKeyDispatch func(key Key) bool

// IsAnyModalOpenFn, if set, reports whether any modal is currently open.
// Used to determine if root content should be blocked from keyboard input.
//
// Nil by default. Set by window package.
var IsAnyModalOpenFn func() bool

// exitCh receives a signal when Exit() is called, requesting graceful shutdown.
// Buffered (size 1) to prevent goroutine blocking if called multiple times.
var exitCh = make(chan struct{}, 1)

// CurrentKey is the most recent keyboard input, available to all hook functions.
// Cleared after each render frame so hooks must check it during their frame.
//
// NOT thread-safe; read/write from the event loop goroutine only.
var CurrentKey Key

// CurrentTick is the state of the periodic tick (alternates every 500ms).
// Available to hook functions for reactive effects (e.g., blinking cursor).
//
// NOT thread-safe; read/write from the event loop goroutine only.
var CurrentTick bool

// ============================================================================
// Global Shutdown Control
// ============================================================================

// Exit requests the running application to stop gracefully.
// Safe to call from any goroutine; uses buffered channel to avoid blocking.
// Calling Exit() multiple times is safe.
func Exit() {
	select {
	case exitCh <- struct{}{}:
	default:
		// Channel full; exit already requested
	}
}

// ============================================================================
// Event Loop & Rendering
// ============================================================================

// Run starts the application's event loop and renders the UI.
//
// Event Loop:
//
// Run manages a select loop that multiplexes several input sources:
//   - Keyboard input (os.Stdin)
//   - Timer tick (500ms interval)
//   - Terminal resize (SIGWINCH)
//   - Explicit exit request (Exit() or Ctrl+C)
//
// Each event triggers a render cycle via fn(props) and optional hook processing.
//
// Rendering Pipeline (Three Passes):
//
// Each render frame executes three passes:
//
//	Pass 1 (Hook Setup): fn(props) is called with CurrentKey set (if not blocked by modal).
//	The element tree is built and any hook functions (State, Effect, Render) run.
//	RootRenderWrap is called for pass effects only (return value discarded).
//	This pass mutates hook state (StateCursor, EffectCursor).
//
//	Pass 2 (Render): fn(props) is called again with CurrentKey cleared.
//	New element tree is built with updated hook state.
//	RootRenderWrap wraps the tree (return value used).
//	Tree is rendered to screen via Renderer.Render() and flushed.
//	Effect hooks (RunEffects) are executed, may request another render.
//
//	Pass 3 (Re-render): If RunEffects set pendingRender, Pass 2 repeats
//	with CurrentKey and hook cursors reset.
//
// This pipeline ensures:
//   - Hook state is computed before rendering
//   - Rendering sees the latest state
//   - Effects can schedule a re-render in the same frame
//   - Key input is consumed immediately (not seen by effects)
//
// Modal Blocking:
//
// When IsAnyModalOpenFn() returns true:
//   - Root content's CurrentKey is cleared (receives no input)
//   - Modal layer receives CurrentKey via WindowKeyDispatch
//   - Tab key has special handling to prevent root interference
//   - Modal can consume keys and the root never sees them
//
// Shutdown:
//
// Run exits gracefully when:
//   - Ctrl+C is pressed (requestQuit called)
//   - os.Stdin read fails (requestQuit called)
//   - Exit() is called (exitCh signal, requestQuit called)
//   - quit channel is closed (select branch returns)
//
// Before returning, defer block calls a.screen.Stop() to restore the terminal
// to its original state.
//
// Parameters:
//
//	fn:    Render function that builds the element tree given props.
//	       Called three times per frame (hook setup, render, re-render if needed).
//
//	props: Props passed to fn on every call. Not modified by Run.
//
// Example:
//
//	func render(props Props) Element {
//	    return Element{
//	        Type:     ElementBox,
//	        Layout:   LayoutProps{Direction: Row},
//	        Children: []Element{...},
//	    }
//	}
//
//	app.Run(render, Props{})
func (a *App) Run(fn func(props Props) Element, props Props) {

	// Start screen for real.
	a.screen.Start()
	defer a.screen.Stop()

	quit := make(chan struct{})
	var quitOnce sync.Once

	requestQuit := func() {
		quitOnce.Do(func() {
			close(quit)
		})
	}

	resize, stopResize := newResizeChan()
	defer stopResize()

	tickerCh := make(chan bool, 1)

	go func() {
		tick := false

		for {
			time.Sleep(500 * time.Millisecond)
			tick = !tick

			select {
			case tickerCh <- tick:
			case <-quit:
				return
			default:
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		scanner := KeyScanner{}

		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				requestQuit()
				return
			}

			for _, key := range scanner.Feed(buf[:n]) {
				if key.Code == KeyCtrlC {
					requestQuit()
					return
				}

				Keys <- key
			}
		}
	}()

	// Initial render.
	a.renderFrame(fn, props)

	// Main event loop.
	for {
		select {
		case <-quit:
			return

		case <-exitCh:
			requestQuit()

		case key := <-Keys:

			// ---------------------------------------------------------
			// RetUI built-in debug panel
			// ---------------------------------------------------------
			//
			// F12 is handled by RetUI itself so applications do not
			// need to register a debug key handler.
			if key.Code == KeyF12 {
				Infof("KEY: code=%v", key.Code)
				ToggleDebugPanel()
				a.renderFrame(fn, props)
				continue
			}

			// ---------------------------------------------------------
			// Normal application key handling
			// ---------------------------------------------------------

			modalOpen := IsAnyModalOpenFn != nil &&
				IsAnyModalOpenFn()

			if key.Code == KeyTab && modalOpen {
				if WindowKeyDispatch != nil {
					WindowKeyDispatch(key)
				}

				a.renderFrame(fn, props)
				continue
			}

			CurrentKey = key

			consumed := false

			if WindowKeyDispatch != nil {
				consumed = WindowKeyDispatch(key)
			}

			if consumed {
				CurrentKey = Key{}
			}

			a.renderFrame(fn, props)

		case tick := <-tickerCh:
			CurrentTick = tick

			a.renderFrame(fn, props)

		case <-resize:
			a.screen.HandleResize()
			a.screen.ForceMarkAllDirty()

			a.renderFrame(fn, props)
		}
	}
}

// Render executes one complete render cycle (three-pass system).
//
// The three-pass system ensures hooks see updated state and effects can
// trigger immediate re-renders without requiring another event loop cycle.
//
// Pass 1 (Hook Setup):
//   - Element tree is built via fn(props)
//   - Hooks run in order (State, then Effects from prior frame)
//   - RootRenderWrap receives the tree for modal pass effects
//   - Return value from RootRenderWrap is discarded (side effects only)
//   - StateCursor and EffectCursor reset at frame start and advance as hooks run
//
// Pass 2 (Render):
//   - CurrentKey is cleared so root doesn't see keys consumed by modal
//   - Element tree is built again via fn(props) with updated hook state
//   - RootRenderWrap wraps the tree for rendering (return value used)
//   - Tree is rendered to screen and flushed
//   - RunEffects() executes effect hooks; may request another render
//
// Pass 3 (Re-render on Effect):
//   - If RunEffects() set pendingRender, immediately repeat Pass 2
//   - Allows effects to update state and see the result in the same frame
//   - Helps keep UI responsive without waiting for next event
//
// Modal Blocking:
//   - If a modal is open, root's CurrentKey is cleared in Pass 1
//   - Modal layer handles the key via WindowKeyDispatch
//   - Root component never sees the key
//
// Parameters:
//
//	fn:    Render function; builds element tree given props
//	props: Props passed to fn; typically includes app state
func (a *App) Render(fn func(props Props) Element, props Props) {

	CurrentScreenWidth = a.screen.Width()
	CurrentScreenHeight = a.screen.Height()

	modalOpen := IsAnyModalOpenFn != nil && IsAnyModalOpenFn()
	realKey := CurrentKey

	// ========================================================================
	// Pass 1: Hook Setup (build tree, run hooks for state setup)
	// ========================================================================
	StateCursor = 0
	EffectCursor = 0
	if modalOpen {
		// Root is blocked by modal; don't expose CurrentKey
		CurrentKey = Key{}
	}
	root := fn(props)
	if RootRenderWrap != nil {
		// Let modal layer process the key (side effects only; return discarded)
		CurrentKey = realKey
		RootRenderWrap(root)
	}

	// ========================================================================
	// Pass 2: Render (build updated tree, render to screen, run effects)
	// ========================================================================
	CurrentKey = Key{} // Clear for root (already consumed if modal handled it)
	StateCursor = 0
	EffectCursor = 0
	next := fn(props)

	if RootRenderWrap != nil {
		next = RootRenderWrap(next)
	}

	a.renderer.Render(next)
	a.screen.Flush()

	// Run effect hooks; may request another render
	pendingRender = false
	RunEffects()

	// ========================================================================
	// Pass 3: Re-render if Effects Requested (repeat Pass 2)
	// ========================================================================
	if pendingRender {
		CurrentKey = Key{}
		StateCursor = 0
		EffectCursor = 0
		next := fn(props)
		if RootRenderWrap != nil {
			next = RootRenderWrap(next)
		}
		a.renderer.Render(next)
		a.screen.Flush()
	}
}

func (a *App) renderFrame(
	fn func(props Props) Element,
	props Props,
) {
	if debug.Enabled() {
		debug.BeginFrame()
	}

	if DebugPanelVisible() {
		a.Render(
			func(props Props) Element {
				return DebugPanel()
			},
			props,
		)
	} else {
		a.Render(fn, props)
	}

	if debug.Enabled() {
		debug.EndFrame()
	}
}
