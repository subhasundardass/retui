package components

import (
	"strconv"
	"strings"

	"github.com/subhasundardass/retui/retui"
)

type InputConfig struct {
	ID          string
	Value       string
	Placeholder string
	Width       int
	Style       retui.Style
	Prefix      string
	Suffix      string
	MinLength   int // Minimum length (0 = no limit)
	MaxLength   int // Maximum length (0 = no limit)
	OnChange    func(id string, value string)
	OnKeyPress  func(id string, key retui.Key) bool
	OnFocus     func(id string)
	OnBlur      func(id string)
	OnSubmit    func(id string, value string)
}

type InputField struct {
	config  InputConfig
	focused bool
}

// ─── Builder Methods ──────────────────────────────────────────────────────

func TextInput() *InputField {
	return &InputField{
		config: InputConfig{
			ID:          "",
			Value:       "",
			Placeholder: "",
			Width:       30,
			Style:       retui.NewStyle(),
			Prefix:      "",
			Suffix:      "",
			MinLength:   0,
			MaxLength:   0,
			OnChange:    nil,
			OnKeyPress:  nil,
			OnFocus:     nil,
			OnBlur:      nil,
			OnSubmit:    nil,
		},
		focused: false,
	}
}

func (i *InputField) ID(v string) *InputField {
	i.config.ID = v
	return i
}

func (i *InputField) Value(v string) *InputField {
	i.config.Value = v
	return i
}

func (i *InputField) Placeholder(v string) *InputField {
	i.config.Placeholder = v
	return i
}

func (i *InputField) Width(w int) *InputField {
	i.config.Width = w
	return i
}

func (i *InputField) Prefix(v string) *InputField {
	i.config.Prefix = v
	return i
}

func (i *InputField) Suffix(v string) *InputField {
	i.config.Suffix = v
	return i
}

func (i *InputField) Style(s retui.Style) *InputField {
	i.config.Style = s
	return i
}

func (i *InputField) MinLength(v int) *InputField {
	i.config.MinLength = v
	return i
}

func (i *InputField) MaxLength(v int) *InputField {
	i.config.MaxLength = v
	return i
}

func (i *InputField) Focused(v bool) *InputField {
	i.focused = v
	return i
}

func (i *InputField) OnChange(fn func(string, string)) *InputField {
	i.config.OnChange = fn
	return i
}

func (i *InputField) OnKeyPress(fn func(string, retui.Key) bool) *InputField {
	i.config.OnKeyPress = fn
	return i
}

func (i *InputField) OnFocus(fn func(string)) *InputField {
	i.config.OnFocus = fn
	return i
}

func (i *InputField) OnBlur(fn func(string)) *InputField {
	i.config.OnBlur = fn
	return i
}

func (i *InputField) OnSubmit(fn func(string, string)) *InputField {
	i.config.OnSubmit = fn
	return i
}

// ─── Render Method ──────────────────────────────────────────────────────

func (i *InputField) Render() retui.Element {
	return renderInput(i.focused, &i.config)
}

// ─── Core Rendering Function ────────────────────────────────────────────

