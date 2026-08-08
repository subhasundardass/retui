package components

import (
	"sync"
	"time"

	"github.com/subhasundardass/retui/retui"
)

// =======================================================================
// Types
// =======================================================================

type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

type ToastPosition int

const (
	ToastBottomRight ToastPosition = iota
	ToastBottomLeft
	ToastTopRight
	ToastTopLeft
)

type Toast struct {
	ID       int64
	Message  string
	Type     ToastType
	Duration time.Duration // 0 = persists until Dismiss()
	Created  time.Time
	Position ToastPosition
}

func (t *Toast) expired(now time.Time) bool {
	return t.Duration > 0 && now.Sub(t.Created) >= t.Duration
}

// Title returns the label shown before the message, based on type.
func (t *Toast) Title() string {
	switch t.Type {
	case ToastSuccess:
		return "SUCCESS"
	case ToastWarning:
		return "WARNING"
	case ToastError:
		return "ERROR"
	default:
		return "INFO"
	}
}

// Style returns the color styling for this toast's type.
func (t *Toast) Style() retui.Style {
	style := retui.NewStyle().Foreground(retui.White).Bold(true)

	switch t.Type {
	case ToastSuccess:
		style = style.Background(retui.Green)
	case ToastWarning:
		style = style.Background(retui.Yellow).Foreground(retui.Black)
	case ToastError:
		style = style.Background(retui.Red)
	default:
		style = style.Background(retui.Blue)
	}
	return style
}

// Width returns the rendered width: title + message + horizontal padding.
func (t *Toast) Width() int {
	label := t.Title() + ": " + t.Message
	w := 0
	for _, r := range label {
		w += retui.RuneWidth(r)
	}
	return w + 2
}

// Height is fixed at 1 row (flat bar, no border).
func (t *Toast) Height() int {
	return 1
}

// =======================================================================
// Options — per-toast (used with Show / ShowSuccess / ShowError / etc.)
// =======================================================================

type ToastOption func(*Toast)

func WithType(t ToastType) ToastOption         { return func(to *Toast) { to.Type = t } }
func WithDuration(d time.Duration) ToastOption { return func(to *Toast) { to.Duration = d } }
func WithPosition(p ToastPosition) ToastOption { return func(to *Toast) { to.Position = p } }

// =======================================================================
// Options — manager-level defaults (used once at app startup)
// =======================================================================

type ManagerOption func(*ToastManager)

// WithDefaultPosition sets the corner new toasts appear in by default.
func WithDefaultPosition(p ToastPosition) ManagerOption {
	return func(m *ToastManager) { m.defaults.Position = p }
}

// WithDefaultDuration sets how long new toasts stay visible by default.
// 0 means toasts persist until Dismiss() is called explicitly.
func WithDefaultDuration(d time.Duration) ManagerOption {
	return func(m *ToastManager) { m.defaults.Duration = d }
}

// =======================================================================
// Manager
// =======================================================================

type ToastManager struct {
	mu       sync.Mutex
	toasts   []*Toast
	nextID   int64
	onChange func()
	defaults Toast
}

func NewToastManager() *ToastManager {
	return &ToastManager{
		defaults: Toast{
			Type:     ToastInfo,
			Duration: 3 * time.Second,
			Position: ToastBottomLeft,
		},
	}
}

// Configure sets this manager's defaults (position, duration) for all
// future Show()/ShowX() calls that don't explicitly override them via
// a ToastOption. Typically called once at app startup.
func (m *ToastManager) Configure(opts ...ManagerOption) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, opt := range opts {
		opt(m)
	}
}

// OnChange registers a callback fired whenever the toast set changes
// (added, dismissed, or expired-on-sweep). Wire this to your app's
// redraw/tick mechanism so toasts appear and disappear promptly instead
// of waiting for an unrelated input event to trigger the next render.
func (m *ToastManager) OnChange(fn func()) {
	m.mu.Lock()
	m.onChange = fn
	m.mu.Unlock()
}

func (m *ToastManager) Show(msg string, opts ...ToastOption) int64 {
	m.mu.Lock()
	t := m.defaults
	t.Message = msg
	t.Created = time.Now()
	for _, opt := range opts {
		opt(&t)
	}
	m.nextID++
	t.ID = m.nextID

	toast := t
	m.toasts = append(m.toasts, &toast)
	m.mu.Unlock()

	m.notify()

	if t.Duration > 0 {
		time.AfterFunc(t.Duration, m.sweepAndNotify)
	}
	return toast.ID
}

func (m *ToastManager) Dismiss(id int64) {
	m.mu.Lock()
	out := m.toasts[:0]
	for _, t := range m.toasts {
		if t.ID != id {
			out = append(out, t)
		}
	}
	m.toasts = out
	m.mu.Unlock()
	m.notify()
}

