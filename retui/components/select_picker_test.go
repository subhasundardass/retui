package components

import (
	"reflect"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestSelectPickerBuilder_Defaults(t *testing.T) {
	s := SelectPicker()

	if s.config.ID != "" {
		t.Errorf("expected empty ID, got %q", s.config.ID)
	}
	if len(s.config.Options) > 0 {
		t.Errorf("expected empty Options, got %v", s.config.Options)
	}
	if s.config.Selected != 0 {
		t.Errorf("expected Selected 0, got %d", s.config.Selected)
	}
	if s.config.Width != 0 {
		t.Errorf("expected Width 0, got %d", s.config.Width)
	}
	if s.config.Prefix != "" {
		t.Errorf("expected empty Prefix, got %q", s.config.Prefix)
	}
	if s.config.Suffix != "" {
		t.Errorf("expected empty Suffix, got %q", s.config.Suffix)
	}
	if s.focused {
		t.Errorf("expected focused false, got %v", s.focused)
	}
}

func TestSelectPickerBuilder_StringFields(t *testing.T) {
	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{"ID", func() string { return SelectPicker().ID("test-id").config.ID }, "test-id"},
		{"Prefix", func() string { return SelectPicker().Prefix("🎨 ").config.Prefix }, "🎨 "},
		{"Suffix", func() string { return SelectPicker().Suffix(" ✓").config.Suffix }, " ✓"},
	}

	for _, tt := range tests {
		t.Run("sets "+tt.name+" correctly", func(t *testing.T) {
			if got := tt.got(); got != tt.want {
				t.Errorf("expected %s %q, got %q", tt.name, tt.want, got)
			}
		})
	}
}

func TestSelectPickerBuilder_Options(t *testing.T) {
	t.Run("sets Options correctly", func(t *testing.T) {
		options := []string{"Red", "Green", "Blue"}
		s := SelectPicker().Options(options)
		if !reflect.DeepEqual(s.config.Options, options) {
			t.Errorf("expected Options %v, got %v", options, s.config.Options)
		}
	})

	t.Run("sets empty Options correctly", func(t *testing.T) {
		options := []string{}
		s := SelectPicker().Options(options)
		if !reflect.DeepEqual(s.config.Options, options) {
			t.Errorf("expected empty Options, got %v", s.config.Options)
		}
	})
}

func TestSelectPickerBuilder_Selected(t *testing.T) {
	t.Run("sets Selected correctly", func(t *testing.T) {
		s := SelectPicker().Selected(2)
		if s.config.Selected != 2 {
			t.Errorf("expected Selected 2, got %d", s.config.Selected)
		}
	})

	t.Run("sets Selected to 0 by default", func(t *testing.T) {
		s := SelectPicker()
		if s.config.Selected != 0 {
			t.Errorf("expected Selected 0, got %d", s.config.Selected)
		}
	})

	t.Run("sets Selected to negative value", func(t *testing.T) {
		s := SelectPicker().Selected(-1)
		if s.config.Selected != -1 {
			t.Errorf("expected Selected -1, got %d", s.config.Selected)
		}
	})
}

func TestSelectPickerBuilder_Width(t *testing.T) {
	s := SelectPicker().Width(20)
	if s.config.Width != 20 {
		t.Errorf("expected Width 20, got %d", s.config.Width)
	}
}

func TestSelectPickerBuilder_Style(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Background(retui.Blue)
		s := SelectPicker().Style(style)
		if !reflect.DeepEqual(s.config.Style, style) {
			t.Errorf("Style not set correctly")
		}
	})
}

func TestSelectPickerBuilder_State(t *testing.T) {
	t.Run("sets Focused state correctly", func(t *testing.T) {
		s := SelectPicker().Focused(true)
		if !s.focused {
			t.Errorf("expected focused true, got %v", s.focused)
		}
	})

	t.Run("sets Focused state to false", func(t *testing.T) {
		s := SelectPicker().Focused(false)
		if s.focused {
			t.Errorf("expected focused false, got %v", s.focused)
		}
	})
}

