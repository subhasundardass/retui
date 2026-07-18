package window

import (
	"fmt"
	"sync"

	"github.com/subhasundardass/retui/retui"
)

var (
	DefaultScreenWidth  = 140
	DefaultScreenHeight = 40
)

// SetDefaultScreenSize sets the default screen size for auto-centering
func SetDefaultScreenSize(width, height int) {
	DefaultScreenWidth = width
	DefaultScreenHeight = height
}

// Window represents a single window in the application.
// It holds state, content, and rendering information.
type Window struct {
	mu            sync.RWMutex // Protects window state
	ID            string
	Title         string
	Width         int
	Height        int
	X             int           // Position on screen
	Y             int           // Position on screen
	Modal         bool          // Whether this window blocks background interaction
	StaticContent retui.Element // The content to display inside the window
	visible       bool          // Whether the window is currently shown
	focused       bool          // Whether this window has focus
	onKeyPress    func(key retui.Key) bool
	RenderFn      func() retui.Element
}

// NewWindow creates a new window with given content.
// Window starts hidden - you must call Show() to display it.
func NewWindow() *Window {

	width := 40
	height := 15

	return &Window{
		ID:      generateID(),
		Title:   "Window",
		Width:   40,
		Height:  15,
		X:       (DefaultScreenWidth - width) / 2,
		Y:       (DefaultScreenHeight - height) / 2,
		Modal:   false,
		visible: false,
		focused: false,
	}
}

// ========================================
// CONFIGURATION METHODS (Fluent API)
// ========================================

// SetTitle sets the window title.
func (w *Window) SetTitle(title string) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Title = title
	return w
}

// SetSize sets window dimensions.
func (w *Window) SetSize(width, height int) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Width = width
	w.Height = height
	return w
}

// SetPosition sets window coordinates on screen.
func (w *Window) SetPosition(x, y int) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.X = x
	w.Y = y
	return w
}

// SetModal marks this window as modal (blocks background input).
func (w *Window) SetModal(modal bool) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Modal = modal
	return w
}

func (w *Window) SetRenderFn(fn func() retui.Element) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.RenderFn = fn
	return w
}

// ========================================
// LIFECYCLE METHODS
// ========================================

// Show makes the window visible and adds it to the manager.
// Triggers a re-render automatically.
func (w *Window) Show() *Window {

	w.mu.Lock()
	w.visible = true
	w.mu.Unlock()

	globalManager.AddWindow(w)
	return w
}

// Hide makes the window invisible but keeps it in the registry.
// The window can be shown again later with Show().
func (w *Window) Hide() *Window {
	w.mu.Lock()
	w.visible = false
	w.mu.Unlock()

	// Trigger re-render
	go globalManager.triggerRender()
	return w
}

// Close removes the window from the manager and hides it.
// The window cannot be shown again after closing.
func (w *Window) Close() {
	retui.Infof("Closing window: %s", w.ID)
	w.mu.Lock()
	w.visible = false
	w.mu.Unlock()

	globalManager.RemoveWindow(w.ID)
}

// ToggleVisibility toggles the window between visible and hidden.
// Returns the new visibility state.
func (w *Window) ToggleVisibility() bool {
	w.mu.Lock()
	w.visible = !w.visible
	newState := w.visible
	w.mu.Unlock()

	if newState {
		globalManager.AddWindow(w)
	} else {
		globalManager.triggerRender()
	}
	return newState
}

// OnKeyPress registers a callback invoked whenever this window has focus
// and receives a key event. Returns the window for chaining.
func (w *Window) OnKeyPress(fn func(key retui.Key) bool) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onKeyPress = fn
	return w
}

// HandleKey is called by the framework's key-dispatch loop when this
func (w *Window) HandleKey(key retui.Key) bool {

	w.mu.RLock()
	fn := w.onKeyPress
	w.mu.RUnlock()

	if fn != nil {
		return fn(key)
	}

	return false
}

func GetFocusedID() string {
	return globalManager.GetFocused()
}

func GetWindowByID(id string) *Window {
	return globalManager.GetWindow(id)
}

// ========================================
// FOCUS AND Z-ORDER
// ========================================

// Focus brings this window to the front (top of Z-order).
func (w *Window) Focus() {
	globalManager.BringToFront(w.ID)
}

// IsFocused returns whether this window is currently focused.
func (w *Window) IsFocused() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.focused
}

// ========================================
// QUERY METHODS
// ========================================

// IsVisible returns whether the window is currently displayed.
func (w *Window) IsVisible() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.visible
}

// IsModal returns whether this window is modal.
func (w *Window) IsModal() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.Modal
}

// IsActive returns whether this window is focused (alias for IsFocused).
func (w *Window) IsActive() bool {
	return w.IsFocused()
}

// GetTitle returns the window title.
func (w *Window) GetTitle() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.Title
}

// GetSize returns the window dimensions as [width, height].
func (w *Window) GetSize() [2]int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return [2]int{w.Width, w.Height}
}

