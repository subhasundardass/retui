package retui

import (
	"fmt"
	"sync"
)

// ============================================================================
// Focus Management
// ============================================================================

// FocusManager manages keyboard focus and event routing for UI components.
//
// Focus System:
//   - Current: The component that receives keyboard input
//   - Order: Tab navigation order for cycling through focusable components
//   - Stack: Modal/dialog stack for nested focus contexts
//   - Capture: A component that intercepts all keyboard events (e.g., dropdown)
//
// Focus Flow:
//  1. A component is set as current (focused)
//  2. User presses Tab/Shift+Tab to navigate focus order
//  3. A component can capture focus to handle all keys (dropdown, modal)
//  4. Modals push/pop focus on the stack
//  5. UseFocusedKey hook checks if current component should handle the key
//
// Thread Safety:
//
//	FocusManager is thread-safe via RWMutex. All operations are protected.
//	Typical use is single-threaded (TUI event loop), but concurrent access
//	from multiple goroutines is safe.
type FocusManager struct {
	mu sync.RWMutex

	// current is the ID of the component that currently has focus.
	current string

	// order is the tab navigation order (list of focusable component IDs).
	// Used by Next() and Prev() to cycle focus through components.
	order []string

	// stack is the focus stack for modals/dialogs.
	// When a modal opens, current is pushed to stack and modal becomes current.
	// When modal closes, stack is popped and previous focus is restored.
	stack []string

	// capture is the ID of the component that has captured focus.
	// When set, that component receives all keyboard input (no other component
	// can process keys). Used for dropdowns, autocomplete, etc.
	capture string
}

// NewFocusManager creates a new focus manager with no initial focus.
//
// Returns: A fully initialized FocusManager ready for use.
func NewFocusManager() *FocusManager {
	return &FocusManager{
		order: []string{},
		stack: []string{},
	}
}

// ============================================================================
// Focus Control
// ============================================================================

// SetFocus sets the current focus to the specified component ID.
//
// Parameters:
//
//	id: The component ID to focus. Must be in the focus order.
//
// If id is not in the order list, it still becomes current, but Tab navigation
// won't cycle through it. This is allowed for components that are focusable
// but not tab-navigable (e.g., dynamically added components).
func (f *FocusManager) SetFocus(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.current = id
}

// Current returns the ID of the component that currently has focus.
//
// Returns: The ID of the focused component, or empty string if no focus.
func (f *FocusManager) Current() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current
}

// IsFocused checks if the specified component ID is the current focus.
//
// Parameters:
//
//	id: The component ID to check.
//
// Returns: True if id is currently focused, false otherwise.
func (f *FocusManager) IsFocused(id string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current == id
}

// ============================================================================
// Tab Navigation
// ============================================================================

// SetOrder sets the tab navigation order for cycling through components.
//
// Parameters:
//
//	order: A list of component IDs in the order they should be visited
//	       when the user presses Tab or Shift+Tab.
//
// Validation:
//   - All IDs must be non-empty
//   - No duplicate IDs allowed
//   - Returns error if validation fails
//
// If SetOrder fails, the previous order remains unchanged.
func (f *FocusManager) SetOrder(order []string) error {
	// Validate: no empty strings, no duplicates
	seen := make(map[string]bool)
	for _, id := range order {
		if id == "" {
			return fmt.Errorf("empty component ID in focus order")
		}
		if seen[id] {
			return fmt.Errorf("duplicate component ID in focus order: %s", id)
		}
		seen[id] = true
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.order = order
	return nil
}

// Next moves focus to the next component in the tab order.
//
// Behavior:
//   - If current is at the end, wraps to the start (circular)
//   - Does nothing if order is empty
//   - Does nothing if capture is active (dropdown/modal has focus lock)
//   - Typically called when user presses Tab
//
// Returns: The ID of the new focus, or empty string if no change.
func (f *FocusManager) Next() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Capture blocks navigation
	if f.capture != "" {
		return f.current
	}

	if len(f.order) == 0 {
		return f.current
	}

	idx := f.indexOf(f.current)
	if idx < 0 {
		// Current not in order; jump to first
		f.current = f.order[0]
		return f.current
	}

	// Move to next, wrap to start if at end
	nextIdx := (idx + 1) % len(f.order)
	f.current = f.order[nextIdx]
	return f.current
}

