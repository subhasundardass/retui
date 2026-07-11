package retui

import (
	"strings"
)

var Renderer = NewRenderer(StdOutScreen)

// rawLines returns the text split on '\n' only, with no word wrapping.
func rawLines(text string) []string {
	return strings.Split(text, "\n")
}

// wrappedLines splits text on '\n' then word-wraps each segment to fit
// within maxWidth columns.
func wrappedLines(text string, maxWidth int) []string {
	var out []string
	for _, seg := range rawLines(text) {
		out = append(out, wrapText(seg, maxWidth)...)
	}
	return out
}

// wrapText breaks a single line (no '\n') into one or more lines so
// each line's total cell width is <= maxWidth, preferring word
// boundaries. Words longer than maxWidth are hard-broken as a fallback.
func wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var line strings.Builder
	lineWidth := 0

	for _, word := range words {
		wordWidth := len([]rune(word))

		if lineWidth == 0 {
			for wordWidth > maxWidth {
				runes := []rune(word)
				line.WriteString(string(runes[:maxWidth]))
				lines = append(lines, line.String())
				line.Reset()
				word = string(runes[maxWidth:])
				wordWidth = len([]rune(word))
			}
			line.WriteString(word)
			lineWidth = wordWidth
		} else if lineWidth+1+wordWidth <= maxWidth {
			line.WriteByte(' ')
			line.WriteString(word)
			lineWidth += 1 + wordWidth
		} else {
			lines = append(lines, line.String())
			line.Reset()
			lineWidth = 0

			for wordWidth > maxWidth {
				runes := []rune(word)
				line.WriteString(string(runes[:maxWidth]))
				lines = append(lines, line.String())
				line.Reset()
				word = string(runes[maxWidth:])
				wordWidth = len([]rune(word))
			}
			line.WriteString(word)
			lineWidth = wordWidth
		}
	}

	lines = append(lines, line.String())
	return lines
}

// ComponentRenderer owns the Screen and drives layout + paint on every
// frame. It contains no dirty channel — scheduling is handled entirely
// by App.Run's select loop and the dirty-cell tracking in Screen.
type ComponentRenderer struct {
	screen *Screen
}

func NewRenderer(screen *Screen) *ComponentRenderer {
	return &ComponentRenderer{screen: screen}
}

// Render runs a full layout + paint pass for the given element tree.
// Cell-level diffing in SetCell ensures that only genuinely changed
// cells are marked dirty, so the subsequent Flush call emits the
// minimum possible ANSI output even though paint visits every cell.
func (r *ComponentRenderer) Render(next Element) {
	layoutRoot := buildLayoutTree(next)

	contentW, contentH := IntrinsicSize(layoutRoot)
	screenH := r.screen.Height()
	if contentH > screenH {
		r.screen.Resize(r.screen.Width(), contentH)
	}

	// Respect the root's sizing intent. A Fit root occupies only its
	// intrinsic size; Grow/Fixed roots adopt the full screen rect.
	availW := r.screen.Width()
	availH := r.screen.Height()
	if layoutRoot.WidthSizing.Mode == SizingFit && contentW < availW {
		availW = contentW
	}
	if layoutRoot.HeightSizing.Mode == SizingFit && contentH < availH {
		availH = contentH
	}

	available := Rect{X: 0, Y: 0, Width: availW, Height: availH}
	rects := ComputeLayout(layoutRoot, available)

	// After ComputeLayout, intrinsicHeight reflects any reflow passes
	// (e.g. wrapped text expanding the tree). Use that for scrollback
	// bookkeeping instead of the pre-reflow value from earlier.
	finalH := layoutRoot.intrinsicHeight
	if finalH > r.screen.Height() {
		r.screen.Resize(r.screen.Width(), finalH)
		rects = ComputeLayout(layoutRoot, Rect{X: 0, Y: 0, Width: availW, Height: finalH})
	}

	r.screen.Clear()
	paint(next, rects, 0, r.screen, Style{})

	// If content overflows the terminal viewport, write rows inline so
	// the terminal scrolls older content into scrollback. Must run after
	// paint because it reads from the cell grid.
	r.screen.EnsureRoom(finalH)
}

