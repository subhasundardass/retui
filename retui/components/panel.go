package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

type panelBuilder struct {
	width         retui.Sizing
	height        retui.Sizing
	children      []retui.Element
	header        retui.Element
	hasHeader     bool
	headerGap     int
	contentGap    int
	style         *retui.Style
	fixedWidth    int
	isFixed       bool
	fixedHeight   int
	isFixedHeight bool
}

// Panel starts an empty panel builder. Width defaults to Grow(1).
func Panel() *panelBuilder {
	return &panelBuilder{
		width:      retui.Grow(1),
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

// Height sets the layout sizing (retui.Fixed(n) or retui.Grow(n)) for the
// panel's total height. Most callers want FixedHeight instead — a plain
// Height doesn't tell buildPanel how to split the total between chrome
// (borders/header) and content, so a Grow/Fixed here without FixedHeight
// won't pad the interior; use it only if you're managing that split
// yourself.
func (p *panelBuilder) Height(h retui.Sizing) *panelBuilder {
	p.height = h
	return p
}

// FixedHeight sets a fixed *total* height for the panel. Top border,
// header row, header gap, the divider under the header, content, and the
// bottom border all have to fit inside this many lines. If content is
// shorter than what's available, buildPanel pads the remainder with a
// bordered (but empty) filler row, so the side walls and bottom border
// land in the right place instead of stopping wherever the content
// happened to end.
func (p *panelBuilder) FixedHeight(height int) *panelBuilder {
	p.height = retui.Fixed(height)
	p.fixedHeight = height
	p.isFixedHeight = true
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
	return retui.NewStyle().Foreground(retui.Gray(1))
}

// buildPanel assembles the complete panel
func (p *panelBuilder) buildPanel(borderStyle retui.Style) retui.Element {
	var innerWidth int
	if p.isFixed && p.fixedWidth > 0 {
		innerWidth = p.fixedWidth - 2
		if innerWidth < 0 {
			innerWidth = 0
		}
	}

	headerRow, headerRowHeight := p.buildHeaderRow(borderStyle, innerWidth)

	// Build content rows with side borders, tracking how many lines of
	// actual content we've accounted for so a FixedHeight panel knows how
	// much filler (if any) is needed to reach the requested total height.
	contentRows := []retui.Element{}
	actualContentHeight := 0

	for i, child := range p.children {

		if i > 0 && p.contentGap > 0 {
			contentRows = append(contentRows, retui.Box(
				retui.Props{Height: retui.Fixed(p.contentGap)},
				retui.NewStyle(),
			))
			actualContentHeight += p.contentGap
		}

		var contentWidth retui.Sizing
		if p.isFixed {
			contentWidth = retui.Fixed(innerWidth)
		} else {
			contentWidth = retui.Grow(1)
		}

		rowHeight := measureHeight(child)
		actualContentHeight += rowHeight

		contentRows = append(contentRows, retui.Box(
			retui.Props{
				Direction: retui.Row,
				Width:     p.width,
			},
			retui.NewStyle(),
			buildVerticalBorder("│", borderStyle, rowHeight),
			retui.Box(
				retui.Props{Width: contentWidth},
				retui.NewStyle(),
				child,
			),
			buildVerticalBorder("│", borderStyle, rowHeight),
		))
	}

	elements := []retui.Element{
		p.buildBorderLine("┌", "─", "┐", borderStyle, innerWidth),
		headerRow,
	}

	chromeHeight := 1 /* top border */ + headerRowHeight + 1 /* ├─┤ */ + 1 /* bottom border */

	if p.headerGap > 0 {
		elements = append(elements, retui.Box(
			retui.Props{Height: retui.Fixed(p.headerGap)},
			retui.NewStyle(),
		))
		chromeHeight += p.headerGap
	}

	elements = append(elements, p.buildBorderLine("├", "─", "┤", borderStyle, innerWidth))

	// FixedHeight: work out how much vertical space is left for content
	// after all the chrome, and pad with a bordered-but-empty filler row
	// so the walls reach the bottom border even when content is shorter
	// than requested.
	if p.isFixedHeight {
		remainingForContent := p.fixedHeight - chromeHeight
		if remainingForContent < 0 {
			remainingForContent = 0
		}

		if leftover := remainingForContent - actualContentHeight; leftover > 0 {
			var fillerWidth retui.Sizing
			if p.isFixed {
				fillerWidth = retui.Fixed(innerWidth)
			} else {
				fillerWidth = retui.Grow(1)
			}

			contentRows = append(contentRows, retui.Box(
				retui.Props{
					Direction: retui.Row,
					Width:     p.width,
				},
				retui.NewStyle(),
				buildVerticalBorder("│", borderStyle, leftover),
				retui.Box(
					retui.Props{Width: fillerWidth, Height: retui.Fixed(leftover)},
					retui.NewStyle(),
				),
				buildVerticalBorder("│", borderStyle, leftover),
			))
			actualContentHeight += leftover
		}
		// If content already exceeds remainingForContent, it overflows —
		// this file has no clipping, so the panel renders taller than
		// FixedHeight rather than cutting content off.
	}

	if len(contentRows) > 0 {
		wrapperProps := retui.Props{
			Direction: retui.Column,
			Width:     p.width,
			Gap:       0,
		}
		if p.isFixedHeight {
			wrapperProps.Height = retui.Fixed(actualContentHeight)
		}
		elements = append(elements, retui.Box(wrapperProps, retui.NewStyle(), contentRows...))
	}

	elements = append(elements, p.buildBorderLine("└", "─", "┘", borderStyle, innerWidth))

	outerProps := retui.Props{
		Direction: retui.Column,
		Width:     p.width,
		Gap:       0,
	}
	if p.isFixedHeight {
		outerProps.Height = p.height
	}

	return retui.Box(outerProps, retui.NewStyle(), elements...)
}

// buildHeaderRow creates the header section. It also returns the header's
// measured height, which buildPanel needs to budget space when FixedHeight
// is set.
func (p *panelBuilder) buildHeaderRow(borderStyle retui.Style, innerWidth int) (retui.Element, int) {
	var headerInner retui.Element

	if p.hasHeader {
		headerInner = p.header
	} else {
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

	rowHeight := measureHeight(headerInner)

	row := retui.Box(
		retui.Props{
			Direction: retui.Row,
			Width:     p.width,
		},
		retui.NewStyle(),
		buildVerticalBorder("│", borderStyle, rowHeight),
		retui.Box(
			retui.Props{
				Width: retui.Grow(1),
			},
			retui.NewStyle(),
			headerInner,
		),
		buildVerticalBorder("│", borderStyle, rowHeight),
	)

	return row, rowHeight
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
	// An explicit fixed height is a hard constraint the real layout engine
	// already enforces on this box — it is NOT derived from children.
	// Ignoring it (as the code below does) is what let a Height: Fixed(10)
	// box measure as 1 when its child content was shorter than 10 lines.
	// if el.Layout.Height.Type == retui.SizingFixed {
	// 	return el.Layout.Height.Value
	// }

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

// buildVerticalBorder draws a side-border character repeated down `height`
// lines. A single Text node's rect can be stretched to any height by
// AlignStretch, but paint() only draws text on its own first line — so a
// multi-line row needs `height` separate Text nodes stacked in a Column,
// one per row, not one Text node asked to be tall.
func buildVerticalBorder(ch string, style retui.Style, height int) retui.Element {
	if height <= 1 {
		return retui.Text(ch, style)
	}
	lines := make([]retui.Element, height)
	for i := range lines {
		lines[i] = retui.Text(ch, style)
	}
	return retui.Box(
		retui.Props{Direction: retui.Column, Gap: 0},
		retui.NewStyle(),
		lines...,
	)
}
