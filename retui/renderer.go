package retui

import (
	"strings"
)

// ============================================================================
// Global Renderer
// ============================================================================

// Renderer is the default global renderer instance.
// Initialize via NewRenderer(screen) and reuse for all Render calls.
// Note: Renderer is not thread-safe; synchronize access via your app's event loop.
var screen = StdOutScreen()
var Renderer = NewRenderer(screen)

// ============================================================================
// Text Wrapping
// ============================================================================

// rawLines splits text on '\n' only, with no word wrapping or processing.
// Returns a slice of line strings; empty input produces []string{""}.
func rawLines(text string) []string {
	return strings.Split(text, "\n")
}

// wrappedLines splits text on '\n', then word-wraps each segment to fit
// within maxWidth columns, handling multi-rune characters correctly.
//
// Word wrapping uses space boundaries preferentially; words exceeding maxWidth
// are hard-broken at maxWidth boundary.
//
// If maxWidth <= 0, returns text as-is (single line with no wrapping).
// Empty input produces []string{""}.
func wrappedLines(text string, maxWidth int) []string {
	var out []string
	for _, seg := range rawLines(text) {
		out = append(out, wrapText(seg, maxWidth)...)
	}
	return out
}

// wrapText breaks a single unwrapped line into one or more lines, each with
// cell width <= maxWidth. Prefers word boundaries (space-separated); words
// longer than maxWidth are hard-broken at the boundary as a fallback.
//
// Cell width accounts for multi-rune characters via RuneWidth (e.g. wide
// characters count as 2 columns).
//
// maxWidth <= 0 is treated as "no wrap"; the entire text is returned as a
// single line. Empty input produces []string{""}.
//
// Example:
//
//	wrapText("hello world test", 10) → []string{"hello", "world test"}
//	wrapText("verylongword", 5) → []string{"verylo", "ngwor", "d"}
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
		wordWidth := countRuneWidth(word)

		if lineWidth == 0 {
			// Line is empty: add word, hard-break if necessary.
			for wordWidth > maxWidth {
				runes := []rune(word)
				addRunesToBuilder(&line, runes, 0, maxWidth)
				lines = append(lines, line.String())
				line.Reset()
				word = string(runes[maxWidth:])
				wordWidth = countRuneWidth(word)
			}
			line.WriteString(word)
			lineWidth = wordWidth
		} else if lineWidth+1+wordWidth <= maxWidth {
			// Word fits with space separator.
			line.WriteByte(' ')
			line.WriteString(word)
			lineWidth += 1 + wordWidth
		} else {
			// Word doesn't fit: flush line and start fresh.
			lines = append(lines, line.String())
			line.Reset()
			// lineWidth = 0

			// Hard-break oversized word if necessary.
			for wordWidth > maxWidth {
				runes := []rune(word)
				addRunesToBuilder(&line, runes, 0, maxWidth)
				lines = append(lines, line.String())
				line.Reset()
				word = string(runes[maxWidth:])
				wordWidth = countRuneWidth(word)
			}
			line.WriteString(word)
			lineWidth = wordWidth
		}
	}

	// Append any remaining partial line.
	lines = append(lines, line.String())
	return lines
}

// countRuneWidth returns the cell width (display columns) of a string,
// accounting for multi-rune characters via RuneWidth.
func countRuneWidth(text string) int {
	width := 0
	for _, r := range text {
		width += RuneWidth(r)
	}
	return width
}

// addRunesToBuilder appends up to maxWidth columns of runes from [start..),
// accounting for multi-rune character widths.
func addRunesToBuilder(b *strings.Builder, runes []rune, start, maxWidth int) {
	width := 0
	for i := start; i < len(runes) && width < maxWidth; i++ {
		b.WriteRune(runes[i])
		width += RuneWidth(runes[i])
	}
}

// ============================================================================
// Renderer
// ============================================================================