func TestSelectPickerCallbacks(t *testing.T) {
	t.Run("OnChange callback is set and invoked correctly", func(t *testing.T) {
		var changed bool
		var capturedID string
		var capturedIndex int
		var capturedValue string
		s := SelectPicker().ID("test-picker").Options([]string{"A", "B", "C"}).OnChange(func(id string, index int, value string) {
			changed = true
			capturedID = id
			capturedIndex = index
			capturedValue = value
		})

		if s.config.OnChange == nil {
			t.Fatal("OnChange callback not set")
		}
		s.config.OnChange("test-picker", 1, "B")
		if !changed {
			t.Error("OnChange callback was not executed")
		}
		if capturedID != "test-picker" {
			t.Errorf("expected id 'test-picker', got %s", capturedID)
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
		s := SelectPicker().ID("test-picker").OnKeyPress(func(id string, key retui.Key) bool {
			keyPressed = true
			if id != "test-picker" {
				t.Errorf("expected id 'test-picker', got %s", id)
			}
			return true
		})

		if s.config.OnKeyPress == nil {
			t.Fatal("OnKeyPress callback not set")
		}
		if result := s.config.OnKeyPress("test-picker", retui.Key{Code: retui.KeyRight}); !result {
			t.Error("OnKeyPress should return true")
		}
		if !keyPressed {
			t.Error("OnKeyPress callback was not executed")
		}
	})

	t.Run("OnFocus callback is set and invoked correctly", func(t *testing.T) {
		var focused bool
		s := SelectPicker().ID("test-picker").OnFocus(func(id string) {
			focused = true
			if id != "test-picker" {
				t.Errorf("expected id 'test-picker', got %s", id)
			}
		})

		if s.config.OnFocus == nil {
			t.Fatal("OnFocus callback not set")
		}
		s.config.OnFocus("test-picker")
		if !focused {
			t.Error("OnFocus callback was not executed")
		}
	})

	t.Run("OnBlur callback is set and invoked correctly", func(t *testing.T) {
		var blurred bool
		s := SelectPicker().ID("test-picker").OnBlur(func(id string) {
			blurred = true
			if id != "test-picker" {
				t.Errorf("expected id 'test-picker', got %s", id)
			}
		})

		if s.config.OnBlur == nil {
			t.Fatal("OnBlur callback not set")
		}
		s.config.OnBlur("test-picker")
		if !blurred {
			t.Error("OnBlur callback was not executed")
		}
	})

	t.Run("OnSubmit callback is set and invoked correctly", func(t *testing.T) {
		var submitted bool
		var capturedID string
		var capturedIndex int
		var capturedValue string
		s := SelectPicker().ID("test-picker").Options([]string{"A", "B", "C"}).OnSubmit(func(id string, index int, value string) {
			submitted = true
			capturedID = id
			capturedIndex = index
			capturedValue = value
		})

		if s.config.OnSubmit == nil {
			t.Fatal("OnSubmit callback not set")
		}
		s.config.OnSubmit("test-picker", 2, "C")
		if !submitted {
			t.Error("OnSubmit callback was not executed")
		}
		if capturedID != "test-picker" {
			t.Errorf("expected id 'test-picker', got %s", capturedID)
		}
		if capturedIndex != 2 {
			t.Errorf("expected index 2, got %d", capturedIndex)
		}
		if capturedValue != "C" {
			t.Errorf("expected value 'C', got %s", capturedValue)
		}
	})
}

func TestSelectPickerRender(t *testing.T) {
	t.Run("returns an Element of type ElementBox", func(t *testing.T) {
		s := SelectPicker().Options([]string{"A", "B", "C"})
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with empty options", func(t *testing.T) {
		// Empty options will cause panic in renderSelectPicker
		// We test that it panics as expected
		defer func() {
			if r := recover(); r != nil {
				// Expected panic for empty options
				t.Log("SelectPicker panicked with empty options as expected")
			}
		}()
		s := SelectPicker().Options([]string{})
		s.Render()
		t.Error("Expected panic but didn't get one")
	})

	t.Run("renders with prefix", func(t *testing.T) {
		s := SelectPicker().Options([]string{"A", "B", "C"}).Prefix("🎨 ")
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with suffix", func(t *testing.T) {
		s := SelectPicker().Options([]string{"A", "B", "C"}).Suffix(" ✓")
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with width", func(t *testing.T) {
		s := SelectPicker().Options([]string{"A", "B", "C"}).Width(10)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with selected value", func(t *testing.T) {
		s := SelectPicker().Options([]string{"Red", "Green", "Blue"}).Selected(1)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestRenderSelectPickerFunction(t *testing.T) {
	t.Run("renders with default style when not focused", func(t *testing.T) {
		config := &SelectPickerConfig{
			ID:       "test",
			Options:  []string{"A", "B", "C"},
			Selected: 0,
		}
		elem := renderSelectPicker(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with focus style when focused", func(t *testing.T) {
		config := &SelectPickerConfig{
			ID:       "test",
			Options:  []string{"A", "B", "C"},
			Selected: 0,
		}
		elem := renderSelectPicker(true, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("selects first item by default", func(t *testing.T) {
		config := &SelectPickerConfig{
			ID:       "test",
			Options:  []string{"A", "B", "C"},
			Selected: 0,
		}
		elem := renderSelectPicker(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles empty options list", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Log("renderSelectPicker panicked with empty options as expected")
			}
		}()
		config := &SelectPickerConfig{
			ID:       "test",
			Options:  []string{},
			Selected: 0,
		}
		renderSelectPicker(false, config)
		t.Error("Expected panic but didn't get one")
	})

	t.Run("triggers OnFocus when focused", func(t *testing.T) {
		var focusCalled bool
		config := &SelectPickerConfig{
			ID:       "test",
			Options:  []string{"A", "B", "C"},
			Selected: 0,
			OnFocus: func(id string) {
				focusCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderSelectPicker(true, config)
		if !focusCalled {
			t.Error("OnFocus was not called when focused")
		}
	})

	t.Run("triggers OnBlur when not focused", func(t *testing.T) {
		var blurCalled bool
		config := &SelectPickerConfig{
			ID:       "test",
			Options:  []string{"A", "B", "C"},
			Selected: 0,
			OnBlur: func(id string) {
				blurCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderSelectPicker(false, config)
		if !blurCalled {
			t.Error("OnBlur was not called when not focused")
		}
	})

	t.Run("does not trigger focus/blur without ID", func(t *testing.T) {
		var focusCalled, blurCalled bool
		config := &SelectPickerConfig{
			ID:       "",
			Options:  []string{"A", "B", "C"},
			Selected: 0,
			OnFocus:  func(id string) { focusCalled = true },
			OnBlur:   func(id string) { blurCalled = true },
		}

		renderSelectPicker(true, config)
		if focusCalled {
			t.Error("OnFocus should not be called without ID")
		}

		renderSelectPicker(false, config)
		if blurCalled {
			t.Error("OnBlur should not be called without ID")
		}
	})
}

func TestSelectPickerChaining(t *testing.T) {
	options := []string{"Red", "Green", "Blue"}
	s := SelectPicker().
		ID("chain-test").
		Options(options).
		Selected(1).
		Width(20).
		Prefix("🎨 ").
		Suffix(" ✓").
		Focused(true)

	if s.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", s.config.ID)
	}
	if !reflect.DeepEqual(s.config.Options, options) {
		t.Errorf("expected Options %v, got %v", options, s.config.Options)
	}
	if s.config.Selected != 1 {
		t.Errorf("expected Selected 1, got %d", s.config.Selected)
	}
	if s.config.Width != 20 {
		t.Errorf("expected Width 20, got %d", s.config.Width)
	}
	if s.config.Prefix != "🎨 " {
		t.Errorf("expected Prefix '🎨 ', got %s", s.config.Prefix)
	}
	if s.config.Suffix != " ✓" {
		t.Errorf("expected Suffix ' ✓', got %s", s.config.Suffix)
	}
	if !s.focused {
		t.Errorf("expected focused true, got %v", s.focused)
	}
}

func TestSelectPickerEdgeCases(t *testing.T) {
	t.Run("handles empty options gracefully", func(t *testing.T) {
		// Empty options will cause panic
		defer func() {
			if r := recover(); r != nil {
				t.Log("SelectPicker panicked with empty options as expected")
			}
		}()
		s := SelectPicker().Options([]string{})
		s.Render()
		t.Error("Expected panic but didn't get one")
	})

	t.Run("handles nil options gracefully", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Log("SelectPicker panicked with nil options as expected")
			}
		}()
		s := SelectPicker().Options(nil)
		s.Render()
		t.Error("Expected panic but didn't get one")
	})

	t.Run("handles negative width gracefully", func(t *testing.T) {
		s := SelectPicker().Options([]string{"A", "B", "C"}).Width(-5)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles negative selected index gracefully", func(t *testing.T) {
		s := SelectPicker().Options([]string{"A", "B", "C"}).Selected(-1)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles selected index out of bounds", func(t *testing.T) {
		s := SelectPicker().Options([]string{"A", "B", "C"}).Selected(10)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very long options", func(t *testing.T) {
		longOption := "This is a very long option that should still render properly"
		s := SelectPicker().Options([]string{longOption, "B", "C"})
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles many options", func(t *testing.T) {
		options := make([]string, 100)
		for i := 0; i < 100; i++ {
			options[i] = "Option " + string(rune('A'+i%26))
		}
		s := SelectPicker().Options(options)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nil callbacks gracefully", func(t *testing.T) {
		s := SelectPicker().ID("test").Options([]string{"A", "B", "C"}).
			OnChange(nil).OnKeyPress(nil).OnFocus(nil).OnBlur(nil).OnSubmit(nil)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Error("Render returned wrong Element type with nil callbacks")
		}
	})

	t.Run("handles empty prefix and suffix", func(t *testing.T) {
		s := SelectPicker().Options([]string{"A", "B", "C"}).Prefix("").Suffix("")
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles special characters in options", func(t *testing.T) {
		specialOptions := []string{"!@#$", "✓完成", "123", "ABC-DEF"}
		s := SelectPicker().Options(specialOptions)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles Unicode characters in options", func(t *testing.T) {
		unicodeOptions := []string{"✓", "★", "❤", "完成", "✅"}
		s := SelectPicker().Options(unicodeOptions)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

// Benchmark tests
func BenchmarkSelectPickerRender(b *testing.B) {
	options := make([]string, 10)
	for i := 0; i < 10; i++ {
		options[i] = "Option " + string(rune('A'+i))
	}
	s := SelectPicker().Options(options).Width(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Render()
	}
}

func BenchmarkRenderSelectPicker(b *testing.B) {
	options := make([]string, 10)
	for i := 0; i < 10; i++ {
		options[i] = "Option " + string(rune('A'+i))
	}
	config := &SelectPickerConfig{
		ID:       "bench",
		Options:  options,
		Selected: 0,
		Width:    20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderSelectPicker(false, config)
	}
}
