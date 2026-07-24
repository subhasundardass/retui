package retui

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

type Direction int

const (
	Row Direction = iota
	Column
)

type SizingMode int

const (
	SizingFixed SizingMode = iota
	SizingGrow
	SizingFit
	SizingPercent
)

type Sizing struct {
	Mode  SizingMode
	Value int
}

type LayoutNode struct {
	Direction       Direction
	WidthSizing     Sizing
	HeightSizing    Sizing
	Children        []*LayoutNode
	intrinsicHeight int
	intrinsicWidth  int
	paddingTop      int
	paddingBottom   int
	paddingLeft     int
	paddingRight    int
	gap             int
	alignment       Alignment
	justify         Justify
	// reflow lets a node recompute its main-axis size once its cross-axis
	// size is known during the layout pass. Used by content whose height
	// depends on width (e.g. word-wrapped text). Currently invoked only when
	// the parent's Direction is Column.
	reflow func(crossSize int) int
}

func NewLayout() *LayoutNode {
	return &LayoutNode{}
}

func (l LayoutNode) WithDirection(dir Direction) *LayoutNode {
	l.Direction = dir
	return &l
}

func Fixed(n int) Sizing { return Sizing{Mode: SizingFixed, Value: n} }

func Grow(n int) Sizing { return Sizing{Mode: SizingGrow, Value: n} }

func Fit() Sizing { return Sizing{Mode: SizingFit} }

// Percent returns a Sizing that resolves to a percentage (0-100) of the
// parent's resolved size along that axis. If the parent's size along that
// axis is itself indefinite (e.g. a SizingFit parent trying to shrink-wrap
// its children), the percentage cannot be resolved during the intrinsic
// measure pass and is treated as 0 for that pass, mirroring how SizingGrow
// is handled — it contributes nothing to the parent's shrink-wrap size. The
// real value is resolved later during ComputeLayout, once the parent has a
// concrete Rect.
func Percent(n int) Sizing { return Sizing{Mode: SizingPercent, Value: n} }

func (l *LayoutNode) WithSize(w, h Sizing) *LayoutNode {
	l.WidthSizing = w
	l.HeightSizing = h
	return l
}

func (l *LayoutNode) WithChildren(children ...*LayoutNode) *LayoutNode {
	l.Children = children
	return l
}

func (l *LayoutNode) WithPadding(top, right, bottom, left int) *LayoutNode {
	l.paddingTop, l.paddingRight, l.paddingBottom, l.paddingLeft = top, right, bottom, left
	return l
}

func (l *LayoutNode) WithGap(value int) *LayoutNode {
	l.gap = value
	return l
}

type Alignment int

const (
	AlignStretch Alignment = iota
	AlignStart
	AlignCenter
	AlignEnd
)

func (l *LayoutNode) WithAlign(alignment Alignment) *LayoutNode {
	l.alignment = alignment
	return l
}

type Justify int

const (
	JustifyStart Justify = iota
	JustifyEnd
	JustifyCenter
	JustifySpaceBetween
	JustifySpaceAround
)

func (l *LayoutNode) WithJustify(value Justify) *LayoutNode {
	l.justify = value
	return l
}

// ============
// IntrinsicSize returns the natural (unconstrained) dimensions of a layout tree.
func IntrinsicSize(root *LayoutNode) (width, height int) {
	return measure(root)
}

