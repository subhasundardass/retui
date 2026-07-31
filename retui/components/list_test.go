package components

import (
	"reflect"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestListBuilder_Defaults(t *testing.T) {
	l := List()

	if l.config.ID != "" {
		t.Errorf("expected empty ID, got %q", l.config.ID)
	}
	if len(l.config.Items) > 0 {
		t.Errorf("expected empty Items, got %v", l.config.Items)
	}
	if l.config.Selected != 0 {
		t.Errorf("expected Selected 0, got %d", l.config.Selected)
	}
	if l.config.Width != 0 {
		t.Errorf("expected Width 0, got %d", l.config.Width)
	}
	if l.config.Prefix != "" {
		t.Errorf("expected empty Prefix, got %q", l.config.Prefix)
	}
	if l.config.Suffix != "" {
		t.Errorf("expected empty Suffix, got %q", l.config.Suffix)
	}
	if l.focused {
		t.Errorf("expected focused false, got %v", l.focused)
	}
}

func TestListBuilder_StringFields(t *testing.T) {
	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{"ID", func() string { return List().ID("test-id").config.ID }, "test-id"},
		{"Prefix", func() string { return List().Prefix("📋 ").config.Prefix }, "📋 "},
		{"Suffix", func() string { return List().Suffix(" ✓").config.Suffix }, " ✓"},
	}

	for _, tt := range tests {
		t.Run("sets "+tt.name+" correctly", func(t *testing.T) {
			if got := tt.got(); got != tt.want {
				t.Errorf("expected %s %q, got %q", tt.name, tt.want, got)
			}
		})
	}
}

func TestListBuilder_Items(t *testing.T) {
	t.Run("sets Items correctly", func(t *testing.T) {
		items := []string{"Apple", "Banana", "Orange"}
		l := List().Items(items)
		if !reflect.DeepEqual(l.config.Items, items) {
			t.Errorf("expected Items %v, got %v", items, l.config.Items)
		}
	})

	t.Run("sets empty Items correctly", func(t *testing.T) {
		items := []string{}
		l := List().Items(items)
		if !reflect.DeepEqual(l.config.Items, items) {
			t.Errorf("expected empty Items, got %v", l.config.Items)
		}
	})
}

func TestListBuilder_Selected(t *testing.T) {
	t.Run("sets Selected correctly", func(t *testing.T) {
		l := List().Selected(2)
		if l.config.Selected != 2 {
			t.Errorf("expected Selected 2, got %d", l.config.Selected)
		}
	})

	t.Run("sets Selected to 0 by default", func(t *testing.T) {
		l := List()
		if l.config.Selected != 0 {
			t.Errorf("expected Selected 0, got %d", l.config.Selected)
		}
	})
}

func TestListBuilder_Width(t *testing.T) {
	l := List().Width(20)
	if l.config.Width != 20 {
		t.Errorf("expected Width 20, got %d", l.config.Width)
	}
}

func TestListBuilder_Styles(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Background(retui.Blue)
		l := List().Style(style)
		if !reflect.DeepEqual(l.config.Style, style) {
			t.Errorf("Style not set correctly")
		}
	})

	t.Run("sets SelectedStyle correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Green).Bold(true)
		l := List().SelectedStyle(style)
		if !reflect.DeepEqual(l.config.SelectedStyle, style) {
			t.Errorf("SelectedStyle not set correctly")
		}
	})

	t.Run("sets UnselectedStyle correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.BrightBlack)
		l := List().UnselectedStyle(style)
		if !reflect.DeepEqual(l.config.UnselectedStyle, style) {
			t.Errorf("UnselectedStyle not set correctly")
		}
	})
}

func TestListBuilder_State(t *testing.T) {
	t.Run("sets Focused state correctly", func(t *testing.T) {
		l := List().Focused(true)
		if !l.focused {
			t.Errorf("expected focused true, got %v", l.focused)
		}
	})

	t.Run("sets Focused state to false", func(t *testing.T) {
		l := List().Focused(false)
		if l.focused {
			t.Errorf("expected focused false, got %v", l.focused)
		}
	})
}

