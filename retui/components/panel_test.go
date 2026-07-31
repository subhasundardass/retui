package components

import (
	"reflect"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestPanelBuilder_Defaults(t *testing.T) {
	p := Panel()

	if p.width != retui.Grow(1) {
		t.Errorf("expected width Grow(1), got %v", p.width)
	}
	if p.fixedWidth != 0 {
		t.Errorf("expected fixedWidth 0, got %d", p.fixedWidth)
	}
	if p.isFixed {
		t.Errorf("expected isFixed false, got %v", p.isFixed)
	}
	if p.contentGap != 0 {
		t.Errorf("expected contentGap 0, got %d", p.contentGap)
	}
	if p.style != nil {
		t.Errorf("expected style nil, got %v", p.style)
	}
	if p.hasHeader {
		t.Errorf("expected hasHeader false, got %v", p.hasHeader)
	}
	if len(p.children) != 0 {
		t.Errorf("expected empty children, got %d", len(p.children))
	}
}

func TestPanelBuilder_Width(t *testing.T) {
	t.Run("sets Width with Grow sizing", func(t *testing.T) {
		p := Panel().Width(retui.Grow(2))
		if p.width != retui.Grow(2) {
			t.Errorf("expected width Grow(2), got %v", p.width)
		}
		if p.isFixed {
			t.Errorf("expected isFixed false, got %v", p.isFixed)
		}
	})

	t.Run("sets Width with Fixed sizing", func(t *testing.T) {
		p := Panel().Width(retui.Fixed(50))
		if p.width != retui.Fixed(50) {
			t.Errorf("expected width Fixed(50), got %v", p.width)
		}
		if p.isFixed {
			t.Errorf("expected isFixed false, got %v", p.isFixed)
		}
	})
}

func TestPanelBuilder_FixedWidth(t *testing.T) {
	t.Run("sets FixedWidth correctly", func(t *testing.T) {
		p := Panel().FixedWidth(50)
		if p.width != retui.Fixed(50) {
			t.Errorf("expected width Fixed(50), got %v", p.width)
		}
		if p.fixedWidth != 50 {
			t.Errorf("expected fixedWidth 50, got %d", p.fixedWidth)
		}
		if !p.isFixed {
			t.Errorf("expected isFixed true, got %v", p.isFixed)
		}
	})

	t.Run("sets FixedWidth to 0", func(t *testing.T) {
		p := Panel().FixedWidth(0)
		if p.width != retui.Fixed(0) {
			t.Errorf("expected width Fixed(0), got %v", p.width)
		}
		if p.fixedWidth != 0 {
			t.Errorf("expected fixedWidth 0, got %d", p.fixedWidth)
		}
		if !p.isFixed {
			t.Errorf("expected isFixed true, got %v", p.isFixed)
		}
	})
}

func TestPanelBuilder_Height(t *testing.T) {
	t.Run("sets Height with Grow sizing", func(t *testing.T) {
		p := Panel().Height(retui.Grow(1))
		if p.height != retui.Grow(1) {
			t.Errorf("expected height Grow(1), got %v", p.height)
		}
		if p.isFixedHeight {
			t.Errorf("expected isFixedHeight false, got %v", p.isFixedHeight)
		}
	})

	t.Run("sets Height with Fixed sizing", func(t *testing.T) {
		p := Panel().Height(retui.Fixed(30))
		if p.height != retui.Fixed(30) {
			t.Errorf("expected height Fixed(30), got %v", p.height)
		}
		if p.isFixedHeight {
			t.Errorf("expected isFixedHeight false, got %v", p.isFixedHeight)
		}
	})
}

func TestPanelBuilder_FixedHeight(t *testing.T) {
	t.Run("sets FixedHeight correctly", func(t *testing.T) {
		p := Panel().FixedHeight(30)
		if p.height != retui.Fixed(30) {
			t.Errorf("expected height Fixed(30), got %v", p.height)
		}
		if p.fixedHeight != 30 {
			t.Errorf("expected fixedHeight 30, got %d", p.fixedHeight)
		}
		if !p.isFixedHeight {
			t.Errorf("expected isFixedHeight true, got %v", p.isFixedHeight)
		}
	})

	t.Run("sets FixedHeight to 0", func(t *testing.T) {
		p := Panel().FixedHeight(0)
		if p.height != retui.Fixed(0) {
			t.Errorf("expected height Fixed(0), got %v", p.height)
		}
		if p.fixedHeight != 0 {
			t.Errorf("expected fixedHeight 0, got %d", p.fixedHeight)
		}
		if !p.isFixedHeight {
			t.Errorf("expected isFixedHeight true, got %v", p.isFixedHeight)
		}
	})
}

func TestPanelBuilder_Style(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Cyan).Bold(true)
		p := Panel().Style(style)
		if p.style == nil {
			t.Error("expected style not nil")
		}
		if !reflect.DeepEqual(*p.style, style) {
			t.Errorf("Style not set correctly")
		}
	})
}

