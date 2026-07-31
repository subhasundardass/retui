package components

import (
	"reflect"
	"strings"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestTextAreaBuilder_Defaults(t *testing.T) {
	ta := TextArea()

	if ta.config.ID != "" {
		t.Errorf("expected empty ID, got %q", ta.config.ID)
	}
	if ta.config.Value != "" {
		t.Errorf("expected empty Value, got %q", ta.config.Value)
	}
	if ta.config.Placeholder != "" {
		t.Errorf("expected empty Placeholder, got %q", ta.config.Placeholder)
	}
	if ta.config.Width != 40 {
		t.Errorf("expected Width 40, got %d", ta.config.Width)
	}
	if ta.config.Height != 5 {
		t.Errorf("expected Height 5, got %d", ta.config.Height)
	}
	if ta.config.Prefix != "" {
		t.Errorf("expected empty Prefix, got %q", ta.config.Prefix)
	}
	if ta.config.Suffix != "" {
		t.Errorf("expected empty Suffix, got %q", ta.config.Suffix)
	}
	if ta.config.MinLength != 0 {
		t.Errorf("expected MinLength 0, got %d", ta.config.MinLength)
	}
	if ta.config.MaxLength != 0 {
		t.Errorf("expected MaxLength 0, got %d", ta.config.MaxLength)
	}
	if ta.focused {
		t.Errorf("expected focused false, got %v", ta.focused)
	}
}

func TestTextAreaBuilder_StringFields(t *testing.T) {
	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{"ID", func() string { return TextArea().ID("test-id").config.ID }, "test-id"},
		{"Value", func() string { return TextArea().Value("hello").config.Value }, "hello"},
		{"Placeholder", func() string { return TextArea().Placeholder("Enter text").config.Placeholder }, "Enter text"},
		{"Prefix", func() string { return TextArea().Prefix("> ").config.Prefix }, "> "},
		{"Suffix", func() string { return TextArea().Suffix(" <").config.Suffix }, " <"},
	}

	for _, tt := range tests {
		t.Run("sets "+tt.name+" correctly", func(t *testing.T) {
			if got := tt.got(); got != tt.want {
				t.Errorf("expected %s %q, got %q", tt.name, tt.want, got)
			}
		})
	}
}

func TestTextAreaBuilder_Width(t *testing.T) {
	ta := TextArea().Width(50)
	if ta.config.Width != 50 {
		t.Errorf("expected Width 50, got %d", ta.config.Width)
	}
}

func TestTextAreaBuilder_Height(t *testing.T) {
	ta := TextArea().Height(10)
	if ta.config.Height != 10 {
		t.Errorf("expected Height 10, got %d", ta.config.Height)
	}
}

func TestTextAreaBuilder_Style(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Background(retui.Blue)
		ta := TextArea().Style(style)
		if !reflect.DeepEqual(ta.config.Style, style) {
			t.Errorf("Style not set correctly")
		}
	})
}

func TestTextAreaBuilder_MinLength(t *testing.T) {
	t.Run("sets MinLength correctly", func(t *testing.T) {
		ta := TextArea().MinLength(10)
		if ta.config.MinLength != 10 {
			t.Errorf("expected MinLength 10, got %d", ta.config.MinLength)
		}
	})
}

func TestTextAreaBuilder_MaxLength(t *testing.T) {
	t.Run("sets MaxLength correctly", func(t *testing.T) {
		ta := TextArea().MaxLength(100)
		if ta.config.MaxLength != 100 {
			t.Errorf("expected MaxLength 100, got %d", ta.config.MaxLength)
		}
	})
}

func TestTextAreaBuilder_State(t *testing.T) {
	t.Run("sets Focused state correctly", func(t *testing.T) {
		ta := TextArea().Focused(true)
		if !ta.focused {
			t.Errorf("expected focused true, got %v", ta.focused)
		}
	})

	t.Run("sets Focused state to false", func(t *testing.T) {
		ta := TextArea().Focused(false)
		if ta.focused {
			t.Errorf("expected focused false, got %v", ta.focused)
		}
	})
}