func TestListCallbacks(t *testing.T) {
	t.Run("OnSelect callback is set and invoked correctly", func(t *testing.T) {
		var selected bool
		var capturedID string
		var capturedIndex int
		var capturedValue string
		l := List().ID("test-list").Items([]string{"A", "B", "C"}).OnSelect(func(id string, index int, value string) {
			selected = true
			capturedID = id
			capturedIndex = index
			capturedValue = value
		})

		if l.config.OnSelect == nil {
			t.Fatal("OnSelect callback not set")
		}
		l.config.OnSelect("test-list", 1, "B")
		if !selected {
			t.Error("OnSelect callback was not executed")
		}
		if capturedID != "test-list" {
			t.Errorf("expected id 'test-list', got %s", capturedID)
		}
		if capturedIndex != 1 {
			t.Errorf("expected index 1, got %d", capturedIndex)
		}
		if capturedValue != "B" {
			t.Errorf("expected value 'B', got %s", capturedValue)
		}
	})

	t.Run("OnKeyPress callback is set and invoked correctly", func(t *testing.T) {
		var keyPressed bool
		l := List().ID("test-list").OnKeyPress(func(id string, key retui.Key) bool {
			keyPressed = true
			if id != "test-list" {
				t.Errorf("expected id 'test-list', got %s", id)
			}
			return true
		})

		if l.config.OnKeyPress == nil {
			t.Fatal("OnKeyPress callback not set")
		}
		if result := l.config.OnKeyPress("test-list", retui.Key{Code: retui.KeyDown}); !result {
			t.Error("OnKeyPress should return true")
		}
		if !keyPressed {
			t.Error("OnKeyPress callback was not executed")
		}
	})

	t.Run("OnFocus callback is set and invoked correctly", func(t *testing.T) {
		var focused bool
		l := List().ID("test-list").OnFocus(func(id string) {
			focused = true
			if id != "test-list" {
				t.Errorf("expected id 'test-list', got %s", id)
			}
		})

		if l.config.OnFocus == nil {
			t.Fatal("OnFocus callback not set")
		}
		l.config.OnFocus("test-list")
		if !focused {
			t.Error("OnFocus callback was not executed")
		}
	})

	t.Run("OnBlur callback is set and invoked correctly", func(t *testing.T) {
		var blurred bool
		l := List().ID("test-list").OnBlur(func(id string) {
			blurred = true
			if id != "test-list" {
				t.Errorf("expected id 'test-list', got %s", id)
			}
		})

		if l.config.OnBlur == nil {
			t.Fatal("OnBlur callback not set")
		}
		l.config.OnBlur("test-list")
		if !blurred {
			t.Error("OnBlur callback was not executed")
		}
	})

	t.Run("OnSubmit callback is set and invoked correctly", func(t *testing.T) {
		var submitted bool
		var capturedID string
		var capturedIndex int
		var capturedValue string
		l := List().ID("test-list").Items([]string{"A", "B", "C"}).OnSubmit(func(id string, index int, value string) {
			submitted = true
			capturedID = id
			capturedIndex = index
			capturedValue = value
		})

		if l.config.OnSubmit == nil {
			t.Fatal("OnSubmit callback not set")
		}
		l.config.OnSubmit("test-list", 2, "C")
		if !submitted {
			t.Error("OnSubmit callback was not executed")
		}
		if capturedID != "test-list" {
			t.Errorf("expected id 'test-list', got %s", capturedID)
		}
		if capturedIndex != 2 {
			t.Errorf("expected index 2, got %d", capturedIndex)
		}
		if capturedValue != "C" {
			t.Errorf("expected value 'C', got %s", capturedValue)
		}
	})
}