func TestPanelBuilder_Header(t *testing.T) {
	t.Run("sets Header correctly", func(t *testing.T) {
		header := retui.Text("Header", retui.NewStyle())
		p := Panel().Header(header)
		if !p.hasHeader {
			t.Error("expected hasHeader true, got false")
		}
		if !reflect.DeepEqual(p.header, header) {
			t.Error("Header not set correctly")
		}
	})
}

func TestPanelBuilder_ContentGap(t *testing.T) {
	t.Run("sets ContentGap correctly", func(t *testing.T) {
		p := Panel().ContentGap(2)
		if p.contentGap != 2 {
			t.Errorf("expected contentGap 2, got %d", p.contentGap)
		}
	})

	t.Run("sets ContentGap to 0", func(t *testing.T) {
		p := Panel().ContentGap(0)
		if p.contentGap != 0 {
			t.Errorf("expected contentGap 0, got %d", p.contentGap)
		}
	})
}

func TestPanelBuilder_Children(t *testing.T) {
	t.Run("adds children correctly", func(t *testing.T) {
		child1 := retui.Text("Child 1", retui.NewStyle())
		child2 := retui.Text("Child 2", retui.NewStyle())
		p := Panel().Children(child1, child2)

		if len(p.children) != 2 {
			t.Errorf("expected 2 children, got %d", len(p.children))
		}
		if !reflect.DeepEqual(p.children[0], child1) {
			t.Error("First child not set correctly")
		}
		if !reflect.DeepEqual(p.children[1], child2) {
			t.Error("Second child not set correctly")
		}
	})

	t.Run("ignores empty children", func(t *testing.T) {
		empty := retui.Element{}
		child := retui.Text("Child", retui.NewStyle())
		p := Panel().Children(empty, child)

		if len(p.children) != 1 {
			t.Errorf("expected 1 child, got %d", len(p.children))
		}
		if !reflect.DeepEqual(p.children[0], child) {
			t.Error("Child not set correctly")
		}
	})
}

func TestPanelBuilder_Divider(t *testing.T) {
	t.Run("adds Divider correctly", func(t *testing.T) {
		p := Panel().Divider()
		if len(p.children) != 1 {
			t.Errorf("expected 1 child, got %d", len(p.children))
		}
	})

	t.Run("adds Divider with FixedWidth", func(t *testing.T) {
		p := Panel().FixedWidth(20).Divider()
		if len(p.children) != 1 {
			t.Errorf("expected 1 child, got %d", len(p.children))
		}
	})
}