// GetPosition returns the window position as [x, y].
func (w *Window) GetPosition() [2]int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return [2]int{w.X, w.Y}
}

// GetBounds returns the window's position and size as [x, y, width, height].
func (w *Window) GetBounds() [4]int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return [4]int{w.X, w.Y, w.Width, w.Height}
}

// ========================================
// POSITIONING HELPERS
// ========================================
// Center automatically centers the window using the manager's screen dimensions
func (w *Window) Center() *Window {
	w.mu.Lock()
	defer w.mu.Unlock()

	screenWidth, screenHeight := globalManager.GetScreenSize()

	// If width or height is 0, set defaults
	if w.Width == 0 {
		w.Width = 40
	}
	if w.Height == 0 {
		w.Height = 15
	}

	w.X = (screenWidth - w.Width) / 2
	w.Y = (screenHeight - w.Height) / 2

	// Clamp to 0 when window is larger than screen
	if w.X < 0 {
		w.X = 0
	}
	if w.Y < 0 {
		w.Y = 0
	}

	// Ensure window doesn't overflow the screen (if possible)
	if w.X+w.Width > screenWidth && w.Width <= screenWidth {
		w.X = screenWidth - w.Width
	}
	if w.Y+w.Height > screenHeight && w.Height <= screenHeight {
		w.Y = screenHeight - w.Height
	}

	return w
}

// CenterHorizontally centers the window horizontally on screen.
func (w *Window) CenterHorizontally(screenWidth int) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.X = (screenWidth - w.Width) / 2
	if w.X < 0 {
		w.X = 0
	}
	if w.X+w.Width > screenWidth {
		w.X = screenWidth - w.Width
	}
	return w
}

// CenterVertically centers the window vertically on screen.
func (w *Window) CenterVertically(screenHeight int) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.Y = (screenHeight - w.Height) / 2
	if w.Y < 0 {
		w.Y = 0
	}
	if w.Y+w.Height > screenHeight {
		w.Y = screenHeight - w.Height
	}
	return w
}

// MoveTo moves the window to the specified position.
func (w *Window) MoveTo(x, y int) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.X = x
	w.Y = y
	return w
}

// MoveBy moves the window by the specified offset.
func (w *Window) MoveBy(dx, dy int) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.X += dx
	w.Y += dy
	return w
}

// ResizeTo resizes the window to the specified dimensions.
func (w *Window) ResizeTo(width, height int) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Width = width
	w.Height = height
	return w
}

// ========================================
// CONTENT MANAGEMENT
// ========================================

// SetContent updates the window's content.
// Triggers a re-render automatically.
func (w *Window) SetContent(content retui.Element) *Window {
	w.mu.Lock()
	w.StaticContent = content
	w.mu.Unlock()

	go globalManager.triggerRender()
	return w
}

// GetContent returns the window's current content.
func (w *Window) GetContent() retui.Element {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.StaticContent
}

// Render returns the retui.Element representing this window's UI.
// Called by the OverlayRenderer during render cycle.
func (w *Window) Render() retui.Element {
	w.mu.RLock()
	fn := w.RenderFn
	content := w.StaticContent
	w.mu.RUnlock()

	if fn != nil {
		return fn()
	}

	return content
}

// ========================================
// UTILITY METHODS
// ========================================

// String returns a string representation of the window.
func (w *Window) String() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return fmt.Sprintf("Window{ID: %s, Title: %s, Size: %dx%d, Pos: (%d,%d), Visible: %v, Modal: %v}",
		w.ID, w.Title, w.Width, w.Height, w.X, w.Y, w.visible, w.Modal)
}

// Clone creates a copy of the window with a new ID.
// Useful for duplicating windows.
func (w *Window) Clone() *Window {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return &Window{
		ID:            generateID(),
		Title:         w.Title + " (Copy)",
		Width:         w.Width,
		Height:        w.Height,
		X:             w.X + 5, // Offset slightly
		Y:             w.Y + 5,
		Modal:         w.Modal,
		StaticContent: w.StaticContent,
		visible:       false,
		focused:       false,
	}
}

// ========================================
// ID GENERATION
// ========================================

var (
	windowCounter = 0
	counterMu     sync.Mutex
)

func generateID() string {
	counterMu.Lock()
	defer counterMu.Unlock()

	windowCounter++
	return fmt.Sprintf("win-%d", windowCounter)
}

// closeFocusedWindow closes the currently focused window
func CloseFocusedWindow() {
	// Try focused window first
	focusedID := globalManager.GetFocused()
	if focusedID != "" {
		win := globalManager.GetWindow(focusedID)
		if win != nil {
			retui.Debugf("Closing focused window: %s", win.Title)
			win.Close()
			return
		}
	}

	// If no focused window, close the top window
	zorder := globalManager.GetZOrder()
	if len(zorder) > 0 {
		topID := zorder[len(zorder)-1]
		win := globalManager.GetWindow(topID)
		if win != nil {
			retui.Debugf("Closing top window: %s", win.Title)
			win.Close()
		}
	}
}