// ComputeLayout runs the layout algorithm on the given tree and returns a
// flat slice of Rects in depth-first, pre-order (parent before children).
// The root is given the provided available Rect as its bounds.
//
// Nodes with a reflow callback (e.g. wrapped text) need a follow-up
// measure+layout pass: the first layout pass calls reflow once widths are
// known, updating leaf heights; the second measure picks up those heights
// and propagates them through ancestor boxes so that bordered/padded
// containers around wrapped text grow to fit their content.
func ComputeLayout(root *LayoutNode, available Rect) []Rect {
	root.intrinsicWidth, root.intrinsicHeight = measure(root)
	var out []Rect
	layout(root, available, &out)

	if hasReflow(root) {
		root.intrinsicWidth, root.intrinsicHeight = measure(root)
		out = out[:0]
		layout(root, available, &out)
	}
	return out
}

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
func measure(n *LayoutNode) (int, int) {
	for _, child := range n.Children {
		child.intrinsicWidth, child.intrinsicHeight = measure(child)
	}

	width, height := 0, 0

	gaps := 0
	if len(n.Children) > 1 {
		gaps = n.gap * (len(n.Children) - 1)
	}

	switch n.WidthSizing.Mode {
	case SizingFixed:
		width = n.WidthSizing.Value

	case SizingFit:
		if n.Direction == Row {
			for _, child := range n.Children {
				width += child.intrinsicWidth
			}
			width += gaps
		} else {
			for _, child := range n.Children {
				if child.intrinsicWidth > width {
					width = child.intrinsicWidth
				}
			}
		}

	case SizingGrow, SizingPercent:
		// Neither can be resolved without knowing the parent's actual
		// allocated size, which isn't available during this bottom-up
		// pass. Both contribute 0 here; SizingGrow is resolved during
		// layout() from remaining space, SizingPercent from the parent's
		// concrete Rect.
		width = 0
	}

	switch n.HeightSizing.Mode {
	case SizingFixed:
		height = n.HeightSizing.Value

	case SizingFit:
		if n.Direction != Row {
			for _, child := range n.Children {
				height += child.intrinsicHeight
			}
			height += gaps
		} else {
			for _, child := range n.Children {
				if child.intrinsicHeight > height {
					height = child.intrinsicHeight
				}
			}
		}

	case SizingGrow, SizingPercent:
		height = 0

	}

	if n.WidthSizing.Mode == SizingFit {
		width += n.paddingLeft + n.paddingRight
	}
	if n.HeightSizing.Mode == SizingFit {
		height += n.paddingTop + n.paddingBottom
	}

	// Preserve a previously-reflowed height. Reflow runs during the layout
	// pass with the actual allocated width — more accurate than anything
	// measure can derive from the (childless) reflow leaf itself. On a
	// second measure pass this keeps the wrapped line count visible to
	// ancestors so containers can grow.
	if n.reflow != nil && n.intrinsicHeight > height {
		height = n.intrinsicHeight
	}

	return width, height
}

func mainSize(r Rect, dir Direction) int {
	if dir == Row {
		return r.Width
	}
	return r.Height
}

func setMainSize(r *Rect, dir Direction, value int) {
	if dir == Row {
		r.Width = value
	} else {
		r.Height = value
	}
}

func setCrossSize(r *Rect, dir Direction, value int) {
	if dir == Row {
		r.Height = value
	} else {
		r.Width = value
	}
}

func mainStart(r Rect, dir Direction) int {
	if dir == Row {
		return r.X
	}
	return r.Y
}

func crossAvailableSize(r Rect, dir Direction) int {
	if dir == Row {
		return r.Height
	}
	return r.Width
}

// clampMax returns v bounded to at most max. A negative max means "no
// bound" (used where the available space isn't meaningfully limited, e.g.
// a root with no parent). This is the enforcement point that stops a
// SizingFixed/SizingFit child's intrinsic size from silently overflowing
// past what its parent actually has to offer.
func clampMax(v, max int) int {
	if max >= 0 && v > max {
		return max
	}
	return v
}