// ComponentRenderer owns a Screen and drives layout + paint on every render frame.
// Render() computes layout and paints the element tree to the screen atomically.
// Cell-level diffing in Screen ensures only changed cells emit ANSI output.
//
// Renderer is not thread-safe; coordinate all Render calls via your app's
// event loop or use explicit synchronization.
type ComponentRenderer struct {
	screen *Screen
}

// NewRenderer creates a new renderer backed by the given screen.
// screen must be non-nil; panics if screen is nil.
//
// Multiple renderers can exist; typically your app uses a single global Renderer.
func NewRenderer(screen *Screen) *ComponentRenderer {
	if screen == nil {
		panic("NewRenderer: screen is nil")
	}
	return &ComponentRenderer{screen: screen}
}

// pendingOverlay represents an overlay element deferred until after the main tree
// is painted, ensuring overlays always render on the topmost visual layer.
type pendingOverlay struct {
	element     Element
	parentStyle Style
}

// Render computes layout and paints element tree to the screen in one atomic pass.
//
// Steps:
// 1. Resolve deferred content (ContentBuilder nodes) with their measured sizes.
// 2. Build layout tree from element tree.
// 3. Measure intrinsic size and adjust screen height if needed.
// 4. Compute layout (position and size each node).
// 5. Paint main tree, deferring overlays.
// 6. Paint deferred overlays on top.
// 7. Ensure scrollback room if content overflows terminal viewport.
//
// Cell-level diffing (SetCell) marks only genuinely changed cells dirty,
// so Flush emits minimal ANSI output even though paint visits every cell.
//
// If content's wrapped text (reflow) causes height to expand after initial
// layout, Render re-measures and re-layouts automatically to propagate the
// new height to containers.
func (r *ComponentRenderer) Render(next Element) {
	// Step 1: Resolve deferred (size-aware) content BEFORE building real layout tree.
	// ContentBuilder nodes need their resolved width/height from parent constraints,
	// not terminal size guesses. This pass gives them the true size they'll actually get.
	next = resolveDeferred(next, r.screen.Width(), r.screen.Height())

	// Step 2: Build layout tree from element tree.
	layoutRoot := buildLayoutTree(next)

	// Step 3: Measure intrinsic size; expand screen if needed for content height.
	contentW, contentH := IntrinsicSize(layoutRoot)
	screenH := r.screen.Height()
	if contentH > screenH {
		r.screen.Resize(r.screen.Width(), contentH)
	}

	// Respect root's sizing intent:
	// - Fit root: occupy only intrinsic size
	// - Grow/Fixed root: adopt full screen/allocated space
	availW := r.screen.Width()
	availH := r.screen.Height()
	if layoutRoot.WidthSizing.Mode == SizingFit && contentW < availW {
		availW = contentW
	}
	if layoutRoot.HeightSizing.Mode == SizingFit && contentH < availH {
		availH = contentH
	}

	// Step 4a: Compute layout with current constraints.
	available := Rect{X: 0, Y: 0, Width: availW, Height: availH}
	rects := ComputeLayout(layoutRoot, available)

	// Step 4b: If reflow callbacks fired (e.g. wrapped text), tree height may have changed.
	// Re-measure and re-layout to propagate new height to ancestors (containers around
	// wrapped text must grow to fit).
	finalH := layoutRoot.intrinsicHeight
	if finalH > r.screen.Height() {
		r.screen.Resize(r.screen.Width(), finalH)
		rects = ComputeLayout(layoutRoot, Rect{X: 0, Y: 0, Width: availW, Height: finalH})
	}

	r.screen.Clear()

	// Steps 5-6: Paint main tree and collect overlays for deferred final pass.
	var pending []pendingOverlay
	paint(next, rects, 0, r.screen, Style{}, &pending)

	// Second pass: paint overlays LAST, on top of everything.
	// This ensures overlays never get stamped over by siblings/cousins that
	// paint later in the main tree traversal.
	for _, po := range pending {
		paintOverlayChildren(po.element, r.screen, po.parentStyle)
	}

	// Step 7: Ensure scrollback room if content overflows terminal viewport.
	// Must run after paint (reads from cell grid).
	r.screen.EnsureRoom(finalH)
}

