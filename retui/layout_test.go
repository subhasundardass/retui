package retui

import "testing"

// --- Grow distribution -------------------------------------------------

func TestGrowWeightDistribution(t *testing.T) {
	// Two children, Grow(2) and Grow(1), inside a Fixed(30) Row.
	// Expect a 20/10 split (2:1), computed via the remaining-weight
	// peel-off in layout() rather than a naive upfront ratio.
	root := NewLayout().
		WithDirection(Row).
		WithChildren(
			NewLayout().WithSize(Grow(2), Fit()),
			NewLayout().WithSize(Grow(1), Fit()),
		)

	rects := ComputeLayout(root, Rect{X: 0, Y: 0, Width: 30, Height: 5})

	// rects[0] = root, rects[1] = child0, rects[2] = child1 (pre-order).
	if len(rects) != 3 {
		t.Fatalf("expected 3 rects (root + 2 children), got %d", len(rects))
	}
	if got := rects[1].Width; got != 20 {
		t.Errorf("Grow(2) child width = %d, want 20", got)
	}
	if got := rects[2].Width; got != 10 {
		t.Errorf("Grow(1) child width = %d, want 10", got)
	}
}

func TestGrowDistributionNoRoundingLossAtLastChild(t *testing.T) {
	// Three equal-weight Grow(1) children over a width that doesn't divide
	// evenly (31 / 3). Confirms the peel-off approach doesn't dump all the
	// remainder onto one child unpredictably, and that widths sum to the
	// full available space (no leaked cells).
	root := NewLayout().
		WithDirection(Row).
		WithChildren(
			NewLayout().WithSize(Grow(1), Fit()),
			NewLayout().WithSize(Grow(1), Fit()),
			NewLayout().WithSize(Grow(1), Fit()),
		)

	rects := ComputeLayout(root, Rect{Width: 31, Height: 3})

	sum := rects[1].Width + rects[2].Width + rects[3].Width
	if sum != 31 {
		t.Errorf("child widths sum to %d, want 31 (no leaked/duplicated cells)", sum)
	}
}

// --- Justify -------------------------------------------------------------

func TestJustifySpaceBetweenPacksEdges(t *testing.T) {
	root := NewLayout().
		WithDirection(Row).
		WithJustify(JustifySpaceBetween).
		WithChildren(
			NewLayout().WithSize(Fixed(2), Fixed(1)),
			NewLayout().WithSize(Fixed(2), Fixed(1)),
		)

	rects := ComputeLayout(root, Rect{Width: 20, Height: 3})

	if rects[1].X != 0 {
		t.Errorf("first child should hug start (x=0), got x=%d", rects[1].X)
	}
	if want := 18; rects[2].X != want { // 20 - width(2)
		t.Errorf("last child should hug end (x=%d), got x=%d", want, rects[2].X)
	}
}

func TestJustifySpaceBetweenIntegerDivisionSlack(t *testing.T) {
	// 23-wide row, 3 children of width 2 each. extraSpace = 23-6 = 17,
	// divided by (childCount-1)=2 gaps -> gap=8 (17/2, remainder 1 lost).
	// This means the last child's right edge lands one cell short of the
	// container's right edge (22, not 23) instead of hugging flush.
	//
	// This test PINS the current behavior, it doesn't assert it's correct.
	// If the algorithm changes to distribute remainder cells to the last
	// gap(s) instead of dropping them, this test's expected values need
	// updating — that would be a deliberate, desirable fix.
	root := NewLayout().
		WithDirection(Row).
		WithJustify(JustifySpaceBetween).
		WithChildren(
			NewLayout().WithSize(Fixed(2), Fixed(1)),
			NewLayout().WithSize(Fixed(2), Fixed(1)),
			NewLayout().WithSize(Fixed(2), Fixed(1)),
		)

	rects := ComputeLayout(root, Rect{Width: 23, Height: 3})

	lastRightEdge := rects[3].X + rects[3].Width
	if lastRightEdge != 22 {
		t.Errorf("last child right edge = %d, want 22 (current rounding behavior); "+
			"if this now equals 23, the rounding fix landed — update this test to assert 23 and delete this comment",
			lastRightEdge)
	}
}

// --- Align / stretch -----------------------------------------------------

