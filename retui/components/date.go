package components

import (
	"strconv"
	"strings"
	"time"

	"github.com/subhasundardass/retui/retui"
)

// ─── Date Input Configuration ────────────────────────────────────────────

// DateConfig holds the state and callbacks for a DateInputField.
//
// Value, Min, and Max are all expected to be formatted according to
// Format (e.g. "2024-01-15" for the default "YYYY-MM-DD"). Min/Max are
// only enforced once the entered date is complete and calendrically
// valid; partial input is never rejected while typing.
type DateConfig struct {
	ID          string
	Value       string
	Format      string
	Placeholder string
	Width       int
	Style       retui.Style
	Prefix      string
	Suffix      string
	Min         string // Minimum allowed date, formatted per Format ("" = no limit)
	Max         string // Maximum allowed date, formatted per Format ("" = no limit)
	OnChange    func(id string, value string)
	OnKeyPress  func(id string, key retui.Key) bool
	OnFocus     func(id string)
	OnBlur      func(id string)
	OnSubmit    func(id string, value string)
}

// DateInputField is a builder for a masked date-entry component.
type DateInputField struct {
	config  DateConfig
	focused bool
}

// ─── Builder Methods ──────────────────────────────────────────────────────

func DateInput() *DateInputField {
	return &DateInputField{
		config: DateConfig{
			ID:          "",
			Value:       "",
			Format:      "YYYY-MM-DD",
			Placeholder: "",
			Width:       20,
			Style:       retui.NewStyle(),
			Prefix:      "[ ",
			Suffix:      " ]",
			Min:         "",
			Max:         "",
			OnChange:    nil,
			OnKeyPress:  nil,
			OnFocus:     nil,
			OnBlur:      nil,
			OnSubmit:    nil,
		},
		focused: false,
	}
}

func (d *DateInputField) ID(v string) *DateInputField {
	d.config.ID = v
	return d
}

func (d *DateInputField) Value(v string) *DateInputField {
	d.config.Value = v
	return d
}

func (d *DateInputField) Format(v string) *DateInputField {
	d.config.Format = v
	return d
}

func (d *DateInputField) Placeholder(v string) *DateInputField {
	d.config.Placeholder = v
	return d
}

func (d *DateInputField) Width(w int) *DateInputField {
	d.config.Width = w
	return d
}

func (d *DateInputField) Prefix(v string) *DateInputField {
	d.config.Prefix = v
	return d
}

func (d *DateInputField) Suffix(v string) *DateInputField {
	d.config.Suffix = v
	return d
}

func (d *DateInputField) Style(s retui.Style) *DateInputField {
	d.config.Style = s
	return d
}

// Min sets the earliest allowed date, formatted per Format.
func (d *DateInputField) Min(v string) *DateInputField {
	d.config.Min = v
	return d
}

// Max sets the latest allowed date, formatted per Format.
func (d *DateInputField) Max(v string) *DateInputField {
	d.config.Max = v
	return d
}

func (d *DateInputField) Focused(v bool) *DateInputField {
	d.focused = v
	return d
}

func (d *DateInputField) OnChange(fn func(string, string)) *DateInputField {
	d.config.OnChange = fn
	return d
}

func (d *DateInputField) OnKeyPress(fn func(string, retui.Key) bool) *DateInputField {
	d.config.OnKeyPress = fn
	return d
}

func (d *DateInputField) OnFocus(fn func(string)) *DateInputField {
	d.config.OnFocus = fn
	return d
}

func (d *DateInputField) OnBlur(fn func(string)) *DateInputField {
	d.config.OnBlur = fn
	return d
}

func (d *DateInputField) OnSubmit(fn func(string, string)) *DateInputField {
	d.config.OnSubmit = fn
	return d
}

// ─── Render Method ──────────────────────────────────────────────────────

func (d *DateInputField) Render() retui.Element {
	return renderDateInput(d.focused, &d.config)
}

// ─── Helper Functions ──────────────────────────────────────────────────

