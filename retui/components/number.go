package components

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/subhasundardass/retui/retui"
)

// ─── Number Input Configuration ──────────────────────────────────────────

// NumberInputConfig holds the state and callbacks for a NumberInputField.
//
// Bounds: HasMin/HasMax gate whether Min/Max are enforced, so a legitimate
// bound of exactly 0 (e.g. Min(0) for an age field, or Max(0) for a field
// that must be non-positive) is respected. Setting Min/Max via the builder
// methods always sets the corresponding HasMin/HasMax flag.
//
// Emptiness: Empty distinguishes "no value entered" from "value is 0".
// It starts true and is cleared automatically the first time Value() is
// called on the builder, or the first time the user commits a parseable
// number. When Empty is true, the placeholder is shown instead of "0".
//
// Select-all-on-focus: like a typical web <input>, gaining focus selects
// the entire current value. While selected, typing a character replaces
// the whole value instead of inserting at the cursor, and Backspace/
// Delete clear it entirely. Left/Right/Home/End instead just collapse the
// selection to that boundary, matching standard browser behavior. Set
// SelectAllOnFocus to false to disable this and keep the cursor wherever
// it was, like a plain text field.
type NumberInputConfig struct {
	ID          string
	Value       float64
	Empty       bool
	Placeholder string
	Width       int
	Style       retui.Style
	Prefix      string
	Suffix      string
	HasMin      bool
	Min         float64
	HasMax      bool
	Max         float64
	Step        float64 // Increment/decrement step (default: 1)
	Decimals    int     // Number of decimal places allowed (0 = integer only)
	ArrowStep   bool    // If false, Up/Down do not change the value (default: true).
	// Set to false when the surrounding form/list also binds Up/Down to
	// move focus between fields, so a single keypress can't both step
	// this field's value and move focus off it in the same instant.

	// SelectAllOnFocus enables web-style select-all behavior on focus.
	// Defaults to true.
	SelectAllOnFocus bool

	OnChange   func(id string, value float64)
	OnKeyPress func(id string, key retui.Key) bool
	OnFocus    func(id string)
	OnBlur     func(id string)
	OnSubmit   func(id string, value float64)
}

// NumberInputField is a builder for a numeric text input component.
type NumberInputField struct {
	config  NumberInputConfig
	focused bool
}

// ─── Builder Methods ──────────────────────────────────────────────────────

// NumberInput creates a new NumberInputField with sensible defaults.
// The field starts Empty (no value) and displays Placeholder until the
// user types something or Value() is called explicitly.
func NumberInput() *NumberInputField {
	return &NumberInputField{
		config: NumberInputConfig{
			ID:               "",
			Value:            0,
			Empty:            true,
			Placeholder:      "0",
			Width:            30,
			Style:            retui.NewStyle(),
			Prefix:           "",
			Suffix:           "",
			HasMin:           false,
			Min:              0,
			HasMax:           false,
			Max:              0,
			Step:             1,
			Decimals:         0,
			ArrowStep:        false,
			SelectAllOnFocus: true,
			OnChange:         nil,
			OnKeyPress:       nil,
			OnFocus:          nil,
			OnBlur:           nil,
			OnSubmit:         nil,
		},
		focused: false,
	}
}

func (n *NumberInputField) ID(v string) *NumberInputField {
	n.config.ID = v
	return n
}

// Value sets an explicit starting value and marks the field as non-empty,
// even if v is 0. Use this to pre-fill a field with a real zero value.
func (n *NumberInputField) Value(v float64) *NumberInputField {
	n.config.Value = v
	n.config.Empty = false
	return n
}

// Empty explicitly marks the field as having no value, showing the
// placeholder regardless of the current Value. Useful for reset/clear.
func (n *NumberInputField) Empty(v bool) *NumberInputField {
	n.config.Empty = v
	return n
}

func (n *NumberInputField) Placeholder(v string) *NumberInputField {
	n.config.Placeholder = v
	return n
}

func (n *NumberInputField) Width(w int) *NumberInputField {
	n.config.Width = w
	return n
}

func (n *NumberInputField) Prefix(v string) *NumberInputField {
	n.config.Prefix = v
	return n
}

