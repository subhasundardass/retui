package retui

import (
	"testing"
)

// available is the root rect used by every test unless stated otherwise.
var available = Rect{X: 0, Y: 0, Width: 80, Height: 24}

// ---- helpers ----------------------------------------------------------------

func rectsEqual(a, b Rect) bool {
	return a.X == b.X && a.Y == b.Y && a.Width == b.Width && a.Height == b.Height
}

func checkRects(t *testing.T, got, want []Rect) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d rects, want %d", len(got), len(want))
	}
	for i := range want {
		if !rectsEqual(got[i], want[i]) {
			t.Errorf("rect[%d]: got {X:%d Y:%d W:%d H:%d}, want {X:%d Y:%d W:%d H:%d}",
				i,
				got[i].X, got[i].Y, got[i].Width, got[i].Height,
				want[i].X, want[i].Y, want[i].Width, want[i].Height,
			)
		}
	}
}

// ---- 1. Fixed children stacking --------------------------------------------

func TestColumn_ThreeFixedChildren_StackVertically(t *testing.T) {
	// A column root with three Fixed(height=3) children.
	// Expected: children placed at y=0, y=3, y=6 each spanning full width.
	root := NewLayout().
		WithDirection(Column).
		WithSize(Fixed(80), Fixed(24)).
		WithChildren(
			NewLayout().WithSize(Grow(1), Fixed(3)),
			NewLayout().WithSize(Grow(1), Fixed(3)),
			NewLayout().WithSize(Grow(1), Fixed(3)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24}, // root
		{X: 0, Y: 0, Width: 80, Height: 3},  // child 0
		{X: 0, Y: 3, Width: 80, Height: 3},  // child 1
		{X: 0, Y: 6, Width: 80, Height: 3},  // child 2
	}
	checkRects(t, got, want)
}

func TestRow_ThreeFixedChildren_StackHorizontally(t *testing.T) {
	// A row root with three Fixed(width=10) children.
	// Expected: children placed at x=0, x=10, x=20 each spanning full height.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithChildren(
			NewLayout().WithSize(Fixed(10), Grow(1)),
			NewLayout().WithSize(Fixed(10), Grow(1)),
			NewLayout().WithSize(Fixed(10), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},  // root
		{X: 0, Y: 0, Width: 10, Height: 24},  // child 0
		{X: 10, Y: 0, Width: 10, Height: 24}, // child 1
		{X: 20, Y: 0, Width: 10, Height: 24}, // child 2
	}
	checkRects(t, got, want)
}

// ---- 2. Grow proportional split --------------------------------------------

func TestRow_FixedAndGrow_SplitsRemainingSpace(t *testing.T) {
	// Row with one Fixed(20) and one Grow(1) child inside 80 columns.
	// Remaining space after fixed = 60, all goes to Grow child.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithChildren(
			NewLayout().WithSize(Fixed(20), Grow(1)),
			NewLayout().WithSize(Grow(1), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 0, Y: 0, Width: 20, Height: 24},
		{X: 20, Y: 0, Width: 60, Height: 24},
	}
	checkRects(t, got, want)
}

func TestRow_TwoGrowChildren_EqualSplit(t *testing.T) {
	// Two Grow(1) children share 80 columns equally → each gets 40.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithChildren(
			NewLayout().WithSize(Grow(1), Grow(1)),
			NewLayout().WithSize(Grow(1), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 0, Y: 0, Width: 40, Height: 24},
		{X: 40, Y: 0, Width: 40, Height: 24},
	}
	checkRects(t, got, want)
}

func TestRow_WeightedGrowChildren_ProportionalSplit(t *testing.T) {
	// Grow(1) and Grow(2) share 90 columns → 30 and 60.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(90), Fixed(24)).
		WithChildren(
			NewLayout().WithSize(Grow(1), Grow(1)),
			NewLayout().WithSize(Grow(2), Grow(1)),
		)

	got := ComputeLayout(root, Rect{X: 0, Y: 0, Width: 90, Height: 24})

	want := []Rect{
		{X: 0, Y: 0, Width: 90, Height: 24},
		{X: 0, Y: 0, Width: 30, Height: 24},
		{X: 30, Y: 0, Width: 60, Height: 24},
	}
	checkRects(t, got, want)
}