// Prev moves focus to the previous component in the tab order.
//
// Behavior:
//   - If current is at the start, wraps to the end (circular)
//   - Does nothing if order is empty
//   - Does nothing if capture is active
//   - Typically called when user presses Shift+Tab
//
// Returns: The ID of the new focus, or empty string if no change.
func (f *FocusManager) Prev() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Capture blocks navigation
	if f.capture != "" {
		return f.current
	}

	if len(f.order) == 0 {
		return f.current
	}

	idx := f.indexOf(f.current)
	if idx < 0 {
		// Current not in order; jump to last
		f.current = f.order[len(f.order)-1]
		return f.current
	}

	// Move to previous, wrap to end if at start
	nextIdx := idx - 1
	if nextIdx < 0 {
		nextIdx = len(f.order) - 1
	}
	f.current = f.order[nextIdx]
	return f.current
}

// indexOf returns the index of id in the focus order, or -1 if not found.
func (f *FocusManager) indexOf(id string) int {
	for i, v := range f.order {
		if v == id {
			return i
		}
	}
	return -1
}

// ============================================================================
// Modal/Dialog Stack
// ============================================================================

// PushFocus pushes the current focus onto the stack and sets a new focus.
//
// Used when opening a modal dialog:
//   - Current focus is saved to the stack
//   - New focus (modal) becomes current
//   - Focus capture is automatically released (modal doesn't inherit capture)
//
// Parameters:
//
//	id: The component ID for the modal/dialog.
//
// When the modal closes, call PopFocus() to restore the previous focus.
func (f *FocusManager) PushFocus(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.current != "" {
		f.stack = append(f.stack, f.current)
	}

	f.current = id

	// CRITICAL: Release capture when modal opens
	// Otherwise, dropdown capture would block modal input
	f.capture = ""
}

// PopFocus pops the focus stack and restores the previous focus.
//
// Used when closing a modal dialog:
//   - Current focus (modal) is discarded
//   - Previous focus is restored from the stack
//   - Returns to normal focus flow
//
// If the stack is empty, does nothing (stays at current focus).
//
// Returns: The ID of the restored focus, or current if stack is empty.
func (f *FocusManager) PopFocus() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.stack) == 0 {
		return f.current
	}

	// Pop from stack
	f.current = f.stack[len(f.stack)-1]
	f.stack = f.stack[:len(f.stack)-1]

	return f.current
}

// ============================================================================
// Focus Capture (Dropdown, Autocomplete)
// ============================================================================

// CaptureFocus sets a component as the exclusive keyboard handler.
//
// Used for dropdowns, autocomplete, and other input-capturing components:
//   - All keyboard input goes to the capturing component
//   - Other components don't receive keys
//   - Tab/Shift+Tab navigation is blocked
//   - Typically released when dropdown closes or user presses Escape
//
// Parameters:
//
//	id: The component ID to capture focus.
//
// Only one component can have capture at a time. Calling CaptureFocus again
// replaces the previous capture.
func (f *FocusManager) CaptureFocus(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.capture = id
}

// ReleaseFocus releases keyboard capture.
//
// Called when the capturing component (dropdown, etc.) closes.
// Restores normal focus flow to the current component.
//
// If no component has capture, does nothing.
func (f *FocusManager) ReleaseFocus() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.capture = ""
}

// CapturedBy returns the ID of the component with keyboard capture, or empty string.
//
// Returns: The ID of the capturing component, or empty string if no capture.
func (f *FocusManager) CapturedBy() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.capture
}

// HasCapture checks if a specific component has keyboard capture.
//
// Parameters:
//
//	id: The component ID to check.
//
// Returns: True if id has capture, false otherwise.
func (f *FocusManager) HasCapture(id string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.capture == id
}