func TestPanelBuilder_DividerWithText(t *testing.T) {
	t.Run("adds DividerWithText correctly", func(t *testing.T) {
		p := Panel().DividerWithText("Section")
		if len(p.children) != 1 {
			t.Errorf("expected 1 child, got %d", len(p.children))
		}
	})

	t.Run("adds DividerWithText with FixedWidth", func(t *testing.T) {
		p := Panel().FixedWidth(20).DividerWithText("Section")
		if len(p.children) != 1 {
			t.Errorf("expected 1 child, got %d", len(p.children))
		}
	})

	t.Run("handles long text in DividerWithText", func(t *testing.T) {
		longText := "This is a very long section title that should be truncated"
		p := Panel().FixedWidth(10).DividerWithText(longText)
		if len(p.children) != 1 {
			t.Errorf("expected 1 child, got %d", len(p.children))
		}
	})
}

func TestPanelRender(t *testing.T) {
	t.Run("returns an Element of type ElementBox", func(t *testing.T) {
		p := Panel()
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with default style", func(t *testing.T) {
		p := Panel()
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with custom style", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Cyan).Bold(true)
		p := Panel().Style(style)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with header", func(t *testing.T) {
		header := retui.Text("Header", retui.NewStyle())
		p := Panel().Header(header)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with children", func(t *testing.T) {
		child := retui.Text("Child", retui.NewStyle())
		p := Panel().Children(child)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with FixedWidth", func(t *testing.T) {
		p := Panel().FixedWidth(30)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with FixedHeight", func(t *testing.T) {
		p := Panel().FixedHeight(20)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with Divider", func(t *testing.T) {
		p := Panel().Divider()
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with DividerWithText", func(t *testing.T) {
		p := Panel().DividerWithText("Section")
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestPanelChaining(t *testing.T) {
	style := retui.NewStyle().Foreground(retui.Cyan)
	header := retui.Text("Header", retui.NewStyle())
	child := retui.Text("Child", retui.NewStyle())

	p := Panel().
		Width(retui.Grow(2)).
		FixedWidth(50).
		Height(retui.Grow(1)).
		FixedHeight(30).
		Style(style).
		Header(header).
		ContentGap(2).
		Children(child).
		Divider().
		DividerWithText("Section")

	// Check that FixedWidth overrides Width
	if p.fixedWidth != 50 {
		t.Errorf("expected fixedWidth 50, got %d", p.fixedWidth)
	}
	if !p.isFixed {
		t.Errorf("expected isFixed true, got %v", p.isFixed)
	}

	// Check that FixedHeight overrides Height
	if p.fixedHeight != 30 {
		t.Errorf("expected fixedHeight 30, got %d", p.fixedHeight)
	}
	if !p.isFixedHeight {
		t.Errorf("expected isFixedHeight true, got %v", p.isFixedHeight)
	}

	if p.style == nil {
		t.Error("expected style not nil")
	}
	if !reflect.DeepEqual(*p.style, style) {
		t.Errorf("Style not set correctly")
	}
	if !p.hasHeader {
		t.Error("expected hasHeader true, got false")
	}
	if !reflect.DeepEqual(p.header, header) {
		t.Error("Header not set correctly")
	}
	if p.contentGap != 2 {
		t.Errorf("expected contentGap 2, got %d", p.contentGap)
	}
	if len(p.children) != 3 {
		t.Errorf("expected 3 children, got %d", len(p.children))
	}
}

func TestPanelHelperFunctions(t *testing.T) {
	t.Run("measureHeight measures Element height correctly", func(t *testing.T) {
		tests := []struct {
			name string
			el   retui.Element
			want int
		}{
			{"empty Element", retui.Element{}, 1},
			{"Text element", retui.Text("Hello", retui.NewStyle()), 1},
			{"Text with newline", retui.Text("Hello\nWorld", retui.NewStyle()), 2},
			{"Box with children", retui.Box(
				retui.Props{Direction: retui.Column},
				retui.NewStyle(),
				retui.Text("Line 1", retui.NewStyle()),
				retui.Text("Line 2", retui.NewStyle()),
			), 2},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := measureHeight(tt.el); got != tt.want {
					t.Errorf("measureHeight() = %d, want %d", got, tt.want)
				}
			})
		}
	})

	t.Run("buildVerticalBorder builds vertical border correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Cyan)

		tests := []struct {
			ch     string
			height int
			want   int // Expected number of children
		}{
			{"│", 1, 1},
			{"│", 3, 3},
			{"│", 5, 5},
		}
		for _, tt := range tests {
			elem := buildVerticalBorder(tt.ch, style, tt.height)
			if tt.height == 1 {
				if elem.Type != retui.ElementText {
					t.Errorf("expected ElementText for height 1, got %v", elem.Type)
				}
			} else {
				if elem.Type != retui.ElementBox {
					t.Errorf("expected ElementBox for height > 1, got %v", elem.Type)
				}
				if len(elem.Children) != tt.want {
					t.Errorf("expected %d children, got %d", tt.want, len(elem.Children))
				}
			}
		}
	})

	t.Run("getBorderStyle returns default style when nil", func(t *testing.T) {
		p := Panel()
		style := p.getBorderStyle()
		if reflect.DeepEqual(style, retui.Style{}) {
			t.Error("expected non-empty default style")
		}
	})

	t.Run("getBorderStyle returns custom style when set", func(t *testing.T) {
		customStyle := retui.NewStyle().Foreground(retui.Cyan).Bold(true)
		p := Panel().Style(customStyle)
		style := p.getBorderStyle()
		if !reflect.DeepEqual(style, customStyle) {
			t.Error("expected custom style")
		}
	})
}

