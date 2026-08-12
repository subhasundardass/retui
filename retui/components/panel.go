// Package components provides reusable terminal UI components for RetUI.
//
// Panel is a bordered box component with an optional header, dividers,
// fixed or growing dimensions, and automatic height padding.
package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

// ─────────────────────────────────────────────────────────────────────────────
// panelBuilder
// ─────────────────────────────────────────────────────────────────────────────

// panelBuilder is the internal state accumulated by the fluent Panel API.
// Construct it with Panel() and call Render() to produce a retui.Element.
//
// fixedW and fixedH replace the original four fields
// (fixedWidth, fixedHeight, isFixed, isFixedHeight):
//
//	nil     → sizing is Grow(1); no exact column/row count is available,
//	          so border strings use a Box+Grow+2000-char fallback.
//	non-nil → FixedWidth/FixedHeight was called; *fixedW or *fixedH is
//	          the declared size and is used for exact string construction
//	          and height-padding calculations.
type panelBuilder struct {
	// Layout sizing passed to the outer retui.Box.
	width  retui.Sizing
	height retui.Sizing

	// fixedW is set when FixedWidth(n) is called; *fixedW == n.
	// nil means width is Grow — no pixel count available.
	fixedW *int

	// fixedH is set when FixedHeight(n) is called; *fixedH == n.
	// nil means height is unspecified or Grow.
	fixedH *int

	// Content children added via Children() or Divider()/DividerWithText().
	children []retui.Element

	// Optional header element set via Header().
	header    retui.Element
	hasHeader bool

	// headerGap adds blank rows between the header and the ├─┤ divider.
	headerGap int

	// contentGap adds blank rows between consecutive children.
	contentGap int

	// style overrides the default border/divider style when non-nil.
	style *retui.Style
}

// ─────────────────────────────────────────────────────────────────────────────
// Constructor
// ─────────────────────────────────────────────────────────────────────────────

