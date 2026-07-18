package retui

import "sync"

// ScreenParams contains navigation parameters passed to a screen.
//
// Unlike component Props, ScreenParams belong to the navigation system
// and are associated with a screen while it is active.
//
// Example:
//
//	retui.PushScreen("ledger_list", retui.ScreenParams{
//	    "groupId": 10,
//	})
type ScreenParams map[string]any

// ScreenRoute represents one entry in the navigation stack.
//
// Each entry stores both the screen identifier and any navigation
// parameters supplied when the screen was pushed.
type ScreenRoute struct {
	ID     string
	Params ScreenParams
}

// ScreenStack manages the navigation history for a RetUI application.
//
// ScreenStack is safe for concurrent use.
type ScreenStack struct {
	mu          sync.RWMutex
	stack       []ScreenRoute
	currentPage ScreenRoute
}

// NewScreenStack creates a new navigation stack.
//
// The supplied screen becomes the root screen.
//
// Example:
//
//	nav := retui.NewScreenStack("home")
func NewScreenStack(rootScreen string) *ScreenStack {
	root := ScreenRoute{
		ID:     rootScreen,
		Params: nil,
	}

	return &ScreenStack{
		stack:       []ScreenRoute{root},
		currentPage: root,
	}
}

// PushScreen pushes a new screen onto the navigation stack.
//
// Params may be nil.
//
// Example:
//
//	nav.PushScreen("ledger_list", retui.ScreenParams{
//	    "groupId": 10,
//	})
func (n *ScreenStack) PushScreen(screenID string, params ScreenParams) {
	n.mu.Lock()
	defer n.mu.Unlock()

	route := ScreenRoute{
		ID:     screenID,
		Params: params,
	}

	n.stack = append(n.stack, route)
	n.currentPage = route
}

// PopScreen removes the top screen.
//
// The root screen cannot be removed.
func (n *ScreenStack) PopScreen() string {
	n.mu.Lock()
	defer n.mu.Unlock()

	if len(n.stack) <= 1 {
		return n.stack[0].ID
	}

	n.stack = n.stack[:len(n.stack)-1]
	n.currentPage = n.stack[len(n.stack)-1]

	return n.currentPage.ID
}

// ReplaceScreen replaces the current screen.
//
// Navigation history is preserved except for the top entry.
func (n *ScreenStack) ReplaceScreen(screenID string, params ScreenParams) {
	n.mu.Lock()
	defer n.mu.Unlock()

	route := ScreenRoute{
		ID:     screenID,
		Params: params,
	}

	if len(n.stack) == 0 {
		n.stack = []ScreenRoute{route}
	} else {
		n.stack[len(n.stack)-1] = route
	}

	n.currentPage = route
}

// Current returns the current screen ID.
func (n *ScreenStack) Current() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.stack) == 0 {
		return ""
	}

	return n.currentPage.ID
}

// CurrentParams returns the navigation parameters for the current screen.
//
// Returns nil if no parameters were supplied.
func (n *ScreenStack) CurrentParams() ScreenParams {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.stack) == 0 {
		return nil
	}

	return n.currentPage.Params
}

// Stack returns the navigation history.
//
// A copy is returned so callers cannot modify the internal stack.
func (n *ScreenStack) Stack() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	cp := make([]string, len(n.stack))

	for i, route := range n.stack {
		cp[i] = route.ID
	}

	return cp
}

// Size returns the number of screens currently on the stack.
func (n *ScreenStack) Size() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return len(n.stack)
}

// CanPop reports whether the current screen can navigate back.
func (n *ScreenStack) CanPop() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return len(n.stack) > 1
}

// Reset clears the stack and creates a new root screen.
func (n *ScreenStack) Reset(rootScreenID string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	root := ScreenRoute{
		ID: rootScreenID,
	}

	n.stack = []ScreenRoute{root}
	n.currentPage = root
}

// IsEmpty reports whether the navigation stack is empty.
func (n *ScreenStack) IsEmpty() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return len(n.stack) == 0
}

// Clear removes every screen from the navigation stack.
//
// Normally Reset should be preferred.
func (n *ScreenStack) Clear() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.stack = nil
	n.currentPage = ScreenRoute{}
}

var globalScreens = NewScreenStack("home")

// SetInitialScreen sets the root screen.
func SetInitialScreen(id string) {
	globalScreens.Reset(id)
}

// PushScreen pushes a new screen onto the global navigation stack.
//
// The params argument is optional.
//
// Example:
//
//	retui.PushScreen("ledger_list")
//
//	retui.PushScreen("ledger_list", retui.ScreenParams{
//	    "groupId": 5,
//	})
func PushScreen(id string, params ...ScreenParams) {
	var p ScreenParams

	if len(params) > 0 {
		p = params[0]
	}

	globalScreens.PushScreen(id, p)
	onScreenChanged()
}

// PopScreen navigates back.
func PopScreen() string {
	before := globalScreens.Current()
	top := globalScreens.PopScreen()

	if before != top {
		onScreenChanged()
	}

	return top
}

// ReplaceScreen replaces the current screen.
func ReplaceScreen(id string, params ...ScreenParams) {
	var p ScreenParams

	if len(params) > 0 {
		p = params[0]
	}

	globalScreens.ReplaceScreen(id, p)
	onScreenChanged()
}

// ResetScreenStack clears navigation history.
func ResetScreenStack(rootID string) {
	globalScreens.Reset(rootID)
	onScreenChanged()
}

// CurrentScreen returns the current screen ID.
func CurrentScreen() string {
	return globalScreens.Current()
}

// CurrentScreenParams returns the parameters associated with the
// currently active screen.
func CurrentScreenParams() ScreenParams {
	return globalScreens.CurrentParams()
}

// ScreenStackSnapshot returns a copy of the navigation history.
func ScreenStackSnapshot() []string {
	return globalScreens.Stack()
}

// ScreenStackSize returns the number of screens on the stack.
func ScreenStackSize() int {
	return globalScreens.Size()
}

// CanPopScreen reports whether Back navigation is possible.
func CanPopScreen() bool {
	return globalScreens.CanPop()
}

func onScreenChanged() {
	ResetComponentState()
	pendingRender = true
}