func buildLayoutTree(element Element) *LayoutNode {
	p := element.Layout
	b := element.Style.border

	padTop, padRight, padBottom, padLeft := p.PaddingTop, p.PaddingRight, p.PaddingBottom, p.PaddingLeft
	if b.Top {
		padTop++
	}
	if b.Right {
		padRight++
	}
	if b.Bottom {
		padBottom++
	}
	if b.Left {
		padLeft++
	}

	l := &LayoutNode{
		Direction:     p.Direction,
		WidthSizing:   p.WidthSizing,
		HeightSizing:  p.HeightSizing,
		paddingTop:    padTop,
		paddingRight:  padRight,
		paddingBottom: padBottom,
		paddingLeft:   padLeft,
		gap:           p.Gap,
		alignment:     p.Align,
		justify:       p.Justify,
	}

	switch element.Type {
	case ElementOverlay:
		// An overlay occupies zero space in the flow layout — its position
		// is absolute (OverlayX/OverlayY on the Element), so it must not
		// push siblings or contribute to the parent's measured size.
		// Children are still added so the rects slice stays in sync with
		// the paint traversal order.
		l.WidthSizing = Fixed(0)
		l.HeightSizing = Fixed(0)

	case ElementText:
		w := 0
		for _, ch := range element.Text {
			w += RuneWidth(ch)
		}
		l.WidthSizing = Fixed(w)
		l.HeightSizing = Fixed(1)

	case ElementMultilineText:
		if element.Wrap {
			l.WidthSizing = Grow(1)
			l.HeightSizing = Fit()
			text := element.Text
			l.reflow = func(width int) int {
				if width <= 0 {
					return 1
				}
				return len(wrappedLines(text, width))
			}
		} else {
			lines := rawLines(element.Text)
			widest := 0
			for _, line := range lines {
				w := 0
				for _, ch := range line {
					w += RuneWidth(ch)
				}
				if w > widest {
					widest = w
				}
			}
			l.WidthSizing = Fixed(widest)
			l.HeightSizing = Fixed(len(lines))
		}

	case ElementMarkdown:
		l.WidthSizing = Grow(1)
		l.HeightSizing = Fit()
		markdownText := element.MarkdownText
		baseStyle := element.Style
		l.reflow = func(width int) int {
			if width <= 0 || markdownText == "" {
				return 1
			}
			lines := renderMarkdownLines(markdownText, width, baseStyle)
			return len(lines)
		}
		l.intrinsicHeight = 1
		if len(element.Markdown.Lines) > 0 {
			l.intrinsicHeight = len(element.Markdown.Lines)
		}
	}

	for _, child := range element.Children {
		l.Children = append(l.Children, buildLayoutTree(child))
	}
	return l
}

// paint walks the element tree in depth-first pre-order, matching the
// traversal order ComputeLayout uses to produce rects. parentStyle is
// inherited from ancestors; each element merges its own Style onto it
// before painting and before passing it to children.
func paint(element Element, rects []Rect, idx int, screen *Screen, parentStyle Style) int {
	rect := rects[idx]
	idx++

	effective := mergeStyles(parentStyle, element.Style)

	switch element.Type {
	case ElementOverlay:
		// Ignore the flow-assigned rect entirely. Paint children at the
		// absolute screen position stored on the element, so the overlay
		// floats on top of whatever flow-positioned content is underneath.
		paintOverlayChildren(element, screen, effective)

	case ElementBox:
		for x := rect.X; x < rect.X+rect.Width; x++ {
			for y := rect.Y; y < rect.Y+rect.Height; y++ {
				screen.SetCell(x, y, ' ', effective)
			}
		}
		paintBorder(screen, rect, effective, element.Style.border)

	case ElementText:
		x := rect.X
		for _, ch := range element.Text {
			if x >= rect.X+rect.Width {
				break
			}
			screen.SetCell(x, rect.Y, ch, effective)
			x += RuneWidth(ch)
		}

	case ElementMultilineText:
		var lines []string
		if element.Wrap {
			lines = wrappedLines(element.Text, rect.Width)
		} else {
			lines = rawLines(element.Text)
		}
		for i, line := range lines {
			y := rect.Y + i
			if y >= rect.Y+rect.Height {
				break
			}
			x := rect.X
			for _, ch := range line {
				if x >= rect.X+rect.Width {
					break
				}
				screen.SetCell(x, y, ch, effective)
				x += RuneWidth(ch)
			}
		}

	case ElementMarkdown:
		lines := element.Markdown.Lines
		if element.MarkdownText != "" && rect.Width > 0 {
			lines = renderMarkdownLines(element.MarkdownText, rect.Width, element.Style)
		}
		for i, line := range lines {
			y := rect.Y + i
			if y >= rect.Y+rect.Height {
				break
			}
			x := rect.X
			for _, cell := range line {
				if x >= rect.X+rect.Width {
					break
				}
				cellStyle := mergeStyles(effective, cell.style)
				screen.SetCell(x, y, cell.r, cellStyle)
				x += RuneWidth(cell.r)
			}
		}
	}

	// Overlay children are painted via paintOverlayChildren above and do
	// not participate in the rects traversal — skip their idx slots.
	if element.Type != ElementOverlay {
		for _, child := range element.Children {
			idx = paint(child, rects, idx, screen, effective)
		}
	} else {
		// Still need to advance idx past the slots ComputeLayout allocated
		// for the overlay's children so subsequent siblings read the right rect.
		idx = skipRects(element, idx)
	}
	return idx
}