// Panel creates a new panel builder with defaults:
//   - Width:  Grow(1) — fills available horizontal space
//   - Height: unset   — sized by content
//   - Style:  Gray(1) border
//   - No header, no children, no gaps
//
// Chain builder methods and call Render() to produce a retui.Element.
//
// Example — minimal panel:
//
//	Panel().
//	    FixedWidth(40).
//	    Children(retui.Text("Hello", retui.NewStyle())).
//	    Render()
//
// Example — full-featured panel:
//
//	Panel().
//	    FixedWidth(60).
//	    FixedHeight(20).
//	    Header(myHeaderElement).
//	    ContentGap(1).
//	    Children(row1, row2, row3).
//	    Divider().
//	    Children(row4).
//	    Render()
func Panel() *panelBuilder {
	return &panelBuilder{
		width: retui.Grow(1),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Builder methods
// ─────────────────────────────────────────────────────────────────────────────

// Width sets the layout sizing for the panel's outer box.
// Use retui.Fixed(n) or retui.Grow(n).
//
// Prefer FixedWidth when you need exact border characters — Width alone
// does not give the panel a pixel count to compute inner border widths.
//
// Example:
//
//	Panel().Width(retui.Grow(1))
func (p *panelBuilder) Width(w retui.Sizing) *panelBuilder {
	p.width = w
	return p
}

// FixedWidth sets the panel to exactly width columns.
// Stores the raw integer so border lines and dividers can be constructed
// as exact strings (left-border + inner-fill + right-border) rather than
// using the Box+Grow fallback.
//
// Example:
//
//	Panel().FixedWidth(60)  // total width including borders
func (p *panelBuilder) FixedWidth(width int) *panelBuilder {
	p.width = retui.Fixed(width)
	p.fixedW = &width
	return p
}

// Height sets the layout sizing for the panel's outer box.
// Use FixedHeight instead when you want the panel to pad its content to
// fill a specific number of rows — Height alone does not trigger padding.
//
// Example:
//
//	Panel().Height(retui.Grow(1))
func (p *panelBuilder) Height(h retui.Sizing) *panelBuilder {
	p.height = h
	return p
}

// FixedHeight sets the panel to exactly height rows.
//
// The panel accounts for all chrome rows (top border, header, header gap,
// ├─┤ divider, bottom border) and pads the content area with blank bordered
// rows so the side walls and bottom border land at exactly the right line
// even when content is shorter than the declared height.
//
// If content is TALLER than the declared height the panel overflows rather
// than clipping — this component does not implement scrolling.
//
// Example:
//
//	Panel().FixedHeight(24)  // total height including borders and header
func (p *panelBuilder) FixedHeight(height int) *panelBuilder {
	p.height = retui.Fixed(height)
	p.fixedH = &height
	return p
}

// Style sets a custom retui.Style for all border and divider characters.
// Defaults to retui.NewStyle().Foreground(retui.Gray(1)) when not set.
//
// Example:
//
//	Panel().Style(retui.NewStyle().Foreground(retui.Cyan))
func (p *panelBuilder) Style(style retui.Style) *panelBuilder {
	p.style = &style
	return p
}

// Header sets a custom retui.Element rendered inside the top border row.
// When Header is not called the header row is an empty padded box.
//
// Example:
//
//	Panel().Header(
//	    retui.Text("Ledger List", retui.NewStyle().Bold(true)),
//	)
func (p *panelBuilder) Header(el retui.Element) *panelBuilder {
	p.header = el
	p.hasHeader = true
	return p
}

// HeaderGap inserts n blank rows between the header and the ├─┤ divider.
// Use to add visual breathing room below the header.
//
// Example:
//
//	Panel().Header(titleEl).HeaderGap(1)
func (p *panelBuilder) HeaderGap(n int) *panelBuilder {
	p.headerGap = n
	return p
}

// ContentGap sets the number of blank rows inserted between consecutive
// children. Defaults to 0.
//
// Example:
//
//	Panel().ContentGap(1).Children(row1, row2, row3)
func (p *panelBuilder) ContentGap(gap int) *panelBuilder {
	p.contentGap = gap
	return p
}

// Children appends one or more content elements to the panel body.
// Empty elements (zero-value retui.Element with no text and no children)
// are silently discarded to avoid invisible blank rows from nil-guard
// patterns in calling code.
//
// Children may be called multiple times — each call appends.
//
// Example:
//
//	Panel().
//	    Children(headerRow).
//	    Divider().
//	    Children(row1, row2)
func (p *panelBuilder) Children(els ...retui.Element) *panelBuilder {
	for _, el := range els {
		if el.Type != 0 || len(el.Children) > 0 || el.Text != "" {
			p.children = append(p.children, el)
		}
	}
	return p
}

// Divider adds a full-width horizontal rule inside the panel body.
//
// With FixedWidth the rule is an exact string of ─ characters:
//
//	│─────────────────────────────────────────────────────────────│
//
// Without FixedWidth a Box+Grow layout is used so the rule fills
// available width at render time.
//
// Example:
//
//	Panel().Children(topContent).Divider().Children(bottomContent)
func (p *panelBuilder) Divider() *panelBuilder {
	borderStyle := p.borderStyle()
	var el retui.Element

	if p.fixedW != nil {
		innerWidth := *p.fixedW - 2
		if innerWidth < 0 {
			innerWidth = 0
		}
		el = retui.Text(strings.Repeat("─", innerWidth), borderStyle)
	} else {
		const maxFill = 2000
		el = retui.Box(
			retui.Props{Direction: retui.Row, Width: p.width},
			retui.NewStyle(),
			retui.Box(
				retui.Props{Width: retui.Grow(1)},
				retui.NewStyle(),
				retui.Text(strings.Repeat("─", maxFill), borderStyle),
			),
		)
	}

	p.children = append(p.children, el)
	return p
}

// DividerWithText adds a horizontal rule with a label centered inside it.
//
// With FixedWidth:
//
//	│───── Section Title ─────────────────────────────────────────│
//
// With Grow width the label is placed between two growing fill boxes.
// If the label is longer than the inner width it is truncated.
//
// Example:
//
//	Panel().DividerWithText("Totals")
func (p *panelBuilder) DividerWithText(text string) *panelBuilder {
	borderStyle := p.borderStyle()
	var el retui.Element

	if p.fixedW != nil {
		innerWidth := *p.fixedW - 2
		if innerWidth < 0 {
			innerWidth = 0
		}
		textLen := len([]rune(text))
		if textLen > innerWidth {
			text = string([]rune(text)[:innerWidth])
			textLen = innerWidth
		}
		total := innerWidth - textLen
		leftFill := total / 2
		rightFill := total - leftFill
		el = retui.Text(
			strings.Repeat("─", leftFill)+text+strings.Repeat("─", rightFill),
			borderStyle,
		)
	} else {
		const maxFill = 2000
		el = retui.Box(
			retui.Props{Direction: retui.Row, Width: p.width},
			retui.NewStyle(),
			retui.Box(
				retui.Props{Width: retui.Grow(1)},
				retui.NewStyle(),
				retui.Text(strings.Repeat("─", maxFill), borderStyle),
			),
			retui.Text(text, borderStyle),
			retui.Box(
				retui.Props{Width: retui.Grow(1)},
				retui.NewStyle(),
				retui.Text(strings.Repeat("─", maxFill), borderStyle),
			),
		)
	}

	p.children = append(p.children, el)
	return p
}

// Render assembles and returns the final retui.Element.
// Call this once at the end of the builder chain.
//
// Example:
//
//	el := Panel().FixedWidth(60).Header(title).Children(rows...).Render()
func (p *panelBuilder) Render() retui.Element {
	return p.build()
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal build
// ─────────────────────────────────────────────────────────────────────────────

// build assembles the complete panel element from accumulated builder state.
func (p *panelBuilder) build() retui.Element {
	bs := p.borderStyle()

	// Inner width = total width − left border − right border.
	// 0 when width is Grow (unknown at build time).
	innerWidth := 0
	if p.fixedW != nil {
		innerWidth = *p.fixedW - 2
		if innerWidth < 0 {
			innerWidth = 0
		}
	}

	headerRow, headerH := p.buildHeaderRow(bs, innerWidth)

	// ── Content rows ─────────────────────────────────────────────────────────
	contentRows := []retui.Element{}
	actualContentH := 0

	for i, child := range p.children {
		// Gap between consecutive children
		if i > 0 && p.contentGap > 0 {
			contentRows = append(contentRows, retui.Box(
				retui.Props{Height: retui.Fixed(p.contentGap)},
				retui.NewStyle(),
			))
			actualContentH += p.contentGap
		}

		var cw retui.Sizing
		if p.fixedW != nil {
			cw = retui.Fixed(innerWidth)
		} else {
			cw = retui.Grow(1)
		}

		rowH := measureHeight(child)
		actualContentH += rowH

		contentRows = append(contentRows, retui.Box(
			retui.Props{Direction: retui.Row, Width: p.width},
			retui.NewStyle(),
			buildBorderCol("│", bs, rowH),
			retui.Box(retui.Props{Width: cw}, retui.NewStyle(), child),
			buildBorderCol("│", bs, rowH),
		))
	}

	// ── Chrome height accounting ──────────────────────────────────────────────
	// top-border(1) + header(headerH) + headerGap + ├─┤(1) + bottom-border(1)
	chromeH := 1 + headerH + p.headerGap + 1 + 1

	// ── FixedHeight padding ───────────────────────────────────────────────────
	// When FixedHeight is set, pad with blank bordered rows so side walls
	// extend to the bottom border even when content is shorter.
	if p.fixedH != nil {
		available := *p.fixedH - chromeH
		if available < 0 {
			available = 0
		}
		if leftover := available - actualContentH; leftover > 0 {
			var fw retui.Sizing
			if p.fixedW != nil {
				fw = retui.Fixed(innerWidth)
			} else {
				fw = retui.Grow(1)
			}
			contentRows = append(contentRows, retui.Box(
				retui.Props{Direction: retui.Row, Width: p.width},
				retui.NewStyle(),
				buildBorderCol("│", bs, leftover),
				retui.Box(
					retui.Props{Width: fw, Height: retui.Fixed(leftover)},
					retui.NewStyle(),
				),
				buildBorderCol("│", bs, leftover),
			))
			actualContentH += leftover
		}
	}

	// ── Assemble ──────────────────────────────────────────────────────────────
	elements := []retui.Element{
		p.buildBorderLine("┌", "─", "┐", bs, innerWidth),
		headerRow,
	}

	if p.headerGap > 0 {
		elements = append(elements,
			retui.Box(retui.Props{Height: retui.Fixed(p.headerGap)}, retui.NewStyle()),
		)
	}

	elements = append(elements, p.buildBorderLine("├", "─", "┤", bs, innerWidth))

	if len(contentRows) > 0 {
		wp := retui.Props{Direction: retui.Column, Width: p.width, Gap: 0}
		if p.fixedH != nil {
			wp.Height = retui.Fixed(actualContentH)
		}
		elements = append(elements, retui.Box(wp, retui.NewStyle(), contentRows...))
	}

	elements = append(elements, p.buildBorderLine("└", "─", "┘", bs, innerWidth))

	op := retui.Props{Direction: retui.Column, Width: p.width, Gap: 0}
	if p.fixedH != nil {
		op.Height = p.height
	}
	return retui.Box(op, retui.NewStyle(), elements...)
}

// buildHeaderRow returns the bordered header row and its measured height.
func (p *panelBuilder) buildHeaderRow(bs retui.Style, innerWidth int) (retui.Element, int) {
	var inner retui.Element
	if p.hasHeader {
		inner = p.header
	} else {
		var cw retui.Sizing
		if p.fixedW != nil {
			cw = retui.Fixed(innerWidth)
		} else {
			cw = retui.Grow(1)
		}
		inner = retui.Box(
			retui.Props{Width: cw, Padding: [4]int{0, 1, 0, 1}},
			retui.NewStyle(),
		)
	}

	h := measureHeight(inner)
	row := retui.Box(
		retui.Props{Direction: retui.Row, Width: p.width},
		retui.NewStyle(),
		buildBorderCol("│", bs, h),
		retui.Box(retui.Props{Width: retui.Grow(1)}, retui.NewStyle(), inner),
		buildBorderCol("│", bs, h),
	)
	return row, h
}

// buildBorderLine returns a single horizontal border line:
// left + strings.Repeat(fill, innerWidth) + right (fixed)
// or Box+Grow (non-fixed).
func (p *panelBuilder) buildBorderLine(left, fill, right string, bs retui.Style, innerWidth int) retui.Element {
	if p.fixedW != nil {
		if innerWidth < 0 {
			innerWidth = 0
		}
		return retui.Text(left+strings.Repeat(fill, innerWidth)+right, bs)
	}
	const maxFill = 2000
	return retui.Box(
		retui.Props{Direction: retui.Row, Width: p.width},
		retui.NewStyle(),
		retui.Text(left, bs),
		retui.Box(
			retui.Props{Width: retui.Grow(1)},
			retui.NewStyle(),
			retui.Text(strings.Repeat(fill, maxFill), bs),
		),
		retui.Text(right, bs),
	)
}

// borderStyle returns the effective border style.
func (p *panelBuilder) borderStyle() retui.Style {
	if p.style != nil {
		return *p.style
	}
	return retui.NewStyle().Foreground(retui.Gray(1))
}

// ─────────────────────────────────────────────────────────────────────────────
// Layout measurement helpers
// ─────────────────────────────────────────────────────────────────────────────

// measureHeight estimates how many terminal rows an Element will occupy.
// It is a best-effort approximation used for content height accounting and
// FixedHeight padding — it does not replicate the full retui layout engine.
//
// Known limitation: ElementBox with Height: Fixed(n) where n > intrinsic
// content height returns the intrinsic height, not n. The layout engine
// enforces Fixed height at paint time, but this function would need to
// inspect the Sizing type to replicate that behaviour.
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

// measureBoxHeight computes the height of a Box element from its children,
// gap, padding, and direction — mirroring the subset of layout rules that
// affect height without a full layout pass.
func measureBoxHeight(el retui.Element) int {
	pad := el.Layout.PaddingTop + el.Layout.PaddingBottom

	// Respect an explicit Fixed height instead of deriving it from
	// children — the layout engine enforces Fixed height at paint time
	// (see LayoutNode.HeightSizing / SizingFixed), so measurement must
	// agree or the border columns built from this value end up too short.
	if el.Layout.HeightSizing.Mode == retui.SizingFixed {
		return el.Layout.HeightSizing.Value
	}

	if len(el.Children) == 0 {
		return 1 + pad
	}

	if el.Layout.Direction == retui.Row {
		// Row: height = tallest child
		max := 0
		for _, c := range el.Children {
			if h := measureHeight(c); h > max {
				max = h
			}
		}
		return max + pad
	}

	// Column: height = sum of children + gaps
	total := el.Layout.Gap * (len(el.Children) - 1)
	for _, c := range el.Children {
		total += measureHeight(c)
	}
	return total + pad
}

// buildBorderCol returns a vertical border element height rows tall.
//
// A single Text node stretched to height lines would only paint on its
// first row — the retui painter does not repeat characters vertically.
// Instead we stack height individual Text nodes in a Column so each row
// gets its own border character.
func buildBorderCol(ch string, style retui.Style, height int) retui.Element {
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
