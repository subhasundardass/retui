package components

import (
	"strconv"
	"strings"

	"github.com/subhasundardass/retui/retui"
)

type TextAreaConfig struct {
	ID          string
	Value       string
	Placeholder string
	Width       int
	Height      int
	Style       retui.Style
	Prefix      string
	Suffix      string
	MinLength   int
	MaxLength   int
	Disabled    bool
	ReadOnly    bool
	Hidden      bool
	OnChange    func(id string, value string)
	OnKeyPress  func(id string, key retui.Key) bool
	OnFocus     func(id string)
	OnBlur      func(id string)
	OnSubmit    func(id string, value string)
}

type TextAreaField struct {
	config  TextAreaConfig
	focused bool
}

// ─── Builder Methods ──────────────────────────────────────────────────────

func TextArea() *TextAreaField {
	return &TextAreaField{
		config: TextAreaConfig{
			ID:          "",
			Value:       "",
			Placeholder: "",
			Width:       40,
			Height:      5,
			Style:       retui.NewStyle(),
			Prefix:      "",
			Suffix:      "",
			MinLength:   0,
			MaxLength:   0,
			Disabled:    false,
			ReadOnly:    false,
			Hidden:      false,
			OnChange:    nil,
			OnKeyPress:  nil,
			OnFocus:     nil,
			OnBlur:      nil,
			OnSubmit:    nil,
		},
		focused: false,
	}
}

func (t *TextAreaField) ID(v string) *TextAreaField {
	t.config.ID = v
	return t
}

func (t *TextAreaField) Value(v string) *TextAreaField {
	t.config.Value = v
	return t
}

func (t *TextAreaField) Placeholder(v string) *TextAreaField {
	t.config.Placeholder = v
	return t
}

func (t *TextAreaField) Width(w int) *TextAreaField {
	t.config.Width = w
	return t
}

func (t *TextAreaField) Height(h int) *TextAreaField {
	t.config.Height = h
	return t
}

func (t *TextAreaField) Prefix(v string) *TextAreaField {
	t.config.Prefix = v
	return t
}

func (t *TextAreaField) Suffix(v string) *TextAreaField {
	t.config.Suffix = v
	return t
}

func (t *TextAreaField) Style(s retui.Style) *TextAreaField {
	t.config.Style = s
	return t
}

func (t *TextAreaField) MinLength(v int) *TextAreaField {
	t.config.MinLength = v
	return t
}

func (t *TextAreaField) MaxLength(v int) *TextAreaField {
	t.config.MaxLength = v
	return t
}

// --
func (i *TextAreaField) Disable(v bool) *TextAreaField {
	i.config.Disabled = v
	return i
}
func (i *TextAreaField) ReadOnly(v bool) *TextAreaField {
	i.config.ReadOnly = v
	return i
}
func (i *TextAreaField) Hidden(v bool) *TextAreaField {
	i.config.Hidden = v
	return i
}

//--

func (t *TextAreaField) Focused(v bool) *TextAreaField {
	t.focused = v
	return t
}

func (t *TextAreaField) OnChange(fn func(string, string)) *TextAreaField {
	t.config.OnChange = fn
	return t
}

func (t *TextAreaField) OnKeyPress(fn func(string, retui.Key) bool) *TextAreaField {
	t.config.OnKeyPress = fn
	return t
}

func (t *TextAreaField) OnFocus(fn func(string)) *TextAreaField {
	t.config.OnFocus = fn
	return t
}

func (t *TextAreaField) OnBlur(fn func(string)) *TextAreaField {
	t.config.OnBlur = fn
	return t
}

func (t *TextAreaField) OnSubmit(fn func(string, string)) *TextAreaField {
	t.config.OnSubmit = fn
	return t
}

// ─── Render Method ──────────────────────────────────────────────────────

func (t *TextAreaField) Render() retui.Element {
	return renderTextArea(t.focused, &t.config)
}

// ─── Core Rendering Function ────────────────────────────────────────────

