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
//   - Disabled/Hidden/ReadOnly: per-component state flags that affect focus
//
// Component State Flags:
//
//	Disabled: Component cannot receive focus at all. Skipped by Next()/Prev(),
//	          rejected by SetFocus(). Existing capture is released if the
//	          captured component becomes disabled.
//	Hidden:   Same focus behavior as Disabled (invisible components shouldn't
//	          receive keyboard input), tracked separately so callers can query
//	          "why" a component isn't focusable.
//	ReadOnly: Does NOT affect focus at all. A read-only component can still be
//	          tabbed to (e.g. to scroll/select/copy its content); it's up to
//	          the component's own key handler to consult IsReadOnly() and
//	          refuse to mutate its value.
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

	// disabled tracks which component IDs cannot receive focus at all.
	disabled map[string]bool

	// hidden tracks which component IDs are not visible and cannot receive focus.
	hidden map[string]bool

	// readonly tracks which component IDs are focusable but not editable.
	// This does not affect focus routing; components consult it themselves.
	readonly map[string]bool
}

// NewFocusManager creates a new focus manager with no initial focus.
//
// Returns: A fully initialized FocusManager ready for use.
func NewFocusManager() *FocusManager {
	return &FocusManager{
		order:    []string{},
		stack:    []string{},
		disabled: make(map[string]bool),
		hidden:   make(map[string]bool),
		readonly: make(map[string]bool),
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
//
// If id is disabled or hidden, focus is NOT changed and false is returned.
//
// Returns: True if focus was granted, false if id is disabled/hidden.
func (f *FocusManager) SetFocus(id string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.disabled[id] || f.hidden[id] {
		return false
	}

	f.current = id
	return true
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
// Component State: Disabled / ReadOnly / Hidden
// ============================================================================

// SetDisabled marks a component as disabled or enabled.
//
// A disabled component cannot receive focus: it is skipped by Next()/Prev()
// and rejected by SetFocus(). If the currently focused component is disabled,
// focus automatically advances to the next focusable component in the tab
// order (or clears to "" if none remain). Any active keyboard capture held by
// a component being disabled is also released.
//
// Triggers a re-render automatically if the flag actually changes (no-op,
// no render, if disabled already matched the current state).
func (f *FocusManager) SetDisabled(id string, disabled bool) {
	f.mu.Lock()
	changed := f.disabled[id] != disabled
	if disabled {
		f.disabled[id] = true
	} else {
		delete(f.disabled, id)
	}
	if disabled {
		f.handleUnfocusable(id)
	}
	f.mu.Unlock()

	if changed {
		requestRender()
	}
}

// IsDisabled checks if a component is disabled.
func (f *FocusManager) IsDisabled(id string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.disabled[id]
}

// SetHidden marks a component as hidden or visible.
//
// A hidden component cannot receive focus: it is skipped by Next()/Prev()
// and rejected by SetFocus(). If the currently focused component becomes
// hidden, focus automatically advances to the next focusable component (or
// clears to "" if none remain). Any active keyboard capture held by a
// component being hidden is also released.
//
// Triggers a re-render automatically if the flag actually changes (no-op,
// no render, if hidden already matched the current state).
func (f *FocusManager) SetHidden(id string, hidden bool) {
	f.mu.Lock()
	changed := f.hidden[id] != hidden
	if hidden {
		f.hidden[id] = true
	} else {
		delete(f.hidden, id)
	}
	if hidden {
		f.handleUnfocusable(id)
	}
	f.mu.Unlock()

	if changed {
		requestRender()
	}
}

// IsHidden checks if a component is hidden.
func (f *FocusManager) IsHidden(id string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.hidden[id]
}

// SetReadOnly marks a component as read-only or editable.
//
// ReadOnly does NOT affect focus routing — a read-only component remains
// focusable via Tab/SetFocus so its content can still be viewed, selected,
// or scrolled. Components should consult IsReadOnly() in their own key
// handlers (typically via UseFocusedKey) and refuse to mutate state when true.
//
// Triggers a re-render automatically if the flag actually changes (no-op,
// no render, if readonly already matched the current state) — a field's
// visual styling (e.g. dimmed/lock icon) typically depends on this flag.
func (f *FocusManager) SetReadOnly(id string, readonly bool) {
	f.mu.Lock()
	changed := f.readonly[id] != readonly
	if readonly {
		f.readonly[id] = true
	} else {
		delete(f.readonly, id)
	}
	f.mu.Unlock()

	if changed {
		requestRender()
	}
}

// IsReadOnly checks if a component is read-only.
func (f *FocusManager) IsReadOnly(id string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.readonly[id]
}

// handleUnfocusable is called (while already holding the lock) after a
// component transitions to disabled or hidden. If that component currently
// holds focus or capture, it releases capture and advances focus to the next
// focusable component in tab order.
func (f *FocusManager) handleUnfocusable(id string) {
	if f.capture == id {
		f.capture = ""
	}

	if f.current != id {
		return
	}

	// Try to advance to the next focusable component in order.
	next := f.nextFocusableFrom(id, +1)
	f.current = next
}

// isFocusable reports whether id can currently receive focus.
// Caller must hold at least a read lock.
func (f *FocusManager) isFocusable(id string) bool {
	return !f.disabled[id] && !f.hidden[id]
}

// nextFocusableFrom scans the tab order starting after (or before, if
// direction is -1) startID and returns the first focusable component ID.
// Returns "" if no focusable component exists in the order.
// Caller must hold the write lock.
func (f *FocusManager) nextFocusableFrom(startID string, direction int) string {
	if len(f.order) == 0 {
		return ""
	}

	idx := f.indexOf(startID)
	if idx < 0 {
		idx = 0
	}

	for i := 1; i <= len(f.order); i++ {
		candidate := (idx + direction*i) % len(f.order)
		if candidate < 0 {
			candidate += len(f.order)
		}
		id := f.order[candidate]
		if f.isFocusable(id) {
			return id
		}
	}

	return ""
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

// Next moves focus to the next focusable component in the tab order.
//
// Behavior:
//   - Skips any component that is Disabled or Hidden
//   - If current is at the end, wraps to the start (circular)
//   - Does nothing if order is empty, or if every component is unfocusable
//   - Does nothing if capture is active (dropdown/modal has focus lock)
//   - Typically called when user presses Tab
//
// Returns: The ID of the new focus, or the current (unchanged) focus if
// navigation could not proceed.
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
		// Current not in order; jump to first focusable
		if next := f.nextFocusableFrom(f.order[len(f.order)-1], +1); next != "" {
			f.current = next
		}
		return f.current
	}

	next := f.nextFocusableFrom(f.current, +1)
	if next != "" {
		f.current = next
	}
	return f.current
}

// Prev moves focus to the previous focusable component in the tab order.
//
// Behavior:
//   - Skips any component that is Disabled or Hidden
//   - If current is at the start, wraps to the end (circular)
//   - Does nothing if order is empty, or if every component is unfocusable
//   - Does nothing if capture is active
//   - Typically called when user presses Shift+Tab
//
// Returns: The ID of the new focus, or the current (unchanged) focus if
// navigation could not proceed.
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
		// Current not in order; jump to last focusable
		if prev := f.nextFocusableFrom(f.order[0], -1); prev != "" {
			f.current = prev
		}
		return f.current
	}

	prev := f.nextFocusableFrom(f.current, -1)
	if prev != "" {
		f.current = prev
	}
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
// Note: PushFocus does not check Disabled/Hidden — modals are assumed to be
// deliberately opened by the caller. If you need that guard, check
// IsDisabled/IsHidden before calling.
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
// If the restored focus is now disabled/hidden, advances to the next
// focusable component in tab order instead.
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
	restored := f.stack[len(f.stack)-1]
	f.stack = f.stack[:len(f.stack)-1]
	f.current = restored

	if !f.isFocusable(restored) {
		if next := f.nextFocusableFrom(restored, +1); next != "" {
			f.current = next
		}
	}

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
// replaces the previous capture. Disabled/Hidden components cannot capture
// focus; the call is a no-op in that case.
func (f *FocusManager) CaptureFocus(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.disabled[id] || f.hidden[id] {
		return
	}

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
//     1. Component is not Disabled (safety net; disabled components should
//     never hold focus, but this guards against stale state)
//     2. Component has focus (current)
//     3. Key hasn't been consumed by another component (Key.consumed is false)
//   - Marks the key as consumed (Key.consumed = true) when claimed
//   - Prevents key duplication: each key is handled by exactly one component
//   - Returns empty key and false if component shouldn't handle input
//
// Note on ReadOnly: UseFocusedKey does NOT check IsReadOnly. A read-only
// component should still receive navigation/selection keys (arrows, copy,
// etc.); it's the component's responsibility to check
// GlobalFocus's IsReadOnly(componentID) before acting on mutating keys
// like typing or delete.
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
func UseFocusedKey(componentID string, isFocused bool) (key Key, isMine bool) {
	key = CurrentKey
	if key == (Key{}) {
		return key, false
	}

	// Safety net: a disabled component should never claim a key, even if it
	// somehow still holds focus/capture (e.g. stale state during a transition).
	if globalFocus.IsDisabled(componentID) {
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
// Returns false if the component is disabled or hidden.
func SetFocus(id string) bool {
	return globalFocus.SetFocus(id)
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

// FocusNext moves focus to the next focusable component in the tab order.
func FocusNext() {
	globalFocus.Next()
}

// FocusPrev moves focus to the previous focusable component in the tab order.
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

// SetDisabled marks a global component as disabled or enabled.
func SetDisabled(id string, disabled bool) {
	globalFocus.SetDisabled(id, disabled)
}

// IsDisabled checks if a global component is disabled.
func IsDisabled(id string) bool {
	return globalFocus.IsDisabled(id)
}

// SetHidden marks a global component as hidden or visible.
func SetHidden(id string, hidden bool) {
	globalFocus.SetHidden(id, hidden)
}

// IsHidden checks if a global component is hidden.
func IsHidden(id string) bool {
	return globalFocus.IsHidden(id)
}

// SetReadOnly marks a global component as read-only or editable.
func SetReadOnly(id string, readonly bool) {
	globalFocus.SetReadOnly(id, readonly)
}

// IsReadOnly checks if a global component is read-only.
func IsReadOnly(id string) bool {
	return globalFocus.IsReadOnly(id)
}
