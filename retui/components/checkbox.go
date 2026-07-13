package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

// ─── Checkbox Configuration ──────────────────────────────────────────────

type CheckboxConfig struct {
	ID             string
	Checked        bool
	Label          string
	Width          int
	Style          retui.Style
	CheckedStyle   retui.Style
	UncheckedStyle retui.Style
	OnChange       func(id string, checked bool)
	OnKeyPress     func(id string, key retui.Key) bool
	OnFocus        func(id string)
	OnBlur         func(id string)
}

type CheckboxField struct {
	config  CheckboxConfig
	focused bool
}

// ─── Builder Methods ──────────────────────────────────────────────────────

func Checkbox() *CheckboxField {
	return &CheckboxField{
		config: CheckboxConfig{
			ID:             "",
			Checked:        false,
			Label:          "",
			Width:          0,
			Style:          retui.NewStyle(),
			CheckedStyle:   retui.NewStyle().Foreground(retui.Green).Bold(true),
			UncheckedStyle: retui.NewStyle().Foreground(retui.BrightBlack),
			OnChange:       nil,
			OnKeyPress:     nil,
			OnFocus:        nil,
			OnBlur:         nil,
		},
		focused: false,
	}
}

func (c *CheckboxField) ID(v string) *CheckboxField {
	c.config.ID = v
	return c
}

func (c *CheckboxField) Checked(v bool) *CheckboxField {
	c.config.Checked = v
	return c
}

func (c *CheckboxField) Label(v string) *CheckboxField {
	c.config.Label = v
	return c
}

func (c *CheckboxField) Width(w int) *CheckboxField {
	c.config.Width = w
	return c
}

func (c *CheckboxField) Style(s retui.Style) *CheckboxField {
	c.config.Style = s
	return c
}

func (c *CheckboxField) CheckedStyle(s retui.Style) *CheckboxField {
	c.config.CheckedStyle = s
	return c
}

func (c *CheckboxField) UncheckedStyle(s retui.Style) *CheckboxField {
	c.config.UncheckedStyle = s
	return c
}

func (c *CheckboxField) Focused(v bool) *CheckboxField {
	c.focused = v
	return c
}

func (c *CheckboxField) OnChange(fn func(string, bool)) *CheckboxField {
	c.config.OnChange = fn
	return c
}

func (c *CheckboxField) OnKeyPress(fn func(string, retui.Key) bool) *CheckboxField {
	c.config.OnKeyPress = fn
	return c
}

func (c *CheckboxField) OnFocus(fn func(string)) *CheckboxField {
	c.config.OnFocus = fn
	return c
}

func (c *CheckboxField) OnBlur(fn func(string)) *CheckboxField {
	c.config.OnBlur = fn
	return c
}

// ─── Render Method ──────────────────────────────────────────────────────

func (c *CheckboxField) Render() retui.Element {
	return renderCheckbox(c.focused, &c.config)
}

// ─── Core Rendering Function ────────────────────────────────────────────

func renderCheckbox(focused bool, config *CheckboxConfig) retui.Element {
	// Track checked state
	checked, setChecked := retui.UseState(config.Checked)

	// Sync with external config changes
	if checked != config.Checked {
		setChecked(config.Checked)
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
		case retui.KeySpace:
			newChecked := !checked
			setChecked(newChecked)
			config.Checked = newChecked
			if config.OnChange != nil && config.ID != "" {
				config.OnChange(config.ID, newChecked)
			}

		case retui.KeyEnter:
			// Toggle on enter as well for accessibility
			newChecked := !checked
			setChecked(newChecked)
			config.Checked = newChecked
			if config.OnChange != nil && config.ID != "" {
				config.OnChange(config.ID, newChecked)
			}
		}
	}

render:
	// Build the checkbox display
	box := "[ ]"
	if checked {
		box = "[✓]"
	}

	// Apply styles
	var boxStyle retui.Style
	if checked {
		boxStyle = boxStyle.
			Foreground(retui.Green).
			Bold(true)
	} else {
		boxStyle = boxStyle.
			Foreground(retui.BrightBlack)
	}

	// Override with focus style
	if focused {
		boxStyle = boxStyle.
			Foreground(retui.Cyan).
			Bold(true)
	}

	// Build the display text
	display := box
	if config.Label != "" {
		display = box + " " + config.Label
	}

	// Pad to width if specified
	if config.Width > 0 {
		displayLen := len([]rune(display))
		if displayLen < config.Width {
			padding := strings.Repeat(" ", config.Width-displayLen)
			display = display + padding
		}
	}

	// Return as a text element
	return retui.Text(display, boxStyle)
}

// ─── Example Usage ──────────────────────────────────────────────────────

// func ExampleCheckboxUsage() retui.Element {
// 	// Single checkbox
// 	agreeCheckbox := Checkbox().
// 		ID("agree").
// 		Label("I agree to the terms").
// 		OnChange(func(id string, checked bool) {
// 			println("Checkbox", id, "changed to:", checked)
// 		})
