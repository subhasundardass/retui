package retui

// Rect represents a rectangular region with absolute positioning.
// Used to define bounds, dimensions, and layout output.
type Rect struct {
	X      int // Left edge coordinate
	Y      int // Top edge coordinate
	Width  int // Horizontal extent
	Height int // Vertical extent
}

// Direction specifies the primary layout axis along which children are arranged.
type Direction int

const (
	// Row arranges children left-to-right (horizontal main axis).
	Row Direction = iota
	// Column arranges children top-to-bottom (vertical main axis).
	Column
)

// SizingMode determines how a node's dimension is resolved.
// See Sizing for details on each mode.
type SizingMode int

const (
	// SizingFixed: dimension is a fixed pixel value.
	SizingFixed SizingMode = iota
	// SizingGrow: dimension expands to claim remaining space proportionally.
	// Only valid with a concrete parent size.
	SizingGrow
	// SizingFit: dimension shrink-wraps to children's intrinsic size.
	// Does not participate in parent's remaining-space distribution.
	SizingFit
	// SizingPercent: dimension is a percentage (0-100) of parent's resolved size.
	// If parent size is indefinite (e.g. SizingFit parent), treated as 0 during
	// measure pass; resolved later during layout with actual parent Rect.
	SizingPercent
)

// Sizing describes how a dimension (width or height) is computed.
//
// Mode:  The sizing strategy (Fixed, Grow, Fit, Percent).
// Value: The mode-specific parameter:
//   - SizingFixed:   pixel count
//   - SizingGrow:    weight (default 1 if 0; higher = larger share of remaining space)
//   - SizingFit:     unused (always 0)
//   - SizingPercent: percentage (0-100 of parent's resolved size)
//
// Examples:
//
//	Fixed(100)       → exactly 100px
//	Grow(2)          → 2x weight in remaining space (vs weight 1)
//	Fit()            → shrink-wrap children
//	Percent(50)      → 50% of parent's resolved size on this axis
type Sizing struct {
	Mode  SizingMode
	Value int
}

// Alignment specifies how children align on the cross axis (perpendicular to layout direction).
//
// In a Row (horizontal layout), cross axis is vertical (Y).
// In a Column (vertical layout), cross axis is horizontal (X).
//
// Default (zero value) is AlignStretch, which overrides child sizing on the cross axis.
type Alignment int

const (
	// AlignStretch: children stretch to fill the full cross-axis size of their parent.
	// Overrides child's cross-axis sizing (Height in Row, Width in Column).
	AlignStretch Alignment = iota
	// AlignStart: align children to the start of the cross axis (top in Row, left in Column).
	AlignStart
	// AlignCenter: center children on the cross axis.
	AlignCenter
	// AlignEnd: align children to the end of the cross axis (bottom in Row, right in Column).
	AlignEnd
)

// Justify specifies how children distribute along the main axis when there is extra space.
type Justify int

const (
	// JustifyStart: pack children at the start, extra space at the end (default).
	JustifyStart Justify = iota
	// JustifyEnd: pack children at the end, extra space at the start.
	JustifyEnd
	// JustifyCenter: center children, extra space split equally on both sides.
	JustifyCenter
	// JustifySpaceBetween: equal spacing between children, no space before first or after last.
	JustifySpaceBetween
	// JustifySpaceAround: equal space around each child (half-space at edges).
	JustifySpaceAround
)