func renderTextArea(focused bool, config *TextAreaConfig) retui.Element {

	// Hidden
	if config.Hidden {
		return retui.Element{}
	}

	runes := []rune(config.Value)

	// Single absolute cursor position, same as TextInput — no separate
	// "currentLine" state, so cursor and line navigation can never drift
	// out of sync with each other.
	pos, setPos := retui.UseState(len(runes))

	if pos > len(runes) {
		pos = len(runes)
		setPos(pos)
	}
	if pos < 0 {
		pos = 0
		setPos(pos)
	}

	if !config.Disabled && focused && config.OnFocus != nil && config.ID != "" {
		config.OnFocus(config.ID)
	}

	if !config.Disabled && !focused && config.OnBlur != nil && config.ID != "" {
		config.OnBlur(config.ID)
	}

	if focused && !config.Disabled {
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

		case retui.KeyUp:
			if newPos, ok := movePosVertically(runes, pos, -1); ok {
				setPos(newPos)
			}

		case retui.KeyDown:
			if newPos, ok := movePosVertically(runes, pos, 1); ok {
				setPos(newPos)
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
			setPos(lineStart(runes, pos))

		case retui.KeyEnd:
			setPos(lineEnd(runes, pos))

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
				return renderTextAreaError(config, "Minimum length is "+strconv.Itoa(config.MinLength)+" characters")
			}
			newRunes := append(append(append([]rune{}, runes[:pos]...), '\n'), runes[pos:]...)
			newValue := string(newRunes)
			config.Value = newValue
			if config.OnChange != nil && config.ID != "" {
				config.OnChange(config.ID, newValue)
			}
			setPos(pos + 1)
			if config.OnSubmit != nil && config.ID != "" {
				config.OnSubmit(config.ID, newValue)
			}

		default:
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
	runes = []rune(config.Value)
	if pos > len(runes) {
		pos = len(runes)
	}

	isValid := true
	if config.MinLength > 0 && len(runes) < config.MinLength {
		isValid = false
	}

	display := config.Value
	if display == "" && config.Placeholder != "" && !focused {
		display = config.Placeholder
	}

	// Same style rules as TextInput: solid bg block, not a border.
	textStyle := config.Style
	if focused {
		textStyle = textStyle.
			Foreground(retui.White).Bold(true).
			Background(retui.Gray(2)).
			Bold(true)
	} else {
		textStyle = textStyle.
			Foreground(retui.White).Bold(true)
	}

	bracketStyle := retui.NewStyle()
	if focused {
		borderColor := retui.Cyan
		if !isValid {
			borderColor = retui.Red
		}
		bracketStyle = bracketStyle.Foreground(borderColor).Bold(true)
	} else {
		bracketStyle = bracketStyle.Foreground(retui.BrightBlack)
	}

	// Wrap to width, then place cursor by converting absolute pos -> (line, col)
	lines := wrapTextArea(display, config.Width)
	for len(lines) < config.Height {
		lines = append(lines, "")
	}

	cursorLine, cursorCol := -1, -1
	if focused {
		cursorLine, cursorCol = posToLineCol(display, config.Width, pos)
	}

	textStyleForDisplay := textStyle
	if !isValid && !focused {
		textStyleForDisplay = textStyleForDisplay.Foreground(retui.Red)
	}

	contentLines := make([]retui.Element, 0, len(lines))
	for lineIdx, lineContent := range lines {
		lineDisplay := lineContent

		if focused && lineIdx == cursorLine {
			lr := []rune(lineDisplay)
			if cursorCol < len(lr) {
				lineDisplay = string(lr[:cursorCol]) + "█" + string(lr[cursorCol:])
			} else {
				lineDisplay = string(lr) + "█"
			}
		}

		// Pad each line to width, same as TextInput pads its single line,
		// so the background color fills the whole field consistently.
		lineLen := len([]rune(lineDisplay))
		if lineLen < config.Width {
			lineDisplay += strings.Repeat(" ", config.Width-lineLen)
		}

		contentLines = append(contentLines, retui.Text(lineDisplay, textStyleForDisplay))
	}

	elements := []retui.Element{}

	if config.Prefix != "" {
		elements = append(elements, retui.Text(config.Prefix, bracketStyle))
	}

	elements = append(elements, retui.Box(
		retui.Props{
			Direction: retui.Column,
			Width:     retui.Fixed(config.Width),
			Height:    retui.Fixed(config.Height),
		},
		retui.NewStyle(),
		contentLines...,
	))

	if config.Suffix != "" {
		elements = append(elements, retui.Text(config.Suffix, bracketStyle))
	}

	return retui.Box(
		retui.Props{
			Direction: retui.Row,
		},
		retui.NewStyle(),
		elements...,
	)
}

// ─── Helpers ──────────────────────────────────────────────────────────────

func renderTextAreaError(config *TextAreaConfig, msg string) retui.Element {
	return retui.Box(
		retui.Props{Direction: retui.Row},
		retui.NewStyle(),
		retui.Text(msg, retui.NewStyle().Foreground(retui.Red)),
	)
}

// lineStart returns the absolute index of the start of the line containing pos.
func lineStart(runes []rune, pos int) int {
	for i := pos - 1; i >= 0; i-- {
		if runes[i] == '\n' {
			return i + 1
		}
	}
	return 0
}

// lineEnd returns the absolute index of the end of the line containing pos.
func lineEnd(runes []rune, pos int) int {
	for i := pos; i < len(runes); i++ {
		if runes[i] == '\n' {
			return i
		}
	}
	return len(runes)
}

// movePosVertically moves pos up (dir=-1) or down (dir=1) one raw line
// (split on '\n' only — matches lineStart/lineEnd, not visual wrapping).
func movePosVertically(runes []rune, pos int, dir int) (int, bool) {
	start := lineStart(runes, pos)
	col := pos - start

	if dir < 0 {
		if start == 0 {
			return 0, false
		}
		prevEnd := start - 1 // the '\n' before this line
		prevStart := lineStart(runes, prevEnd)
		prevLen := prevEnd - prevStart
		if col > prevLen {
			col = prevLen
		}
		return prevStart + col, true
	}

	end := lineEnd(runes, pos)
	if end >= len(runes) {
		return pos, false
	}
	nextStart := end + 1
	nextEnd := lineEnd(runes, nextStart)
	nextLen := nextEnd - nextStart
	if col > nextLen {
		col = nextLen
	}
	return nextStart + col, true
}

// posToLineCol converts an absolute rune index in raw text into a
// (wrapped-line, column) position within the wrapped display, so the
// cursor lands in the right spot after word-wrapping.
func posToLineCol(text string, width int, pos int) (int, int) {
	wrapped := wrapTextArea(text[:min(pos, len(text))], width)
	if len(wrapped) == 0 {
		return 0, 0
	}
	lastLine := wrapped[len(wrapped)-1]
	return len(wrapped) - 1, len([]rune(lastLine))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// wrapTextArea wraps text to max width
func wrapTextArea(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	paragraphs := strings.Split(text, "\n")
	var allLines []string

	for _, para := range paragraphs {
		if para == "" {
			allLines = append(allLines, "")
			continue
		}

		words := strings.Fields(para)
		if len(words) == 0 {
			allLines = append(allLines, "")
			continue
		}

		var line strings.Builder
		lineWidth := 0

		for _, word := range words {
			wordWidth := len([]rune(word))
			if lineWidth == 0 {
				line.WriteString(word)
				lineWidth = wordWidth
			} else if lineWidth+1+wordWidth <= maxWidth {
				line.WriteByte(' ')
				line.WriteString(word)
				lineWidth += 1 + wordWidth
			} else {
				allLines = append(allLines, line.String())
				line.Reset()
				line.WriteString(word)
				lineWidth = wordWidth
			}
		}

		if line.Len() > 0 {
			allLines = append(allLines, line.String())
		}
	}

	return allLines
}