func TestColumn_WeightedGrowChildren_ProportionalSplit(t *testing.T) {
	// Grow(1) and Grow(3) share 24 rows → 6 and 18.
	root := NewLayout().
		WithDirection(Column).
		WithSize(Fixed(80), Fixed(24)).
		WithChildren(
			NewLayout().WithSize(Grow(1), Grow(1)),
			NewLayout().WithSize(Grow(1), Grow(3)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 0, Y: 0, Width: 80, Height: 6},
		{X: 0, Y: 6, Width: 80, Height: 18},
	}
	checkRects(t, got, want)
}

// ---- 3. Nested flex composition --------------------------------------------

func TestNested_HeaderSidebarMain(t *testing.T) {
	// Classic terminal layout:
	//   Column root
	//     ├── header: full width, Fixed(3)
	//     └── body row: Grow(1)
	//           ├── sidebar: Fixed(20), full height
	//           └── main: Grow(1), full height
	root := NewLayout().
		WithDirection(Column).
		WithSize(Fixed(80), Fixed(24)).
		WithChildren(
			NewLayout().
				WithDirection(Row).
				WithSize(Grow(1), Fixed(3)),
			NewLayout().
				WithDirection(Row).
				WithSize(Grow(1), Grow(1)).
				WithChildren(
					NewLayout().WithSize(Fixed(20), Grow(1)),
					NewLayout().WithSize(Grow(1), Grow(1)),
				),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},  // root
		{X: 0, Y: 0, Width: 80, Height: 3},   // header
		{X: 0, Y: 3, Width: 80, Height: 21},  // body
		{X: 0, Y: 3, Width: 20, Height: 21},  // sidebar
		{X: 20, Y: 3, Width: 60, Height: 21}, // main
	}
	checkRects(t, got, want)
}

func TestNested_GrowInsideGrow_ProportionalSlice(t *testing.T) {
	// Outer row has Grow(1) and Grow(1) → each gets 40 cols.
	// Inner left column has Grow(1) and Grow(1) → each gets 12 rows (of 24).
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithChildren(
			NewLayout().
				WithDirection(Column).
				WithSize(Grow(1), Grow(1)).
				WithChildren(
					NewLayout().WithSize(Grow(1), Grow(1)),
					NewLayout().WithSize(Grow(1), Grow(1)),
				),
			NewLayout().WithSize(Grow(1), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},  // root
		{X: 0, Y: 0, Width: 40, Height: 24},  // left column
		{X: 0, Y: 0, Width: 40, Height: 12},  // left-top
		{X: 0, Y: 12, Width: 40, Height: 12}, // left-bottom
		{X: 40, Y: 0, Width: 40, Height: 24}, // right
	}
	checkRects(t, got, want)
}

// ---- 4. Padding ------------------------------------------------------------

func TestColumn_PaddingShrinkChildSpace(t *testing.T) {
	// Root column with padding 1 on all sides.
	// One Grow(1) child fills the inner area: 78×22.
	root := NewLayout().
		WithDirection(Column).
		WithSize(Fixed(80), Fixed(24)).
		WithPadding(1, 1, 1, 1).
		WithChildren(
			NewLayout().WithSize(Grow(1), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 1, Y: 1, Width: 78, Height: 22},
	}
	checkRects(t, got, want)
}

func TestColumn_AsymmetricPadding(t *testing.T) {
	// Padding: top=2, right=4, bottom=1, left=3.
	// Inner area: x=3, y=2, width=80-3-4=73, height=24-2-1=21.
	root := NewLayout().
		WithDirection(Column).
		WithSize(Fixed(80), Fixed(24)).
		WithPadding(2, 4, 1, 3).
		WithChildren(
			NewLayout().WithSize(Grow(1), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 3, Y: 2, Width: 73, Height: 21},
	}
	checkRects(t, got, want)
}

func TestRow_PaddingWithMultipleChildren(t *testing.T) {
	// Row with padding 1 all sides, two equal Grow children.
	// Inner width = 80-2 = 78, split equally → 39 each.
	// Inner height = 24-2 = 22.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithPadding(1, 1, 1, 1).
		WithChildren(
			NewLayout().WithSize(Grow(1), Grow(1)),
			NewLayout().WithSize(Grow(1), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 1, Y: 1, Width: 39, Height: 22},
		{X: 40, Y: 1, Width: 39, Height: 22},
	}
	checkRects(t, got, want)
}