func TestPanelEdgeCases(t *testing.T) {
	t.Run("handles FixedWidth with negative value", func(t *testing.T) {
		p := Panel().FixedWidth(-5)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles FixedHeight with negative value", func(t *testing.T) {
		p := Panel().FixedHeight(-5)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles many children", func(t *testing.T) {
		p := Panel()
		for i := 0; i < 50; i++ {
			p = p.Children(retui.Text("Child", retui.NewStyle()))
		}
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles empty header", func(t *testing.T) {
		p := Panel().Header(retui.Element{})
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles multiple dividers", func(t *testing.T) {
		p := Panel().
			Divider().
			Divider().
			DividerWithText("Middle").
			Divider()
		if len(p.children) != 4 {
			t.Errorf("expected 4 children, got %d", len(p.children))
		}
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles ContentGap with no children", func(t *testing.T) {
		p := Panel().ContentGap(2)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles FixedHeight with content overflow", func(t *testing.T) {
		p := Panel().FixedHeight(5)
		for i := 0; i < 10; i++ {
			p = p.Children(retui.Text("Line "+string(rune('A'+i)), retui.NewStyle()))
		}
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

// Benchmark tests
func BenchmarkPanelRender(b *testing.B) {
	p := Panel().
		FixedWidth(40).
		FixedHeight(20).
		Header(retui.Text("Header", retui.NewStyle())).
		Children(
			retui.Text("Line 1", retui.NewStyle()),
			retui.Text("Line 2", retui.NewStyle()),
			retui.Text("Line 3", retui.NewStyle()),
		).
		Divider().
		DividerWithText("Section")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Render()
	}
}

func BenchmarkPanelRenderLarge(b *testing.B) {
	p := Panel().FixedWidth(80).FixedHeight(50)
	for i := 0; i < 20; i++ {
		p = p.Children(retui.Text("Line "+string(rune('A'+i%26)), retui.NewStyle()))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Render()
	}
}

// Example usage test
func ExamplePanel() {
	// Create a panel with header and content
	panel := Panel().
		FixedWidth(40).
		FixedHeight(15).
		Header(retui.Text(" Panel Title ", retui.NewStyle().Bold(true))).
		Children(
			retui.Text("Content line 1", retui.NewStyle()),
			retui.Text("Content line 2", retui.NewStyle()),
		).
		Divider().
		DividerWithText("Section").
		Children(
			retui.Text("More content", retui.NewStyle()),
		).
		Render()
	_ = panel
}