// extractDigitsOnly extracts only digit characters from a string.
func extractDigitsOnly(value string) string {
	var result strings.Builder
	for _, ch := range value {
		if ch >= '0' && ch <= '9' {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// countDigitsInMask counts how many digit placeholders are in the mask.
func countDigitsInMask(mask string) int {
	count := 0
	for _, ch := range mask {
		if ch == 'Y' || ch == 'M' || ch == 'D' {
			count++
		}
	}
	return count
}

// applyMaskFormat applies the raw digits to the mask format, leaving any
// not-yet-entered slots as their literal placeholder letter (e.g. "MM").
func applyMaskFormat(rawDigits string, mask string) string {
	if rawDigits == "" {
		return ""
	}

	var result strings.Builder
	digitIdx := 0
	runes := []rune(rawDigits)

	for _, maskChar := range mask {
		if maskChar == 'Y' || maskChar == 'M' || maskChar == 'D' {
			if digitIdx < len(runes) {
				result.WriteRune(runes[digitIdx])
				digitIdx++
			} else {
				result.WriteRune(maskChar)
			}
		} else {
			result.WriteRune(maskChar)
		}
	}

	return result.String()
}

// calculateCursorDisplayPos calculates where to place the cursor in the
// formatted display, given a cursor position expressed as a digit index.
func calculateCursorDisplayPos(mask string, digitCursorPos int) int {
	displayPos := 0
	digitCount := 0

	for _, maskChar := range mask {
		if maskChar == 'Y' || maskChar == 'M' || maskChar == 'D' {
			if digitCount == digitCursorPos {
				return displayPos
			}
			digitCount++
		}
		displayPos++
	}

	return displayPos
}

// buildDateDisplayString builds the display string with proper formatting
// and, when focused, an inline cursor glyph.
func buildDateDisplayString(rawDigits string, mask string, placeholder string, focused bool, cursorPos int) string {
	if rawDigits == "" {
		display := placeholder
		if display == "" {
			display = mask
		}
		if focused {
			return "█" + display
		}
		return display
	}

	formatted := applyMaskFormat(rawDigits, mask)

	if focused {
		cursorDisplayPos := calculateCursorDisplayPos(mask, cursorPos)

		runes := []rune(formatted)
		if cursorDisplayPos < len(runes) {
			return string(runes[:cursorDisplayPos]) + "█" + string(runes[cursorDisplayPos:])
		}
		return string(runes) + "█"
	}

	return formatted
}

// dateFieldSpec locates a Y/M/D group within the raw-digit index space
// (not the mask's character index space).
type dateFieldSpec struct {
	start, length int
}

// maskFieldSpecs finds where the year, month, and day digit groups fall
// within the digit sequence a mask expects, so masks with any field
// ordering (e.g. "MM/DD/YYYY" as well as "YYYY-MM-DD") work correctly.
func maskFieldSpecs(mask string) (y, m, d dateFieldSpec) {
	digitIdx := 0
	var curChar rune
	length := 0
	curStart := 0

	flush := func(ch rune, start, length int) {
		switch ch {
		case 'Y':
			y = dateFieldSpec{start, length}
		case 'M':
			m = dateFieldSpec{start, length}
		case 'D':
			d = dateFieldSpec{start, length}
		}
	}

	for _, ch := range mask {
		if ch == 'Y' || ch == 'M' || ch == 'D' {
			if length == 0 {
				curChar = ch
				curStart = digitIdx
			} else if ch != curChar {
				flush(curChar, curStart, length)
				curChar = ch
				curStart = digitIdx
				length = 0
			}
			length++
			digitIdx++
		} else if length > 0 {
			flush(curChar, curStart, length)
			length = 0
		}
	}
	if length > 0 {
		flush(curChar, curStart, length)
	}
	return
}

// extractDateParts pulls the year/month/day integers out of rawDigits
// according to mask. complete is false if any field isn't fully entered.
func extractDateParts(rawDigits, mask string) (year, month, day int, complete bool) {
	ySpec, mSpec, dSpec := maskFieldSpecs(mask)
	runes := []rune(rawDigits)

	get := func(spec dateFieldSpec) (int, bool) {
		if spec.length == 0 || spec.start+spec.length > len(runes) {
			return 0, false
		}
		v, err := strconv.Atoi(string(runes[spec.start : spec.start+spec.length]))
		if err != nil {
			return 0, false
		}
		return v, true
	}

	var okY, okM, okD bool
	year, okY = get(ySpec)
	month, okM = get(mSpec)
	day, okD = get(dSpec)
	complete = okY && okM && okD
	return
}

// isValidCalendarDate reports whether year/month/day form a real date,
// correctly rejecting things like day 30 in February (including on leap
// years, since time.Date's normalization is used rather than a fixed
// days-in-month table).
func isValidCalendarDate(year, month, day int) bool {
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return false
	}
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return t.Year() == year && int(t.Month()) == month && t.Day() == day
}

// dateSortKey turns a validated date into a value that sorts/compares
// correctly regardless of field order in the mask.
func dateSortKey(year, month, day int) int {
	return year*10000 + month*100 + day
}

// parseFormattedDate extracts year/month/day from an already-formatted
// date string (such as config.Min or config.Max) using the same mask.
func parseFormattedDate(value, mask string) (year, month, day int, ok bool) {
	digits := extractDigitsOnly(value)
	y, m, d, complete := extractDateParts(digits, mask)
	if !complete || !isValidCalendarDate(y, m, d) {
		return 0, 0, 0, false
	}
	return y, m, d, true
}

// evaluateDate checks the currently entered digits against calendar
// validity and the configured Min/Max bounds. ok is true only when the
// date is complete, valid, and in range; msg explains why not otherwise.
// A date that's simply incomplete is not treated as an error (ok=false,
// msg=""), since the user may still be typing.
func evaluateDate(config *DateConfig, rawDigits, mask string) (ok bool, msg string) {
	year, month, day, complete := extractDateParts(rawDigits, mask)
	if !complete {
		return false, ""
	}
	if !isValidCalendarDate(year, month, day) {
		return false, "Enter a valid date"
	}

	entered := dateSortKey(year, month, day)

	if config.Min != "" {
		if my, mm, md, ok := parseFormattedDate(config.Min, mask); ok && entered < dateSortKey(my, mm, md) {
			return false, "Date is before the minimum allowed"
		}
	}
	if config.Max != "" {
		if My, Mm, Md, ok := parseFormattedDate(config.Max, mask); ok && entered > dateSortKey(My, Mm, Md) {
			return false, "Date is after the maximum allowed"
		}
	}

	return true, ""
}

// ─── Core Rendering Function ────────────────────────────────────────────

func renderDateInput(focused bool, config *DateConfig) retui.Element {
	mask := config.Format
	if mask == "" {
		mask = "YYYY-MM-DD"
	}

	// State for raw digits and cursor position (expressed as a digit index,
	// not a display-string index).
	rawDigits, setRawDigits := retui.UseState(extractDigitsOnly(config.Value))
	cursorPos, setCursorPos := retui.UseState(len([]rune(rawDigits)))

	maxDigits := countDigitsInMask(mask)

	// Sync from an external Value change. Compare against the *digits* of
	// config.Value, not the formatted string itself - rawDigits never
	// contains separators, so comparing against the formatted value would
	// almost always differ and force a resync (snapping the cursor to the
	// end) on every render, including plain cursor moves. This only fires
	// on a genuine external change, e.g. the parent resetting the field or
	// setting an initial value; a component-driven OnChange round-trip
	// produces the same digits and is a no-op here.
	externalDigits := extractDigitsOnly(config.Value)
	if externalDigits != rawDigits {
		rawDigits = externalDigits
		setRawDigits(rawDigits)
		cursorPos = len([]rune(rawDigits))
		setCursorPos(cursorPos)
	}

	// Clamp cursor - applied to the local variable too, so it takes effect
	// in this render rather than only the next one.
	if maxPos := len([]rune(rawDigits)); cursorPos > maxPos {
		cursorPos = maxPos
		setCursorPos(cursorPos)
	}
	if cursorPos < 0 {
		cursorPos = 0
		setCursorPos(cursorPos)
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
			if cursorPos > 0 {
				setCursorPos(cursorPos - 1)
			}

		case retui.KeyRight:
			if cursorPos < len([]rune(rawDigits)) {
				setCursorPos(cursorPos + 1)
			}

		case retui.KeyBackspace:
			if cursorPos > 0 && len(rawDigits) > 0 {
				runes := []rune(rawDigits)
				newRunes := append(append([]rune{}, runes[:cursorPos-1]...), runes[cursorPos:]...)
				newRaw := string(newRunes)
				setRawDigits(newRaw)
				setCursorPos(cursorPos - 1)
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, applyMaskFormat(newRaw, mask))
				}
			}

		case retui.KeyHome:
			setCursorPos(0)

		case retui.KeyEnd:
			setCursorPos(len([]rune(rawDigits)))

		case retui.KeyDelete:
			runes := []rune(rawDigits)
			if cursorPos < len(runes) {
				newRunes := append(append([]rune{}, runes[:cursorPos]...), runes[cursorPos+1:]...)
				newRaw := string(newRunes)
				setRawDigits(newRaw)
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, applyMaskFormat(newRaw, mask))
				}
			}

		case retui.KeyEnter:
			ok, msg := evaluateDate(config, rawDigits, mask)
			if !ok {
				if msg != "" {
					return renderDateError(config, msg)
				}
				// Incomplete date: nothing to submit yet, but not an error.
				break
			}
			if config.OnSubmit != nil && config.ID != "" {
				config.OnSubmit(config.ID, applyMaskFormat(rawDigits, mask))
			}

		default:
			// Only accept digits, and only while there's room in the mask.
			if key.Rune >= '0' && key.Rune <= '9' && len([]rune(rawDigits)) < maxDigits {
				runes := []rune(rawDigits)
				newRunes := make([]rune, 0, len(runes)+1)
				newRunes = append(newRunes, runes[:cursorPos]...)
				newRunes = append(newRunes, key.Rune)
				newRunes = append(newRunes, runes[cursorPos:]...)
				newRaw := string(newRunes)

				setRawDigits(newRaw)
				setCursorPos(cursorPos + 1)

				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, applyMaskFormat(newRaw, mask))
				}
			}
		}
	}

