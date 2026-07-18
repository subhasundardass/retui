package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

// ─────────────────────────────────────────────────────────────────────────────
// SelectDropdown - A basic dropdown select with overlay
// ─────────────────────────────────────────────────────────────────────────────

type SelectOption struct {
	Label    string
	Value    string
	Disabled bool
}

type SelectOptionFunc func(*SelectConfig)

type SelectConfig struct {
	ID           string
	Label        string
	Options      []SelectOption
	Selected     string
	Width        int
	Height       int
	Style        retui.Style
	OnChange     func(id string, value string)
	OnKeyPress   func(id string, key retui.Key) bool
	OnFocus      func(id string)
	OnBlur       func(id string)
	OnOpenChange func(id string, isOpen bool)
	Placeholder  string
	Disabled     bool
}

func SelectWithID(id string) SelectOptionFunc {
	return func(c *SelectConfig) { c.ID = id }
}

func SelectWithLabel(label string) SelectOptionFunc {
	return func(c *SelectConfig) { c.Label = label }
}

func SelectWithOptions(options []SelectOption) SelectOptionFunc {
	return func(c *SelectConfig) { c.Options = options }
}

func SelectWithSelected(value string) SelectOptionFunc {
	return func(c *SelectConfig) { c.Selected = value }
}

func SelectWithWidth(width int) SelectOptionFunc {
	return func(c *SelectConfig) { c.Width = width }
}

func SelectWithHeight(height int) SelectOptionFunc {
	return func(c *SelectConfig) { c.Height = height }
}

func SelectWithStyle(style retui.Style) SelectOptionFunc {
	return func(c *SelectConfig) { c.Style = style }
}

func SelectWithOnChange(fn func(id string, value string)) SelectOptionFunc {
	return func(c *SelectConfig) { c.OnChange = fn }
}

func SelectWithOnKeyPress(fn func(id string, key retui.Key) bool) SelectOptionFunc {
	return func(c *SelectConfig) { c.OnKeyPress = fn }
}

func SelectWithOnFocus(fn func(id string)) SelectOptionFunc {
	return func(c *SelectConfig) { c.OnFocus = fn }
}

func SelectWithOnBlur(fn func(id string)) SelectOptionFunc {
	return func(c *SelectConfig) { c.OnBlur = fn }
}

func SelectWithOnOpenChange(fn func(id string, isOpen bool)) SelectOptionFunc {
	return func(c *SelectConfig) { c.OnOpenChange = fn }
}

func SelectWithPlaceholder(placeholder string) SelectOptionFunc {
	return func(c *SelectConfig) { c.Placeholder = placeholder }
}

func SelectWithDisabled(disabled bool) SelectOptionFunc {
	return func(c *SelectConfig) { c.Disabled = disabled }
}

func overlayFocusID(config *SelectConfig) string {
	if config.ID == "" {
		return "__unnamed_select_overlay__"
	}
	return config.ID + "__overlay"
}

// ─── SelectDropdown ─────────────────────────────────────────────────────────