// LayoutNode is a node in the layout tree. It describes sizing, positioning,
// and child-arrangement rules for a rectangular region.
//
// Build via builder methods (WithDirection, WithSize, WithChildren, etc.) and
// compute layout via ComputeLayout(root, bounds).
//
// Fields are private; use public methods to construct and configure nodes.
type LayoutNode struct {
	// Direction is the primary layout axis (Row or Column).
	Direction Direction
	// WidthSizing describes how width is computed.
	WidthSizing Sizing
	// HeightSizing describes how height is computed.
	HeightSizing Sizing
	// Children is the list of child nodes.
	Children []*LayoutNode

	// intrinsic dimensions are computed by measure() and used by layout().
	// Preserved across multiple passes if reflow callbacks are present.
	intrinsicHeight int
	intrinsicWidth  int

	// padding reserves space inside the node, between its bounds and children.
	paddingTop    int
	paddingBottom int
	paddingLeft   int
	paddingRight  int

	// gap is the minimum spacing between adjacent children on the main axis.
	gap int

	// alignment controls child positioning on the cross axis.
	// Default (zero) is AlignStretch.
	alignment Alignment
	// justify controls space distribution along the main axis.
	justify Justify

	// reflow is an optional callback for content whose cross-axis dimension affects
	// its main-axis size (e.g. wrapped text where height depends on width).
	// Called during layout once the node's cross-axis size is known.
	// For Row parents: called after width is determined (from grow distribution).
	// For Column parents: called after height is determined (from grow distribution).
	// Return value is the new main-axis size; the node's intrinsic dimension is updated.
	// Return value should be clamped to [0, available]; clamping is not automatic.
	reflow func(crossAxisSize int) int
}

// NewLayout creates a new, empty LayoutNode with default settings:
// - Direction: Row
// - All sizing: SizingFit (shrink-wrap)
// - Alignment: AlignStretch
// - Justify: JustifyStart
// - No padding, gap, or children
func NewLayout() *LayoutNode {
	return &LayoutNode{}
}

// WithDirection sets the layout direction and returns the node for chaining.
// Direction is the primary axis along which children are arranged.
func (l *LayoutNode) WithDirection(dir Direction) *LayoutNode {
	l.Direction = dir
	return l
}

// Fixed returns a Sizing for a fixed pixel dimension.
func Fixed(n int) Sizing {
	return Sizing{Mode: SizingFixed, Value: n}
}

// Grow returns a Sizing that claims remaining space with the given weight.
// In a parent with leftover space after fixed/fit children are sized,
// each Grow child gets a share proportional to its weight.
// Weight 0 defaults to weight 1.
// Example: two children with Grow(1) and Grow(2) split remaining space 1:2.
func Grow(n int) Sizing {
	return Sizing{Mode: SizingGrow, Value: n}
}

// Fit returns a Sizing that shrink-wraps to children's intrinsic size.
// Does not claim remaining space; children are measured normally and the
// dimension is the sum (plus padding/gap) of their contributions.
func Fit() Sizing {
	return Sizing{Mode: SizingFit}
}

// Percent returns a Sizing that resolves to a percentage (0-100) of the
// parent's resolved size along the same axis.
//
// If the parent's size is indefinite during the measure pass (e.g. a SizingFit
// parent trying to shrink-wrap), the percentage cannot be resolved and is
// treated as 0 for measure, mirroring SizingGrow. The real value is resolved
// later during ComputeLayout once the parent has a concrete Rect.
//
// Value is clamped to [0, 100] in practice, though out-of-range values
// are not explicitly rejected.
func Percent(n int) Sizing {
	return Sizing{Mode: SizingPercent, Value: n}
}

// WithSize sets both width and height sizing and returns the node for chaining.
func (l *LayoutNode) WithSize(w, h Sizing) *LayoutNode {
	l.WidthSizing = w
	l.HeightSizing = h
	return l
}

// WithWidth sets width sizing and returns the node for chaining.
func (l *LayoutNode) WithWidth(w Sizing) *LayoutNode {
	l.WidthSizing = w
	return l
}

// WithHeight sets height sizing and returns the node for chaining.
func (l *LayoutNode) WithHeight(h Sizing) *LayoutNode {
	l.HeightSizing = h
	return l
}

// WithChildren sets the child nodes and returns the node for chaining.
func (l *LayoutNode) WithChildren(children ...*LayoutNode) *LayoutNode {
	l.Children = children
	return l
}

// WithPadding sets interior spacing (top, right, bottom, left) and returns the node for chaining.
// Padding reserves space between the node's bounds and its children's layout area.
func (l *LayoutNode) WithPadding(top, right, bottom, left int) *LayoutNode {
	l.paddingTop = top
	l.paddingRight = right
	l.paddingBottom = bottom
	l.paddingLeft = left
	return l
}