render:
	// Live validity, used for border/text coloring - does not block typing,
	// only flags a complete-but-invalid or out-of-range date.
	isValid := true
	if valid, msg := evaluateDate(config, rawDigits, mask); msg != "" {
		isValid = valid
	}

	// Build display with formatting
	display := buildDateDisplayString(rawDigits, mask, config.Placeholder, focused, cursorPos)

	// Pad to width using rune count for proper width calculation
	paddedDisplay := display
	displayLen := len([]rune(paddedDisplay))
	if displayLen < config.Width {
		padding := strings.Repeat(" ", config.Width-displayLen)
		paddedDisplay = paddedDisplay + padding
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
	if !isValid && !focused {
		textStyle = textStyle.Foreground(retui.Red)
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

	prefixStyle := retui.NewStyle().Foreground(retui.BrightBlack)
	suffixStyle := retui.NewStyle().Foreground(retui.BrightBlack)
	if focused {
		prefixStyle = retui.NewStyle().Foreground(borderColor).Bold(true)
		suffixStyle = retui.NewStyle().Foreground(borderColor).Bold(true)
	}

	// Build elements
	elements := []retui.Element{}

	if config.Prefix != "" {
		elements = append(elements, retui.Text(config.Prefix, prefixStyle))
	}

	elements = append(elements, retui.Text(paddedDisplay, textStyle))

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

// ─── Helper: Render Error State ────────────────────────────────────────

func renderDateError(config *DateConfig, msg string) retui.Element {
	return retui.Box(
		retui.Props{Direction: retui.Row},
		retui.NewStyle(),
		retui.Text(msg, retui.NewStyle().Foreground(retui.Red)),
	)
}

// ─── Example Usage ──────────────────────────────────────────────────────

// func ExampleDateUsage() retui.Element {
// 	// Date input with format and a range constraint.
// 	dateInput := DateInput().
// 		ID("birthday").
// 		Format("YYYY-MM-DD").
// 		Placeholder("Enter date").
// 		Width(25).
// 		Prefix("📅 ").
// 		Suffix(" ✓").
// 		Min("1900-01-01").
// 		Max("2026-12-31").
// 		OnChange(func(id string, value string) {
// 			fmt.Printf("Date changed to: %s\n", value)
// 		}).
// 		OnSubmit(func(id string, value string) {
// 			fmt.Printf("Date submitted: %s\n", value)
// 		})
//
// 	// Date input with a different field order.
// 	dateInput2 := DateInput().
// 		ID("meeting").
// 		Format("MM/DD/YYYY").
// 		Placeholder("MM/DD/YYYY").
// 		Width(20).
// 		Prefix("📆 ")
//
// 	return retui.Box(
// 		retui.Props{
// 			Direction: retui.Column,
// 		},
// 		retui.NewStyle(),
// 		dateInput.Render(),
// 		retui.Text(" ", retui.NewStyle()),
// 		dateInput2.Render(),
// 	)
// }