func TestAlignStretchOverridesExplicitChildCrossSize(t *testing.T) {
	// Default alignment is AlignStretch (zero value). A child with an
	// explicit Fixed height inside a Row still gets stretched to the
	// parent's full cross-axis size, because resolveCrossSize checks
	// parent.alignment == AlignStretch BEFORE looking at the child's own
	// HeightSizing mode.
	//
	// This diverges from CSS flexbox, where align-items: stretch only
	// applies to children with an auto (unspecified) cross size — an
	// explicit height wins over stretch there. This test documents retui's
	// current (different) behavior so it can't regress silently, and so a
	// future decision to match CSS semantics is a deliberate, visible diff.
	root := NewLayout().
		WithDirection(Row).
		WithChildren(
			NewLayout().WithSize(Fixed(3), Fixed(3)), // explicit height=3
		)

	rects := ComputeLayout(root, Rect{Width: 10, Height: 10})

	if got := rects[1].Height; got != 10 {
		t.Errorf("stretched child height = %d, want 10 (parent's full cross size); "+
			"if this now equals 3, stretch behavior changed to respect explicit sizing — "+
			"update this test and note it in CHANGELOG.md as a behavior change", got)
	}
}

func TestAlignStartRespectsExplicitChildCrossSize(t *testing.T) {
	// Sanity check contrasting the stretch case above: with AlignStart,
	// the explicit Fixed(3) height should be honored, not overridden.
	root := NewLayout().
		WithDirection(Row).
		WithAlign(AlignStart).
		WithChildren(
			NewLayout().WithSize(Fixed(3), Fixed(3)),
		)

	rects := ComputeLayout(root, Rect{Width: 10, Height: 10})

	if got := rects[1].Height; got != 3 {
		t.Errorf("AlignStart child height = %d, want 3 (explicit size honored)", got)
	}
}

// --- Zero-value Sizing risk ----------------------------------------------

func TestZeroValueSizingDefaultsToFixedZero(t *testing.T) {
	// SizingFixed is iota 0, so a bare Sizing{} (or a LayoutNode built
	// without WithSize) silently means "Fixed(0)", not "Fit()" — even
	// though DOCS.md documents Fit() as Box's default sizing. If the
	// Props->LayoutNode conversion (wherever that lives, likely node.go
	// or elements.go) ever constructs a LayoutNode without explicitly
	// calling WithSize, the box collapses to zero size instead of hugging
	// its content. This test pins the raw layout.go behavior; a
	// corresponding test belongs in node_test.go / elements_test.go
	// confirming the conversion always sets WithSize explicitly.
	var s Sizing
	if s.Mode != SizingFixed || s.Value != 0 {
		t.Fatalf("zero-value Sizing{} = %+v, want {SizingFixed 0}", s)
	}

	n := &LayoutNode{}
	if n.WidthSizing.Mode != SizingFixed || n.HeightSizing.Mode != SizingFixed {
		t.Fatalf("zero-value LayoutNode has Fixed(0) sizing on both axes, not Fit() — "+
			"got WidthSizing=%+v HeightSizing=%+v", n.WidthSizing, n.HeightSizing)
	}
}

// --- Reflow timing (Row vs Column) ---------------------------------------

func TestReflowFiresBeforeGrowResolutionInColumn(t *testing.T) {
	// In a Column, width is the cross axis and is known immediately via
	// resolveCrossSize — reflow should fire in the per-child loop, before
	// any grow distribution happens. This pins WHEN reflow is called
	// relative to the rest of the pass.
	var gotWidth int
	leaf := NewLayout().WithSize(Grow(1), Fit())
	leaf.reflow = func(crossSize int) int {
		gotWidth = crossSize
		return 2 // pretend this content wraps to 2 lines
	}

	root := NewLayout().
		WithDirection(Column).
		WithSize(Fixed(15), Fit()).
		WithChildren(leaf)

	_ = ComputeLayout(root, Rect{Width: 15, Height: 10})

	if gotWidth != 15 {
		t.Errorf("reflow received cross width %d, want 15 (parent's resolved width)", gotWidth)
	}
}

func TestReflowFiresAfterGrowResolutionInRow(t *testing.T) {
	// In a Row, width is the main axis — a Grow child's width isn't known
	// until grow distribution runs. reflow must fire AFTER that, using the
	// actual allocated width, not 0 or the pre-grow intrinsic width.
	var gotWidth int
	leaf := NewLayout().WithSize(Grow(1), Fit())
	leaf.reflow = func(crossSize int) int {
		gotWidth = crossSize
		return 2
	}

	root := NewLayout().
		WithDirection(Row).
		WithChildren(
			NewLayout().WithSize(Fixed(10), Fixed(1)),
			leaf, // gets whatever's left after the Fixed(10) sibling
		)

	_ = ComputeLayout(root, Rect{Width: 30, Height: 10})

	if gotWidth != 20 { // 30 - 10
		t.Errorf("reflow received width %d, want 20 (post-grow-distribution width)", gotWidth)
	}
}