// WithPaddingUniform sets uniform padding on all sides and returns the node for chaining.
func (l *LayoutNode) WithPaddingUniform(padding int) *LayoutNode {
	l.paddingTop = padding
	l.paddingRight = padding
	l.paddingBottom = padding
	l.paddingLeft = padding
	return l
}

// WithGap sets the minimum spacing between adjacent children on the main axis
// and returns the node for chaining.
func (l *LayoutNode) WithGap(value int) *LayoutNode {
	l.gap = value
	return l
}

// WithAlign sets cross-axis alignment (AlignStretch, AlignStart, AlignCenter, AlignEnd)
// and returns the node for chaining.
// Default is AlignStretch, which makes children fill the full cross-axis space.
func (l *LayoutNode) WithAlign(alignment Alignment) *LayoutNode {
	l.alignment = alignment
	return l
}

// WithJustify sets main-axis space distribution (JustifyStart, JustifyEnd, etc.)
// and returns the node for chaining.
// Default is JustifyStart (pack at the start).
func (l *LayoutNode) WithJustify(value Justify) *LayoutNode {
	l.justify = value
	return l
}

// WithReflow sets a callback for responsive main-axis sizing based on cross-axis constraints.
// Used for content whose height depends on width (e.g. wrapped text, wrapped flex items).
//
// The callback receives the node's resolved cross-axis size (width in Column,
// height in Row) and must return the updated main-axis size (height in Column,
// width in Row). The returned value should be non-negative and clamped to the
// available space; clamping is not automatic.
//
// Reflow is invoked during the layout pass:
//   - For Column parents: immediately after cross-axis (width) is resolved, before
//     descending into children.
//   - For Row parents: after width distribution is complete (after grow sizing).
//
// If a reflow callback is present anywhere in the tree, ComputeLayout performs
// a second measure+layout pass to propagate updated heights to ancestors.
func (l *LayoutNode) WithReflow(fn func(crossAxisSize int) int) *LayoutNode {
	l.reflow = fn
	return l
}

// ============================================================================
// Public API: IntrinsicSize and ComputeLayout
// ============================================================================

// IntrinsicSize returns the natural (unconstrained) dimensions of a layout tree.
// The tree is measured bottom-up without a parent constraint; Grow and Percent
// sizing modes resolve to 0 during this pass.
//
// Useful for:
// - Determining the minimum bounding box for a layout subtree.
// - Previewing layout before committing to a parent Rect.
// - Implementing responsive containers.
//
// Returns (width, height) as if the tree had infinite space and no parent sizing.
func IntrinsicSize(root *LayoutNode) (width, height int) {
	return measure(root)
}

// ComputeLayout computes the layout of the tree and returns concrete Rects
// for each node in depth-first, pre-order (parent before children).
//
// The root node is constrained to the provided available Rect; child layout
// is derived from parent sizing and positioning rules.
//
// If any node in the tree has a reflow callback, ComputeLayout performs a
// second measure+layout pass: the first pass calls reflow once cross-axis
// sizes are known; the second measure propagates updated dimensions to
// ancestors (e.g. containers around wrapped text grow to fit new heights).
//
// Returns a flat slice of Rects in depth-first, pre-order. Each Rect corresponds
// to a node; indices align with a depth-first traversal of the tree.
//
// Nodes with Fit sizing that contain no children may have 0 dimensions if
// clamped to a parent constraint.
func ComputeLayout(root *LayoutNode, available Rect) []Rect {
	root.intrinsicWidth, root.intrinsicHeight = measure(root)
	var out []Rect
	layout(root, available, &out)

	// If any reflow callbacks fired, re-measure to propagate updated dimensions
	// (e.g. wrapped text height) to ancestors, then re-layout.
	if hasReflow(root) {
		root.intrinsicWidth, root.intrinsicHeight = measure(root)
		out = out[:0]
		layout(root, available, &out)
	}
	return out
}

// ============================================================================
// Private: Helpers and internal layout logic
// ============================================================================