// ---- 5. Gap ----------------------------------------------------------------

func TestColumn_GapBetweenChildren(t *testing.T) {
	// Column with gap=1, three Fixed(height=3) children.
	// Positions: y=0, y=4, y=8 (gap inserted between, not at edges).
	root := NewLayout().
		WithDirection(Column).
		WithSize(Fixed(80), Fixed(24)).
		WithGap(1).
		WithChildren(
			NewLayout().WithSize(Grow(1), Fixed(3)),
			NewLayout().WithSize(Grow(1), Fixed(3)),
			NewLayout().WithSize(Grow(1), Fixed(3)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 0, Y: 0, Width: 80, Height: 3},
		{X: 0, Y: 4, Width: 80, Height: 3},
		{X: 0, Y: 8, Width: 80, Height: 3},
	}
	checkRects(t, got, want)
}

func TestRow_GapBetweenChildren(t *testing.T) {
	// Row with gap=2, two Fixed(width=10) children.
	// Positions: x=0, x=12.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithGap(2).
		WithChildren(
			NewLayout().WithSize(Fixed(10), Grow(1)),
			NewLayout().WithSize(Fixed(10), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 0, Y: 0, Width: 10, Height: 24},
		{X: 12, Y: 0, Width: 10, Height: 24},
	}
	checkRects(t, got, want)
}

func TestColumn_GapReducesSpaceForGrowChildren(t *testing.T) {
	// Column gap=2, two Grow(1) children. Total height=24, gaps=2 (one gap
	// between two children), remaining for grow=22, split → 11 each.
	root := NewLayout().
		WithDirection(Column).
		WithSize(Fixed(80), Fixed(24)).
		WithGap(2).
		WithChildren(
			NewLayout().WithSize(Grow(1), Grow(1)),
			NewLayout().WithSize(Grow(1), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 0, Y: 0, Width: 80, Height: 11},
		{X: 0, Y: 13, Width: 80, Height: 11},
	}
	checkRects(t, got, want)
}

// ---- 6. Cross-axis alignment -----------------------------------------------

func TestRow_AlignCenter_NarrowChild(t *testing.T) {
	// Row, cross axis = vertical (height). Parent height=24, child Fixed(height=6).
	// AlignCenter → child y = (24-6)/2 = 9.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithAlign(AlignCenter).
		WithChildren(
			NewLayout().WithSize(Fixed(20), Fixed(6)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 0, Y: 9, Width: 20, Height: 6},
	}
	checkRects(t, got, want)
}

func TestRow_AlignEnd_NarrowChild(t *testing.T) {
	// AlignEnd → child y = 24-6 = 18.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithAlign(AlignEnd).
		WithChildren(
			NewLayout().WithSize(Fixed(20), Fixed(6)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 0, Y: 18, Width: 20, Height: 6},
	}
	checkRects(t, got, want)
}

func TestColumn_AlignCenter_NarrowChild(t *testing.T) {
	// Column, cross axis = horizontal (width). Parent width=80, child Fixed(width=20).
	// AlignCenter → child x = (80-20)/2 = 30.
	root := NewLayout().
		WithDirection(Column).
		WithSize(Fixed(80), Fixed(24)).
		WithAlign(AlignCenter).
		WithChildren(
			NewLayout().WithSize(Fixed(20), Fixed(6)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 30, Y: 0, Width: 20, Height: 6},
	}
	checkRects(t, got, want)
}

func TestRow_AlignStretch_ChildFillsCrossAxis(t *testing.T) {
	// AlignStretch (default): child height = parent height regardless of Fixed.
	// The child's cross-axis sizing is overridden to fill the parent.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithAlign(AlignStretch).
		WithChildren(
			NewLayout().WithSize(Fixed(20), Fixed(6)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 0, Y: 0, Width: 20, Height: 24},
	}
	checkRects(t, got, want)
}

// ---- 7. Justify (main-axis distribution) -----------------------------------

