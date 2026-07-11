package retui

import (
	"sync"
)

// ScreenStack manages the navigation state for a TUI application.
// It provides thread-safe operations for pushing, popping, and
// managing screens.
//
// ScreenStack is safe for concurrent use.
type ScreenStack struct {
	mu          sync.RWMutex
	stack       []string
	currentPage string
}

// NewScreenStack creates a new screen stack with the given root screen.
//
// Example:
//
//	nav := retui.NewScreenStack("home")
//	nav.PushScreen("settings")
//	current := nav.Current() // "settings"
func NewScreenStack(rootScreen string) *ScreenStack {
	return &ScreenStack{
		stack:       []string{rootScreen},
		currentPage: rootScreen,
	}
}

// PushScreen adds a screen to the top of the stack and makes it current.
//
// Example:
//
//	nav.PushScreen("profile")
func (n *ScreenStack) PushScreen(screenID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.stack = append(n.stack, screenID)
	n.currentPage = screenID
}

// PopScreen removes the top screen and returns the new current screen ID.
// If the stack has only one entry it stays — you can never pop the root.
//
// Example:
//
//	if previous := nav.PopScreen(); previous != "home" {
//	    // Handle return
//	}
func (n *ScreenStack) PopScreen() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.stack) <= 1 {
		return n.stack[0]
	}
	n.stack = n.stack[:len(n.stack)-1]
	top := n.stack[len(n.stack)-1]
	n.currentPage = top
	return top
}

// ReplaceScreen replaces the top of the stack with a new screen.
// Use this for redirects where you do not want the user to go back.
//
// Example:
//
//	nav.ReplaceScreen("dashboard") // Replaces current with dashboard
func (n *ScreenStack) ReplaceScreen(screenID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.stack) == 0 {
		n.stack = []string{screenID}
	} else {
		n.stack[len(n.stack)-1] = screenID
	}
	n.currentPage = screenID
}

// Current returns the ID of the currently active screen.
//
// Example:
//
//	current := nav.Current()
//	if current == "settings" {
//	    // Render settings
//	}
func (n *ScreenStack) Current() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if len(n.stack) == 0 {
		return ""
	}
	return n.stack[len(n.stack)-1]
}

// Stack returns a copy of the current stack.
// A copy is returned so callers cannot mutate internal state.
//
// Example:
//
//	history := nav.Stack()
//	for i, screen := range history {
//	    fmt.Printf("%d: %s\n", i, screen)
//	}
func (n *ScreenStack) Stack() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	cp := make([]string, len(n.stack))
	copy(cp, n.stack)
	return cp
}

// Size returns how many screens are on the stack.
//
// Example:
//
//	if nav.Size() > 1 {
//	    // Show back button
//	}
func (n *ScreenStack) Size() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.stack)
}

// CanPop returns true when there is more than one screen on the stack.
// Use this to decide whether to show a back button.
//
// Example:
//
//	if nav.CanPop() {
//	    renderBackButton()
//	}
func (n *ScreenStack) CanPop() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.stack) > 1
}

// Reset clears the stack and sets a single root screen.
// Use this on logout or when navigating to a completely new flow.
//
// Example:
//
//	nav.Reset("login") // Clears everything and goes to login
func (n *ScreenStack) Reset(rootScreenID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.stack = []string{rootScreenID}
	n.currentPage = rootScreenID
}

// IsEmpty returns true if the stack is empty.
func (n *ScreenStack) IsEmpty() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.stack) == 0
}

// Clear empties the stack completely.
// Use with caution - this leaves the stack in an invalid state
// until a new screen is pushed or reset.
func (n *ScreenStack) Clear() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.stack = []string{}
	n.currentPage = ""
}

// ─── Global default instance ───────────────────────────────────────────
//
// Mirrors the FocusManager pattern already used in this package: the type
// itself stays instantiable/testable (see NewScreenStack above), while a
// default global instance backs the exported free functions below, so
// Root() and the rest of the app can call retui.PushScreen(...) directly
// without threading a *ScreenStack through every call site by hand.

var globalScreens = NewScreenStack("home")

// SetInitialScreen sets the root screen on the global stack. Call this
// once during startup, before the first Render — e.g.
//
//	retui.SetInitialScreen(cfg.DefaultPage)
//
// Skip this if "home" is already your intended root screen ID.
func SetInitialScreen(id string) {
	globalScreens.Reset(id)
}

// PushScreen adds a screen to the top of the global stack and makes it
// current. Resets component hook state and schedules a re-render so the
// new screen appears immediately.
func PushScreen(id string) {
	globalScreens.PushScreen(id)
	onScreenChanged()
}

// PopScreen removes the top screen from the global stack and returns the
// new current screen ID. No-op (returns the unchanged root) if only one
// screen remains — in that case no reset/re-render is triggered either,
// since nothing actually changed.
func PopScreen() string {
	before := globalScreens.Current()
	top := globalScreens.PopScreen()
	if top != before {
		onScreenChanged()
	}
	return top
}

// ReplaceScreen swaps the top of the global stack for a new screen. Use
// this for redirects where the user shouldn't be able to navigate Back
// to whatever screen is being replaced.
func ReplaceScreen(id string) {
	globalScreens.ReplaceScreen(id)
	onScreenChanged()
}

// ResetScreenStack clears all history on the global stack and sets a
// single root screen. Use this on logout, or when starting an entirely
// new navigation flow.
func ResetScreenStack(rootID string) {
	globalScreens.Reset(rootID)
	onScreenChanged()
}

// CurrentScreen returns the ID of the currently active screen on the
// global stack.
func CurrentScreen() string {
	return globalScreens.Current()
}

// ScreenStackSnapshot returns a copy of the global navigation stack,
// root first.
func ScreenStackSnapshot() []string {
	return globalScreens.Stack()
}

// ScreenStackSize returns how many screens are on the global stack.
func ScreenStackSize() int {
	return globalScreens.Size()
}

// CanPopScreen reports whether there is a screen to go Back to on the
// global stack. Use this to decide whether to show a back button.
func CanPopScreen() bool {
	return globalScreens.CanPop()
}

// onScreenChanged resets per-component hook state — so the incoming
// screen never inherits stale UseState/UseEffect slots from whatever was
// showing before — and schedules an extra render pass so the new screen
// appears immediately, rather than waiting for the next key/tick/resize
// event. Every mutating free function above calls this, so a manual
// UseScreenReset(currentID) call in Root() is no longer necessary as long
// as all screen transitions go through this file.
func onScreenChanged() {
	ResetComponentState()
	pendingRender = true
}