func (m *ToastManager) Clear() {
	m.mu.Lock()
	m.toasts = nil
	m.mu.Unlock()
	m.notify()
}

func (m *ToastManager) Toasts() []*Toast {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sweepLocked()
	out := make([]*Toast, len(m.toasts))
	copy(out, m.toasts)
	return out
}

func (m *ToastManager) HasToast() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sweepLocked()
	return len(m.toasts) > 0
}

func (m *ToastManager) sweepAndNotify() {
	m.mu.Lock()
	changed := m.sweepLocked()
	m.mu.Unlock()
	if changed {
		m.notify()
	}
}

// sweepLocked removes expired toasts. Caller must hold m.mu.
func (m *ToastManager) sweepLocked() bool {
	now := time.Now()
	out := m.toasts[:0]
	changed := false
	for _, t := range m.toasts {
		if t.expired(now) {
			changed = true
			continue
		}
		out = append(out, t)
	}
	m.toasts = out
	return changed
}

func (m *ToastManager) notify() {
	m.mu.Lock()
	cb := m.onChange
	m.mu.Unlock()
	if cb != nil {
		cb()
	}
}

// =======================================================================
// Package-level default manager — global ergonomic API
// =======================================================================

var defaultToasts = NewToastManager()

// ConfigureToasts sets app-wide defaults for the global toast manager.
// Call once at startup, before any ShowToast/ShowSuccess/etc. calls:
//
//	components.ConfigureToasts(
//	    components.WithDefaultPosition(components.ToastTopRight),
//	    components.WithDefaultDuration(4*time.Second),
//	)
//
// Show error toast — auto-dismisses after default duration
// components.ShowError(err.Error())

func ConfigureToasts(opts ...ManagerOption) {
	defaultToasts.Configure(opts...)
}

func ShowToast(msg string, opts ...ToastOption) int64 {
	return defaultToasts.Show(msg, opts...)
}
func ShowSuccess(msg string, opts ...ToastOption) int64 {
	return ShowToast(msg, append(opts, WithType(ToastSuccess))...)
}
func ShowWarning(msg string, opts ...ToastOption) int64 {
	return ShowToast(msg, append(opts, WithType(ToastWarning))...)
}
func ShowError(msg string, opts ...ToastOption) int64 {
	return ShowToast(msg, append(opts, WithType(ToastError))...)
}
func DismissToast(id int64) { defaultToasts.Dismiss(id) }
func ClearToasts()          { defaultToasts.Clear() }

// =======================================================================
// Layer — builds Overlay elements from the current toast queue
// =======================================================================

// ToastLayer returns one Overlay element per visible toast, positioned
// using the current terminal dimensions tracked by the window package.
// Append its result to the children of your root layout element:
//
//	children := append([]retui.Element{final}, components.ToastLayer(nil)...)
func ToastLayer(mgr *ToastManager) []retui.Element {
	if mgr == nil {
		mgr = defaultToasts
	}

	toasts := mgr.Toasts()
	if len(toasts) == 0 {
		return nil
	}

	screenW := retui.CurrentScreenWidth
	screenH := retui.CurrentScreenHeight

	byPos := map[ToastPosition][]*Toast{}
	for _, t := range toasts {
		byPos[t.Position] = append(byPos[t.Position], t)
	}

	var overlays []retui.Element
	for pos, list := range byPos {
		offset := 0
		for _, t := range list {
			x, y := anchor(screenW, screenH, pos, t.Width(), t.Height(), offset)
			overlays = append(overlays, retui.Overlay(x, y, toastElement(t)))
			offset += t.Height() + 1
		}
	}
	return overlays
}

func anchor(screenW, screenH int, pos ToastPosition, w, h, offset int) (x, y int) {
	switch pos {
	case ToastBottomRight:
		return screenW - w - 1, screenH - h - 1 - offset
	case ToastBottomLeft:
		return 1, screenH - h - 1 - offset
	case ToastTopRight:
		return screenW - w - 1, 1 + offset
	default: // ToastTopLeft
		return 1, 1 + offset
	}
}

// toastElement builds the visual Box for a single toast: a flat colored
// bar with no border, sized to Toast.Width()/Height().
func toastElement(t *Toast) retui.Element {
	return retui.Box(
		retui.Props{
			Direction: retui.Column,
			Width:     retui.Fixed(t.Width()),
			Height:    retui.Fixed(t.Height()),
			Justify:   retui.JustifyStart,
		},
		t.Style(),
		retui.Text(" "+t.Title()+": "+t.Message+" ", t.Style()),
	)
}
