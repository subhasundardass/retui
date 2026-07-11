package retui

const (
	ElementBox ElementType = iota
	ElementText
	ElementMultilineText
	ElementMarkdown
	ElementComponent
	ElementOverlay // floating absolute-position container — see Overlay()
)

// Box creates a container element that arranges its children using a
// flexbox-style layout.
//
// Box is the primary layout primitive in  A box can arrange its
// children either horizontally (Row) or vertically (Column), apply
// padding, spacing, alignment, justification, sizing, and styling.
//
// If Width or Height is not specified, the box defaults to Fit(),
// meaning it sizes itself to its content.
//
// Parameters:
//
//   - props: Layout configuration such as direction, sizing, padding,
//     alignment, and justification.
//   - style: Visual styling including borders, foreground/background
//     colors, and text attributes.
//   - children: Child elements to render inside the box.
//
// Common use cases:
//
//   - Vertical page layouts
//   - Header, sidebar, and footer layouts
//   - Forms and dashboards
//   - Nested flexbox-style UIs
//

func Box(props Props, style Style, children ...Element) Element {
	width := props.Width
	if width == (Sizing{}) {
		width = Fit()
	}
	height := props.Height
	if height == (Sizing{}) {
		height = Fit()
	}

	return Element{
		Type: ElementBox,
		Layout: LayoutProps{
			Direction:     props.Direction,
			WidthSizing:   width,
			HeightSizing:  height,
			Gap:           props.Gap,
			PaddingTop:    props.Padding[0],
			PaddingRight:  props.Padding[1],
			PaddingBottom: props.Padding[2],
			PaddingLeft:   props.Padding[3],
			Align:         props.Align,
			Justify:       props.Justify,
		},
		Style:    style,
		Children: children,
	}
}

func Text(text string, style Style) Element {
	return Element{
		Type:  ElementText,
		Text:  text,
		Style: style,
	}
}

// MultilineText renders text that may contain '\n' line breaks. Each '\n'
// starts a new row at the element's left edge; the intrinsic width is the
// longest line and the intrinsic height is the line count.
func MultilineText(text string, style Style) Element {
	return Element{
		Type:  ElementMultilineText,
		Text:  text,
		Style: style,
	}
}

// WrappedText renders text that fills its container's available width and
// breaks on word boundaries to fit. Existing '\n' characters still force a
// new row. Width grows to the parent's cross-axis size; height adapts to the
// resulting wrapped line count.
func WrappedText(text string, style Style) Element {
	return Element{
		Type:  ElementMultilineText,
		Text:  text,
		Wrap:  true,
		Style: style,
	}
}

// If returns choice1 when condition is true and choice2 otherwise. It is a
// ternary-style helper for composing element trees inline, since both
// branches are evaluated before the call, use it for picking between
// already-constructed elements rather than for guarding expensive work.
func If(condition bool, choice1 Element, choice2 Element) Element {
	if condition {
		return choice1
	}
	return choice2
}

// ColorIf returns color1 if condition is true, otherwise color2
func ColorIf(condition bool, color1, color2 Color) Color {
	if condition {
		return color1
	}
	return color2
}

// Markdown renders a markdown string with rich formatting including headers,
// bold, italic, code, links, lists, blockquotes, and tables. The content fills
// its container's available width and wraps automatically.
func Markdown(markdown string, style Style) Element {
	width := 80 // default width, will be recalculated during layout
	lines := renderMarkdownLines(markdown, width, style)
	return Element{
		Type:         ElementMarkdown,
		Markdown:     MarkdownContent{Lines: lines},
		MarkdownText: markdown,
		Style:        style,
	}
}

// Overlay places children at absolute screen position (x, y), completely
// outside normal flow layout. The node itself is zero-size so it does not
// push siblings or affect the parent's dimensions. Children are painted
// after all flow-positioned siblings, so they always appear on top.
//
// Use for modal dialogs, tooltips, popups — anything that should float
// above the rest of the UI at a fixed screen position. Position in the
// element tree does not matter — it always paints at (x, y).
func Overlay(x, y int, children ...Element) Element {
	return Element{
		Type:     ElementOverlay,
		OverlayX: x,
		OverlayY: y,
		Children: children,
	}
}
