package components

import (
	"reflect"
	"strings"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestTextInputBuilder_Defaults(t *testing.T) {
	inp := TextInput()

	if inp.config.ID != "" {
		t.Errorf("expected empty ID, got %q", inp.config.ID)
	}
	if inp.config.Value != "" {
		t.Errorf("expected empty Value, got %q", inp.config.Value)
	}
	if inp.config.Placeholder != "" {
		t.Errorf("expected empty Placeholder, got %q", inp.config.Placeholder)
	}
	if inp.config.Width != 30 {
		t.Errorf("expected Width 30, got %d", inp.config.Width)
	}
	if inp.config.Prefix != "" {
		t.Errorf("expected empty Prefix, got %q", inp.config.Prefix)
	}
	if inp.config.Suffix != "" {
		t.Errorf("expected empty Suffix, got %q", inp.config.Suffix)
	}
	if inp.config.MinLength != 0 {
		t.Errorf("expected MinLength 0, got %d", inp.config.MinLength)
	}
	if inp.config.MaxLength != 0 {
		t.Errorf("expected MaxLength 0, got %d", inp.config.MaxLength)
	}
	if inp.focused {
		t.Errorf("expected focused false, got %v", inp.focused)
	}
}

func TestTextInputBuilder_StringFields(t *testing.T) {
	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{"ID", func() string { return TextInput().ID("test-id").config.ID }, "test-id"},
		{"Value", func() string { return TextInput().Value("hello").config.Value }, "hello"},
		{"Placeholder", func() string { return TextInput().Placeholder("Enter text").config.Placeholder }, "Enter text"},
		{"Prefix", func() string { return TextInput().Prefix("> ").config.Prefix }, "> "},
		{"Suffix", func() string { return TextInput().Suffix(" <").config.Suffix }, " <"},
	}

	for _, tt := range tests {
		t.Run("sets "+tt.name+" correctly", func(t *testing.T) {
			if got := tt.got(); got != tt.want {
				t.Errorf("expected %s %q, got %q", tt.name, tt.want, got)
			}
		})
	}
}

func TestTextInputBuilder_Width(t *testing.T) {
	inp := TextInput().Width(25)
	if inp.config.Width != 25 {
		t.Errorf("expected Width 25, got %d", inp.config.Width)
	}
}

func TestTextInputBuilder_Style(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Background(retui.Blue)
		inp := TextInput().Style(style)
		if !reflect.DeepEqual(inp.config.Style, style) {
			t.Errorf("Style not set correctly")
		}
	})
}

func TestTextInputBuilder_MinLength(t *testing.T) {
	t.Run("sets MinLength correctly", func(t *testing.T) {
		inp := TextInput().MinLength(5)
		if inp.config.MinLength != 5 {
			t.Errorf("expected MinLength 5, got %d", inp.config.MinLength)
		}
	})
}

func TestTextInputBuilder_MaxLength(t *testing.T) {
	t.Run("sets MaxLength correctly", func(t *testing.T) {
		inp := TextInput().MaxLength(10)
		if inp.config.MaxLength != 10 {
			t.Errorf("expected MaxLength 10, got %d", inp.config.MaxLength)
		}
	})
}

func TestTextInputBuilder_State(t *testing.T) {
	t.Run("sets Focused state correctly", func(t *testing.T) {
		inp := TextInput().Focused(true)
		if !inp.focused {
			t.Errorf("expected focused true, got %v", inp.focused)
		}
	})

	t.Run("sets Focused state to false", func(t *testing.T) {
		inp := TextInput().Focused(false)
		if inp.focused {
			t.Errorf("expected focused false, got %v", inp.focused)
		}
	})
}

