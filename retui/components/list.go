package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

// ─── List Configuration ──────────────────────────────────────────────────

type ListConfig struct {
	ID              string
	Items           []string
	Selected        int
	Width           int
	Style           retui.Style
	SelectedStyle   retui.Style
	UnselectedStyle retui.Style
	Prefix          string
	Suffix          string
	OnSelect        func(id string, index int, value string)
	OnKeyPress      func(id string, key retui.Key) bool
	OnFocus         func(id string)
	OnBlur          func(id string)
	OnSubmit        func(id string, index int, value string)
}

type ListField struct {
	config  ListConfig
	focused bool
}

// ─── Builder Methods ──────────────────────────────────────────────────────

func List() *ListField {
	return &ListField{
		config: ListConfig{
			ID:              "",
			Items:           []string{},
			Selected:        0,
			Width:           0,
			Style:           retui.NewStyle(),
			SelectedStyle:   retui.NewStyle().Foreground(retui.Cyan).Bold(true),
			UnselectedStyle: retui.NewStyle().Foreground(retui.BrightBlack),
			Prefix:          "",
			Suffix:          "",
			OnSelect:        nil,
			OnKeyPress:      nil,
			OnFocus:         nil,
			OnBlur:          nil,
			OnSubmit:        nil,
		},
		focused: false,
	}
}

func (l *ListField) ID(v string) *ListField {
	l.config.ID = v
	return l
}

func (l *ListField) Items(v []string) *ListField {
	l.config.Items = v
	return l
}

func (l *ListField) Selected(v int) *ListField {
	l.config.Selected = v
	return l
}

func (l *ListField) Width(w int) *ListField {
	l.config.Width = w
	return l
}

func (l *ListField) Prefix(v string) *ListField {
	l.config.Prefix = v
	return l
}

func (l *ListField) Suffix(v string) *ListField {
	l.config.Suffix = v
	return l
}

func (l *ListField) Style(s retui.Style) *ListField {
	l.config.Style = s
	return l
}

func (l *ListField) SelectedStyle(s retui.Style) *ListField {
	l.config.SelectedStyle = s
	return l
}

func (l *ListField) UnselectedStyle(s retui.Style) *ListField {
	l.config.UnselectedStyle = s
	return l
}

func (l *ListField) Focused(v bool) *ListField {
	l.focused = v
	return l
}

func (l *ListField) OnSelect(fn func(string, int, string)) *ListField {
	l.config.OnSelect = fn
	return l
}

func (l *ListField) OnKeyPress(fn func(string, retui.Key) bool) *ListField {
	l.config.OnKeyPress = fn
	return l
}

func (l *ListField) OnFocus(fn func(string)) *ListField {
	l.config.OnFocus = fn
	return l
}

func (l *ListField) OnBlur(fn func(string)) *ListField {
	l.config.OnBlur = fn
	return l
}

func (l *ListField) OnSubmit(fn func(string, int, string)) *ListField {
	l.config.OnSubmit = fn
	return l
}

// ─── Render Method ──────────────────────────────────────────────────────

func (l *ListField) Render() retui.Element {
	return renderList(l.focused, &l.config)
}

// ─── Core Rendering Function ────────────────────────────────────────────

func renderList(focused bool, config *ListConfig) retui.Element {
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
	if selected >= len(config.Items) {
		setSelected(len(config.Items) - 1)
		selected = len(config.Items) - 1
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
		case retui.KeyDown:
			if selected < len(config.Items)-1 {
				setSelected(selected + 1)
				config.Selected = selected + 1
				if config.OnSelect != nil && config.ID != "" {
					config.OnSelect(config.ID, selected+1, config.Items[selected+1])
				}
			}

		case retui.KeyUp:
			if selected > 0 {
				setSelected(selected - 1)
				config.Selected = selected - 1
				if config.OnSelect != nil && config.ID != "" {
					config.OnSelect(config.ID, selected-1, config.Items[selected-1])
				}
			}

		case retui.KeyHome:
			setSelected(0)
			config.Selected = 0
			if config.OnSelect != nil && config.ID != "" {
				config.OnSelect(config.ID, 0, config.Items[0])
			}

		case retui.KeyEnd:
			setSelected(len(config.Items) - 1)
			config.Selected = len(config.Items) - 1
			if config.OnSelect != nil && config.ID != "" {
				config.OnSelect(config.ID, len(config.Items)-1, config.Items[len(config.Items)-1])
			}

		case retui.KeyEnter:
			if config.OnSubmit != nil && config.ID != "" {
				config.OnSubmit(config.ID, selected, config.Items[selected])
			}
		}
	}

