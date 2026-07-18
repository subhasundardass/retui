package window

import (
	"sync"

	"github.com/subhasundardass/retui/retui"
)

// WindowManager manages all active windows in the application.
// It maintains the registry, Z-order (stacking), and focus state.
// All operations are thread-safe using a mutex.
type WindowManager struct {
	mu            sync.RWMutex
	windows       map[string]*Window
	stack         []string // Z-order: stack[0]=bottom, last=top
	modalStack    []string // modal IDs in order (bottom → top)
	focused       string   // Currently focused window ID
	renderTrigger func()   // Callback to trigger re-renders in retui
	screenWidth   int
	screenHeight  int
}

// NewWindowManager creates a new window manager instance.
func NewWindowManager() *WindowManager {
	return &WindowManager{
		windows:       make(map[string]*Window),
		stack:         make([]string, 0),
		modalStack:    make([]string, 0),
		focused:       "",
		renderTrigger: nil,
		screenWidth:   DefaultScreenWidth,
		screenHeight:  DefaultScreenHeight,
	}
}

func init() {
	retui.IsAnyModalOpenFn = func() bool {
		return globalManager.IsAnyModalOpen()
	}
	retui.RootRenderWrap = func(root retui.Element) retui.Element {
		return RenderWindowsOverlay(root)
	}
	retui.WindowKeyDispatch = func(key retui.Key) bool {

		if key.Code == retui.KeyTab && globalManager.IsAnyModalOpen() {
			if globalManager.GetActiveModal() != "" && len(globalManager.modalStack) <= 1 {
				if win := globalManager.GetWindow(globalManager.GetActiveModal()); win != nil {
					return win.HandleKey(key)
				}
				return true
			}
			globalManager.FocusNext()
			return true
		}

		// Added: default Escape behavior for modals
		if key.Code == retui.KeyEscape && globalManager.IsAnyModalOpen() {
			activeID := globalManager.GetActiveModal()
			if activeID != "" {
				win := globalManager.GetWindow(activeID)
				if win != nil {
					if win.HandleKey(key) {
						return true
					}
					win.Close()
				}
			}
			return true // ← always hit, regardless of what happened above
		}

		id := globalManager.GetFocused()
		if id == "" {
			return false
		}
		if win := globalManager.GetWindow(id); win != nil {
			return win.HandleKey(key)
		}

		return false
	}
}

// ========================================
// RENDER TRIGGER
// ========================================
// In WindowManager - SetScreenSize updates screen dimensions
func (wm *WindowManager) SetScreenSize(width, height int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.screenWidth = width
	wm.screenHeight = height
}

// GetScreenSize returns current screen dimensions
func (wm *WindowManager) GetScreenSize() (int, int) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.screenWidth, wm.screenHeight
}

// SetRenderTrigger sets the callback function that triggers re-renders.
// This should be called from your main App to connect the window system
// with the retui render loop.
func (wm *WindowManager) SetRenderTrigger(trigger func()) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.renderTrigger = trigger
}

// triggerRender calls the render trigger if it's set.
// This is called internally when window state changes.
func (wm *WindowManager) triggerRender() {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	if wm.renderTrigger != nil {
		wm.renderTrigger()
	}
}

// ========================================
// WINDOW LIFECYCLE
// ========================================

// AddWindow adds a window to the registry and Z-order stack.
// Called when window.Show() is invoked.
func (wm *WindowManager) AddWindow(w *Window) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, exists := wm.windows[w.ID]; exists {
		return
	}
	wm.windows[w.ID] = w
	wm.stack = append(wm.stack, w.ID)

	if w.IsModal() {
		wm.modalStack = append(wm.modalStack, w.ID)
		wm.bringToFrontLocked(w.ID)
		wm.setFocusedLocked(w.ID)
	} else {
		wm.reorderStackLocked()
		if len(wm.modalStack) == 0 {
			wm.setFocusedLocked(w.ID)
		}
	}
	go wm.triggerRender()
}

// RemoveWindow removes a window from the registry and Z-order.
// Called when window.Close() is invoked.
func (wm *WindowManager) RemoveWindow(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, exists := wm.windows[id]; !exists {
		return
	}
	delete(wm.windows, id)

	// Remove from stack
	for i, wid := range wm.stack {
		if wid == id {
			wm.stack = append(wm.stack[:i], wm.stack[i+1:]...)
			break
		}
	}

	// Remove from modalStack if present
	for i, wid := range wm.modalStack {
		if wid == id {
			wm.modalStack = append(wm.modalStack[:i], wm.modalStack[i+1:]...)
			break
		}
	}

	// Reassign focus if needed
	if wm.focused == id {
		if active := wm.getTopModalLocked(); active != "" {
			wm.setFocusedLocked(active)
		} else if len(wm.stack) > 0 {
			top := wm.stack[len(wm.stack)-1]
			wm.setFocusedLocked(top)
		} else {
			wm.setFocusedLocked("") // no window focused
		}
	}
	go wm.triggerRender()
}