func TestTextInputCallbacks(t *testing.T) {
	t.Run("OnChange callback is set and invoked correctly", func(t *testing.T) {
		var changed bool
		var capturedID string
		var capturedValue string
		inp := TextInput().ID("test-input").OnChange(func(id string, value string) {
			changed = true
			capturedID = id
			capturedValue = value
		})

		if inp.config.OnChange == nil {
			t.Fatal("OnChange callback not set")
		}
		inp.config.OnChange("test-input", "new value")
		if !changed {
			t.Error("OnChange callback was not executed")
		}
		if capturedID != "test-input" {
			t.Errorf("expected id 'test-input', got %s", capturedID)
		}
		if capturedValue != "new value" {
			t.Errorf("expected value 'new value', got %s", capturedValue)
		}
	})

	t.Run("OnKeyPress callback is set and invoked correctly", func(t *testing.T) {
		var keyPressed bool
		inp := TextInput().ID("test-input").OnKeyPress(func(id string, key retui.Key) bool {
			keyPressed = true
			if id != "test-input" {
				t.Errorf("expected id 'test-input', got %s", id)
			}
			return true
		})

		if inp.config.OnKeyPress == nil {
			t.Fatal("OnKeyPress callback not set")
		}
		if result := inp.config.OnKeyPress("test-input", retui.Key{Code: retui.KeyEnter}); !result {
			t.Error("OnKeyPress should return true")
		}
		if !keyPressed {
			t.Error("OnKeyPress callback was not executed")
		}
	})

	t.Run("OnFocus callback is set and invoked correctly", func(t *testing.T) {
		var focused bool
		inp := TextInput().ID("test-input").OnFocus(func(id string) {
			focused = true
			if id != "test-input" {
				t.Errorf("expected id 'test-input', got %s", id)
			}
		})

		if inp.config.OnFocus == nil {
			t.Fatal("OnFocus callback not set")
		}
		inp.config.OnFocus("test-input")
		if !focused {
			t.Error("OnFocus callback was not executed")
		}
	})

	t.Run("OnBlur callback is set and invoked correctly", func(t *testing.T) {
		var blurred bool
		inp := TextInput().ID("test-input").OnBlur(func(id string) {
			blurred = true
			if id != "test-input" {
				t.Errorf("expected id 'test-input', got %s", id)
			}
		})

		if inp.config.OnBlur == nil {
			t.Fatal("OnBlur callback not set")
		}
		inp.config.OnBlur("test-input")
		if !blurred {
			t.Error("OnBlur callback was not executed")
		}
	})

	t.Run("OnSubmit callback is set and invoked correctly", func(t *testing.T) {
		var submitted bool
		var capturedID string
		var capturedValue string
		inp := TextInput().ID("test-input").OnSubmit(func(id string, value string) {
			submitted = true
			capturedID = id
			capturedValue = value
		})

		if inp.config.OnSubmit == nil {
			t.Fatal("OnSubmit callback not set")
		}
		inp.config.OnSubmit("test-input", "submitted")
		if !submitted {
			t.Error("OnSubmit callback was not executed")
		}
		if capturedID != "test-input" {
			t.Errorf("expected id 'test-input', got %s", capturedID)
		}
		if capturedValue != "submitted" {
			t.Errorf("expected value 'submitted', got %s", capturedValue)
		}
	})
}

