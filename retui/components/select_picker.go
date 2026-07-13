package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

// ─── Select Picker Configuration ─────────────────────────────────────────

type SelectPickerConfig struct {
	ID         string
	Options    []string
	Selected   int // Index of selected option
	Width      int
	Style      retui.Style
	Prefix     string
	Suffix     string
	OnChange   func(id string, selected int, value string)
	OnKeyPress func(id string, key retui.Key) bool
	OnFocus    func(id string)
	OnBlur     func(id string)
	OnSubmit   func(id string, selected int, value string)
}

type SelectPickerField struct {
	config  SelectPickerConfig
	focused bool
}

// ─── Builder Methods ──────────────────────────────────────────────────────

func SelectPicker() *SelectPickerField {
	return &SelectPickerField{
		config: SelectPickerConfig{
			ID:         "",
			Options:    []string{},
			Selected:   0,
			Width:      0,
			Style:      retui.NewStyle(),
			Prefix:     "",
			Suffix:     "",
			OnChange:   nil,
			OnKeyPress: nil,
			OnFocus:    nil,
			OnBlur:     nil,
			OnSubmit:   nil,
		},
		focused: false,
	}
}

func (s *SelectPickerField) ID(v string) *SelectPickerField {
	s.config.ID = v
	return s
}

func (s *SelectPickerField) Options(v []string) *SelectPickerField {
	s.config.Options = v
	return s
}

func (s *SelectPickerField) Selected(v int) *SelectPickerField {
	s.config.Selected = v
	return s
}

func (s *SelectPickerField) Width(w int) *SelectPickerField {
	s.config.Width = w
	return s
}

func (s *SelectPickerField) Prefix(v string) *SelectPickerField {
	s.config.Prefix = v
	return s
}

func (s *SelectPickerField) Suffix(v string) *SelectPickerField {
	s.config.Suffix = v
	return s
}

func (s *SelectPickerField) Style(style retui.Style) *SelectPickerField {
	s.config.Style = style
	return s
}

func (s *SelectPickerField) Focused(v bool) *SelectPickerField {
	s.focused = v
	return s
}

func (s *SelectPickerField) OnChange(fn func(string, int, string)) *SelectPickerField {
	s.config.OnChange = fn
	return s
}

func (s *SelectPickerField) OnKeyPress(fn func(string, retui.Key) bool) *SelectPickerField {
	s.config.OnKeyPress = fn
	return s
}

func (s *SelectPickerField) OnFocus(fn func(string)) *SelectPickerField {
	s.config.OnFocus = fn
	return s
}

func (s *SelectPickerField) OnBlur(fn func(string)) *SelectPickerField {
	s.config.OnBlur = fn
	return s
}

func (s *SelectPickerField) OnSubmit(fn func(string, int, string)) *SelectPickerField {
	s.config.OnSubmit = fn
	return s
}

// ─── Render Method ──────────────────────────────────────────────────────

func (s *SelectPickerField) Render() retui.Element {
	return renderSelectPicker(s.focused, &s.config)
}

// ─── Core Rendering Function ────────────────────────────────────────────

func renderSelectPicker(focused bool, config *SelectPickerConfig) retui.Element {
	// Track selected index
	selected, setSelected := retui.UseState(config.Selected)

	// Sync with external config changes
	if selected != config.Selected {
		setSelected(config.Selected)
	}

	// Clamp selected index
	if selected < 0 {
		setSelected(0)
		selected = 0
	}
	if selected >= len(config.Options) {
		setSelected(len(config.Options) - 1)
		selected = len(config.Options) - 1
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
		case retui.KeyLeft:
			if selected > 0 {
				setSelected(selected - 1)
				config.Selected = selected - 1
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, selected-1, config.Options[selected-1])
				}
			}

		case retui.KeyRight:
			if selected < len(config.Options)-1 {
				setSelected(selected + 1)
				config.Selected = selected + 1
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, selected+1, config.Options[selected+1])
				}
			}

		case retui.KeyEnter:
			if config.OnSubmit != nil && config.ID != "" {
				config.OnSubmit(config.ID, selected, config.Options[selected])
			}
		}
	}

render:
	// Get the current selected option
	label := config.Options[selected]

	// Apply styles
	var textStyle retui.Style
	if focused {
		textStyle = config.Style.
			Foreground(retui.Cyan).
			Bold(true)
	} else {
		textStyle = config.Style.
			Foreground(retui.White)
	}

	// Build the display with arrows
	display := "< " + label + " >"

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
	if focused {
		prefixStyle = prefixStyle.Foreground(retui.Cyan).Bold(true)
	} else {
		prefixStyle = prefixStyle.Foreground(retui.BrightBlack)
	}

	suffixStyle := retui.NewStyle()
	if focused {
		suffixStyle = suffixStyle.Foreground(retui.Cyan).Bold(true)
	} else {
		suffixStyle = suffixStyle.Foreground(retui.BrightBlack)
	}

	// Build elements
	elements := []retui.Element{}

	if config.Prefix != "" {
		elements = append(elements, retui.Text(config.Prefix, prefixStyle))
	}

	elements = append(elements, retui.Text(display, textStyle))

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

func ExampleSelectPickerUsage() retui.Element {
	// Simple select picker
	colorPicker := SelectPicker().
		ID("color").
		Options([]string{"Red", "Green", "Blue", "Yellow"}).
		Selected(0).
		Width(20).
		Prefix("🎨 ").
		Suffix(" ✓").
		OnChange(func(id string, selected int, value string) {
			println("Selected:", value, "at index:", selected)
		}).
		OnSubmit(func(id string, selected int, value string) {
			println("Submitted:", value)
		})

	// Size picker
	sizePicker := SelectPicker().
		ID("size").
		Options([]string{"Small", "Medium", "Large", "XL"}).
		Selected(1).
		Prefix("📏 ")

	return retui.Box(
		retui.Props{
			Direction: retui.Column,
		},
		retui.NewStyle(),
		colorPicker.Render(),
		retui.Text(" ", retui.NewStyle()),
		sizePicker.Render(),
	)
}