// ============================================================================
// Frame-Level Key Consumption
// ============================================================================

// Note: Key consumption is tracked via the Key.consumed field directly.
// No separate tracking needed; the Key struct manages consumption state.

// ============================================================================
// Hook: UseFocusedKey
// ============================================================================

// UseFocusedKey is a hook for components to check if a keyboard key is theirs.
//
// Usage:
//
//	In your component render function:
//
//	func myComponent(props Element) Element {
//	    key, isMine := UseFocusedKey("my-component-id", isFocused)
//	    if isMine && key.Code == KeyEnter {
//	        handleEnter()
//	    }
//	    // ...
//	}
//
// Behavior:
//   - Returns the current key and whether the component should handle it
//   - Component only handles the key if:
//     1. Component has focus (current)
//     2. Key hasn't been consumed by another component (Key.consumed is false)
//   - Marks the key as consumed (Key.consumed = true) when claimed
//   - Prevents key duplication: each key is handled by exactly one component
//   - Returns empty key and false if component shouldn't handle input
//
// Parameters:
//
//	componentID: Unique identifier for this component
//	isFocused:   True if this component is visually focused
//
// Returns:
//
//	key:   The current keyboard input (or empty if not applicable)
//	isMine: True if this component should handle the key, false otherwise
//
// Key Consumption:
//   - Uses Key.consumed field to track if key was already claimed
//   - When a component claims a key, Key.consumed is set to true
//   - Other components see Key.consumed=true and don't claim it
//   - CurrentKey is cleared after each render frame, resetting all keys
//
// Example:
//
//	func render(props Props) Element {
//	    key, isMine := UseFocusedKey("input", currentFocus == "input")
//	    if isMine {
//	        switch key.Code {
//	        case KeyEnter:
//	            return handleEnter()
//	        case KeyEscape:
//	            return handleEscape()
//	        }
//	    }
//	    return Element{...}
//	}
func UseFocusedKey(componentID string, isFocused bool) (key Key, isMine bool) {
	key = CurrentKey
	if key == (Key{}) {
		return key, false
	}

	// If someone captured the keyboard, only they receive keys.
	if captured := CapturedFocus(); captured != "" {
		if captured != componentID {
			return key, false
		}
	} else {
		// Otherwise normal focus rules apply.
		if !isFocused {
			return key, false
		}
	}

	// Prevent multiple components from handling the same key.
	if key.Consumed {
		return key, false
	}

	CurrentKey.Consumed = true
	return key, true
}

// ============================================================================
// Global Convenience Functions
// ============================================================================

var globalFocus = NewFocusManager()

// SetFocus sets the global focus to a component.
func SetFocus(id string) {
	globalFocus.SetFocus(id)
}

// CurrentFocus returns the currently focused component ID.
func CurrentFocus() string {
	return globalFocus.Current()
}

// IsFocused checks if a component has global focus.
func IsFocused(id string) bool {
	return globalFocus.IsFocused(id)
}

// SetFocusOrder sets the global tab navigation order.
func SetFocusOrder(order []string) error {
	return globalFocus.SetOrder(order)
}

// FocusNext moves focus to the next component in the tab order.
func FocusNext() {
	globalFocus.Next()
}

// FocusPrev moves focus to the previous component in the tab order.
func FocusPrev() {
	globalFocus.Prev()
}

// PushFocus pushes focus onto the modal stack.
func PushFocus(id string) {
	globalFocus.PushFocus(id)
}

// PopFocus pops focus from the modal stack.
func PopFocus() {
	globalFocus.PopFocus()
}

// CaptureFocus captures keyboard input for a component.
func CaptureFocus(id string) {
	globalFocus.CaptureFocus(id)
}

// ReleaseFocus releases keyboard capture.
func ReleaseFocus() {
	globalFocus.ReleaseFocus()
}

// CapturedFocus returns the ID of the component with capture.
func CapturedFocus() string {
	return globalFocus.CapturedBy()
}

// HasFocusCapture checks if a component has capture.
func HasFocusCapture(id string) bool {
	return globalFocus.HasCapture(id)
}
