package components

import (
	"strconv"
	"strings"

	"github.com/subhasundardass/retui/retui"
)

// ─── Password Input Configuration ─────────────────────────────────────────

type PasswordConfig struct {
	ID           string
	Value        string
	Placeholder  string
	Width        int
	Style        retui.Style
	Prefix       string
	Suffix       string
	MinLength    int    // Minimum length (0 = no limit)
	MaxLength    int    // Maximum length (0 = no limit)
	MaskChar     string // Character to mask input (default: "•")
	ShowLastChar bool   // Show last character unmasked (default: true)
	OnChange     func(id string, value string)
	OnKeyPress   func(id string, key retui.Key) bool
	OnFocus      func(id string)
	OnBlur       func(id string)
	OnSubmit     func(id string, value string)
}

type PasswordFeild struct {
	config  PasswordConfig
	focused bool
}

// ─── Builder Methods ──────────────────────────────────────────────────────

func Password() *PasswordFeild {
	return &PasswordFeild{
		config: PasswordConfig{
			ID:           "",
			Value:        "",
			Placeholder:  "Enter password",
			Width:        30,
			Style:        retui.NewStyle(),
			Prefix:       "[ ",
			Suffix:       " ]",
			MinLength:    0,
			MaxLength:    0,
			MaskChar:     "•",
			ShowLastChar: true,
			OnChange:     nil,
			OnKeyPress:   nil,
			OnFocus:      nil,
			OnBlur:       nil,
			OnSubmit:     nil,
		},
		focused: false,
	}
}

func (p *PasswordFeild) ID(v string) *PasswordFeild {
	p.config.ID = v
	return p
}

func (p *PasswordFeild) Value(v string) *PasswordFeild {
	p.config.Value = v
	return p
}

func (p *PasswordFeild) Placeholder(v string) *PasswordFeild {
	p.config.Placeholder = v
	return p
}

func (p *PasswordFeild) Width(w int) *PasswordFeild {
	p.config.Width = w
	return p
}

func (p *PasswordFeild) Prefix(v string) *PasswordFeild {
	p.config.Prefix = v
	return p
}

func (p *PasswordFeild) Suffix(v string) *PasswordFeild {
	p.config.Suffix = v
	return p
}

func (p *PasswordFeild) Style(s retui.Style) *PasswordFeild {
	p.config.Style = s
	return p
}

func (p *PasswordFeild) MinLength(v int) *PasswordFeild {
	p.config.MinLength = v
	return p
}

func (p *PasswordFeild) MaxLength(v int) *PasswordFeild {
	p.config.MaxLength = v
	return p
}

func (p *PasswordFeild) MaskChar(v string) *PasswordFeild {
	p.config.MaskChar = v
	return p
}

func (p *PasswordFeild) ShowLastChar(v bool) *PasswordFeild {
	p.config.ShowLastChar = v
	return p
}

func (p *PasswordFeild) Focused(v bool) *PasswordFeild {
	p.focused = v
	return p
}

func (p *PasswordFeild) OnChange(fn func(string, string)) *PasswordFeild {
	p.config.OnChange = fn
	return p
}

func (p *PasswordFeild) OnKeyPress(fn func(string, retui.Key) bool) *PasswordFeild {
	p.config.OnKeyPress = fn
	return p
}

func (p *PasswordFeild) OnFocus(fn func(string)) *PasswordFeild {
	p.config.OnFocus = fn
	return p
}

func (p *PasswordFeild) OnBlur(fn func(string)) *PasswordFeild {
	p.config.OnBlur = fn
	return p
}

func (p *PasswordFeild) OnSubmit(fn func(string, string)) *PasswordFeild {
	p.config.OnSubmit = fn
	return p
}

// ─── Render Method ──────────────────────────────────────────────────────

func (p *PasswordFeild) Render() retui.Element {
	return renderPassword(p.focused, &p.config)
}

// ─── Helper Functions ──────────────────────────────────────────────────