func TestTextInputRender(t *testing.T) {
	t.Run("returns an Element of type ElementBox", func(t *testing.T) {
		inp := TextInput()
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with placeholder when empty", func(t *testing.T) {
		inp := TextInput().Placeholder("Enter text")
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with value", func(t *testing.T) {
		inp := TextInput().Value("hello")
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with prefix", func(t *testing.T) {
		inp := TextInput().Prefix("> ").Value("hello")
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with suffix", func(t *testing.T) {
		inp := TextInput().Suffix(" <").Value("hello")
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with width", func(t *testing.T) {
		inp := TextInput().Value("hello").Width(20)
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestRenderInputFunction(t *testing.T) {
	t.Run("renders with default style when not focused", func(t *testing.T) {
		config := &InputConfig{
			ID:    "test",
			Value: "hello",
			Width: 20,
		}
		elem := renderInput(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with focus style when focused", func(t *testing.T) {
		config := &InputConfig{
			ID:    "test",
			Value: "hello",
			Width: 20,
		}
		elem := renderInput(true, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with placeholder when empty", func(t *testing.T) {
		config := &InputConfig{
			ID:          "test",
			Value:       "",
			Placeholder: "Enter text",
			Width:       20,
		}
		elem := renderInput(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("triggers OnFocus when focused", func(t *testing.T) {
		var focusCalled bool
		config := &InputConfig{
			ID:    "test",
			Value: "hello",
			OnFocus: func(id string) {
				focusCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderInput(true, config)
		if !focusCalled {
			t.Error("OnFocus was not called when focused")
		}
	})

	t.Run("triggers OnBlur when not focused", func(t *testing.T) {
		var blurCalled bool
		config := &InputConfig{
			ID:    "test",
			Value: "hello",
			OnBlur: func(id string) {
				blurCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderInput(false, config)
		if !blurCalled {
			t.Error("OnBlur was not called when not focused")
		}
	})

	t.Run("does not trigger focus/blur without ID", func(t *testing.T) {
		var focusCalled, blurCalled bool
		config := &InputConfig{
			ID:      "",
			Value:   "hello",
			OnFocus: func(id string) { focusCalled = true },
			OnBlur:  func(id string) { blurCalled = true },
		}

		renderInput(true, config)
		if focusCalled {
			t.Error("OnFocus should not be called without ID")
		}

		renderInput(false, config)
		if blurCalled {
			t.Error("OnBlur should not be called without ID")
		}
	})
}

func TestTextInputChaining(t *testing.T) {
	inp := TextInput().
		ID("chain-test").
		Value("hello").
		Placeholder("Enter text").
		Width(25).
		Prefix("> ").
		Suffix(" <").
		MinLength(3).
		MaxLength(10).
		Focused(true)

	if inp.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", inp.config.ID)
	}
	if inp.config.Value != "hello" {
		t.Errorf("expected Value 'hello', got %s", inp.config.Value)
	}
	if inp.config.Placeholder != "Enter text" {
		t.Errorf("expected Placeholder 'Enter text', got %s", inp.config.Placeholder)
	}
	if inp.config.Width != 25 {
		t.Errorf("expected Width 25, got %d", inp.config.Width)
	}
	if inp.config.Prefix != "> " {
		t.Errorf("expected Prefix '> ', got %s", inp.config.Prefix)
	}
	if inp.config.Suffix != " <" {
		t.Errorf("expected Suffix ' <', got %s", inp.config.Suffix)
	}
	if inp.config.MinLength != 3 {
		t.Errorf("expected MinLength 3, got %d", inp.config.MinLength)
	}
	if inp.config.MaxLength != 10 {
		t.Errorf("expected MaxLength 10, got %d", inp.config.MaxLength)
	}
	if !inp.focused {
		t.Errorf("expected focused true, got %v", inp.focused)
	}
}

func TestTextInputEdgeCases(t *testing.T) {
	t.Run("handles empty value gracefully", func(t *testing.T) {
		inp := TextInput().Value("")
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles negative width gracefully", func(t *testing.T) {
		inp := TextInput().Value("hello").Width(-5)
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very long value", func(t *testing.T) {
		longValue := strings.Repeat("a", 1000)
		inp := TextInput().Value(longValue)
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nil callbacks gracefully", func(t *testing.T) {
		inp := TextInput().ID("test").Value("hello").
			OnChange(nil).OnKeyPress(nil).OnFocus(nil).OnBlur(nil).OnSubmit(nil)
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Error("Render returned wrong Element type with nil callbacks")
		}
	})

	t.Run("handles empty prefix and suffix", func(t *testing.T) {
		inp := TextInput().Value("hello").Prefix("").Suffix("")
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles Unicode characters", func(t *testing.T) {
		unicodeValue := "✓ 完成 ✅"
		inp := TextInput().Value(unicodeValue)
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles special characters", func(t *testing.T) {
		specialValue := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		inp := TextInput().Value(specialValue)
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("validates min length correctly", func(t *testing.T) {
		inp := TextInput().Value("hi").MinLength(5)
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("validates max length correctly", func(t *testing.T) {
		inp := TextInput().Value("too long").MaxLength(5)
		elem := inp.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

// Benchmark tests
func BenchmarkTextInputRender(b *testing.B) {
	inp := TextInput().Value("hello").Width(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inp.Render()
	}
}

func BenchmarkRenderInput(b *testing.B) {
	config := &InputConfig{
		ID:    "bench",
		Value: "hello",
		Width: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderInput(false, config)
	}
}

// Example usage test
func ExampleTextInput() {
	// Simple text input
	textInput := TextInput().
		ID("name").
		Placeholder("Enter your name").
		Width(30).
		MinLength(2).
		MaxLength(50).
		Prefix("👤 ").
		Suffix(" ✏️").
		OnChange(func(id string, value string) {
			// Handle change
		}).
		OnSubmit(func(id string, value string) {
			// Handle submit
		}).
		Render()
	_ = textInput
}
