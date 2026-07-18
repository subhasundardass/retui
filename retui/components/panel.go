package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

type panelBuilder struct {
	width      retui.Sizing
	children   []retui.Element
	header     retui.Element
	hasHeader  bool
	headerGap  int
	contentGap int
	style      *retui.Style
	fixedWidth int
	isFixed    bool
}

// Panel starts an empty panel builder. Width defaults to Grow(1).
func Panel() *panelBuilder {
	return &panelBuilder{
		width:      retui.Grow(1),
		headerGap:  0,
		contentGap: 0,
		style:      nil,
		fixedWidth: 0,
		isFixed:    false,
	}
}

// Width sets the layout sizing (retui.Fixed(n) or retui.Grow(n)).
func (p *panelBuilder) Width(w retui.Sizing) *panelBuilder {
	p.width = w
	return p
}

// FixedWidth sets a fixed width and stores the value for border calculation
func (p *panelBuilder) FixedWidth(width int) *panelBuilder {
	p.width = retui.Fixed(width)
	p.fixedWidth = width
	p.isFixed = true
	return p
}

// Style sets a custom style for the panel borders.
func (p *panelBuilder) Style(style retui.Style) *panelBuilder {
	p.style = &style
	return p
}

// Header sets a custom header Element.
func (p *panelBuilder) Header(el retui.Element) *panelBuilder {
	p.header = el
	p.hasHeader = true
	return p
}

// HeaderGap sets the gap between header and content (default: 0).
func (p *panelBuilder) HeaderGap(gap int) *panelBuilder {
	p.headerGap = gap
	return p
}

// ContentGap sets the gap between content rows (default: 0).
func (p *panelBuilder) ContentGap(gap int) *panelBuilder {
	p.contentGap = gap
	return p
}

// Children appends one or more children.
func (p *panelBuilder) Children(els ...retui.Element) *panelBuilder {
	for _, el := range els {
		if el.Type != 0 || len(el.Children) > 0 || el.Text != "" {
			p.children = append(p.children, el)
		}
	}
	return p
}

// Divider adds a horizontal divider line to the panel
// Example: ├─────────────────────────────────────────────────────────────┤
func (p *panelBuilder) Divider() *panelBuilder {
	borderStyle := p.getBorderStyle()

	var divider retui.Element
	if p.isFixed && p.fixedWidth > 0 {
		// For fixed width, create exact string
		innerWidth := p.fixedWidth - 2
		if innerWidth < 0 {
			innerWidth = 0
		}
		divider = retui.Text(strings.Repeat("─", innerWidth), borderStyle)
	} else {
		// For grow width, use Box approach
		const maxWidth = 2000
		divider = retui.Box(
			retui.Props{
				Direction: retui.Row,
				Width:     p.width,
			},
			retui.NewStyle(),
			retui.Text("", borderStyle),
			retui.Box(
				retui.Props{
					Width: retui.Grow(1),
				},
				retui.NewStyle(),
				retui.Text(strings.Repeat("─", maxWidth), borderStyle),
			),
			retui.Text("", borderStyle),
		)
	}

	p.children = append(p.children, divider)
	return p
}

// DividerWithText adds a horizontal divider with text in the middle
// Example: ├───── Section Title ─────┤
func (p *panelBuilder) DividerWithText(text string) *panelBuilder {
	borderStyle := p.getBorderStyle()

	var divider retui.Element
	if p.isFixed && p.fixedWidth > 0 {
		// For fixed width, create exact string with centered text
		innerWidth := p.fixedWidth - 2
		if innerWidth < 0 {
			innerWidth = 0
		}

		textLen := len(text)
		if textLen > innerWidth {
			text = text[:innerWidth]
			textLen = len(text)
		}

		totalFill := innerWidth - textLen
		leftFill := totalFill / 2
		rightFill := totalFill - leftFill

		divider = retui.Text(
			strings.Repeat("─", leftFill)+text+strings.Repeat("─", rightFill),
			borderStyle,
		)
	} else {
		// For grow width, use Box approach
		const maxWidth = 2000
		leftFill := strings.Repeat("─", maxWidth/2)
		rightFill := strings.Repeat("─", maxWidth/2)

		divider = retui.Box(
			retui.Props{
				Direction: retui.Row,
				Width:     p.width,
			},
			retui.NewStyle(),
			retui.Text("", borderStyle),
			retui.Box(
				retui.Props{
					Width: retui.Grow(1),
				},
				retui.NewStyle(),
				retui.Text(leftFill, borderStyle),
			),
			retui.Text(text, borderStyle),
			retui.Box(
				retui.Props{
					Width: retui.Grow(1),
				},
				retui.NewStyle(),
				retui.Text(rightFill, borderStyle),
			),
			retui.Text("", borderStyle),
		)
	}

	p.children = append(p.children, divider)
	return p
}

// Render builds the final retui.Element.
func (p *panelBuilder) Render() retui.Element {
	borderStyle := p.getBorderStyle()
	return p.buildPanel(borderStyle)
}

func (p *panelBuilder) getBorderStyle() retui.Style {
	if p.style != nil {
		return *p.style
	}
	return retui.NewStyle().Foreground(retui.Hex("#535353"))
}