func (n *NumberInputField) Suffix(v string) *NumberInputField {
	n.config.Suffix = v
	return n
}

func (n *NumberInputField) Style(s retui.Style) *NumberInputField {
	n.config.Style = s
	return n
}

// Min sets a minimum bound and enables its enforcement (including a
// legitimate minimum of exactly 0).
func (n *NumberInputField) Min(v float64) *NumberInputField {
	n.config.Min = v
	n.config.HasMin = true
	return n
}

// Max sets a maximum bound and enables its enforcement (including a
// legitimate maximum of exactly 0).
func (n *NumberInputField) Max(v float64) *NumberInputField {
	n.config.Max = v
	n.config.HasMax = true
	return n
}

// Step sets the increment/decrement applied by the Up/Down keys.
// Non-positive values are ignored and fall back to the default of 1.
func (n *NumberInputField) Step(v float64) *NumberInputField {
	if v <= 0 {
		v = 1
	}
	n.config.Step = v
	return n
}

// Decimals sets how many digits after the decimal point are allowed.
// 0 means the field only accepts integers (no decimal point at all).
func (n *NumberInputField) Decimals(v int) *NumberInputField {
	if v < 0 {
		v = 0
	}
	n.config.Decimals = v
	return n
}

// ArrowStep controls whether Up/Down change the value. Disable this when
// the surrounding form/list also uses Up/Down to move focus between
// fields, so the same keypress can't both step the value and blur the
// field at once.
func (n *NumberInputField) ArrowStep(v bool) *NumberInputField {
	n.config.ArrowStep = v
	return n
}

// SelectAllOnFocus toggles web-style select-all-on-focus behavior.
// Enabled by default; pass false to keep the cursor position instead.
func (n *NumberInputField) SelectAllOnFocus(v bool) *NumberInputField {
	n.config.SelectAllOnFocus = v
	return n
}

func (n *NumberInputField) Focused(v bool) *NumberInputField {
	n.focused = v
	return n
}

func (n *NumberInputField) OnChange(fn func(string, float64)) *NumberInputField {
	n.config.OnChange = fn
	return n
}

func (n *NumberInputField) OnKeyPress(fn func(string, retui.Key) bool) *NumberInputField {
	n.config.OnKeyPress = fn
	return n
}

func (n *NumberInputField) OnFocus(fn func(string)) *NumberInputField {
	n.config.OnFocus = fn
	return n
}

func (n *NumberInputField) OnBlur(fn func(string)) *NumberInputField {
	n.config.OnBlur = fn
	return n
}

func (n *NumberInputField) OnSubmit(fn func(string, float64)) *NumberInputField {
	n.config.OnSubmit = fn
	return n
}

// ─── Render Method ──────────────────────────────────────────────────────

func (n *NumberInputField) Render() retui.Element {
	return renderNumberInput(n.focused, &n.config)
}

// ─── Helper Functions ──────────────────────────────────────────────────

// formatNumber formats a float64 with the specified number of decimals,
// normalizing negative zero (e.g. -0.0001 rounded to 0 decimals) to "0"
// rather than the confusing "-0".
func formatNumber(value float64, decimals int) string {
	if value == 0 {
		value = 0 // collapse -0.0 to +0.0
	}
	return strconv.FormatFloat(value, 'f', decimals, 64)
}

// roundToDecimals rounds value to the given number of decimal places,
// used to avoid floating-point drift when stepping with a fractional Step.
func roundToDecimals(value float64, decimals int) float64 {
	if decimals <= 0 {
		return math.Round(value)
	}
	shift := math.Pow(10, float64(decimals))
	return math.Round(value*shift) / shift
}

// parseNumber parses a string to float64, treating incomplete input
// ("", "-", ".", "-.") as "not yet a number" rather than an error.
func parseNumber(s string) (float64, bool) {
	if s == "" || s == "-" || s == "." || s == "-." {
		return 0, false
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return val, true
}

// clampValue restricts value to [min, max], honoring hasMin/hasMax so a
// bound of exactly 0 is enforced correctly.
func clampValue(value float64, hasMin bool, min float64, hasMax bool, max float64) float64 {
	if hasMin && value < min {
		return min
	}
	if hasMax && value > max {
		return max
	}
	return value
}

// isValidNumber reports whether s is a valid float, or a valid in-progress
// prefix of one (e.g. "-", "12."), respecting the decimal-place limit.
func isValidNumber(s string, decimals int) bool {
	if s == "" || s == "-" || s == "." || s == "-." {
		return true // partial input is allowed while typing
	}

	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return false
	}

	if strings.Contains(s, ".") {
		parts := strings.SplitN(s, ".", 2)
		if len(parts) == 2 && len(parts[1]) > decimals {
			return false
		}
	}

	return true
}