func renderInput(focused bool, config *InputConfig) retui.Element {
	// Work in rune-space throughout so multi-byte UTF-8 characters
	// (e.g. set via Value()) are never sliced through their middle.
	runes := []rune(config.Value)

	// Track cursor position using retui's state
	pos, setPos := retui.UseState(len(runes))

	// Value may have been changed externally (e.g. via the Value() prop)
	// since pos was last set; clamp defensively to avoid an out-of-range
	// slice on the next edit.
	if pos > len(runes) {
		pos = len(runes)
		setPos(pos)
	}
	if pos < 0 {
		pos = 0
		setPos(pos)
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
			if pos > 0 {
				setPos(pos - 1)
			}

		case retui.KeyRight:
			if pos < len(runes) {
				setPos(pos + 1)
			}

		case retui.KeySpace:
			if config.MaxLength == 0 || len(runes) < config.MaxLength {
				newRunes := append(append(append([]rune{}, runes[:pos]...), ' '), runes[pos:]...)
				newValue := string(newRunes)
				config.Value = newValue
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, newValue)
				}
				setPos(pos + 1)
			}

		case retui.KeyBackspace:
			if pos > 0 && len(runes) > 0 {
				newRunes := append(append([]rune{}, runes[:pos-1]...), runes[pos:]...)
				newValue := string(newRunes)
				config.Value = newValue
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, newValue)
				}
				setPos(pos - 1)
			}

		case retui.KeyHome:
			setPos(0)

		case retui.KeyEnd:
			setPos(len(runes))

		case retui.KeyDelete:
			if pos < len(runes) {
				newRunes := append(append([]rune{}, runes[:pos]...), runes[pos+1:]...)
				newValue := string(newRunes)
				config.Value = newValue
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, newValue)
				}
			}

		case retui.KeyEnter:
			if config.MinLength > 0 && len(runes) < config.MinLength {
				// Return error state
				return renderError(config, "Minimum length is "+strconv.Itoa(config.MinLength)+" characters")
			}
			if config.OnSubmit != nil && config.ID != "" {
				config.OnSubmit(config.ID, config.Value)
			}

		default:
			// Insert printable character
			if key.Rune != 0 && key.Rune >= 32 && key.Rune <= 126 {
				if config.MaxLength == 0 || len(runes) < config.MaxLength {
					newRunes := append(append(append([]rune{}, runes[:pos]...), key.Rune), runes[pos:]...)
					newValue := string(newRunes)
					config.Value = newValue
					if config.OnChange != nil && config.ID != "" {
						config.OnChange(config.ID, newValue)
					}
					setPos(pos + 1)
				}
			}
		}
	}

render:
	// Refresh rune view in case config.Value was mutated above.
	runes = []rune(config.Value)
	if pos > len(runes) {
		pos = len(runes)
	}

	// Check validation
	isValid := true
	if config.MinLength > 0 && len(runes) < config.MinLength {
		isValid = false
	}

	// Determine display value
	display := config.Value
	if display == "" && config.Placeholder != "" && !focused {
		display = config.Placeholder
	}

	// Apply styles
	textStyle := config.Style
	if focused {
		textStyle = textStyle.
			Foreground(retui.BrightWhite).
			Background(retui.Blue).
			Bold(true)
	} else {
		textStyle = textStyle.
			Foreground(retui.BrightBlack).
			Background(retui.Hex("#0c0c0c")).
			Bold(true)
	}

	// Border color based on focus and validation
	borderColor := retui.BrightBlack
	if focused {
		if isValid {
			borderColor = retui.Cyan
		} else {
			borderColor = retui.Red
		}
	}

	bracketStyle := retui.NewStyle()
	if focused {
		bracketStyle = bracketStyle.
			Foreground(borderColor).
			Bold(true)
	} else {
		bracketStyle = bracketStyle.
			Foreground(retui.BrightBlack)
	}

	// Add cursor
	cursorDisplay := display
	if focused {
		displayRunes := []rune(display)
		if pos < len(displayRunes) {
			cursorDisplay = string(displayRunes[:pos]) + "█" + string(displayRunes[pos:])
		} else {
			cursorDisplay = string(displayRunes) + "█"
		}
	}

	// Pad to width
	paddedDisplay := cursorDisplay
	displayLen := len([]rune(paddedDisplay))
	if displayLen < config.Width {
		padding := strings.Repeat(" ", config.Width-displayLen)
		paddedDisplay = paddedDisplay + padding
	}

	// Build elements
	elements := []retui.Element{}

	if config.Prefix != "" {
		elements = append(elements, retui.Text(config.Prefix, bracketStyle))
	}

	// Apply validation styling
	textStyleForDisplay := textStyle
	if !isValid && !focused {
		textStyleForDisplay = textStyleForDisplay.Foreground(retui.Red)
	}

	elements = append(elements, retui.Text(paddedDisplay, textStyleForDisplay))

	if config.Suffix != "" {
		elements = append(elements, retui.Text(config.Suffix, bracketStyle))
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

// ─── Helper: Render Error State ────────────────────────────────────────

func renderError(config *InputConfig, msg string) retui.Element {
	return retui.Box(
		retui.Props{Direction: retui.Row},
		retui.NewStyle(),
		retui.Text(msg, retui.NewStyle().Foreground(retui.Red)),
	)
}
