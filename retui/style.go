package retui

import (
	"fmt"
	"strconv"
	"strings"
)

type Style struct {
	bold       bool
	italic     bool
	underline  bool
	background Color
	foreground Color
	border     Border
}

func NewStyle() Style {
	return Style{}
}

func (s Style) Bold(bold bool) Style {
	s.bold = bold
	return s
}

func (s Style) Foreground(color Color) Style {
	s.foreground = color
	return s
}

func (s Style) Background(color Color) Style {
	s.background = color
	return s
}

func (s Style) Italic(italic bool) Style {
	s.italic = italic
	return s
}

func (s Style) Underline(underline bool) Style {
	s.underline = underline
	return s
}

func (s Style) Border(b Border) Style {
	if b.Chars == (BorderChars{}) {
		b.Chars = BorderSharp
	}
	s.border = b
	return s
}

type BorderChars struct {
	Top, Bottom, Left, Right                   rune
	TopLeft, TopRight, BottomLeft, BottomRight rune
}

var (
	BorderSharp = BorderChars{
		Top: '─', Bottom: '─', Left: '│', Right: '│',
		TopLeft: '┌', TopRight: '┐', BottomLeft: '└', BottomRight: '┘',
	}
	BorderRounded = BorderChars{
		Top: '─', Bottom: '─', Left: '│', Right: '│',
		TopLeft: '╭', TopRight: '╮', BottomLeft: '╰', BottomRight: '╯',
	}
	BorderDouble = BorderChars{
		Top: '═', Bottom: '═', Left: '║', Right: '║',
		TopLeft: '╔', TopRight: '╗', BottomLeft: '╚', BottomRight: '╝',
	}
	BorderThick = BorderChars{
		Top: '━', Bottom: '━', Left: '┃', Right: '┃',
		TopLeft: '┏', TopRight: '┓', BottomLeft: '┗', BottomRight: '┛',
	}
)

type Border struct {
	Top, Right, Bottom, Left bool
	Chars                    BorderChars
	Color                    Color

	Title string
}

func (b Border) Any() bool {
	return b.Top || b.Right || b.Bottom || b.Left
}

type ColorType int

const (
	ColorNone ColorType = iota
	ColorANSI16
	ColorANSI256
	ColorRGB
)

type Color struct {
	Type    ColorType
	R, G, B uint8 //for RGB color
	Code    uint8 //for ANSI16 and ANSI256 color
}

func (c Color) RGB() (uint8, uint8, uint8) {
	return c.R, c.G, c.B
}

var (
	Black         = Color{Type: ColorANSI16, Code: 0}
	Red           = Color{Type: ColorANSI16, Code: 1}
	Green         = Color{Type: ColorANSI16, Code: 2}
	Yellow        = Color{Type: ColorANSI16, Code: 3}
	Blue          = Color{Type: ColorANSI16, Code: 4}
	Magenta       = Color{Type: ColorANSI16, Code: 5}
	Cyan          = Color{Type: ColorANSI16, Code: 6}
	White         = Color{Type: ColorANSI16, Code: 7}
	BrightBlack   = Color{Type: ColorANSI16, Code: 8}
	BrightRed     = Color{Type: ColorANSI16, Code: 9}
	BrightGreen   = Color{Type: ColorANSI16, Code: 10}
	BrightYellow  = Color{Type: ColorANSI16, Code: 11}
	BrightBlue    = Color{Type: ColorANSI16, Code: 12}
	BrightMagenta = Color{Type: ColorANSI16, Code: 13}
	BrightCyan    = Color{Type: ColorANSI16, Code: 14}
	BrightWhite   = Color{Type: ColorANSI16, Code: 15}
)

func Hex(color string) Color {
	color = strings.TrimPrefix(color, "#")

	r, _ := strconv.ParseUint(color[0:2], 16, 8)
	g, _ := strconv.ParseUint(color[2:4], 16, 8)
	b, _ := strconv.ParseUint(color[4:6], 16, 8)

	return Color{
		Type: ColorRGB,
		R:    uint8(r),
		G:    uint8(g),
		B:    uint8(b),
	}
}

func ANSI256(color uint8) Color {
	return Color{
		Type: ColorANSI256,
		Code: color,
	}
}

func (s Style) ANSIPrefix() string {
	var b strings.Builder

	// Always start with reset so previous styles don't bleed in
	b.WriteString("\033[0m")

	if s.bold {
		b.WriteString("\033[1m")
	}
	if s.italic {
		b.WriteString("\033[3m")
	}
	if s.underline {
		b.WriteString("\033[4m")
	}

	// Foreground color
	switch s.foreground.Type {
	case ColorANSI16:
		fmt.Fprintf(&b, "\033[%dm", 30+s.foreground.Code)
	case ColorANSI256:
		fmt.Fprintf(&b, "\033[38;5;%dm", s.foreground.Code)
	case ColorRGB:
		fmt.Fprintf(&b, "\033[38;2;%d;%d;%dm", s.foreground.R, s.foreground.G, s.foreground.B)
	}

	// Background color
	switch s.background.Type {
	case ColorANSI16:
		fmt.Fprintf(&b, "\033[%dm", 40+s.background.Code)
	case ColorANSI256:
		fmt.Fprintf(&b, "\033[48;5;%dm", s.background.Code)
	case ColorRGB:
		fmt.Fprintf(&b, "\033[48;2;%d;%d;%dm", s.background.R, s.background.G, s.background.B)
	}

	return b.String()
}

func (s Style) IsBold() bool {
	return s.bold
}

// mergeStyles returns the effective style for a child whose own Style is
// `child`, given an inherited `parent` style. A field is considered
// "unspecified" on the child when its zero value indicates "not set":
// foreground/background use ColorNone as the unset sentinel; bold/italic/
// underline are bools with no distinct "unset" — pick a rule and stick to it.
func mergeStyles(parent, child Style) Style {
	if child.foreground.Type == ColorNone {
		child.foreground = parent.foreground
	}

	if child.background.Type == ColorNone {
		child.background = parent.background
	}

	if parent.bold && !child.bold {
		child.bold = true
	}

	return child
}