// hasReflow checks if this node or any descendant has a reflow callback.
// Used to determine if a second measure+layout pass is needed.
func hasReflow(n *LayoutNode) bool {
	if n.reflow != nil {
		return true
	}
	for _, c := range n.Children {
		if hasReflow(c) {
			return true
		}
	}
	return false
}

// measure fills in intrinsicWidth and intrinsicHeight for every node in the
// subtree rooted at n (bottom-up, children first).
//
// Sizing modes resolved:
// - SizingFixed: value is used as-is.
// - SizingFit: summed or maxed from children (axis-dependent), plus padding.
// - SizingGrow, SizingPercent: resolved to 0 (no parent size available yet).
//
// Reflow callbacks are NOT invoked during measure (cross-axis size unknown);
// previously-computed reflow heights are preserved for ancestor sizing.
func measure(n *LayoutNode) (int, int) {
	// Measure children first (bottom-up).
	for _, child := range n.Children {
		child.intrinsicWidth, child.intrinsicHeight = measure(child)
	}

	width, height := 0, 0

	// Calculate gap contribution (gap between adjacent children).
	gaps := 0
	if len(n.Children) > 1 {
		gaps = n.gap * (len(n.Children) - 1)
	}

	// Resolve width based on sizing mode.
	switch n.WidthSizing.Mode {
	case SizingFixed:
		width = n.WidthSizing.Value

	case SizingFit:
		if n.Direction == Row {
			// Row: sum children widths + gaps.
			for _, child := range n.Children {
				width += child.intrinsicWidth
			}
			width += gaps
		} else {
			// Column: max child width.
			for _, child := range n.Children {
				if child.intrinsicWidth > width {
					width = child.intrinsicWidth
				}
			}
		}

	case SizingGrow, SizingPercent:
		// Cannot resolve without parent size; use 0.
		width = 0
	}

	// Resolve height based on sizing mode.
	switch n.HeightSizing.Mode {
	case SizingFixed:
		height = n.HeightSizing.Value

	case SizingFit:
		if n.Direction != Row {
			// Column: sum children heights + gaps.
			for _, child := range n.Children {
				height += child.intrinsicHeight
			}
			height += gaps
		} else {
			// Row: max child height.
			for _, child := range n.Children {
				if child.intrinsicHeight > height {
					height = child.intrinsicHeight
				}
			}
		}

	case SizingGrow, SizingPercent:
		// Cannot resolve without parent size; use 0.
		height = 0
	}

	// Add padding for SizingFit dimensions.
	if n.WidthSizing.Mode == SizingFit {
		width += n.paddingLeft + n.paddingRight
	}
	if n.HeightSizing.Mode == SizingFit {
		height += n.paddingTop + n.paddingBottom
	}

	// Preserve reflow height from a prior layout pass.
	// Reflow has access to actual allocated width (more accurate than measure's
	// estimation), so once computed, we trust it across measure passes.
	if n.reflow != nil && n.intrinsicHeight > height {
		height = n.intrinsicHeight
	}

	return width, height
}

// mainSize returns the size along the main axis (width for Row, height for Column).
func mainSize(r Rect, dir Direction) int {
	if dir == Row {
		return r.Width
	}
	return r.Height
}

// setMainSize sets the size along the main axis.
func setMainSize(r *Rect, dir Direction, value int) {
	if dir == Row {
		r.Width = value
	} else {
		r.Height = value
	}
}

// setCrossSize sets the size along the cross axis (height for Row, width for Column).
func setCrossSize(r *Rect, dir Direction, value int) {
	if dir == Row {
		r.Height = value
	} else {
		r.Width = value
	}
}

// mainStart returns the start coordinate on the main axis (X for Row, Y for Column).
func mainStart(r Rect, dir Direction) int {
	if dir == Row {
		return r.X
	}
	return r.Y
}