// ============================================================================
// Deferred Content Resolution
// ============================================================================

// hasDeferred reports whether element or any descendant has a ContentBuilder
// that needs resolving.
func hasDeferred(element Element) bool {
	if element.ContentBuilder != nil {
		return true
	}
	for _, c := range element.Children {
		if hasDeferred(c) {
			return true
		}
	}
	return false
}

// resolveDeferred walks element tree and resolves all ContentBuilder callbacks
// with their actual allocated sizes.
//
// How it works:
//  1. Run a throwaway ("placeholder") layout pass where ContentBuilder nodes
//     are treated as childless leaves.
//  2. A ContentBuilder node's LayoutProps sizing (Fixed/Grow/Fit/Percent)
//     determines its placeholder rect exactly as any other leaf would.
//     For example, Grow(1) still correctly receives its share of the parent's
//     remaining space in the placeholder pass.
//  3. That rect's width/height is passed to ContentBuilder, returning the
//     real Element, which replaces the node.
//  4. Recursively resolve any deferred content in the built result.
//
// This mirrors the reflow mechanism in layout.go (height deferred until width
// known) one level higher: here a node's entire content is deferred until its
// size is known, not just a dimension number.
//
// availW, availH are the screen/parent dimensions available for the initial
// placeholder pass.
func resolveDeferred(element Element, availW, availH int) Element {
	if !hasDeferred(element) {
		return element
	}

	// Build placeholder tree (ContentBuilder nodes are leaves).
	placeholderRoot := buildLayoutTree(element)
	rects := ComputeLayout(placeholderRoot, Rect{X: 0, Y: 0, Width: availW, Height: availH})

	idx := 0
	var resolve func(e Element) Element
	resolve = func(e Element) Element {
		rect := rects[idx]
		idx++

		if e.ContentBuilder != nil {
			// Call builder with resolved size.
			built := e.ContentBuilder(rect.Width, rect.Height)

			// Built content may itself contain deferred nodes (unusual but allowed).
			// Resolve recursively, scoped to the space this node was given.
			return resolveDeferred(built, rect.Width, rect.Height)
		}

		if len(e.Children) == 0 {
			return e
		}

		// Resolve children.
		newChildren := make([]Element, len(e.Children))
		for i, c := range e.Children {
			newChildren[i] = resolve(c)
		}
		e.Children = newChildren
		return e
	}

	return resolve(element)
}

// ============================================================================
// Layout Tree Construction
// ============================================================================

// buildLayoutTree converts an Element tree to a LayoutNode tree, extracting
// sizing, padding, border, and gap information.
//
// Element text content (ElementText, ElementMultilineText, ElementMarkdown)
// is converted to fixed or reflow-based sizing:
//   - Plain text: fixed width (summed rune widths) and height 1.
//   - Wrapped multiline: Grow(1) width, Fit() height, with reflow callback
//     to compute height from allocated width.
//   - Unwrapped multiline: fixed width (max line width) and height (line count).
//   - Markdown: Grow(1) width, Fit() height, reflow callback renders markdown
//     to lines given allocated width.
//
// Overlay elements are zero-sized in the flow (position is absolute), but
// children are still added to the tree so the rects slice stays in sync with
// the paint traversal order.
//
// Borders are merged into padding (border.Top etc. add 1 to the corresponding
// padding, since border draws inside the padding region).
func buildLayoutTree(element Element) *LayoutNode {
	p := element.Layout
	b := element.Style.border

	// Merge border into padding (border draws inside).
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
		marginTop:     p.MarginTop,
		marginRight:   p.MarginRight,
		marginBottom:  p.MarginBottom,
		marginLeft:    p.MarginLeft,
		gap:           p.Gap,
		alignment:     p.Align,
		justify:       p.Justify,
	}

	switch element.Type {
	case ElementOverlay:
		// Overlay: zero-size in flow (absolute positioning), but children are still
		// added so rects slice stays in sync with paint.
		l.WidthSizing = Fixed(0)
		l.HeightSizing = Fixed(0)

	case ElementText:
		// Plain text: fixed width and height 1.
		w := countRuneWidth(element.Text)
		l.WidthSizing = Fixed(w)
		l.HeightSizing = Fixed(1)

	case ElementMultilineText:
		if element.Wrap {
			// Wrapped: Grow width (fill available), Fit height with reflow.
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
			// Unwrapped: fixed width (max line width) and height (line count).
			lines := rawLines(element.Text)
			widest := 0
			for _, line := range lines {
				w := countRuneWidth(line)
				if w > widest {
					widest = w
				}
			}
			l.WidthSizing = Fixed(widest)
			l.HeightSizing = Fixed(len(lines))
		}

	case ElementMarkdown:
		// Markdown: Grow width (fill available), Fit height with reflow.
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
		// Preserve intrinsic height from prior render (used by reflow propagation).
		l.intrinsicHeight = 1
		if len(element.Markdown.Lines) > 0 {
			l.intrinsicHeight = len(element.Markdown.Lines)
		}
	}

	// Recursively build children.
	for _, child := range element.Children {
		l.Children = append(l.Children, buildLayoutTree(child))
	}
	return l
}

