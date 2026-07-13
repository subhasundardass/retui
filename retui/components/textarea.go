package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

type TextAreaOption func(*TextAreaConfig)

type TextAreaConfig struct {
	ID          string
	Value       string
	Placeholder string
	Width       int
	Height      int
	Style       retui.Style
	Prefix      string
	Suffix      string
	OnChange    func(id string, value string)
	OnKeyPress  func(id string, key retui.Key) bool
	OnFocus     func(id string)
	OnBlur      func(id string)
	OnSubmit    func(id string, value string)
}

func TextAreaWithID(id string) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.ID = id
	}
}

func TextAreaWithValue(value string) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.Value = value
	}
}

func TextAreaWithPlaceholder(text string) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.Placeholder = text
	}
}

func TextAreaWithWidth(width int) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.Width = width
	}
}

func TextAreaWithHeight(height int) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.Height = height
	}
}

func TextAreaWithStyle(style retui.Style) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.Style = style
	}
}

func TextAreaWithPrefix(prefix string) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.Prefix = prefix
	}
}

func TextAreaWithSuffix(suffix string) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.Suffix = suffix
	}
}

func TextAreaWithOnChange(fn func(id string, value string)) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.OnChange = fn
	}
}

func TextAreaWithOnKeyPress(fn func(id string, key retui.Key) bool) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.OnKeyPress = fn
	}
}

func TextAreaWithOnFocus(fn func(id string)) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.OnFocus = fn
	}
}

func TextAreaWithOnBlur(fn func(id string)) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.OnBlur = fn
	}
}

func TextAreaWithOnSubmit(fn func(id string, value string)) TextAreaOption {
	return func(c *TextAreaConfig) {
		c.OnSubmit = fn
	}
}

// ─── TextArea ──────────────────────────────────────────────────────────────

func TextArea(focused bool, opts ...TextAreaOption) retui.Element {
	config := &TextAreaConfig{
		ID:          "",
		Value:       "",
		Placeholder: "",
		Width:       40,
		Height:      5,
		Style:       retui.NewStyle(),
		Prefix:      "",
		Suffix:      "",
		OnChange:    nil,
		OnKeyPress:  nil,
		OnFocus:     nil,
		OnBlur:      nil,
		OnSubmit:    nil,
	}

	for _, opt := range opts {
		opt(config)
	}

	//State for cursor position
	pos, setPos := retui.UseState(len(config.Value))
	currentLine, setCurrentLine := retui.UseState(0) //FIXED: renamed to currentLine

	if focused && config.OnFocus != nil && config.ID != "" {
		config.OnFocus(config.ID)
	}

	if !focused && config.OnBlur != nil && config.ID != "" {
		config.OnBlur(config.ID)
	}

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
			if pos < len(config.Value) {
				setPos(pos + 1)
			}
		case retui.KeyUp:
			if currentLine > 0 {
				setCurrentLine(currentLine - 1)
			}
		case retui.KeyDown:
			if currentLine < config.Height-1 {
				setCurrentLine(currentLine + 1)
			}
		case retui.KeyBackspace:
			if pos > 0 && len(config.Value) > 0 {
				newValue := config.Value[:pos-1] + config.Value[pos:]
				config.Value = newValue
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, newValue)
				}
				setPos(pos - 1)
			}
		case retui.KeyHome:
			setPos(0)
		case retui.KeyEnd:
			setPos(len(config.Value))
		case retui.KeyDelete:
			if pos < len(config.Value) {
				newValue := config.Value[:pos] + config.Value[pos+1:]
				config.Value = newValue
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, newValue)
				}
			}
		case retui.KeyEnter:
			//Insert newline on Enter
			newValue := config.Value[:pos] + "\n" + config.Value[pos:]
			config.Value = newValue
			if config.OnChange != nil && config.ID != "" {
				config.OnChange(config.ID, newValue)
			}
			setPos(pos + 1)
			if config.OnSubmit != nil && config.ID != "" {
				config.OnSubmit(config.ID, newValue)
			}
		default:
			if key.Rune != 0 && key.Rune >= 32 {
				newValue := config.Value[:pos] + string(key.Rune) + config.Value[pos:]
				config.Value = newValue
				if config.OnChange != nil && config.ID != "" {
					config.OnChange(config.ID, newValue)
				}
				setPos(pos + 1)
			}
		}
	}

render:
	display := config.Value
	if display == "" && config.Placeholder != "" {
		display = config.Placeholder
	}

	//Wrap text to width
	lines := wrapTextArea(display, config.Width)

	//Pad to height
	for len(lines) < config.Height {
		lines = append(lines, "")
	}

	textStyle := config.Style
	if focused {
		textStyle = textStyle.Foreground(retui.White)
	} else {
		textStyle = textStyle.Foreground(retui.BrightBlack)
	}

	borderColor := retui.BrightBlack
	if focused {
		borderColor = retui.Cyan
	}

	bracketStyle := retui.NewStyle()
	if focused {
		bracketStyle = bracketStyle.Foreground(borderColor).Bold(true)
	} else {
		bracketStyle = bracketStyle.Foreground(retui.BrightBlack)
	}

	//Build content lines with cursor
	contentLines := []retui.Element{}
	for lineIdx, lineContent := range lines {
		displayLine := lineContent
		//FIXED: Use currentLine instead of line
		if focused && lineIdx == currentLine {
			runes := []rune(displayLine)
			if pos < len(runes) {
				displayLine = string(runes[:pos]) + "█" + string(runes[pos:])
			} else {
				displayLine = string(runes) + "█"
			}
		}
		contentLines = append(contentLines, retui.Text(displayLine, textStyle))
	}

	elements := []retui.Element{}

	if config.Prefix != "" {
		elements = append(elements, retui.Text(config.Prefix, bracketStyle))
	}

	//Text area content
	elements = append(elements, retui.Box(
		retui.Props{
			Direction: retui.Column,
			Width:     retui.Fixed(config.Width + 2),
			Height:    retui.Fixed(config.Height),
		},
		retui.NewStyle().
			Border(retui.Border{
				Top:    true,
				Right:  true,
				Bottom: true,
				Left:   true,
				Chars:  retui.BorderRounded,
				Color:  borderColor,
			}),
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

// wrapTextArea wraps text to max width
func wrapTextArea(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	// Split by newlines first
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