// maskPassword masks the password string
func maskPassword(value string, maskChar string) string {
	if value == "" {
		return ""
	}

	runes := []rune(value)
	// Mask all characters
	return strings.Repeat(maskChar, len(runes))
}

// ─── Core Rendering Function ────────────────────────────────────────────

func renderPassword(focused bool, config *PasswordConfig) retui.Element {
	// Work in rune-space throughout so multi-byte UTF-8 characters
	// are never sliced through their middle.
	runes := []rune(config.Value)

	// Track cursor position using retui's state
	pos, setPos := retui.UseState(len(runes))

	// Value may have been changed externally since pos was last set;
	// clamp defensively to avoid an out-of-range slice on the next edit.
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

		case retui.KeyBackspace:
			if pos > 0 && len(runes) > 0 {
				newRunes := append(runes[:pos-1], runes[pos:]...)
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
				newRunes := append(runes[:pos], runes[pos+1:]...)
				newValue := string(newRunes)
				config.Value = newValue
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, newValue)
				}
			}

		case retui.KeyEnter:
			if config.MinLength > 0 && len(runes) < config.MinLength {
				// Return error state
				return renderPasswordError(config, "Minimum length is "+strconv.Itoa(config.MinLength)+" characters")
			}
			if config.OnSubmit != nil && config.ID != "" {
				config.OnSubmit(config.ID, config.Value)
			}

		default:
			// Insert printable character
			if key.Rune != 0 && key.Rune >= 32 && key.Rune <= 126 {
				if config.MaxLength == 0 || len(runes) < config.MaxLength {
					newRunes := append(runes[:pos], append([]rune{key.Rune}, runes[pos:]...)...)
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
	display := ""
	if config.Value != "" {
		display = maskPassword(config.Value, config.MaskChar)
	} else if config.Placeholder != "" && !focused {
		display = config.Placeholder
	}

	// Apply styles
	textStyle := config.Style
	if focused {
		textStyle = textStyle.
			Foreground(retui.White).
			Background(retui.Blue).
			Bold(true)
	} else {
		textStyle = textStyle.
			Foreground(retui.BrightBlack).
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
	if focused && config.Value != "" {
		displayRunes := []rune(display)
		// For masked display, cursor position needs to be mapped
		// We show the cursor at the end of the masked string
		if pos < len(displayRunes) {
			cursorDisplay = string(displayRunes[:pos]) + "█" + string(displayRunes[pos:])
		} else {
			cursorDisplay = string(displayRunes) + "█"
		}
	} else if focused && config.Value == "" {
		// Show cursor at start when empty
		cursorDisplay = "█" + display
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

func renderPasswordError(config *PasswordConfig, msg string) retui.Element {
	return retui.Box(
		retui.Props{Direction: retui.Row},
		retui.NewStyle(),
		retui.Text("⚠️ "+msg, retui.NewStyle().Foreground(retui.Red)),
	)
}

// ─── Example Usage ──────────────────────────────────────────────────────

// func ExamplePasswordUsage() retui.Element {
// 	// Simple password input
// 	passwordInput := Password().
// 		ID("password").
// 		Placeholder("Enter password").
// 		Width(30).
// 		MinLength(8).
// 		MaxLength(20).
// 		Prefix("🔒 ").
// 		Suffix(" ✓").
// 		OnChange(func(id string, value string) {
// 			println("Password changed")
// 		}).
// 		OnSubmit(func(id string, value string) {
// 			println("Password submitted")
// 		})

// 	// Password with custom mask
// 	customMask := Password().
// 		ID("pin").
// 		Placeholder("Enter PIN").
// 		Width(20).
// 		MaskChar("*").
// 		ShowLastChar(false).
// 		MaxLength(4).
// 		Prefix("PIN: ")

// 	return retui.Box(
// 		retui.Props{
// 			Direction: retui.Column,
// 		},
// 		retui.NewStyle(),
// 		passwordInput.Render(),
// 		retui.Text(" ", retui.NewStyle()),
// 		customMask.Render(),
// 	)
// }