// applyParsedValue attempts to parse text as a number, clamps it to the
// configured bounds, and updates config.Value (firing OnChange) if it
// changed. An empty string marks the field Empty and resets Value to 0.
// Partial input (e.g. "-", ".") is left uncommitted until it resolves to
// a full number.
func applyParsedValue(config *NumberInputConfig, text string) {
	if text == "" {
		if !config.Empty || config.Value != 0 {
			config.Empty = true
			config.Value = 0
			if config.OnChange != nil && config.ID != "" {
				config.OnChange(config.ID, 0)
			}
		}
		return
	}

	val, ok := parseNumber(text)
	if !ok {
		return
	}

	clamped := clampValue(val, config.HasMin, config.Min, config.HasMax, config.Max)
	if config.Empty || clamped != config.Value {
		config.Empty = false
		config.Value = clamped
		if config.OnChange != nil && config.ID != "" {
			config.OnChange(config.ID, clamped)
		}
	}
}

// ─── Core Rendering Function ────────────────────────────────────────────

func renderNumberInput(focused bool, config *NumberInputConfig) retui.Element {
	// Initial display text, used only to seed the hooks below on first mount.
	displayValue := ""
	if !config.Empty {
		displayValue = formatNumber(config.Value, config.Decimals)
	}

	// Track cursor position and the live editing buffer via retui's state.
	pos, setPos := retui.UseState(len([]rune(displayValue)))
	inputText, setInputText := retui.UseState(displayValue)

	// wasFocused lets us detect the exact frame focus is GAINED (an edge),
	// as opposed to every frame while focused is already true — without
	// this, select-all would re-fire every render and wipe out whatever
	// the user just typed.
	wasFocused, setWasFocused := retui.UseState(false)
	selected, setSelected := retui.UseState(false)

	// While not focused, the display always reflects the committed
	// config.Value (or the placeholder, if Empty). Guard the setters so we
	// don't trigger a state update - and potential re-render - every frame.
	if !focused {
		desired := ""
		if !config.Empty {
			desired = formatNumber(config.Value, config.Decimals)
		}
		if inputText != desired {
			inputText = desired
			setInputText(inputText)
		}
		desiredPos := len([]rune(inputText))
		if pos != desiredPos {
			pos = desiredPos
			setPos(pos)
		}
	}

	// Defensive clamp: config.Value/Empty may have changed externally
	// since pos was last set, so keep the cursor in bounds.
	if maxPos := len([]rune(inputText)); pos > maxPos {
		pos = maxPos
	}
	if pos < 0 {
		pos = 0
	}

	// Select-all-on-focus: fires exactly once, on the frame focus is
	// gained. Selecting nothing for an empty field is harmless (there's
	// nothing to select or replace).
	//
	// justGainedFocus is captured BEFORE wasFocused is updated below, and
	// is used further down to skip key handling entirely on this render.
	// Without that, retui.CurrentKey still holds whatever key caused focus
	// to move here (Tab, Down, etc.) — and since that key almost never
	// matches Left/Right/Home/End/Backspace/Delete, it would fall into the
	// "deselect" default case in the selected-handling switch below,
	// silently undoing the selection in the same render it was set.
	justGainedFocus := focused && !wasFocused

	if focused && !wasFocused {
		wasFocused = true
		setWasFocused(true)
		if config.SelectAllOnFocus {
			selected = true
			setSelected(true)
		}
	} else if !focused && wasFocused {
		wasFocused = false
		setWasFocused(false)
		selected = false
		setSelected(false)
	}

	// Trigger focus/blur events
	if focused && config.OnFocus != nil && config.ID != "" {
		config.OnFocus(config.ID)
	}
	if !focused && config.OnBlur != nil && config.ID != "" {
		config.OnBlur(config.ID)
	}

	// Handle keyboard input when focused. justGainedFocus is excluded so
	// the key that caused focus to move here isn't also treated as an
	// edit/deselect action inside this field (see comment above).
	if focused && !justGainedFocus {
		key := retui.CurrentKey

		if config.OnKeyPress != nil && config.ID != "" {
			if config.OnKeyPress(config.ID, key) {
				goto render
			}
		}

		// While the whole value is selected, keys behave like a web
		// input's selection: navigation keys collapse the selection to a
		// boundary; Backspace/Delete clear everything; any other key
		// (character input, Up/Down, Enter) deselects and — for character
		// input specifically — replaces the whole buffer instead of
		// inserting into it. The main switch below then runs against
		// whatever inputText/pos this leaves behind.
		if selected && key.Code != retui.KeyNone {
			switch key.Code {
			case retui.KeyLeft, retui.KeyHome:
				pos = 0
				setPos(0)
				selected = false
				setSelected(false)

			case retui.KeyRight, retui.KeyEnd:
				pos = len([]rune(inputText))
				setPos(pos)
				selected = false
				setSelected(false)

			case retui.KeyBackspace, retui.KeyDelete:
				inputText = ""
				setInputText("")
				pos = 0
				setPos(0)
				selected = false
				setSelected(false)
				applyParsedValue(config, "")
				goto render

			default:
				if key.Rune != 0 && key.Rune >= 32 && key.Rune <= 126 {
					// Replace-all: collapse to an empty buffer so the
					// character-insertion case below inserts into
					// nothing, instead of editing the old value in place.
					inputText = ""
					pos = 0
				}
				selected = false
				setSelected(false)
			}
		}

		runes := []rune(inputText)

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
				newRunes := append(append([]rune{}, runes[:pos-1]...), runes[pos:]...)
				newText := string(newRunes)
				inputText = newText
				setInputText(newText)
				setPos(pos - 1)
				applyParsedValue(config, newText)
			}

		case retui.KeyHome:
			setPos(0)

		case retui.KeyEnd:
			setPos(len(runes))

		case retui.KeyDelete:
			if pos < len(runes) {
				newRunes := append(append([]rune{}, runes[:pos]...), runes[pos+1:]...)
				newText := string(newRunes)
				inputText = newText
				setInputText(newText)
				applyParsedValue(config, newText)
			}

		case retui.KeyEnter:
			if inputText == "" {
				config.Empty = true
				config.Value = 0
				if config.OnSubmit != nil && config.ID != "" {
					config.OnSubmit(config.ID, 0)
				}
				break
			}

			val, ok := parseNumber(inputText)
			if !ok {
				return renderNumberError(config, "Enter a valid number")
			}

			clamped := clampValue(val, config.HasMin, config.Min, config.HasMax, config.Max)
			config.Empty = false
			config.Value = clamped

			// Reformat to the canonical representation (e.g. drop a
			// trailing ".", apply clamping) so the field reflects the
			// committed value.
			inputText = formatNumber(clamped, config.Decimals)
			setInputText(inputText)
			setPos(len([]rune(inputText)))

			if config.OnSubmit != nil && config.ID != "" {
				config.OnSubmit(config.ID, clamped)
			}

		case retui.KeyUp:
			if !config.ArrowStep {
				break
			}
			newVal := roundToDecimals(config.Value+config.Step, config.Decimals)
			if !config.HasMax || newVal <= config.Max {
				config.Empty = false
				config.Value = newVal
				inputText = formatNumber(newVal, config.Decimals)
				setInputText(inputText)
				setPos(len([]rune(inputText)))
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, newVal)
				}
			}

		case retui.KeyDown:
			if !config.ArrowStep {
				break
			}
			newVal := roundToDecimals(config.Value-config.Step, config.Decimals)
			if !config.HasMin || newVal >= config.Min {
				config.Empty = false
				config.Value = newVal
				inputText = formatNumber(newVal, config.Decimals)
				setInputText(inputText)
				setPos(len([]rune(inputText)))
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, newVal)
				}
			}

		default:
			// Insert printable character (only digits, one decimal point
			// when Decimals > 0, and a leading minus sign).
			if key.Rune != 0 && key.Rune >= 32 && key.Rune <= 126 {
				char := string(key.Rune)
				isValidChar := false

				switch char {
				case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
					isValidChar = true
				case ".":
					if config.Decimals > 0 && !strings.Contains(inputText, ".") {
						isValidChar = true
					}
				case "-":
					if pos == 0 && !strings.Contains(inputText, "-") {
						isValidChar = true
					}
				}

				if !isValidChar {
					break
				}

				newRunes := make([]rune, 0, len(runes)+1)
				newRunes = append(newRunes, runes[:pos]...)
				newRunes = append(newRunes, key.Rune)
				newRunes = append(newRunes, runes[pos:]...)
				newText := string(newRunes)

				if !isValidNumber(newText, config.Decimals) {
					break
				}

				inputText = newText
				setInputText(newText)
				setPos(pos + 1)
				applyParsedValue(config, newText)
			}
		}
	}