// GetWindow retrieves a window by ID.
func (wm *WindowManager) GetWindow(id string) *Window {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.windows[id]
}

// ========================================
// Z-ORDER MANAGEMENT
// ========================================

// BringToFront moves a window to the top of the Z-order (front).
// Also updates focus to this window.
func (wm *WindowManager) BringToFront(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// If a modal is active, only that modal can be brought forward
	if active := wm.getTopModalLocked(); active != "" && active != id {
		return
	}

	// Find and move to end of stack
	idx := -1
	for i, wid := range wm.stack {
		if wid == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}
	wm.stack = append(wm.stack[:idx], wm.stack[idx+1:]...)
	wm.stack = append(wm.stack, id)

	// If it's a modal, push to top of modalStack
	if w, ok := wm.windows[id]; ok && w.IsModal() {
		wm.pushModalToTopLocked(id)
	}

	wm.setFocusedLocked(id)
	go wm.triggerRender()
}

// GetZOrder returns the current Z-order stack (bottom to top).
// Returns a copy to prevent external modification.
func (wm *WindowManager) GetZOrder() []string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	stack := make([]string, len(wm.stack))
	copy(stack, wm.stack)
	return stack
}

// ========================================
// FOCUS MANAGEMENT
// ========================================

// GetFocused returns the ID of the currently focused window.
func (wm *WindowManager) GetFocused() string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	if active := wm.getTopModalLocked(); active != "" {
		return active
	}
	return wm.focused
}

// SetFocus sets which window should receive keyboard input.
func (wm *WindowManager) SetFocus(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if active := wm.getTopModalLocked(); active != "" && active != id {
		return
	}
	if _, exists := wm.windows[id]; !exists {
		return
	}
	wm.setFocusedLocked(id)
	go wm.triggerRender()
}

// IsFocused checks if a window is currently focused.
func (wm *WindowManager) IsFocused(id string) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	if active := wm.getTopModalLocked(); active != "" {
		return active == id
	}
	return wm.focused == id
}

// FocusNext cycles focus to the next window in Z-order.
// Useful for keyboard navigation (e.g., Tab key).
func (wm *WindowManager) FocusNext() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if active := wm.getTopModalLocked(); active != "" {
		if len(wm.modalStack) > 1 {
			idx := -1
			for i, mID := range wm.modalStack {
				if mID == active {
					idx = i
					break
				}
			}
			if idx >= 0 {
				nextIdx := (idx + 1) % len(wm.modalStack)
				wm.setFocusedLocked(wm.modalStack[nextIdx])
				go wm.triggerRender()
			}
		}
		return
	}

	if len(wm.stack) == 0 {
		return
	}
	currentIdx := -1
	for i, id := range wm.stack {
		if id == wm.focused {
			currentIdx = i
			break
		}
	}
	nextIdx := currentIdx + 1
	if nextIdx >= len(wm.stack) {
		nextIdx = 0
	}
	wm.setFocusedLocked(wm.stack[nextIdx])
	go wm.triggerRender()
}

// ========================================
// QUERY METHODS
// ========================================
// SetScreenSize sets the screen dimensions globally
func SetScreenSize(width, height int) {
	globalManager.SetScreenSize(width, height)
	DefaultScreenWidth = width
	DefaultScreenHeight = height
}

// GetScreenSize returns the current global screen dimensions
func GetScreenSize() (int, int) {
	return globalManager.GetScreenSize()
}

// GetAll returns all active windows.
func (wm *WindowManager) GetAll() []*Window {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	windows := make([]*Window, 0, len(wm.windows))
	for _, w := range wm.windows {
		windows = append(windows, w)
	}
	return windows
}

// GetVisible returns all visible windows in Z-order (bottom to top).
func (wm *WindowManager) GetVisible() []*Window {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	windows := make([]*Window, 0)
	for _, id := range wm.stack {
		if w, ok := wm.windows[id]; ok && w.IsVisible() {
			windows = append(windows, w)
		}
	}
	return windows
}

// Count returns the number of open windows.
func (wm *WindowManager) Count() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.windows)
}

// CountVisible returns the number of visible windows.
func (wm *WindowManager) CountVisible() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	count := 0
	for _, w := range wm.windows {
		if w.visible {
			count++
		}
	}
	return count
}

// IsAnyModalOpen returns true if any modal window is currently visible.
func (wm *WindowManager) IsAnyModalOpen() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	for _, w := range wm.windows {
		if w.IsVisible() && w.IsModal() {
			return true
		}
	}
	return false
}

