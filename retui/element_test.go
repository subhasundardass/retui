package retui

import (
	"testing"
)

// ─── TestProps ──────────────────────────────────────────────────────────────

func TestPropsDefault(t *testing.T) {
	p := Props{}
	if p.Direction != 0 {
		t.Errorf("Expected Direction 0, got %v", p.Direction)
	}
	if p.Gap != 0 {
		t.Errorf("Expected Gap 0, got %d", p.Gap)
	}
	if p.Padding != [4]int{0, 0, 0, 0} {
		t.Errorf("Expected Padding [0,0,0,0], got %v", p.Padding)
	}
	if p.Width != (Sizing{}) {
		t.Errorf("Expected Width zero, got %v", p.Width)
	}
	if p.Height != (Sizing{}) {
		t.Errorf("Expected Height zero, got %v", p.Height)
	}
}

func TestPropsGet(t *testing.T) {
	p := Props{
		Values: map[string]any{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		},
	}

	if val := p.Get("key1"); val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
	if val := p.Get("key2"); val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}
	if val := p.Get("key3"); val != true {
		t.Errorf("Expected true, got %v", val)
	}
	if val := p.Get("nonexistent"); val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}

func TestPropsGetWithNilValues(t *testing.T) {
	p := Props{}
	if val := p.Get("key"); val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}

// ─── TestElement ────────────────────────────────────────────────────────────

func TestElementDefault(t *testing.T) {
	e := Element{}
	if e.Type != 0 {
		t.Errorf("Expected Type 0, got %v", e.Type)
	}
	if e.Text != "" {
		t.Errorf("Expected empty Text, got %q", e.Text)
	}
	if e.Children != nil {
		t.Errorf("Expected nil Children, got %v", e.Children)
	}
	if e.Style != (Style{}) {
		t.Errorf("Expected zero Style, got %v", e.Style)
	}
	if e.Id != "" {
		t.Errorf("Expected empty Id, got %q", e.Id)
	}
	if e.Key != "" {
		t.Errorf("Expected empty Key, got %q", e.Key)
	}
}

func TestElementWithID(t *testing.T) {
	e := Element{
		Id:  "test-id",
		Key: "test-key",
	}
	if e.Id != "test-id" {
		t.Errorf("Expected 'test-id', got %q", e.Id)
	}
	if e.Key != "test-key" {
		t.Errorf("Expected 'test-key', got %q", e.Key)
	}
}

func TestElementWithText(t *testing.T) {
	e := Element{
		Text: "Hello, World!",
	}
	if e.Text != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got %q", e.Text)
	}
}

func TestElementWithStyle(t *testing.T) {
	style := NewStyle().Bold(true).Foreground(Cyan)
	e := Element{
		Style: style,
	}
	if !e.Style.IsBold() {
		t.Error("Expected Bold true")
	}
	//Access the foreground field directly (if public)
	if e.Style.foreground != Cyan {
		t.Errorf("Expected Foreground Cyan, got %v", e.Style.foreground)
	}
}

func TestElementWithChildren(t *testing.T) {
	child1 := Element{Text: "child1"}
	child2 := Element{Text: "child2"}
	e := Element{
		Children: []Element{child1, child2},
	}
	if len(e.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(e.Children))
	}
	if e.Children[0].Text != "child1" {
		t.Errorf("Expected child1 text 'child1', got %q", e.Children[0].Text)
	}
	if e.Children[1].Text != "child2" {
		t.Errorf("Expected child2 text 'child2', got %q", e.Children[1].Text)
	}
}

// ─── TestPropsPadding ──────────────────────────────────────────────────────

func TestPropsPadding(t *testing.T) {
	p := Props{
		Padding: [4]int{1, 2, 3, 4},
	}
	if p.Padding[0] != 1 || p.Padding[1] != 2 || p.Padding[2] != 3 || p.Padding[3] != 4 {
		t.Errorf("Expected [1,2,3,4], got %v", p.Padding)
	}
}

// ─── TestLayoutProps ────────────────────────────────────────────────────────

