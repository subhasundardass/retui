package retui

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	screen   *Screen
	renderer *ComponentRenderer
	focus    *FocusManager
}

// NewApp creates and initializes a new terminal application.
//
// NewApp performs the complete application bootstrap process:
//
//  1. Creates a Screen that writes to os.Stdout.
//  2. Puts the terminal into raw mode.
//  3. Hides the cursor and enables bracketed paste mode.
//  4. Detects the actual terminal dimensions.
//  5. Creates the root Renderer.
//  6. Returns a fully initialized App instance.
//
// The width and height arguments act as fallback dimensions when the
// terminal size cannot be determined (for example when stdout is
// redirected or running in certain CI environments).
//
// In normal interactive terminals, the application's dimensions are
// automatically expanded to the real terminal viewport, regardless of
// the values passed to this constructor.
//

// RootRenderWrap, if set, wraps the root element before it's rendered.
// Used by packages like `window` to composite overlays without re
// needing to import them (avoids import cycle).
var RootRenderWrap func(Element) Element // nil by default

func NewApp(width, height int) *App {

	screen := NewScreenWriter(width, height, os.Stdout)
	screen.Start()

	// Prefer the real terminal dimensions over the constructor args so
	// layout fills the actual viewport. The args remain a fallback for
	// environments where term.GetSize fails (e.g. piped output).
	if screen.termCols > 0 && screen.termRows > 0 {
		screen.SetDimensions(screen.termCols, screen.termRows)
	} else {
		screen.SetDimensions(width, height)
	}

	renderer := NewRenderer(screen)

	return &App{
		screen:   screen,
		renderer: renderer,
		focus:    globalFocus,
	}
}

var ticker = make(chan bool, 1)
var CurrentTick bool = false

var exitCh = make(chan struct{}, 1)

// in retui package, app.go or similar
var WindowKeyDispatch func(key Key) // nil by default
var IsAnyModalOpenFn func() bool

// Exit requests the running application to stop gracefully.
func Exit() {
	select {
	case exitCh <- struct{}{}:
	default:
	}
}

func (a *App) Run(fn func(props Props) Element, props Props) {

	//Start the screen
	a.screen.Start()
	defer a.screen.Stop() //ALWAYS restore terminal on exit

	//Exit channel for graceful shutdown
	// exitCh := make(chan struct{})
	quit := make(chan struct{})
	var quitOnce sync.Once
	requestQuit := func() {
		quitOnce.Do(func() { close(quit) })
	}

	resize := make(chan os.Signal, 1)
	signal.Notify(resize, syscall.SIGWINCH)

	//Ticker for periodic updates
	ticker := make(chan bool, 1)
	go func() {
		tick := false
		for {
			time.Sleep(time.Millisecond * 500)
			tick = !tick
			select {
			case ticker <- tick:
			case <-quit: // ← stop goroutine when app exits
				return
			default:
				// Channel full, skip
			}
		}
	}()

	//Keyboard input handler
	go func() {
		buf := make([]byte, 1024)
		var scanner KeyScanner
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

	//Initial render
	a.Render(fn, props)

	//Main event loop
	for {
		select {
		case <-quit:
			return
		case <-exitCh:
			requestQuit()
		case key := <-Keys:

			modalOpen := IsAnyModalOpenFn != nil && IsAnyModalOpenFn()
			if key.Code == KeyTab && modalOpen {
				// Don't expose Tab to CurrentKey at all — prevents any widget
				// hook (parent or otherwise) from reacting to it this frame.
				if WindowKeyDispatch != nil {
					WindowKeyDispatch(key) // your own modal-focus-cycling still runs
				}
				a.Render(fn, props)
				break // out of the select case
			}

			CurrentKey = key
			// dispatch to focused window, if any
			if WindowKeyDispatch != nil {
				WindowKeyDispatch(key)
			}

			a.Render(fn, props)
		case tick := <-ticker:
			CurrentTick = tick
			a.Render(fn, props)
		case <-resize:
			a.screen.HandleResize()
			a.screen.ForceMarkAllDirty()
			a.Render(fn, props)
		}
	}
}

func (a *App) Render(fn func(props Props) Element, props Props) {
	modalOpen := IsAnyModalOpenFn != nil && IsAnyModalOpenFn()
	realKey := CurrentKey

	// Pass 1: process key events and mutate state
	StateCursor = 0
	EffectCursor = 0
	if modalOpen {
		CurrentKey = Key{} // background content never reacts to the key while blocked
	}
	root := fn(props)
	if RootRenderWrap != nil {
		CurrentKey = realKey // window/modal content always gets the real key
		RootRenderWrap(root) // result discarded — this pass is for hook side effects only
	}

	// Pass 2: render with updated state; key is now consumed for everyone
	CurrentKey = Key{}
	StateCursor = 0
	EffectCursor = 0
	next := fn(props)

	if RootRenderWrap != nil {
		next = RootRenderWrap(next)
	}

	a.renderer.Render(next)
	a.screen.Flush()

	pendingRender = false
	RunEffects()

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