// GetTopVisibleModal returns the topmost visible modal window (if any).
func (wm *WindowManager) GetTopVisibleModal() *Window {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	for i := len(wm.stack) - 1; i >= 0; i-- {
		if w, ok := wm.windows[wm.stack[i]]; ok && w.IsVisible() && w.IsModal() {
			return w
		}
	}
	return nil
}

// HasWindow checks if a window with the given ID exists.
func (wm *WindowManager) HasWindow(id string) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	_, exists := wm.windows[id]
	return exists
}

// IsWindowBlocked returns true if the window is blocked by a modal.
func (wm *WindowManager) IsWindowBlocked(id string) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	active := wm.getTopModalLocked()
	if active == "" {
		return false
	}
	return active != id
}

// GetActiveModal returns the topmost modal ID (or empty).
func (wm *WindowManager) GetActiveModal() string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.getTopModalLocked()
}

// ========================================
// INTERNAL HELPERS (must be called with lock held)
// ========================================

// getTopModalLocked returns the topmost modal ID.
func (wm *WindowManager) getTopModalLocked() string {
	if len(wm.modalStack) == 0 {
		return ""
	}
	return wm.modalStack[len(wm.modalStack)-1]
}

// bringToFrontLocked moves a window to the top of the main stack.
func (wm *WindowManager) bringToFrontLocked(id string) {
	for i, wid := range wm.stack {
		if wid == id {
			wm.stack = append(wm.stack[:i], wm.stack[i+1:]...)
			break
		}
	}
	wm.stack = append(wm.stack, id)
}

// reorderStackLocked places all modals after all non‑modals.
func (wm *WindowManager) reorderStackLocked() {
	var nonModals, modals []string
	for _, id := range wm.stack {
		if w, ok := wm.windows[id]; ok && w.IsModal() {
			modals = append(modals, id)
		} else {
			nonModals = append(nonModals, id)
		}
	}
	wm.stack = append(nonModals, modals...)
}

// pushModalToTopLocked moves a modal to the top of modalStack and main stack.
func (wm *WindowManager) pushModalToTopLocked(id string) {
	// Remove from modalStack and re‑append
	for i, mID := range wm.modalStack {
		if mID == id {
			wm.modalStack = append(wm.modalStack[:i], wm.modalStack[i+1:]...)
			break
		}
	}
	wm.modalStack = append(wm.modalStack, id)
	// Also ensure it's on top of main stack
	wm.bringToFrontLocked(id)
}

// setFocusedLocked updates the focused ID and all windows' focused flags.
func (wm *WindowManager) setFocusedLocked(id string) {
	wm.focused = id
	for wid, w := range wm.windows {
		w.mu.Lock()
		w.focused = (wid == id)
		w.mu.Unlock()
	}
}

// ========================================
// GLOBAL INSTANCE
// ========================================

var globalManager = NewWindowManager()

// SetRenderTrigger sets the global render trigger function.
// This connects the window system to the retui render loop.
func SetRenderTrigger(trigger func()) {
	globalManager.SetRenderTrigger(trigger)
}

// GetManager returns the global window manager instance.
func GetManager() *WindowManager {
	return globalManager
}

// CloseAll closes all open windows.
func CloseAll() {
	globalManager.mu.Lock()
	for _, id := range globalManager.stack {
		if w, ok := globalManager.windows[id]; ok {
			w.mu.Lock()
			w.visible = false
			w.mu.Unlock()
		}
		delete(globalManager.windows, id)
	}
	globalManager.stack = make([]string, 0)
	globalManager.modalStack = make([]string, 0)
	globalManager.setFocusedLocked("")
	globalManager.mu.Unlock()
	go globalManager.triggerRender()
}

// ResetGlobalManager resets the global manager state (for testing).
func ResetGlobalManager() {
	globalManager = NewWindowManager()
}

// Count returns the number of open windows globally.
func Count() int {
	return globalManager.Count()
}

// CountVisible returns the number of visible windows globally.
func CountVisible() int {
	return globalManager.CountVisible()
}

// GetFocused returns the globally focused window ID.
func GetFocused() string {
	return globalManager.GetFocused()
}

// IsAnyModalOpen returns true if any modal is currently open.
func IsAnyModalOpen() bool {
	return globalManager.IsAnyModalOpen()
}

// GetVisible returns all visible windows globally.
func GetVisible() []*Window {
	return globalManager.GetVisible()
}

// IsWindowBlocked returns true if window is blocked by modal.
func IsWindowBlocked(id string) bool {
	return globalManager.IsWindowBlocked(id)
}

// GetActiveModal returns the topmost modal ID.
func GetActiveModal() string {
	return globalManager.GetActiveModal()
}
