// Global focus
package retui

import "sync"

// FocusManager controls global keyboard focus in the application.
//
// It ensures that only one component is "active" at a time,
// and only the focused component should respond to keyboard input.
//
// It supports:
//   - Single focus ID (active component)
//   - Focus order (Tab / Shift+Tab navigation)
//   - Focus stack (for modal dialogs)
//   - Focus capture (for popups/dropdowns to trap events)
type FocusManager struct {
	mu sync.RWMutex

	// Current focused component ID
	current string

	// Ordered list of focusable component IDs (Tab navigation)
	order []string

	// Stack used for modal / temporary focus overrides
	stack []string

	// Capture ID - component that has captured all keyboard events
	// When set, ONLY this component receives keyboard events
	capture string
}

// NewFocusManager creates a new focus system.
func NewFocusManager() *FocusManager {
	return &FocusManager{
		order:   make([]string, 0),
		stack:   make([]string, 0),
		capture: "",
	}
}

//
// -----------------------------
// BASIC FOCUS CONTROL
// -----------------------------

// SetFocus sets the active focused component.
func (f *FocusManager) Focus(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if id == "" {
		f.current = ""
		return
	}

	f.current = id
}

// Current returns the currently focused component ID.
func (f *FocusManager) Current() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current
}

// IsFocused checks whether a component is currently focused.
func (f *FocusManager) IsFocused(id string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current == id
}

// global accessor
func (a *App) FocusManager() *FocusManager {
	return a.focus
}

//
// -----------------------------
// FOCUS CAPTURE (for popups/dropdowns)
// -----------------------------

// CaptureFocus captures all keyboard events to the specified component.
// When a component captures focus, ONLY that component receives keyboard events.
// This is useful for dropdowns, modals, and popups that need to trap focus.
func (f *FocusManager) CaptureFocus(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if id == "" {
		return
	}

	f.capture = id
	// Also set as current focus
	f.current = id
	Debug("🔒 CaptureFocus: " + id + " has captured all keyboard events")
}

// ReleaseCapture releases the capture, allowing normal focus routing.
func (f *FocusManager) ReleaseCapture() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.capture != "" {
		Debug("🔓 ReleaseCapture: " + f.capture + " released capture")
		f.capture = ""
	}
}

// Captured returns the ID of the component that has captured focus.
// Returns empty string if no component has captured focus.
func (f *FocusManager) Captured() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.capture
}

// IsCaptured checks whether a specific component has captured focus.
func (f *FocusManager) IsCaptured(id string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.capture == id
}

//
// -----------------------------
// TAB NAVIGATION
// -----------------------------

// SetOrder defines the Tab navigation order.
func (f *FocusManager) SetOrder(order []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.order = order
}

// Next moves focus to next element in order (Tab).
func (f *FocusManager) Next() {
	f.mu.Lock()
	defer f.mu.Unlock()

	// If capture is active, don't navigate
	if f.capture != "" {
		return
	}

	if len(f.order) == 0 {
		return
	}

	// If no current focus, start at first
	if f.current == "" {
		f.current = f.order[0]
		return
	}

	start := f.indexOf(f.current)
	if start == -1 {
		f.current = f.order[0]
		return
	}

	next := (start + 1) % len(f.order)
	f.current = f.order[next]
}

// Prev moves focus to previous element (Shift+Tab).
func (f *FocusManager) Prev() {
	f.mu.Lock()
	defer f.mu.Unlock()

	// If capture is active, don't navigate
	if f.capture != "" {
		return
	}

	if len(f.order) == 0 {
		return
	}

	// If no current focus, start at last
	if f.current == "" {
		f.current = f.order[len(f.order)-1]
		return
	}

	idx := f.indexOf(f.current)
	if idx == -1 {
		f.current = f.order[len(f.order)-1]
		return
	}

	prev := (idx - 1 + len(f.order)) % len(f.order)
	f.current = f.order[prev]
}

// indexOf finds index of current focus in order list.
func (f *FocusManager) indexOf(id string) int {
	for i, v := range f.order {
		if v == id {
			return i
		}
	}
	return -1
}

//
// -----------------------------
// MODAL / STACK SUPPORT
// -----------------------------

// PushFocus stores current focus and switches to new one.
// Used when opening modal/dialog.
func (f *FocusManager) PushFocus(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.current != "" {
		f.stack = append(f.stack, f.current)
	}

	f.current = id
	Debug("📌 PushFocus: " + id + " pushed to stack")
}

// PopFocus restores previous focus.
// Used when closing modal/dialog.
func (f *FocusManager) PopFocus() {
	f.mu.Lock()
	defer f.mu.Unlock()

	n := len(f.stack)
	if n == 0 {
		return
	}

	last := f.stack[n-1]
	f.stack = f.stack[:n-1]

	f.current = last
	Debug("📌 PopFocus: restored to " + last)
}

