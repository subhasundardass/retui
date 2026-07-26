package retui

import (
	"fmt"
	"strconv"
	"strings"
)

type Title struct {
	Text       string
	Foreground Color
	Background Color
	Bold       bool
	Italic     bool
	Align      Alignment
	Underline  bool
}

type Style struct {
	bold      bool
	italic    bool
	underline bool

	foreground Color
	background Color

	border Border
	title  Title
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

func (s Style) Title(title Title) Style {
	s.title = title
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

	title Title
}

func BorderAll() Border {
	return Border{
		Top:    true,
		Right:  true,
		Bottom: true,
		Left:   true,
		Chars:  BorderSharp,
	}
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

// IsZero reports whether the color is unset (the zero value / ColorNone).
func (c Color) IsZero() bool {
	return c.Type == ColorNone
}

// Hex returns the "#rrggbb" hex representation of an RGB color. For
// non-RGB color types (ANSI16, ANSI256, or unset) it returns "".
func (c Color) Hex() string {
	if c.Type != ColorRGB {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

func clampUnit(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

func lerp8(a, b uint8, t float64) uint8 {
	return uint8(float64(a) + (float64(b)-float64(a))*t)
}

// Lighten blends an RGB color toward white by amount (0-1, clamped).
// Non-RGB colors are returned unchanged.
func (c Color) Lighten(amount float64) Color {
	if c.Type != ColorRGB {
		return c
	}
	t := clampUnit(amount)
	return Color{
		Type: ColorRGB,
		R:    lerp8(c.R, 255, t),
		G:    lerp8(c.G, 255, t),
		B:    lerp8(c.B, 255, t),
	}
}

// Darken blends an RGB color toward black by amount (0-1, clamped).
// Non-RGB colors are returned unchanged.
func (c Color) Darken(amount float64) Color {
	if c.Type != ColorRGB {
		return c
	}
	t := clampUnit(amount)
	return Color{
		Type: ColorRGB,
		R:    lerp8(c.R, 0, t),
		G:    lerp8(c.G, 0, t),
		B:    lerp8(c.B, 0, t),
	}
}

// Mix blends two RGB colors together; t=0 returns a, t=1 returns b.
// If either color is not RGB, it returns whichever one is RGB, or the
// zero Color if neither is.
func Mix(a, b Color, t float64) Color {
	if a.Type != ColorRGB && b.Type != ColorRGB {
		return Color{}
	}
	if a.Type != ColorRGB {
		return b
	}
	if b.Type != ColorRGB {
		return a
	}
	t = clampUnit(t)
	return Color{
		Type: ColorRGB,
		R:    lerp8(a.R, b.R, t),
		G:    lerp8(a.G, b.G, t),
		B:    lerp8(a.B, b.B, t),
	}
}

// Standard 16-color ANSI palette.
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

// Extended named colors, approximated from the 256-color palette. These
// give richer named options than the base 16 ANSI colors while still
// rendering correctly on terminals that only support 256 colors (unlike
// arbitrary Hex()/RGB() truecolor values).
var (
	Orange      = ANSI256(208)
	Purple      = ANSI256(129)
	Pink        = ANSI256(213)
	Teal        = ANSI256(30)
	Lime        = ANSI256(154)
	Gold        = ANSI256(220)
	Silver      = ANSI256(250)
	Navy        = ANSI256(17)
	Maroon      = ANSI256(52)
	Olive       = ANSI256(58)
	Indigo      = ANSI256(54)
	Turquoise   = ANSI256(80)
	Coral       = ANSI256(203)
	Salmon      = ANSI256(210)
	Violet      = ANSI256(177)
	Chocolate   = ANSI256(166)
	SkyBlue     = ANSI256(117)
	ForestGreen = ANSI256(28)
	Crimson     = ANSI256(161)
	Khaki       = ANSI256(222)
)

// NoColor explicitly represents "no color set". Equivalent to the zero
// Color, provided for readability at call sites.
var NoColor = Color{Type: ColorNone}

// Hex parses a hex color string such as "#ff8800" or "ff8800" (case
// insensitive) into a truecolor RGB Color. It returns the zero Color if
// the input isn't a valid 6-digit hex string.
func Hex(color string) Color {
	color = strings.TrimPrefix(color, "#")

	if len(color) != 6 {
		return Color{}
	}

	r, err1 := strconv.ParseUint(color[0:2], 16, 8)
	g, err2 := strconv.ParseUint(color[2:4], 16, 8)
	b, err3 := strconv.ParseUint(color[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return Color{}
	}

	return Color{
		Type: ColorRGB,
		R:    uint8(r),
		G:    uint8(g),
		B:    uint8(b),
	}
}

// RGB creates a truecolor color from red, green, and blue components
// (0-255 each).
func RGB(r, g, b uint8) Color {
	return Color{Type: ColorRGB, R: r, G: g, B: b}
}

// ANSI256 creates a color from the 256-color palette (0-255).
func ANSI256(color uint8) Color {
	return Color{
		Type: ColorANSI256,
		Code: color,
	}
}

// Gray returns a grayscale color from the 256-color palette's grayscale
// ramp. n ranges 0 (near-black) to 23 (near-white), corresponding to
// ANSI256 codes 232-255; out-of-range values are clamped.
func Gray(n int) Color {
	if n < 0 {
		n = 0
	}
	if n > 23 {
		n = 23
	}
	return Color{Type: ColorANSI256, Code: uint8(232 + n)}
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