func SelectDropdown(focused bool, opts ...SelectOptionFunc) retui.Element {
	config := &SelectConfig{
		ID:          "",
		Label:       "",
		Options:     []SelectOption{},
		Selected:    "",
		Width:       30,
		Height:      5,
		Style:       retui.NewStyle(),
		Placeholder: "Select...",
		Disabled:    false,
		OnChange:    nil,
		OnFocus:     nil,
		OnBlur:      nil,
	}
	for _, opt := range opts {
		opt(config)
	}

	// ── State ────────────────────────────────────────────────────────────
	isOpen, setIsOpen := retui.UseState(false)
	highlightedIndex, setHighlightedIndex := retui.UseState(0)
	selectedValue, setSelectedValue := retui.UseState(config.Selected)

	overlayID := overlayFocusID(config)

	isFocusedNow := focused
	if isOpen {
		isFocusedNow = retui.IsFocused(overlayID)
	}

	componentID := config.ID
	if isOpen {
		componentID = overlayID
	}
	key, isMine := retui.UseFocusedKey(componentID, isFocusedNow)

	// Find selected index
	selectedIndex := 0
	found := false
	for i, opt := range config.Options {
		if opt.Value == selectedValue {
			selectedIndex = i
			found = true
			break
		}
	}
	if !found && len(config.Options) > 0 {
		selectedIndex = 0
	}

	// Get selected label
	selectedLabel := config.Placeholder
	for _, opt := range config.Options {
		if opt.Value == selectedValue {
			selectedLabel = opt.Label
			break
		}
	}

	// Determine if input should appear focused
	inputFocused := isMine && !isOpen && !config.Disabled

	if inputFocused && config.OnFocus != nil && config.ID != "" {
		config.OnFocus(config.ID)
	}

	// ── Helper: close the overlay ──────────────────────────────────────
	openOverlay := func() {
		setIsOpen(true)
		setHighlightedIndex(selectedIndex)

		retui.SetFocus(config.ID) // anchor baseline so the stack has the right ID
		retui.PushFocus(overlayID)
		retui.CaptureFocus(overlayID)

		if config.OnOpenChange != nil && config.ID != "" {
			config.OnOpenChange(config.ID, true)
		}
	}

	closeOverlay := func() {
		setIsOpen(false)
		retui.ReleaseCaptureFocus()
		retui.PopFocus()          // restores current -> config.ID (now correct)
		retui.SetFocus(config.ID) // belt-and-suspenders: force it explicitly too

		if config.OnBlur != nil && config.ID != "" {
			config.OnBlur(overlayID)
		}
		if config.OnOpenChange != nil && config.ID != "" {
			config.OnOpenChange(config.ID, false)
		}
	}

	// ── Keyboard handling ───────────────────────────────────────────────
	if isMine && !config.Disabled {
		if !isOpen {
			// INPUT MODE - Closed
			switch key.Code {
			case retui.KeyEscape:
				closeOverlay() // closes without touching selectedValue, and refocuses config.ID

			case retui.KeyEnter, retui.KeySpace:
				openOverlay() // Opens and captures focus automatically
			}
		} else {
			// DROPDOWN MODE - Open (all keys captured)
			switch key.Code {
			case retui.KeyEscape:
				closeOverlay()
			case retui.KeyEnter, retui.KeySpace:
				// Select and close
				if highlightedIndex >= 0 && highlightedIndex < len(config.Options) {
					option := config.Options[highlightedIndex]
					if !option.Disabled {
						setSelectedValue(option.Value)
						if config.OnChange != nil && config.ID != "" {
							config.OnChange(config.ID, option.Value)
						}
					}
				}
				closeOverlay()
			case retui.KeyDown:
				next := highlightedIndex + 1
				for next < len(config.Options) && config.Options[next].Disabled {
					next++
				}
				if next < len(config.Options) {
					setHighlightedIndex(next)
				}
			case retui.KeyUp:
				prev := highlightedIndex - 1
				for prev >= 0 && config.Options[prev].Disabled {
					prev--
				}
				if prev >= 0 {
					setHighlightedIndex(prev)
				}
			case retui.KeyTab:
				closeOverlay()
			}
		}
	}

	// Defensive re-sync
	if isOpen && config.ID != "" && !retui.IsFocused(overlayID) {
		setIsOpen(false)
	}

	// ── Render ───────────────────────────────────────────────────────────
	inputElement := buildInput(config, selectedLabel, isOpen, inputFocused)

	var overlayElement retui.Element
	if isOpen && len(config.Options) > 0 {
		overlayElement = buildOverlay(config, selectedValue, highlightedIndex, isOpen)
	} else {
		overlayElement = retui.Text("", retui.NewStyle())
	}

	return retui.Box(
		retui.Props{Direction: retui.Column},
		retui.NewStyle(),
		inputElement,
		overlayElement,
	)
}