render:
	// Determine display value
	display := inputText
	if display == "" && config.Placeholder != "" {
		display = config.Placeholder
	}

	// Apply styles. The selected state gets its own distinct highlight so
	// it reads as "everything will be replaced" rather than a normal
	// cursor position.
	textStyle := config.Style
	if focused {
		if selected {
			textStyle = textStyle.
				Foreground(retui.Black).
				Background(retui.Cyan).
				Bold(true)
		} else {
			textStyle = textStyle.
				Foreground(retui.White).
				Background(retui.Blue).
				Bold(true)
		}
	} else {
		textStyle = textStyle.
			Foreground(retui.BrightBlack).
			Bold(true)
	}

	// Check validation against the committed value (not the in-progress text)
	isValid := true
	if config.HasMin && config.Value < config.Min {
		isValid = false
	}
	if config.HasMax && config.Value > config.Max {
		isValid = false
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

	// Add cursor — skipped while selected, since the whole-text highlight
	// above already communicates "this will be replaced" without a caret.
	cursorDisplay := display
	if focused && !selected {
		displayRunes := []rune(display)
		// If showing the placeholder (field is genuinely empty), park the
		// cursor at the start rather than at whatever pos was left at.
		cursorPos := pos
		if inputText == "" {
			cursorPos = 0
		}

		if cursorPos < len(displayRunes) {
			cursorDisplay = string(displayRunes[:cursorPos]) + "█" + string(displayRunes[cursorPos:])
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

func renderNumberError(config *NumberInputConfig, msg string) retui.Element {
	return retui.Box(
		retui.Props{Direction: retui.Row},
		retui.NewStyle(),
		retui.Text("⚠️ "+msg, retui.NewStyle().Foreground(retui.Red)),
	)
}

// ─── Example Usage ──────────────────────────────────────────────────────

func ExampleNumberUsage() retui.Element {
	// Integer input. Min(0) is enforced correctly, and gaining focus
	// selects the whole value so typing a digit immediately replaces it.
	ageInput := NumberInput().
		ID("age").
		Placeholder("Enter age").
		Width(20).
		Min(0).
		Max(150).
		Step(1).
		Decimals(0).
		Prefix("👤 ").
		Suffix(" years").
		OnChange(func(id string, value float64) {
			fmt.Printf("Age changed to: %v\n", value)
		}).
		OnSubmit(func(id string, value float64) {
			fmt.Printf("Age submitted: %v\n", value)
		})

	// Decimal input. Explicitly submitting 0.00 now displays correctly
	// instead of reverting to the placeholder.
	priceInput := NumberInput().
		ID("price").
		Placeholder("0.00").
		Width(25).
		Min(0).
		Max(9999.99).
		Step(0.01).
		Decimals(2).
		Prefix("$ ").
		Suffix(" USD").
		OnChange(func(id string, value float64) {
			fmt.Printf("Price changed to: %v\n", value)
		}).
		OnSubmit(func(id string, value float64) {
			fmt.Printf("Price submitted: %v\n", value)
		})

	// Return a box with both inputs
	return retui.Box(
		retui.Props{
			Direction: retui.Column,
		},
		retui.NewStyle(),
		ageInput.Render(),
		retui.Text(" ", retui.NewStyle()),
		priceInput.Render(),
	)
}