func TestListRender(t *testing.T) {
	t.Run("returns an Element of type ElementBox", func(t *testing.T) {
		l := List().Items([]string{"A", "B", "C"})
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with empty items", func(t *testing.T) {
		l := List().Items([]string{})
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with prefix", func(t *testing.T) {
		l := List().Items([]string{"A", "B", "C"}).Prefix("📋 ")
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with suffix", func(t *testing.T) {
		l := List().Items([]string{"A", "B", "C"}).Suffix(" ✓")
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with width", func(t *testing.T) {
		l := List().Items([]string{"A", "B", "C"}).Width(10)
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestRenderListFunction(t *testing.T) {
	t.Run("renders with default style when not focused", func(t *testing.T) {
		config := &ListConfig{
			ID:    "test",
			Items: []string{"A", "B", "C"},
		}
		elem := renderList(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with focus style when focused", func(t *testing.T) {
		config := &ListConfig{
			ID:    "test",
			Items: []string{"A", "B", "C"},
		}
		elem := renderList(true, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("selects first item by default", func(t *testing.T) {
		config := &ListConfig{
			ID:    "test",
			Items: []string{"A", "B", "C"},
		}
		elem := renderList(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles empty items list", func(t *testing.T) {
		config := &ListConfig{
			ID:    "test",
			Items: []string{},
		}
		elem := renderList(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("triggers OnFocus when focused", func(t *testing.T) {
		var focusCalled bool
		config := &ListConfig{
			ID:    "test",
			Items: []string{"A", "B", "C"},
			OnFocus: func(id string) {
				focusCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderList(true, config)
		if !focusCalled {
			t.Error("OnFocus was not called when focused")
		}
	})

	t.Run("triggers OnBlur when not focused", func(t *testing.T) {
		var blurCalled bool
		config := &ListConfig{
			ID:    "test",
			Items: []string{"A", "B", "C"},
			OnBlur: func(id string) {
				blurCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderList(false, config)
		if !blurCalled {
			t.Error("OnBlur was not called when not focused")
		}
	})

	t.Run("does not trigger focus/blur without ID", func(t *testing.T) {
		var focusCalled, blurCalled bool
		config := &ListConfig{
			ID:      "",
			Items:   []string{"A", "B", "C"},
			OnFocus: func(id string) { focusCalled = true },
			OnBlur:  func(id string) { blurCalled = true },
		}

		renderList(true, config)
		if focusCalled {
			t.Error("OnFocus should not be called without ID")
		}

		renderList(false, config)
		if blurCalled {
			t.Error("OnBlur should not be called without ID")
		}
	})
}

func TestListChaining(t *testing.T) {
	items := []string{"Apple", "Banana", "Orange"}
	l := List().
		ID("chain-test").
		Items(items).
		Selected(1).
		Width(20).
		Prefix("📋 ").
		Suffix(" ✓").
		Focused(true)

	if l.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", l.config.ID)
	}
	if !reflect.DeepEqual(l.config.Items, items) {
		t.Errorf("expected Items %v, got %v", items, l.config.Items)
	}
	if l.config.Selected != 1 {
		t.Errorf("expected Selected 1, got %d", l.config.Selected)
	}
	if l.config.Width != 20 {
		t.Errorf("expected Width 20, got %d", l.config.Width)
	}
	if l.config.Prefix != "📋 " {
		t.Errorf("expected Prefix '📋 ', got %s", l.config.Prefix)
	}
	if l.config.Suffix != " ✓" {
		t.Errorf("expected Suffix ' ✓', got %s", l.config.Suffix)
	}
	if !l.focused {
		t.Errorf("expected focused true, got %v", l.focused)
	}
}

func TestListDefaultStyles(t *testing.T) {
	t.Run("has default selected style", func(t *testing.T) {
		l := List()
		expected := retui.NewStyle().Foreground(retui.Cyan).Bold(true)

		if !reflect.DeepEqual(l.config.SelectedStyle, expected) {
			t.Errorf("expected selected style %+v, got %+v", expected, l.config.SelectedStyle)
		}
	})

	t.Run("has default unselected style", func(t *testing.T) {
		l := List()
		expected := retui.NewStyle().Foreground(retui.BrightBlack)

		if !reflect.DeepEqual(l.config.UnselectedStyle, expected) {
			t.Errorf("expected unselected style %+v, got %+v", expected, l.config.UnselectedStyle)
		}
	})
}

func TestListEdgeCases(t *testing.T) {
	t.Run("handles empty items gracefully", func(t *testing.T) {
		l := List().Items([]string{})
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles negative width gracefully", func(t *testing.T) {
		l := List().Items([]string{"A", "B", "C"}).Width(-5)
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles negative selected index gracefully", func(t *testing.T) {
		l := List().Items([]string{"A", "B", "C"}).Selected(-1)
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles selected index out of bounds", func(t *testing.T) {
		l := List().Items([]string{"A", "B", "C"}).Selected(10)
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very long items", func(t *testing.T) {
		longItem := "This is a very long list item that should still render properly"
		l := List().Items([]string{longItem, "B", "C"})
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles many items", func(t *testing.T) {
		items := make([]string, 100)
		for i := 0; i < 100; i++ {
			items[i] = "Item " + string(rune('A'+i%26))
		}
		l := List().Items(items)
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nil callbacks gracefully", func(t *testing.T) {
		l := List().ID("test").Items([]string{"A", "B", "C"}).
			OnSelect(nil).OnKeyPress(nil).OnFocus(nil).OnBlur(nil).OnSubmit(nil)
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Error("Render returned wrong Element type with nil callbacks")
		}
	})

	t.Run("handles empty prefix and suffix", func(t *testing.T) {
		l := List().Items([]string{"A", "B", "C"}).Prefix("").Suffix("")
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles special characters in items", func(t *testing.T) {
		specialItems := []string{"!@#$", "✓完成", "123", "ABC-DEF"}
		l := List().Items(specialItems)
		elem := l.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

// Benchmark tests
func BenchmarkListRender(b *testing.B) {
	items := make([]string, 10)
	for i := 0; i < 10; i++ {
		items[i] = "Item " + string(rune('A'+i))
	}
	l := List().Items(items).Width(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Render()
	}
}

func BenchmarkRenderList(b *testing.B) {
	items := make([]string, 10)
	for i := 0; i < 10; i++ {
		items[i] = "Item " + string(rune('A'+i))
	}
	config := &ListConfig{
		ID:    "bench",
		Items: items,
		Width: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderList(false, config)
	}
}

// Example usage test
func ExampleList() {
	// Create a list
	fruitList := List().
		ID("fruits").
		Items([]string{"Apple", "Banana", "Orange", "Grape", "Mango"}).
		Selected(0).
		Width(20).
		Prefix("📋 ").
		Suffix(" ✓").
		OnSelect(func(id string, index int, value string) {
			println("Selected:", value, "at index:", index)
		}).
		OnSubmit(func(id string, index int, value string) {
			println("Submitted:", value)
		}).
		Render()
	_ = fruitList
}