func resolveCrossSize(parent *LayoutNode, child *LayoutNode, into Rect) int {
	if parent.alignment == AlignStretch {
		return crossAvailableSize(into, parent.Direction)
	}

	if parent.Direction == Row {
		switch child.HeightSizing.Mode {
		case SizingFixed, SizingFit:
			return clampMax(child.intrinsicHeight, into.Height)
		case SizingPercent:
			return into.Height * child.HeightSizing.Value / 100
		case SizingGrow:
			return into.Height
		}
	}

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

func applyCrossAlignment(parent *LayoutNode, childRect *Rect, into Rect) {
	if parent.Direction == Row {
		switch parent.alignment {
		case AlignCenter:
			childRect.Y = into.Y + (into.Height-childRect.Height)/2
		case AlignEnd:
			childRect.Y = into.Y + (into.Height - childRect.Height)
		default:
			childRect.Y = into.Y
		}
		return
	}

	switch parent.alignment {
	case AlignCenter:
		childRect.X = into.X + (into.Width-childRect.Width)/2
	case AlignEnd:
		childRect.X = into.X + (into.Width - childRect.Width)
	default:
		childRect.X = into.X
	}
}

func resolveJustify(justify Justify, start, innerMainSize, usedMain, baseGap, childCount int) (int, int) {
	if childCount == 0 {
		return start, 0
	}

	minGapTotal := 0
	if childCount > 1 {
		minGapTotal = baseGap * (childCount - 1)
	}

	extraSpace := innerMainSize - usedMain - minGapTotal
	if extraSpace < 0 {
		extraSpace = 0
	}

	cursor := start
	gap := baseGap

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

// layout assigns a concrete Rect to n given the space offered by its parent,
// then recurses into children (top-down).
func layout(n *LayoutNode, into Rect, out *[]Rect) {
	*out = append(*out, into)

	into.X += n.paddingLeft
	into.Y += n.paddingTop
	into.Width -= n.paddingLeft + n.paddingRight
	into.Height -= n.paddingTop + n.paddingBottom

	if into.Width < 0 {
		into.Width = 0
	}
	if into.Height < 0 {
		into.Height = 0
	}

	if len(n.Children) == 0 {
		return
	}

	innerMainSize := mainSize(into, n.Direction)
	totalGap := n.gap * (len(n.Children) - 1)
	usedByFixedAndFit := 0
	totalGrowWeight := 0
	growIndices := make([]int, 0, len(n.Children))
	childRects := make([]Rect, 0, len(n.Children))

	for _, child := range n.Children {
		var childRect Rect

		setCrossSize(&childRect, n.Direction, resolveCrossSize(n, child, into))

		if n.Direction == Column && child.reflow != nil {
			child.intrinsicHeight = child.reflow(childRect.Width)
		}

		if n.Direction == Row {
			switch child.WidthSizing.Mode {
			case SizingFixed, SizingFit:
				// Clamp: a Fixed/Fit child's intrinsic width must not
				// exceed what the parent actually has left to give on the
				// main axis, or it silently overflows the parent's bounds
				// (e.g. a wide table rendered inside a narrower Box).
				childRect.Width = clampMax(child.intrinsicWidth, max(innerMainSize-usedByFixedAndFit, 0))
				usedByFixedAndFit += childRect.Width
			case SizingPercent:
				childRect.Width = innerMainSize * child.WidthSizing.Value / 100
				usedByFixedAndFit += childRect.Width
			case SizingGrow:
				totalGrowWeight += child.WidthSizing.Value
				growIndices = append(growIndices, len(childRects))
			}
		} else {
			switch child.HeightSizing.Mode {
			case SizingFixed, SizingFit:
				childRect.Height = clampMax(child.intrinsicHeight, max(innerMainSize-usedByFixedAndFit, 0))
				usedByFixedAndFit += childRect.Height
			case SizingPercent:
				childRect.Height = innerMainSize * child.HeightSizing.Value / 100
				usedByFixedAndFit += childRect.Height
			case SizingGrow:
				totalGrowWeight += child.HeightSizing.Value
				growIndices = append(growIndices, len(childRects))
			}
		}

		childRects = append(childRects, childRect)
	}

	remaining := max(innerMainSize-totalGap-usedByFixedAndFit, 0)

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

	// Row direction: child widths are only fully known once the grow
	// distribution above has run. Fire reflow callbacks now so wrapped
	// text inside a Row gets its line count from the allocated width
	// (the Column path handles this earlier in the per-child loop).
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

	for i, child := range n.Children {
		childRect := childRects[i]

		if n.Direction == Row {
			childRect.X = cursor
		} else {
			childRect.Y = cursor
		}

		applyCrossAlignment(n, &childRect, into)
		layout(child, childRect, out)

		cursor += mainSize(childRect, n.Direction)
		if i < len(n.Children)-1 {
			cursor += gap
		}
	}
}
