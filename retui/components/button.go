package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

// ─── Button Configuration ────────────────────────────────────────────────

type ButtonConfig struct {
	ID          string
	Label       string
	Width       int
	Style       retui.Style
	HoverStyle  retui.Style
	ActiveStyle retui.Style
	Prefix      string
	Suffix      string
	OnClick     func(id string)
	OnKeyPress  func(id string, key retui.Key) bool
	OnFocus     func(id string)
	OnBlur      func(id string)
}

type ButtonField struct {
	config  ButtonConfig
	focused bool
	active  bool
}

// ─── Builder Methods ──────────────────────────────────────────────────────

func Button() *ButtonField {
	return &ButtonField{
		config: ButtonConfig{
			ID:          "",
			Label:       "",
			Width:       0,
			Style:       retui.NewStyle(),
			HoverStyle:  retui.NewStyle().Foreground(retui.Cyan).Bold(true),
			ActiveStyle: retui.NewStyle().Foreground(retui.White).Background(retui.Blue).Bold(true),
			Prefix:      "",
			Suffix:      "",
			OnClick:     nil,
			OnKeyPress:  nil,
			OnFocus:     nil,
			OnBlur:      nil,
		},
		focused: false,
		active:  false,
	}
}

func (b *ButtonField) ID(v string) *ButtonField {
	b.config.ID = v
	return b
}

func (b *ButtonField) Label(v string) *ButtonField {
	b.config.Label = v
	return b
}

func (b *ButtonField) Width(w int) *ButtonField {
	b.config.Width = w
	return b
}

func (b *ButtonField) Prefix(v string) *ButtonField {
	b.config.Prefix = v
	return b
}

func (b *ButtonField) Suffix(v string) *ButtonField {
	b.config.Suffix = v
	return b
}

func (b *ButtonField) Style(s retui.Style) *ButtonField {
	b.config.Style = s
	return b
}

func (b *ButtonField) HoverStyle(s retui.Style) *ButtonField {
	b.config.HoverStyle = s
	return b
}

func (b *ButtonField) ActiveStyle(s retui.Style) *ButtonField {
	b.config.ActiveStyle = s
	return b
}

func (b *ButtonField) Focused(v bool) *ButtonField {
	b.focused = v
	return b
}

func (b *ButtonField) Active(v bool) *ButtonField {
	b.active = v
	return b
}

func (b *ButtonField) OnClick(fn func(string)) *ButtonField {
	b.config.OnClick = fn
	return b
}

func (b *ButtonField) OnKeyPress(fn func(string, retui.Key) bool) *ButtonField {
	b.config.OnKeyPress = fn
	return b
}

func (b *ButtonField) OnFocus(fn func(string)) *ButtonField {
	b.config.OnFocus = fn
	return b
}

func (b *ButtonField) OnBlur(fn func(string)) *ButtonField {
	b.config.OnBlur = fn
	return b
}

// ─── Render Method ──────────────────────────────────────────────────────

func (b *ButtonField) Render() retui.Element {
	return renderButton(b.focused, b.active, &b.config)
}

// ─── Core Rendering Function ────────────────────────────────────────────

func renderButton(focused bool, active bool, config *ButtonConfig) retui.Element {
	// Track active state
	isActive, setActive := retui.UseState(active)

	// Sync with external config changes
	if isActive != active {
		setActive(active)
	}

	// Trigger focus/blur events
	if focused && config.OnFocus != nil && config.ID != "" {
		config.OnFocus(config.ID)
	}

	if !focused && config.OnBlur != nil && config.ID != "" {
		config.OnBlur(config.ID)
	}

	// Handle keyboard input when focused
	if focused {
		key := retui.CurrentKey

		if config.OnKeyPress != nil && config.ID != "" {
			if config.OnKeyPress(config.ID, key) {
				goto render
			}
		}

		switch key.Code {
		case retui.KeyEnter, retui.KeySpace:
			setActive(true)
			if config.OnClick != nil && config.ID != "" {
				config.OnClick(config.ID)
			}
			// Reset active state after click
			go func() {
				setActive(false)
			}()
		}
	}

render:
	// Determine button style based on state
	var buttonStyle retui.Style
	if isActive {
		buttonStyle = config.ActiveStyle
	} else if focused {
		buttonStyle = config.HoverStyle
	} else {
		buttonStyle = config.Style
	}

	// Build the button display
	display := config.Label
	if display == "" {
		display = "Button"
	}

	// Add decorative brackets for button appearance
	display = "[ " + display + " ]"

	// Pad to width if specified
	if config.Width > 0 {
		displayLen := len([]rune(display))
		if displayLen < config.Width {
			padding := strings.Repeat(" ", config.Width-displayLen)
			display = display + padding
		}
	}

	// Add prefix/suffix
	prefixStyle := retui.NewStyle()
	if focused || isActive {
		prefixStyle = prefixStyle.Foreground(retui.Cyan).Bold(true)
	} else {
		prefixStyle = prefixStyle.Foreground(retui.BrightBlack)
	}

	suffixStyle := retui.NewStyle()
	if focused || isActive {
		suffixStyle = suffixStyle.Foreground(retui.Cyan).Bold(true)
	} else {
		suffixStyle = suffixStyle.Foreground(retui.BrightBlack)
	}

	// Build elements
	elements := []retui.Element{}

	if config.Prefix != "" {
		elements = append(elements, retui.Text(config.Prefix, prefixStyle))
	}

	elements = append(elements, retui.Text(display, buttonStyle))

	if config.Suffix != "" {
		elements = append(elements, retui.Text(config.Suffix, suffixStyle))
	}

	// Return as a row/box
	return retui.Box(
		retui.Props{
			Direction: retui.Row,
		},
		retui.NewStyle(),
		elements...,
	)
}

// ─── Example Usage ──────────────────────────────────────────────────────

func ExampleButtonUsage() retui.Element {
	// Simple button
	submitBtn := Button().
		ID("submit").
		Label("Submit").
		Width(20).
		Prefix("▶ ").
		OnClick(func(id string) {
			println("Button", id, "clicked!")
		})

	// Button with custom styles
	cancelBtn := Button().
		ID("cancel").
		Label("Cancel").
		Style(retui.NewStyle().Foreground(retui.Red)).
		HoverStyle(retui.NewStyle().Foreground(retui.Red).Bold(true)).
		ActiveStyle(retui.NewStyle().Foreground(retui.White).Background(retui.Red).Bold(true)).
		OnClick(func(id string) {
			println("Button", id, "clicked!")
		})

	// Button with prefix/suffix
	deleteBtn := Button().
		ID("delete").
		Label("Delete").
		Prefix("🗑 ").
		Suffix(" ✓").
		OnClick(func(id string) {
			println("Button", id, "clicked!")
		})

	return retui.Box(
		retui.Props{
			Direction: retui.Column,
		},
		retui.NewStyle(),
		submitBtn.Render(),
		retui.Text(" ", retui.NewStyle()),
		cancelBtn.Render(),
		retui.Text(" ", retui.NewStyle()),
		deleteBtn.Render(),
	)
}