// buildInput builds the input field
func buildInput(config *SelectConfig, selectedLabel string, isOpen bool, focused bool) retui.Element {
	displayText := selectedLabel
	if displayText == "" {
		displayText = config.Placeholder
	}

	// Text style
	textStyle := config.Style
	switch {
	case config.Disabled:
		textStyle = textStyle.Foreground(retui.BrightBlack)
	case isOpen:
		// When dropdown is open, input text dims
		textStyle = textStyle.Foreground(retui.BrightBlack)
	case focused:
		// When focused, text is white on blue background
		textStyle = textStyle.Foreground(retui.White)
	default:
		textStyle = textStyle.Foreground(retui.BrightBlack)
	}

	// Arrow style
	arrowStyle := retui.NewStyle()
	switch {
	case config.Disabled:
		arrowStyle = arrowStyle.Foreground(retui.BrightBlack)
	case isOpen:
		arrowStyle = arrowStyle.Foreground(retui.Cyan).Bold(true)
	case focused:
		arrowStyle = arrowStyle.Foreground(retui.White)
	default:
		arrowStyle = arrowStyle.Foreground(retui.BrightBlack)
	}

	arrow := "▼"
	if isOpen {
		arrow = "▲"
	}

	// Pad display to width (without brackets now)
	paddedDisplay := displayText
	displayLen := len([]rune(paddedDisplay))
	// Width - 2 for the arrow and space
	if displayLen < config.Width-2 {
		padding := strings.Repeat(" ", config.Width-2-displayLen)
		paddedDisplay = paddedDisplay + padding
	}

	// Build the input content
	inputContent := retui.Box(
		retui.Props{
			Direction: retui.Row,
			Width:     retui.Fixed(config.Width),
		},
		retui.NewStyle(),
		retui.Text(paddedDisplay, textStyle),
		retui.Text(" ", retui.NewStyle()),
		retui.Text(arrow, arrowStyle),
	)

	// Apply background to the ENTIRE input when focused
	var inputField retui.Element
	if focused && !isOpen && !config.Disabled {
		// Whole input gets blue background
		inputField = retui.Box(
			retui.Props{
				Direction: retui.Row,
				Width:     retui.Fixed(config.Width),
			},
			retui.NewStyle().Background(retui.Blue),
			inputContent,
		)
	} else {
		inputField = inputContent
	}

	// Add label if present
	if config.Label != "" {
		labelStyle := retui.NewStyle().Foreground(retui.White)
		if config.Disabled || isOpen {
			labelStyle = labelStyle.Foreground(retui.BrightBlack)
		}
		labelElement := retui.Text(config.Label+":", labelStyle)

		return retui.Box(
			retui.Props{Direction: retui.Row, Gap: 1},
			retui.NewStyle(),
			labelElement,
			inputField,
		)
	}

	return inputField
}

// buildOverlay builds the dropdown overlay
func buildOverlay(config *SelectConfig, selectedValue string, highlightedIndex int, isOpen bool) retui.Element {
	if !isOpen || len(config.Options) == 0 {
		return retui.Text("", retui.NewStyle())
	}

	optionElements := []retui.Element{}

	startIdx := 0
	endIdx := len(config.Options)
	if len(config.Options) > config.Height {
		if highlightedIndex >= config.Height {
			startIdx = highlightedIndex - config.Height + 1
		}
		endIdx = startIdx + config.Height
		if endIdx > len(config.Options) {
			endIdx = len(config.Options)
			startIdx = endIdx - config.Height
		}
	}

	if startIdx > 0 {
		optionElements = append(optionElements,
			retui.Text(" ↑ ", retui.NewStyle().Foreground(retui.BrightBlack)))
	}

	for i := startIdx; i < endIdx; i++ {
		opt := config.Options[i]
		style := retui.NewStyle()

		switch {
		case opt.Disabled:
			style = style.Foreground(retui.BrightBlack)
		case i == highlightedIndex:
			style = style.Background(retui.Blue).Foreground(retui.White).Bold(true)
		default:
			style = style.Foreground(retui.White)
		}

		prefix := "  "
		if i == highlightedIndex && !opt.Disabled {
			prefix = "▶ "
		}
		if opt.Value == selectedValue && !opt.Disabled && i != highlightedIndex {
			prefix = "✓ "
		}

		optionElements = append(optionElements, retui.Text(prefix+opt.Label, style))
	}

	if endIdx < len(config.Options) {
		optionElements = append(optionElements,
			retui.Text(" ↓ ", retui.NewStyle().Foreground(retui.BrightBlack)))
	}

	dropdownWidth := config.Width + 2
	dropdownBox := retui.Box(
		retui.Props{
			Direction: retui.Column,
			Width:     retui.Fixed(dropdownWidth),
		},
		retui.NewStyle().
			Background(retui.Black).
			Border(retui.Border{
				Top:    true,
				Right:  true,
				Bottom: true,
				Left:   true,
				Chars:  retui.BorderRounded,
				Color:  retui.Cyan,
			}),
		optionElements...,
	)

	return retui.Overlay(0, 1, dropdownBox)
}