func TestRow_JustifyCenter_SingleChild(t *testing.T) {
	// Row JustifyCenter, one Fixed(20) child in 80 cols.
	// child x = (80-20)/2 = 30.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithJustify(JustifyCenter).
		WithChildren(
			NewLayout().WithSize(Fixed(20), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 30, Y: 0, Width: 20, Height: 24},
	}
	checkRects(t, got, want)
}

func TestRow_JustifyEnd_SingleChild(t *testing.T) {
	// JustifyEnd: child placed at x = 80-20 = 60.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithJustify(JustifyEnd).
		WithChildren(
			NewLayout().WithSize(Fixed(20), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 60, Y: 0, Width: 20, Height: 24},
	}
	checkRects(t, got, want)
}

func TestRow_JustifySpaceBetween_TwoChildren(t *testing.T) {
	// JustifySpaceBetween: two Fixed(10) children in 80 cols.
	// Used = 20, remaining = 60, one gap between → children at x=0 and x=70.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithJustify(JustifySpaceBetween).
		WithChildren(
			NewLayout().WithSize(Fixed(10), Grow(1)),
			NewLayout().WithSize(Fixed(10), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 0, Y: 0, Width: 10, Height: 24},
		{X: 70, Y: 0, Width: 10, Height: 24},
	}
	checkRects(t, got, want)
}

func TestRow_JustifySpaceAround_TwoChildren(t *testing.T) {
	// JustifySpaceAround: two Fixed(10) children in 80 cols.
	// Remaining = 60, split into 2*2=4 segments → each segment = 15.
	// child0 x = 15, child1 x = 15+10+30 = 55.
	root := NewLayout().
		WithDirection(Row).
		WithSize(Fixed(80), Fixed(24)).
		WithJustify(JustifySpaceAround).
		WithChildren(
			NewLayout().WithSize(Fixed(10), Grow(1)),
			NewLayout().WithSize(Fixed(10), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 15, Y: 0, Width: 10, Height: 24},
		{X: 55, Y: 0, Width: 10, Height: 24},
	}
	checkRects(t, got, want)
}

// ---- 8. No children --------------------------------------------------------

func TestLeafNode_NoChildren_ReturnsOnlyRoot(t *testing.T) {
	root := NewLayout().WithSize(Fixed(80), Fixed(24))
	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
	}
	checkRects(t, got, want)
}

// ---- 9. Padding + Gap combined --------------------------------------------

func TestColumn_PaddingAndGap_Combined(t *testing.T) {
	// Padding 1 all sides → inner area: x=1, y=1, w=78, h=22.
	// Gap=1 between two Grow(1) children → available h=22, gaps=1, remaining=21.
	// 21 is odd: floor(21/2)=10, remainder goes to last child → 10 and 11.
	// (Implementation note: standard approach gives first children the floor,
	//  last child absorbs the remainder — test reflects this.)
	root := NewLayout().
		WithDirection(Column).
		WithSize(Fixed(80), Fixed(24)).
		WithPadding(1, 1, 1, 1).
		WithGap(1).
		WithChildren(
			NewLayout().WithSize(Grow(1), Grow(1)),
			NewLayout().WithSize(Grow(1), Grow(1)),
		)

	got := ComputeLayout(root, available)

	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 24},
		{X: 1, Y: 1, Width: 78, Height: 10},
		{X: 1, Y: 12, Width: 78, Height: 11},
	}
	checkRects(t, got, want)
}

// ---- 10. Origin offset -----------------------------------------------------

func TestColumn_NonZeroOrigin_OffsetsAllRects(t *testing.T) {
	// Root available rect starts at x=5, y=3 (e.g. inside a parent panel).
	// All child rects should be offset accordingly.
	offset := Rect{X: 5, Y: 3, Width: 40, Height: 12}
	root := NewLayout().
		WithDirection(Column).
		WithSize(Grow(1), Grow(1)).
		WithChildren(
			NewLayout().WithSize(Grow(1), Fixed(4)),
			NewLayout().WithSize(Grow(1), Grow(1)),
		)

	got := ComputeLayout(root, offset)

	want := []Rect{
		{X: 5, Y: 3, Width: 40, Height: 12},
		{X: 5, Y: 3, Width: 40, Height: 4},
		{X: 5, Y: 7, Width: 40, Height: 8},
	}
	checkRects(t, got, want)
}