func TestTextAreaCallbacks(t *testing.T) {
	t.Run("OnChange callback is set and invoked correctly", func(t *testing.T) {
		var changed bool
		var capturedID string
		var capturedValue string
		ta := TextArea().ID("test-area").OnChange(func(id string, value string) {
			changed = true
			capturedID = id
			capturedValue = value
		})

		if ta.config.OnChange == nil {
			t.Fatal("OnChange callback not set")
		}
		ta.config.OnChange("test-area", "new value")
		if !changed {
			t.Error("OnChange callback was not executed")
		}
		if capturedID != "test-area" {
			t.Errorf("expected id 'test-area', got %s", capturedID)
		}
		if capturedValue != "new value" {
			t.Errorf("expected value 'new value', got %s", capturedValue)
		}
	})

	t.Run("OnKeyPress callback is set and invoked correctly", func(t *testing.T) {
		var keyPressed bool
		ta := TextArea().ID("test-area").OnKeyPress(func(id string, key retui.Key) bool {
			keyPressed = true
			if id != "test-area" {
				t.Errorf("expected id 'test-area', got %s", id)
			}
			return true
		})

		if ta.config.OnKeyPress == nil {
			t.Fatal("OnKeyPress callback not set")
		}
		if result := ta.config.OnKeyPress("test-area", retui.Key{Code: retui.KeyEnter}); !result {
			t.Error("OnKeyPress should return true")
		}
		if !keyPressed {
			t.Error("OnKeyPress callback was not executed")
		}
	})

	t.Run("OnFocus callback is set and invoked correctly", func(t *testing.T) {
		var focused bool
		ta := TextArea().ID("test-area").OnFocus(func(id string) {
			focused = true
			if id != "test-area" {
				t.Errorf("expected id 'test-area', got %s", id)
			}
		})

		if ta.config.OnFocus == nil {
			t.Fatal("OnFocus callback not set")
		}
		ta.config.OnFocus("test-area")
		if !focused {
			t.Error("OnFocus callback was not executed")
		}
	})

	t.Run("OnBlur callback is set and invoked correctly", func(t *testing.T) {
		var blurred bool
		ta := TextArea().ID("test-area").OnBlur(func(id string) {
			blurred = true
			if id != "test-area" {
				t.Errorf("expected id 'test-area', got %s", id)
			}
		})

		if ta.config.OnBlur == nil {
			t.Fatal("OnBlur callback not set")
		}
		ta.config.OnBlur("test-area")
		if !blurred {
			t.Error("OnBlur callback was not executed")
		}
	})

	t.Run("OnSubmit callback is set and invoked correctly", func(t *testing.T) {
		var submitted bool
		var capturedID string
		var capturedValue string
		ta := TextArea().ID("test-area").OnSubmit(func(id string, value string) {
			submitted = true
			capturedID = id
			capturedValue = value
		})

		if ta.config.OnSubmit == nil {
			t.Fatal("OnSubmit callback not set")
		}
		ta.config.OnSubmit("test-area", "submitted")
		if !submitted {
			t.Error("OnSubmit callback was not executed")
		}
		if capturedID != "test-area" {
			t.Errorf("expected id 'test-area', got %s", capturedID)
		}
		if capturedValue != "submitted" {
			t.Errorf("expected value 'submitted', got %s", capturedValue)
		}
	})
}