func TestLayoutPropsDefault(t *testing.T) {
	lp := LayoutProps{}
	if lp.Direction != 0 {
		t.Errorf("Expected Direction 0, got %v", lp.Direction)
	}
	if lp.WidthSizing != (Sizing{}) {
		t.Errorf("Expected zero WidthSizing, got %v", lp.WidthSizing)
	}
	if lp.HeightSizing != (Sizing{}) {
		t.Errorf("Expected zero HeightSizing, got %v", lp.HeightSizing)
	}
	if lp.PaddingTop != 0 || lp.PaddingRight != 0 || lp.PaddingBottom != 0 || lp.PaddingLeft != 0 {
		t.Errorf("Expected all padding 0, got %v", lp)
	}
	if lp.Gap != 0 {
		t.Errorf("Expected Gap 0, got %d", lp.Gap)
	}
	if lp.Align != 0 {
		t.Errorf("Expected Align 0, got %v", lp.Align)
	}
	if lp.Justify != 0 {
		t.Errorf("Expected Justify 0, got %v", lp.Justify)
	}
}

// ─── TestMarkdownContent ──────────────────────────────────────────────────

func TestMarkdownContentDefault(t *testing.T) {
	mc := MarkdownContent{}
	if mc.Lines != nil {
		t.Errorf("Expected nil Lines, got %v", mc.Lines)
	}
}

// ─── TestElementWithMarkdown ─────────────────────────────────────────────

func TestElementWithMarkdown(t *testing.T) {
	e := Element{
		MarkdownText: "# Heading",
		Markdown: MarkdownContent{
			Lines: []markdownLine{},
		},
	}
	if e.MarkdownText != "# Heading" {
		t.Errorf("Expected '# Heading', got %q", e.MarkdownText)
	}
	if e.Markdown.Lines == nil {
		t.Error("Expected non-nil Lines")
	}
}

// ─── TestOverlay ──────────────────────────────────────────────────────────

func TestElementOverlay(t *testing.T) {
	e := Element{
		Type:     ElementOverlay,
		OverlayX: 10,
		OverlayY: 5,
	}
	if e.Type != ElementOverlay {
		t.Errorf("Expected ElementOverlay, got %v", e.Type)
	}
	if e.OverlayX != 10 {
		t.Errorf("Expected OverlayX 10, got %d", e.OverlayX)
	}
	if e.OverlayY != 5 {
		t.Errorf("Expected OverlayY 5, got %d", e.OverlayY)
	}
}

// ─── TestSizing ────────────────────────────────────────────────────────────

func TestSizing(t *testing.T) {
	fixed := Fixed(42)
	if fixed.Mode != SizingFixed {
		t.Errorf("Expected SizingFixed, got %v", fixed.Mode)
	}
	if fixed.Value != 42 {
		t.Errorf("Expected Value 42, got %d", fixed.Value)
	}

	grow := Grow(3)
	if grow.Mode != SizingGrow {
		t.Errorf("Expected SizingGrow, got %v", grow.Mode)
	}
	if grow.Value != 3 {
		t.Errorf("Expected Value 3, got %d", grow.Value)
	}

	fit := Fit()
	if fit.Mode != SizingFit {
		t.Errorf("Expected SizingFit, got %v", fit.Mode)
	}
	if fit.Value != 0 {
		t.Errorf("Expected Value 0, got %d", fit.Value)
	}
}

// ─── TestSizingZeroValue ──────────────────────────────────────────────────

func TestSizingZeroValue(t *testing.T) {
	s := Sizing{}
	if s.Mode != 0 {
		t.Errorf("Expected Mode 0, got %v", s.Mode)
	}
	if s.Value != 0 {
		t.Errorf("Expected Value 0, got %d", s.Value)
	}
}

// ─── Benchmarks ─────────────────────────────────────────────────────────────

func BenchmarkPropsGet(b *testing.B) {
	p := Props{
		Values: map[string]any{
			"key": "value",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Get("key")
	}
}

func BenchmarkElementCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Element{
			Id:    "test",
			Text:  "hello",
			Style: NewStyle().Bold(true).Foreground(Cyan),
			Children: []Element{
				{Text: "child1"},
				{Text: "child2"},
			},
		}
	}
}