// ============================================================================
// Painting
// ============================================================================

// paint walks element tree in depth-first pre-order (matching ComputeLayout's
// rect order) and paints cells to screen. parentStyle is inherited from
// ancestors; each element merges its own Style onto it.
//
// Overlay nodes are NOT painted here — they're appended to pending and painted
// in a final deferred pass (see Render), ensuring overlays never get stamped
// over by siblings/cousins painted later in tree order.
//
// Returns the index after the last rect consumed, so siblings can continue
// from the correct position in the rects slice.
func paint(element Element, rects []Rect, idx int, screen *Screen, parentStyle Style, pending *[]pendingOverlay) int {
	rect := rects[idx]
	idx++

	// Merge styles: element's style layered on parent's.
	effective := mergeStyles(parentStyle, element.Style)

	switch element.Type {
	case ElementOverlay:
		// Defer: collect for end-of-frame pass (painted last, on top).
		*pending = append(*pending, pendingOverlay{element: element, parentStyle: effective})

	case ElementBox:
		// Fill rect with background.
		for x := rect.X; x < rect.X+rect.Width; x++ {
			for y := rect.Y; y < rect.Y+rect.Height; y++ {
				screen.SetCell(x, y, ' ', effective)
			}
		}
		paintBorder(screen, rect, effective, element.Style.border)

	case ElementText:
		// Paint single line of text.
		x := rect.X
		for _, ch := range element.Text {
			if x >= rect.X+rect.Width {
				break
			}
			screen.SetCell(x, rect.Y, ch, effective)
			x += RuneWidth(ch)
		}

	case ElementMultilineText:
		// Paint multiple lines (wrapped or raw).
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
		// Paint markdown-rendered lines (with inline styling).
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

	// Overlay children are deferred (handled in pending above) and do not
	// participate in the rects traversal here. Skip their idx slots.
	if element.Type != ElementOverlay {
		for _, child := range element.Children {
			idx = paint(child, rects, idx, screen, effective, pending)
		}
	} else {
		// Still need to advance idx past the slots ComputeLayout allocated
		// for the overlay's children so subsequent siblings read the correct rect.
		idx = skipRects(element, idx)
	}
	return idx
}

// skipRects advances idx past all rects allocated by ComputeLayout for element
// and its entire subtree, without painting anything.
//
// Used when a subtree is handled by other means (e.g. overlay painting) but
// the rects index must stay in sync for subsequent siblings.
func skipRects(element Element, idx int) int {
	for _, child := range element.Children {
		idx++ // skip child's rect
		idx = skipRects(child, idx)
	}
	return idx
}

// paintOverlayChildren paints element's children at absolute coordinates
// (element.OverlayX, element.OverlayY), bypassing flow layout.
//
// Each child is measured and laid out in its own independent tree (using a
// wrapper Element) so ComputeLayout gives it a fresh rect starting at
// (OverlayX, OverlayY). This allows overlays to be positioned absolutely
// independent of the main flow.
//
// Called only from Render's deferred final pass, guaranteeing overlay paint
// lands on top of the already-completed main tree.
//
// Nested overlays are allowed (unusual); each gets its own deferred pass
// scoped to the call rather than leaking into the outer frame's pending list.
func paintOverlayChildren(element Element, screen *Screen, parentStyle Style) {
	if len(element.Children) == 0 {
		return
	}

	// Wrap children in a container for independent layout.
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

	// Measure intrinsic size; clamp available rect to it so Fit() doesn't
	// expand into leftover screen space.
	contentW, contentH := IntrinsicSize(layoutRoot)

	maxW := screen.Width() - element.OverlayX
	maxH := screen.Height() - element.OverlayY

	// Clamp to screen boundaries; allow negative coordinates (off-screen overlay).
	if maxW < 0 {
		maxW = 0
	}
	if maxH < 0 {
		maxH = 0
	}
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

	// Nested overlays get their own pending slice scoped to this call.
	var nestedPending []pendingOverlay
	paint(wrapper, rects, 0, screen, parentStyle, &nestedPending)
	for _, po := range nestedPending {
		paintOverlayChildren(po.element, screen, po.parentStyle)
	}
}

// paintBorder renders a box border on the screen with optional title text
// in the top edge.
//
// Borders are drawn at the edges of the rect using Unicode box-drawing glyphs.
// Title (if present and non-empty) is centered in the top edge with space
// padding on both sides; if title is too long, it's truncated.
//
// Does nothing if border has no edges enabled or rect has zero dimensions.
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

	// Top edge: draw line, then overlay title if present.
	if b.Top {
		// Draw full top line with horizontal character.
		for x := x0 + 1; x < x1; x++ {
			screen.SetCell(x, y0, c.Top, bs)
		}

		// Overlay title text if provided.
		if b.Title != nil && b.Title.Text != "" {
			inside := x1 - x0 - 1
			if inside > 2 {
				title := " " + b.Title.Text + " "
				runes := []rune(title)

				// Truncate if needed.
				if len(runes) > inside {
					runes = runes[:inside]
				}

				const padding = 1

				var start int
				switch b.Title.Align {
				case AlignStart:
					start = x0 + 1 + padding

				case AlignEnd:
					start = x1 - len(runes) - padding

				case AlignCenter:
					fallthrough
				default:
					start = x0 + 1 + (inside-len(runes))/2
				}

				// Ensure title stays within the border.
				minStart := x0 + 1
				maxStart := x1 - len(runes)

				if start < minStart {
					start = minStart
				}
				if start > maxStart {
					start = maxStart
				}

				titleStyle := mergeStyles(bs, b.Title.Style)

				for i, r := range runes {
					screen.SetCell(start+i, y0, r, titleStyle)
				}
			}
		}
	}

	// Bottom edge
	if b.Bottom && y1 != y0 {
		for x := x0 + 1; x < x1; x++ {
			screen.SetCell(x, y1, c.Bottom, bs)
		}
	}

	// Left edge
	if b.Left {
		for y := y0 + 1; y < y1; y++ {
			screen.SetCell(x0, y, c.Left, bs)
		}
	}

	// Right edge
	if b.Right && x1 != x0 {
		for y := y0 + 1; y < y1; y++ {
			screen.SetCell(x1, y, c.Right, bs)
		}
	}

	// Corners
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

// cornerGlyph selects the correct rune for a box-drawing corner given the
// available edge glyphs and which edges are enabled.
//
// Returns 0 to skip drawing this corner entirely (e.g. if no edges enabled).
//
// Priority: cornerChar (both edges) > single-edge character > skip.
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