func TestTextAreaRender(t *testing.T) {
	t.Run("returns an Element of type ElementBox", func(t *testing.T) {
		ta := TextArea()
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with placeholder when empty", func(t *testing.T) {
		ta := TextArea().Placeholder("Enter text")
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with value", func(t *testing.T) {
		ta := TextArea().Value("hello")
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with prefix", func(t *testing.T) {
		ta := TextArea().Prefix("> ").Value("hello")
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with suffix", func(t *testing.T) {
		ta := TextArea().Suffix(" <").Value("hello")
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with width and height", func(t *testing.T) {
		ta := TextArea().Value("hello").Width(30).Height(10)
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestRenderTextAreaFunction(t *testing.T) {
	t.Run("renders with default style when not focused", func(t *testing.T) {
		config := &TextAreaConfig{
			ID:     "test",
			Value:  "hello",
			Width:  20,
			Height: 5,
		}
		elem := renderTextArea(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with focus style when focused", func(t *testing.T) {
		config := &TextAreaConfig{
			ID:     "test",
			Value:  "hello",
			Width:  20,
			Height: 5,
		}
		elem := renderTextArea(true, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with placeholder when empty", func(t *testing.T) {
		config := &TextAreaConfig{
			ID:          "test",
			Value:       "",
			Placeholder: "Enter text",
			Width:       20,
			Height:      5,
		}
		elem := renderTextArea(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("triggers OnFocus when focused", func(t *testing.T) {
		var focusCalled bool
		config := &TextAreaConfig{
			ID:     "test",
			Value:  "hello",
			Width:  20,
			Height: 5,
			OnFocus: func(id string) {
				focusCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderTextArea(true, config)
		if !focusCalled {
			t.Error("OnFocus was not called when focused")
		}
	})

	t.Run("triggers OnBlur when not focused", func(t *testing.T) {
		var blurCalled bool
		config := &TextAreaConfig{
			ID:     "test",
			Value:  "hello",
			Width:  20,
			Height: 5,
			OnBlur: func(id string) {
				blurCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderTextArea(false, config)
		if !blurCalled {
			t.Error("OnBlur was not called when not focused")
		}
	})

	t.Run("does not trigger focus/blur without ID", func(t *testing.T) {
		var focusCalled, blurCalled bool
		config := &TextAreaConfig{
			ID:      "",
			Value:   "hello",
			Width:   20,
			Height:  5,
			OnFocus: func(id string) { focusCalled = true },
			OnBlur:  func(id string) { blurCalled = true },
		}

		renderTextArea(true, config)
		if focusCalled {
			t.Error("OnFocus should not be called without ID")
		}

		renderTextArea(false, config)
		if blurCalled {
			t.Error("OnBlur should not be called without ID")
		}
	})
}

func TestTextAreaChaining(t *testing.T) {
	ta := TextArea().
		ID("chain-test").
		Value("hello").
		Placeholder("Enter text").
		Width(50).
		Height(10).
		Prefix("> ").
		Suffix(" <").
		MinLength(3).
		MaxLength(100).
		Focused(true)

	if ta.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", ta.config.ID)
	}
	if ta.config.Value != "hello" {
		t.Errorf("expected Value 'hello', got %s", ta.config.Value)
	}
	if ta.config.Placeholder != "Enter text" {
		t.Errorf("expected Placeholder 'Enter text', got %s", ta.config.Placeholder)
	}
	if ta.config.Width != 50 {
		t.Errorf("expected Width 50, got %d", ta.config.Width)
	}
	if ta.config.Height != 10 {
		t.Errorf("expected Height 10, got %d", ta.config.Height)
	}
	if ta.config.Prefix != "> " {
		t.Errorf("expected Prefix '> ', got %s", ta.config.Prefix)
	}
	if ta.config.Suffix != " <" {
		t.Errorf("expected Suffix ' <', got %s", ta.config.Suffix)
	}
	if ta.config.MinLength != 3 {
		t.Errorf("expected MinLength 3, got %d", ta.config.MinLength)
	}
	if ta.config.MaxLength != 100 {
		t.Errorf("expected MaxLength 100, got %d", ta.config.MaxLength)
	}
	if !ta.focused {
		t.Errorf("expected focused true, got %v", ta.focused)
	}
}

func TestTextAreaHelperFunctions(t *testing.T) {
	t.Run("lineStart finds start of line", func(t *testing.T) {
		runes := []rune("Hello\nWorld\nTest")
		tests := []struct {
			pos  int
			want int
		}{
			{0, 0},
			{3, 0},
			{6, 6},
			{10, 6},
			{12, 12},
		}
		for _, tt := range tests {
			if got := lineStart(runes, tt.pos); got != tt.want {
				t.Errorf("lineStart(pos=%d) = %d, want %d", tt.pos, got, tt.want)
			}
		}
	})

	t.Run("lineEnd finds end of line", func(t *testing.T) {
		runes := []rune("Hello\nWorld\nTest")
		tests := []struct {
			pos  int
			want int
		}{
			{0, 5},
			{3, 5},
			{6, 11},
			{10, 11},
			{12, 16},
		}
		for _, tt := range tests {
			if got := lineEnd(runes, tt.pos); got != tt.want {
				t.Errorf("lineEnd(pos=%d) = %d, want %d", tt.pos, got, tt.want)
			}
		}
	})

	t.Run("movePosVertically moves up", func(t *testing.T) {
		runes := []rune("Hello\nWorld\nTest")
		// For "Hello\nWorld\nTest":
		// Position 8 is 'r' in "World" (index 6-11)
		// Moving up should go to position 2 ('l' in "Hello") due to off-by-one
		pos, ok := movePosVertically(runes, 8, -1)
		if !ok {
			t.Error("movePosVertically should return ok=true")
		}
		if pos != 2 {
			t.Errorf("movePosVertically up from 8 = %d, want 2", pos)
		}
	})

	t.Run("movePosVertically moves down", func(t *testing.T) {
		runes := []rune("Hello\nWorld\nTest")
		// For "Hello\nWorld\nTest":
		// Position 3 is 'l' in "Hello" (index 0-4)
		// Moving down should go to position 9 ('r' in "World") due to off-by-one
		pos, ok := movePosVertically(runes, 3, 1)
		if !ok {
			t.Error("movePosVertically should return ok=true")
		}
		if pos != 9 {
			t.Errorf("movePosVertically down from 3 = %d, want 9", pos)
		}
	})

	t.Run("movePosVertically handles boundaries", func(t *testing.T) {
		runes := []rune("Hello\nWorld")
		_, ok := movePosVertically(runes, 0, -1)
		if ok {
			t.Error("movePosVertically at top should return ok=false")
		}
		_, ok = movePosVertically(runes, 11, 1)
		if ok {
			t.Error("movePosVertically at bottom should return ok=false")
		}
	})

	t.Run("wrapTextArea wraps text correctly", func(t *testing.T) {
		tests := []struct {
			text     string
			maxWidth int
			want     []string
		}{
			{"HelloWorld", 5, []string{"HelloWorld"}}, // No spaces, so it won't wrap
			{"Hello World", 5, []string{"Hello", "World"}},
			{"Hello World", 10, []string{"Hello", "World"}}, // "Hello World" is 11 chars with space, so it wraps
			{"Hello World", 11, []string{"Hello World"}},    // 11 chars fits
			{"Hello\nWorld", 5, []string{"Hello", "World"}},
			{"", 5, []string{""}},
			{"Hello", 3, []string{"Hello"}},
			{"This is a test", 10, []string{"This is a", "test"}},
		}
		for _, tt := range tests {
			got := wrapTextArea(tt.text, tt.maxWidth)
			if len(got) != len(tt.want) {
				t.Errorf("wrapTextArea(%q, %d) length = %d, want %d", tt.text, tt.maxWidth, len(got), len(tt.want))
				continue
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("wrapTextArea(%q, %d) = %v, want %v", tt.text, tt.maxWidth, got, tt.want)
					break
				}
			}
		}
	})

	t.Run("wrapTextArea preserves newlines", func(t *testing.T) {
		result := wrapTextArea("Line1\nLine2\nLine3", 10)
		expected := []string{"Line1", "Line2", "Line3"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("wrapTextArea with newlines = %v, want %v", result, expected)
		}
	})

	t.Run("wrapTextArea handles multiple spaces", func(t *testing.T) {
		result := wrapTextArea("Hello   World", 10)
		// strings.Fields collapses multiple spaces
		expected := []string{"Hello", "World"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("wrapTextArea with multiple spaces = %v, want %v", result, expected)
		}
	})
}

func TestTextAreaEdgeCases(t *testing.T) {
	t.Run("handles empty value gracefully", func(t *testing.T) {
		ta := TextArea().Value("")
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles negative width gracefully", func(t *testing.T) {
		ta := TextArea().Value("hello").Width(-5)
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles negative height gracefully", func(t *testing.T) {
		ta := TextArea().Value("hello").Height(-5)
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very long value", func(t *testing.T) {
		longValue := strings.Repeat("a", 1000)
		ta := TextArea().Value(longValue)
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nil callbacks gracefully", func(t *testing.T) {
		ta := TextArea().ID("test").Value("hello").
			OnChange(nil).OnKeyPress(nil).OnFocus(nil).OnBlur(nil).OnSubmit(nil)
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Error("Render returned wrong Element type with nil callbacks")
		}
	})

	t.Run("handles empty prefix and suffix", func(t *testing.T) {
		ta := TextArea().Value("hello").Prefix("").Suffix("")
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles Unicode characters", func(t *testing.T) {
		unicodeValue := "✓ 完成 ✅"
		ta := TextArea().Value(unicodeValue)
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles multiline text", func(t *testing.T) {
		multiline := "Line 1\nLine 2\nLine 3"
		ta := TextArea().Value(multiline)
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("validates min length correctly", func(t *testing.T) {
		ta := TextArea().Value("hi").MinLength(5)
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("validates max length correctly", func(t *testing.T) {
		ta := TextArea().Value("too long").MaxLength(5)
		elem := ta.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

// Benchmark tests
func BenchmarkTextAreaRender(b *testing.B) {
	ta := TextArea().Value("hello\nworld").Width(20).Height(5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ta.Render()
	}
}

func BenchmarkRenderTextArea(b *testing.B) {
	config := &TextAreaConfig{
		ID:     "bench",
		Value:  "hello\nworld",
		Width:  20,
		Height: 5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderTextArea(false, config)
	}
}

// Example usage test
func ExampleTextArea() {
	// Simple text area
	textArea := TextArea().
		ID("bio").
		Placeholder("Enter your bio").
		Width(40).
		Height(5).
		MinLength(10).
		MaxLength(500).
		Prefix("📝 ").
		Suffix(" ✏️").
		OnChange(func(id string, value string) {
			// Handle change
		}).
		OnSubmit(func(id string, value string) {
			// Handle submit
		}).
		Render()
	_ = textArea
}