// buildPanel assembles the complete panel
func (p *panelBuilder) buildPanel(borderStyle retui.Style) retui.Element {
	// For fixed width, calculate inner width once
	var innerWidth int
	if p.isFixed && p.fixedWidth > 0 {
		innerWidth = p.fixedWidth - 2 // Subtract left and right borders
		if innerWidth < 0 {
			innerWidth = 0
		}
	}

	// Build content rows with side borders
	contentRows := []retui.Element{}
	for i, child := range p.children {

		if i > 0 && p.contentGap > 0 {
			contentRows = append(contentRows, retui.Box(
				retui.Props{
					Height: retui.Fixed(p.contentGap),
				},
				retui.NewStyle(),
			))
		}

		// For fixed width, use Fixed(innerWidth) instead of Grow(1)
		var contentWidth retui.Sizing
		if p.isFixed {
			contentWidth = retui.Fixed(innerWidth)
		} else {
			contentWidth = retui.Grow(1)
		}

		contentRows = append(contentRows, retui.Box(
			retui.Props{
				Direction: retui.Row,
				Width:     p.width,
			},
			retui.NewStyle(),
			retui.Text("│", borderStyle),
			retui.Box(
				retui.Props{
					Width: contentWidth,
				},
				retui.NewStyle(),
				child,
			),
			retui.Text("│", borderStyle),
		))
	}

	// Build header row
	headerRow := p.buildHeaderRow(borderStyle, innerWidth)

	// Build the complete panel
	elements := []retui.Element{
		p.buildBorderLine("┌", "─", "┐", borderStyle, innerWidth),
		headerRow,
	}

	if p.headerGap > 0 {
		elements = append(elements, retui.Box(
			retui.Props{
				Height: retui.Fixed(p.headerGap),
			},
			retui.NewStyle(),
		))
	}

	elements = append(elements, p.buildBorderLine("├", "─", "┤", borderStyle, innerWidth))

	if len(contentRows) > 0 {
		elements = append(elements, retui.Box(
			retui.Props{
				Direction: retui.Column,
				Width:     p.width,
				Gap:       0,
			},
			retui.NewStyle(),
			contentRows...,
		))
	}

	elements = append(elements, p.buildBorderLine("└", "─", "┘", borderStyle, innerWidth))

	return retui.Box(
		retui.Props{
			Direction: retui.Column,
			Width:     p.width,
			Gap:       0,
		},
		retui.NewStyle(),
		elements...,
	)
}

// buildHeaderRow creates the header section
func (p *panelBuilder) buildHeaderRow(borderStyle retui.Style, innerWidth int) retui.Element {
	var headerInner retui.Element

	if p.hasHeader {
		headerInner = p.header
	} else {
		// For fixed width, use Fixed sizing for the empty header
		var contentWidth retui.Sizing
		if p.isFixed {
			contentWidth = retui.Fixed(innerWidth)
		} else {
			contentWidth = retui.Grow(1)
		}
		headerInner = retui.Box(
			retui.Props{
				Width:   contentWidth,
				Padding: [4]int{0, 1, 0, 1},
			},
			retui.NewStyle(),
		)
	}

	return retui.Box(
		retui.Props{
			Direction: retui.Row,
			Width:     p.width,
		},
		retui.NewStyle(),
		retui.Text("│", borderStyle),
		retui.Box(
			retui.Props{
				Width: retui.Grow(1),
			},
			retui.NewStyle(),
			headerInner,
		),
		retui.Text("│", borderStyle),
	)
}

// buildBorderLine creates a border line
func (p *panelBuilder) buildBorderLine(left, fill, right string, style retui.Style, innerWidth int) retui.Element {
	// If fixed width was set, create exact string
	if p.isFixed && p.fixedWidth > 0 {
		if innerWidth < 0 {
			innerWidth = 0
		}
		borderStr := left + strings.Repeat(fill, innerWidth) + right
		return retui.Text(borderStr, style)
	}

	// For Grow width, use the Box approach
	const maxWidth = 2000
	return retui.Box(
		retui.Props{
			Direction: retui.Row,
			Width:     p.width,
		},
		retui.NewStyle(),
		retui.Text(left, style),
		retui.Box(
			retui.Props{
				Width: retui.Grow(1),
			},
			retui.NewStyle(),
			retui.Text(strings.Repeat(fill, maxWidth), style),
		),
		retui.Text(right, style),
	)
}

// measureHeight returns how many lines an Element will render as.
func measureHeight(el retui.Element) int {
	if el.Type == 0 && len(el.Children) == 0 && el.Text == "" {
		return 1
	}

	switch el.Type {
	case retui.ElementText:
		if el.Text == "" {
			return 1
		}
		return strings.Count(el.Text, "\n") + 1
	case retui.ElementBox:
		return measureBoxHeight(el)
	default:
		return 1
	}
}

func measureBoxHeight(el retui.Element) int {
	pad := el.Layout.PaddingTop + el.Layout.PaddingBottom

	if len(el.Children) == 0 {
		return 1 + pad
	}

	if el.Layout.Direction == retui.Row {
		max := 0
		for _, c := range el.Children {
			if h := measureHeight(c); h > max {
				max = h
			}
		}
		return max + pad
	}

	total := el.Layout.Gap * (len(el.Children) - 1)
	for _, c := range el.Children {
		total += measureHeight(c)
	}
	return total + pad
}