// ClearStack clears the focus stack (useful for emergency cleanup).
func (f *FocusManager) ClearStack() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stack = make([]string, 0)
}

//
// -----------------------------
// GLOBAL INSTANCE (SIMPLE MODE)
// -----------------------------

var globalFocus = NewFocusManager()

// CURRENT - returns current focus
func CurrentFocus() string {
	return globalFocus.Current()
}

// SETORDER - sets Tab navigation order
func SetFocusOrder(order []string) {
	globalFocus.SetOrder(order)
}

// Focus sets global focus.
func SetFocus(id string) {
	globalFocus.Focus(id)
}

// Blur clears focus.
func Blur() {
	globalFocus.Focus("")
}

// IsFocused checks global focus.
func IsFocused(id string) bool {
	return globalFocus.IsFocused(id)
}

// FocusNext moves to next focusable item.
func FocusNext() {
	globalFocus.Next()
}

// FocusPrev moves to previous focusable item.
func FocusPrev() {
	globalFocus.Prev()
}

// PushFocus opens focus.
func PushFocus(id string) {
	globalFocus.PushFocus(id)
}

// PopFocus closes focus.
func PopFocus() {
	globalFocus.PopFocus()
}

// ─────────────────────────────────────────────────────────────────────────────
// FOCUS CAPTURE GLOBAL FUNCTIONS
// ─────────────────────────────────────────────────────────────────────────────

// CaptureFocus captures all keyboard events to the specified component.
// When a component captures focus, ONLY that component receives keyboard events.
// Useful for dropdowns, popups, and modal dialogs that need to trap focus.
func CaptureFocus(id string) {
	globalFocus.CaptureFocus(id)
}

// ReleaseCapture releases the focus capture.
func ReleaseCaptureFocus() {
	globalFocus.ReleaseCapture()
}

// Captured returns the ID of the component that has captured focus.
// Returns empty string if no component has captured focus.
func CapturedFocus() string {
	return globalFocus.Captured()
}

// IsCaptured checks whether a specific component has captured focus.
func IsCaptured(id string) bool {
	return globalFocus.IsCaptured(id)
}

// ─────────────────────────────────────────────────────────────────────────────
// UseFocusedKey - Handles key events with proper focus and capture checks
// ─────────────────────────────────────────────────────────────────────────────

// UseFocusedKey processes keyboard events with focus and capture awareness.
//
// It checks:
//  1. If there's a key event
//  2. If the component is allowed to process it (focus or capture)
//  3. Prevents duplicate key processing (dedup)
//
// Returns the key and whether this component should process it.
func UseFocusedKey(componentID string, isFocused bool) (key Key, isMine bool) {
	key = CurrentKey

	// No physical key this frame.
	if key == (Key{}) {
		return key, false
	}

	// ──────────────────────────────────────────────────────────────
	// STEP 1: Check if something has captured focus
	// ──────────────────────────────────────────────────────────────
	captured := globalFocus.Captured()
	if captured != "" {
		// Something has captured focus - ONLY the capturer gets keys
		if captured != componentID {
			// This component is NOT the capturer - skip
			return key, false
		}
		// This IS the capturer - it gets the key
		// Continue to dedup check
	} else {
		// ──────────────────────────────────────────────────────────
		// STEP 2: No capture - check normal focus
		// ──────────────────────────────────────────────────────────
		if !isFocused {
			// This component is not focused - skip
			return key, false
		}
	}

	// ──────────────────────────────────────────────────────────────
	// STEP 3: Dedup guard - prevent processing same key twice
	// ──────────────────────────────────────────────────────────────
	lastClaimed, setLastClaimed := UseState(Key{})
	if key == lastClaimed {
		// This key was already processed this tick - skip
		return key, false
	}
	setLastClaimed(key)

	// All checks passed - this component can process the key
	return key, true
}

// UseFocusedKeySimple is a simpler version that doesn't require component ID.
// It only checks if the component is focused.
// Use this for simple components that don't need capture awareness.
func UseFocusedKeySimple(isFocused bool) (key Key, isMine bool) {
	key = CurrentKey

	if key == (Key{}) {
		return key, false
	}

	// NEW: if something else has captured focus, nobody without the
	// matching ID gets keys — capture always wins over local `isFocused`.
	if captured := globalFocus.Captured(); captured != "" {
		return key, false
	}

	if !isFocused {
		return key, false
	}

	// Dedup guard
	lastClaimed, setLastClaimed := UseState(Key{})
	if key == lastClaimed {
		return key, false
	}
	setLastClaimed(key)

	return key, true
}