// skipRects advances idx past all rects that ComputeLayout allocated for
// element and its entire subtree, without painting anything. Used when
// paint has already handled a subtree by other means (e.g. overlay absolute
// painting) but must keep the rects index in sync for subsequent siblings.
func skipRects(element Element, idx int) int {
	for _, child := range element.Children {
		idx++ // skip child's own rect
		idx = skipRects(child, idx)
	}
	return idx
}

// paintOverlayChildren renders element's children at absolute coordinates
// (element.OverlayX, element.OverlayY), bypassing flow layout completely.
// Each child is built into its own independent layout tree so ComputeLayout
// gives it a fresh rect starting at (OverlayX, OverlayY).

func paintOverlayChildren(element Element, screen *Screen, parentStyle Style) {
	if len(element.Children) == 0 {
		return
	}

	wrapper := Element{
		Type:     ElementBox,
		Style:    element.Style,
		Children: element.Children,
		Layout: LayoutProps{
			WidthSizing:  Fit(),
			HeightSizing: Fit(),
		},
	}

	layoutRoot := buildLayoutTree(wrapper)

	// Measure intrinsic size BEFORE computing layout, and clamp the
	// available rect to it so Fit() can't expand into leftover screen space.
	contentW, contentH := IntrinsicSize(layoutRoot)

	maxW := screen.Width() - element.OverlayX
	maxH := screen.Height() - element.OverlayY
	if contentW < maxW {
		maxW = contentW
	}
	if contentH < maxH {
		maxH = contentH
	}

	available := Rect{
		X:      element.OverlayX,
		Y:      element.OverlayY,
		Width:  maxW,
		Height: maxH,
	}
	rects := ComputeLayout(layoutRoot, available)
	paint(wrapper, rects, 0, screen, parentStyle)
}

func paintBorder(screen *Screen, rect Rect, base Style, b Border) {
	if !b.Any() || rect.Width == 0 || rect.Height == 0 {
		return
	}

	bs := base
	if b.Color.Type != ColorNone {
		bs.foreground = b.Color
	}
	c := b.Chars

	x0, y0 := rect.X, rect.Y
	x1, y1 := rect.X+rect.Width-1, rect.Y+rect.Height-1

	// if b.Top {
	// 	for x := x0 + 1; x < x1; x++ {
	// 		screen.SetCell(x, y0, c.Top, bs)
	// 	}
	// }

	//--Added title if avaiable
	if b.Top {
		inside := x1 - x0 - 1

		// Draw full top border first
		for x := x0 + 1; x < x1; x++ {
			screen.SetCell(x, y0, c.Top, bs)
		}

		if b.Title != "" && inside > 2 {
			title := " " + b.Title + " "
			runes := []rune(title)

			if len(runes) > inside {
				runes = runes[:inside]
			}

			start := x0 + 2 // leave one border glyph before title

			for i, r := range runes {
				x := start + i
				if x >= x1 {
					break
				}

				screen.SetCell(x, y0, r, bs)
			}
		}
	}
	if b.Bottom && y1 != y0 {
		for x := x0 + 1; x < x1; x++ {
			screen.SetCell(x, y1, c.Bottom, bs)
		}
	}
	if b.Left {
		for y := y0 + 1; y < y1; y++ {
			screen.SetCell(x0, y, c.Left, bs)
		}
	}
	if b.Right && x1 != x0 {
		for y := y0 + 1; y < y1; y++ {
			screen.SetCell(x1, y, c.Right, bs)
		}
	}

	if g := cornerGlyph(c.TopLeft, c.Top, c.Left, b.Top, b.Left); g != 0 {
		screen.SetCell(x0, y0, g, bs)
	}
	if g := cornerGlyph(c.TopRight, c.Top, c.Right, b.Top, b.Right); g != 0 {
		screen.SetCell(x1, y0, g, bs)
	}
	if g := cornerGlyph(c.BottomLeft, c.Bottom, c.Left, b.Bottom, b.Left); g != 0 {
		screen.SetCell(x0, y1, g, bs)
	}
	if g := cornerGlyph(c.BottomRight, c.Bottom, c.Right, b.Bottom, b.Right); g != 0 {
		screen.SetCell(x1, y1, g, bs)
	}
}

// cornerGlyph picks the rune for a single corner of a box border.
// Returns 0 to skip drawing the corner cell entirely.
func cornerGlyph(cornerChar, hChar, vChar rune, hasH, hasV bool) rune {
	switch {
	case hasH && hasV:
		return cornerChar
	case hasH:
		return hChar
	case hasV:
		return vChar
	default:
		return 0
	}
}