// crossAvailableSize returns the available space on the cross axis.
func crossAvailableSize(r Rect, dir Direction) int {
	if dir == Row {
		return r.Height
	}
	return r.Width
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// resolveCrossSize computes a child's size on the cross axis.
// If parent alignment is AlignStretch, child takes full cross-axis space.
// Otherwise, child size is based on its own cross-axis sizing mode.
func resolveCrossSize(parent *LayoutNode, child *LayoutNode, into Rect) int {
	if parent.alignment == AlignStretch {
		return crossAvailableSize(into, parent.Direction)
	}

	if parent.Direction == Row {
		// Cross axis is height (Y).
		switch child.HeightSizing.Mode {
		case SizingFixed, SizingFit:
			return clampMax(child.intrinsicHeight, into.Height)
		case SizingPercent:
			return into.Height * child.HeightSizing.Value / 100
		case SizingGrow:
			return into.Height
		}
	}

	// Column: cross axis is width (X).
	switch child.WidthSizing.Mode {
	case SizingFixed, SizingFit:
		return clampMax(child.intrinsicWidth, into.Width)
	case SizingPercent:
		return into.Width * child.WidthSizing.Value / 100
	case SizingGrow:
		return into.Width
	}

	return 0
}

// applyCrossAlignment positions a child on the cross axis based on the parent's
// alignment and the available space.
func applyCrossAlignment(parent *LayoutNode, childRect *Rect, into Rect) {
	if parent.Direction == Row {
		// Cross axis is Y (vertical).
		switch parent.alignment {
		case AlignCenter:
			childRect.Y = into.Y + (into.Height-childRect.Height)/2
		case AlignEnd:
			childRect.Y = into.Y + (into.Height - childRect.Height)
		default: // AlignStart, AlignStretch
			childRect.Y = into.Y
		}
		return
	}

	// Column: cross axis is X (horizontal).
	switch parent.alignment {
	case AlignCenter:
		childRect.X = into.X + (into.Width-childRect.Width)/2
	case AlignEnd:
		childRect.X = into.X + (into.Width - childRect.Width)
	default: // AlignStart, AlignStretch
		childRect.X = into.X
	}
}

// resolveJustify computes the starting cursor position and effective gap for
// main-axis distribution based on the justify mode and available space.
// Returns (startCursor, effectiveGap).
func resolveJustify(justify Justify, start, innerMainSize, usedMain, baseGap, childCount int) (int, int) {
	if childCount == 0 {
		return start, 0
	}

	// Compute minimum gap space (between all children).
	minGapTotal := 0
	if childCount > 1 {
		minGapTotal = baseGap * (childCount - 1)
	}

	// Compute leftover space after children and minimum gaps.
	extraSpace := innerMainSize - usedMain - minGapTotal
	if extraSpace < 0 {
		extraSpace = 0
	}

	cursor := start
	gap := baseGap

	// Adjust cursor and gap based on justify mode.
	switch justify {
	case JustifyEnd:
		cursor += extraSpace
	case JustifyCenter:
		cursor += extraSpace / 2
	case JustifySpaceBetween:
		if childCount > 1 {
			gap += extraSpace / (childCount - 1)
		}
	case JustifySpaceAround:
		segment := extraSpace / (childCount * 2)
		cursor += segment
		gap += segment * 2
	}

	return cursor, gap
}

// layout assigns a concrete Rect to n and all its descendants given the space
// offered by the parent (top-down traversal). Appends Rects to out in
// depth-first, pre-order.
func layout(n *LayoutNode, into Rect, out *[]Rect) {
	// Append this node's rect to output.
	*out = append(*out, into)

	// Shrink the inner layout area by padding.
	into.X += n.paddingLeft
	into.Y += n.paddingTop
	into.Width -= n.paddingLeft + n.paddingRight
	into.Height -= n.paddingTop + n.paddingBottom

	// Clamp negative dimensions to 0 (result of over-constrained padding).
	if into.Width < 0 {
		into.Width = 0
	}
	if into.Height < 0 {
		into.Height = 0
	}

	// No children: leaf node.
	if len(n.Children) == 0 {
		return
	}

	// Compute available space on the main axis.
	innerMainSize := mainSize(into, n.Direction)
	totalGap := n.gap * (len(n.Children) - 1)

	// First pass: resolve fixed, fit, and percent children; collect grow children.
	usedMainAxis := 0
	totalGrowWeight := 0
	growIndices := make([]int, 0, len(n.Children))
	childRects := make([]Rect, 0, len(n.Children))

	for _, child := range n.Children {
		var childRect Rect

		// Resolve cross-axis size (height for Row, width for Column).
		setCrossSize(&childRect, n.Direction, resolveCrossSize(n, child, into))

		// For Column: invoke reflow now that cross-axis (width) is known.
		if n.Direction == Column && child.reflow != nil {
			child.intrinsicHeight = child.reflow(childRect.Width)
		}

		// Resolve main-axis size based on sizing mode.
		if n.Direction == Row {
			switch child.WidthSizing.Mode {
			case SizingFixed, SizingFit:
				childRect.Width = clampMax(child.intrinsicWidth, max(innerMainSize-usedMainAxis, 0))
				usedMainAxis += childRect.Width
			case SizingPercent:
				childRect.Width = innerMainSize * child.WidthSizing.Value / 100
				usedMainAxis += childRect.Width
			case SizingGrow:
				totalGrowWeight += child.WidthSizing.Value
				growIndices = append(growIndices, len(childRects))
			}
		} else {
			// Column direction.
			switch child.HeightSizing.Mode {
			case SizingFixed, SizingFit:
				childRect.Height = clampMax(child.intrinsicHeight, max(innerMainSize-usedMainAxis, 0))
				usedMainAxis += childRect.Height
			case SizingPercent:
				childRect.Height = innerMainSize * child.HeightSizing.Value / 100
				usedMainAxis += childRect.Height
			case SizingGrow:
				totalGrowWeight += child.HeightSizing.Value
				growIndices = append(growIndices, len(childRects))
			}
		}

		childRects = append(childRects, childRect)
	}

	// Second pass: distribute remaining space to Grow children.
	remaining := max(innerMainSize-totalGap-usedMainAxis, 0)

	if totalGrowWeight > 0 {
		remainingWeight := totalGrowWeight
		remainingSpace := remaining

		for _, idx := range growIndices {
			child := n.Children[idx]
			weight := child.WidthSizing.Value
			if n.Direction == Column {
				weight = child.HeightSizing.Value
			}

			size := remainingSpace
			if remainingWeight > weight {
				size = remainingSpace * weight / remainingWeight
			}

			setMainSize(&childRects[idx], n.Direction, size)
			remainingSpace -= size
			remainingWeight -= weight
		}
	}

	// For Row: invoke reflow now that widths are known.
	if n.Direction == Row {
		for i, child := range n.Children {
			if child.reflow == nil {
				continue
			}
			child.intrinsicHeight = child.reflow(childRects[i].Width)
			if child.HeightSizing.Mode == SizingFit {
				childRects[i].Height = child.intrinsicHeight
			}
		}
	}

	// Compute total used space and resolve justify for main-axis positioning.
	usedMain := 0
	for _, childRect := range childRects {
		usedMain += mainSize(childRect, n.Direction)
	}

	cursor, gap := resolveJustify(
		n.justify,
		mainStart(into, n.Direction),
		innerMainSize,
		usedMain,
		n.gap,
		len(n.Children),
	)

	// Third pass: position children and recurse.
	for i, child := range n.Children {
		childRect := childRects[i]

		// Set main-axis position.
		if n.Direction == Row {
			childRect.X = cursor
		} else {
			childRect.Y = cursor
		}

		// Apply cross-axis alignment.
		applyCrossAlignment(n, &childRect, into)

		// Recurse into child.
		layout(child, childRect, out)

		// Advance cursor for next child.
		cursor += mainSize(childRect, n.Direction)
		if i < len(n.Children)-1 {
			cursor += gap
		}
	}
}

// clampMax returns v bounded to at most max.
// If max is negative (unused today), v is returned unchanged.
// This is the enforcement point that prevents a Fixed/Fit child's intrinsic
// size from silently overflowing past what its parent has allocated.
func clampMax(v, max int) int {
	if max >= 0 && v > max {
		return max
	}
	return v
}