render:
	// Build list items
	children := make([]retui.Element, len(config.Items))

	for i, item := range config.Items {
		// Determine if this item is selected
		isSelected := i == selected

		// Build the item display
		prefix := " "
		if isSelected {
			prefix = ">"
		}

		// Apply styles
		var itemStyle retui.Style
		if isSelected {
			if focused {
				itemStyle = config.SelectedStyle.
					Background(retui.Blue).
					Foreground(retui.White).
					Bold(true)
			} else {
				itemStyle = config.SelectedStyle.
					Foreground(retui.White).
					Bold(true)
			}
		} else {
			itemStyle = config.UnselectedStyle
		}

		// Pad to width if specified
		display := prefix + item
		if config.Width > 0 {
			displayLen := len([]rune(display))
			if displayLen < config.Width {
				padding := strings.Repeat(" ", config.Width-displayLen)
				display = display + padding
			}
		}

		children[i] = retui.Text(display, itemStyle)
	}

	// Add prefix/suffix elements
	elements := []retui.Element{}

	if config.Prefix != "" {
		prefixStyle := retui.NewStyle()
		if focused {
			prefixStyle = prefixStyle.Foreground(retui.Cyan).Bold(true)
		} else {
			prefixStyle = prefixStyle.Foreground(retui.BrightBlack)
		}
		elements = append(elements, retui.Text(config.Prefix, prefixStyle))
	}

	// Add the list items
	for _, child := range children {
		elements = append(elements, child)
	}

	if config.Suffix != "" {
		suffixStyle := retui.NewStyle()
		if focused {
			suffixStyle = suffixStyle.Foreground(retui.Cyan).Bold(true)
		} else {
			suffixStyle = suffixStyle.Foreground(retui.BrightBlack)
		}
		elements = append(elements, retui.Text(config.Suffix, suffixStyle))
	}

	// Return as a column
	return retui.Box(
		retui.Props{
			Direction: retui.Column,
		},
		retui.NewStyle(),
		elements...,
	)
}

// ─── Example Usage ──────────────────────────────────────────────────────

// func ExampleListUsage() retui.Element {
// 	// Simple list
// 	fruitList := List().
// 		ID("fruits").
// 		Items([]string{"Apple", "Banana", "Orange", "Grape", "Mango"}).
// 		Selected(0).
// 		Width(20).
// 		Prefix("📋 ").
// 		Suffix(" ✓").
// 		OnSelect(func(id string, index int, value string) {
// 			println("Selected:", value, "at index:", index)
// 		}).
// 		OnSubmit(func(id string, index int, value string) {
// 			println("Submitted:", value)
// 		})

// 	// List with custom styles
// 	colorList := List().
// 		ID("colors").
// 		Items([]string{"Red", "Green", "Blue", "Yellow"}).
// 		Selected(2).
// 		SelectedStyle(retui.NewStyle().Foreground(retui.Green).Bold(true)).
// 		UnselectedStyle(retui.NewStyle().Foreground(retui.BrightBlack))

// 	return retui.Box(
// 		retui.Props{
// 			Direction: retui.Column,
// 		},
// 		retui.NewStyle(),
// 		fruitList.Render(),
// 		retui.Text(" ", retui.NewStyle()),
// 		colorList.Render(),
// 	)
// }
