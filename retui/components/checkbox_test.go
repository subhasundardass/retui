package components

import (
	"reflect"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestCheckboxBuilder_Defaults(t *testing.T) {
	cb := Checkbox()

	if cb.config.ID != "" {
		t.Errorf("expected empty ID, got %q", cb.config.ID)
	}
	if cb.config.Label != "" {
		t.Errorf("expected empty Label, got %q", cb.config.Label)
	}
	if cb.config.Width != 0 {
		t.Errorf("expected Width 0, got %d", cb.config.Width)
	}
	if cb.config.Checked {
		t.Errorf("expected Checked false, got %v", cb.config.Checked)
	}
	if cb.focused {
		t.Errorf("expected focused false, got %v", cb.focused)
	}
}

func TestCheckboxBuilder_StringFields(t *testing.T) {
	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{"ID", func() string { return Checkbox().ID("test-id").config.ID }, "test-id"},
		{"Label", func() string { return Checkbox().Label("Click Me").config.Label }, "Click Me"},
	}

	for _, tt := range tests {
		t.Run("sets "+tt.name+" correctly", func(t *testing.T) {
			if got := tt.got(); got != tt.want {
				t.Errorf("expected %s %q, got %q", tt.name, tt.want, got)
			}
		})
	}
}

func TestCheckboxBuilder_Width(t *testing.T) {
	cb := Checkbox().Width(20)
	if cb.config.Width != 20 {
		t.Errorf("expected Width 20, got %d", cb.config.Width)
	}
}

func TestCheckboxBuilder_Checked(t *testing.T) {
	t.Run("sets Checked to true", func(t *testing.T) {
		cb := Checkbox().Checked(true)
		if !cb.config.Checked {
			t.Errorf("expected Checked true, got %v", cb.config.Checked)
		}
	})

	t.Run("sets Checked to false", func(t *testing.T) {
		cb := Checkbox().Checked(false)
		if cb.config.Checked {
			t.Errorf("expected Checked false, got %v", cb.config.Checked)
		}
	})
}

func TestCheckboxBuilder_Styles(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Background(retui.Blue)
		cb := Checkbox().Style(style)
		if !reflect.DeepEqual(cb.config.Style, style) {
			t.Errorf("Style not set correctly")
		}
	})

	t.Run("sets CheckedStyle correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Green).Bold(true)
		cb := Checkbox().CheckedStyle(style)
		if !reflect.DeepEqual(cb.config.CheckedStyle, style) {
			t.Errorf("CheckedStyle not set correctly")
		}
	})

	t.Run("sets UncheckedStyle correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.BrightBlack)
		cb := Checkbox().UncheckedStyle(style)
		if !reflect.DeepEqual(cb.config.UncheckedStyle, style) {
			t.Errorf("UncheckedStyle not set correctly")
		}
	})
}

func TestCheckboxBuilder_State(t *testing.T) {
	t.Run("sets Focused state correctly", func(t *testing.T) {
		cb := Checkbox().Focused(true)
		if !cb.focused {
			t.Errorf("expected focused true, got %v", cb.focused)
		}
	})

	t.Run("sets Focused state to false", func(t *testing.T) {
		cb := Checkbox().Focused(false)
		if cb.focused {
			t.Errorf("expected focused false, got %v", cb.focused)
		}
	})
}

func TestCheckboxCallbacks(t *testing.T) {
	t.Run("OnChange callback is set and invoked correctly", func(t *testing.T) {
		var changed bool
		var capturedID string
		var capturedChecked bool
		cb := Checkbox().ID("test-cb").OnChange(func(id string, checked bool) {
			changed = true
			capturedID = id
			capturedChecked = checked
		})

		if cb.config.OnChange == nil {
			t.Fatal("OnChange callback not set")
		}
		cb.config.OnChange("test-cb", true)
		if !changed {
			t.Error("OnChange callback was not executed")
		}
		if capturedID != "test-cb" {
			t.Errorf("expected id 'test-cb', got %s", capturedID)
		}
		if !capturedChecked {
			t.Error("expected checked true, got false")
		}
	})

	t.Run("OnKeyPress callback is set and invoked correctly", func(t *testing.T) {
		var keyPressed bool
		cb := Checkbox().ID("test-cb").OnKeyPress(func(id string, key retui.Key) bool {
			keyPressed = true
			if id != "test-cb" {
				t.Errorf("expected id 'test-cb', got %s", id)
			}
			return true
		})

		if cb.config.OnKeyPress == nil {
			t.Fatal("OnKeyPress callback not set")
		}
		if result := cb.config.OnKeyPress("test-cb", retui.Key{Code: retui.KeySpace}); !result {
			t.Error("OnKeyPress should return true")
		}
		if !keyPressed {
			t.Error("OnKeyPress callback was not executed")
		}
	})

	t.Run("OnFocus callback is set and invoked correctly", func(t *testing.T) {
		var focused bool
		cb := Checkbox().ID("test-cb").OnFocus(func(id string) {
			focused = true
			if id != "test-cb" {
				t.Errorf("expected id 'test-cb', got %s", id)
			}
		})

		if cb.config.OnFocus == nil {
			t.Fatal("OnFocus callback not set")
		}
		cb.config.OnFocus("test-cb")
		if !focused {
			t.Error("OnFocus callback was not executed")
		}
	})

	t.Run("OnBlur callback is set and invoked correctly", func(t *testing.T) {
		var blurred bool
		cb := Checkbox().ID("test-cb").OnBlur(func(id string) {
			blurred = true
			if id != "test-cb" {
				t.Errorf("expected id 'test-cb', got %s", id)
			}
		})

		if cb.config.OnBlur == nil {
			t.Fatal("OnBlur callback not set")
		}
		cb.config.OnBlur("test-cb")
		if !blurred {
			t.Error("OnBlur callback was not executed")
		}
	})
}

func TestCheckboxRender(t *testing.T) {
	t.Run("returns an Element of type ElementText", func(t *testing.T) {
		elem := Checkbox().Label("Test").Render()
		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
	})

	t.Run("renders unchecked checkbox by default", func(t *testing.T) {
		elem := Checkbox().Label("Test").Render()
		if elem.Text != "[ ] Test" {
			t.Errorf("expected '[ ] Test', got %q", elem.Text)
		}
	})

	t.Run("renders checked checkbox when Checked is true", func(t *testing.T) {
		elem := Checkbox().Checked(true).Label("Test").Render()
		if elem.Text != "[✓] Test" {
			t.Errorf("expected '[✓] Test', got %q", elem.Text)
		}
	})

	t.Run("renders without label when empty", func(t *testing.T) {
		elem := Checkbox().Render()
		if elem.Text != "[ ]" {
			t.Errorf("expected '[ ]', got %q", elem.Text)
		}
	})

	t.Run("renders checked without label", func(t *testing.T) {
		elem := Checkbox().Checked(true).Render()
		if elem.Text != "[✓]" {
			t.Errorf("expected '[✓]', got %q", elem.Text)
		}
	})

	t.Run("renders with width padding", func(t *testing.T) {
		elem := Checkbox().Label("Test").Width(20).Render()
		if len(elem.Text) < 20 {
			t.Errorf("expected text length at least 20, got %d", len(elem.Text))
		}
	})
}

func TestRenderCheckboxFunction(t *testing.T) {
	t.Run("renders unchecked checkbox when not focused", func(t *testing.T) {
		config := &CheckboxConfig{
			ID:      "test",
			Checked: false,
			Label:   "Test",
		}
		elem := renderCheckbox(false, config)
		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		if elem.Text != "[ ] Test" {
			t.Errorf("expected '[ ] Test', got %q", elem.Text)
		}
	})

	t.Run("renders checked checkbox", func(t *testing.T) {
		config := &CheckboxConfig{
			ID:      "test",
			Checked: true,
			Label:   "Test",
		}
		elem := renderCheckbox(false, config)
		if elem.Text != "[✓] Test" {
			t.Errorf("expected '[✓] Test', got %q", elem.Text)
		}
	})

	t.Run("renders with focus style when focused", func(t *testing.T) {
		config := &CheckboxConfig{
			ID:      "test",
			Checked: false,
			Label:   "Test",
		}
		elem := renderCheckbox(true, config)
		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}

		// Check style using reflect.DeepEqual
		expectedStyle := retui.NewStyle().Foreground(retui.Cyan).Bold(true)
		if !reflect.DeepEqual(elem.Style, expectedStyle) {
			t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
		}
	})

	t.Run("renders checked with checked style", func(t *testing.T) {
		config := &CheckboxConfig{
			ID:      "test",
			Checked: true,
			Label:   "Test",
		}
		elem := renderCheckbox(false, config)

		// Should have white foreground and bold
		expectedStyle := retui.NewStyle().Foreground(retui.White).Bold(true)
		if !reflect.DeepEqual(elem.Style, expectedStyle) {
			t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
		}
	})

	t.Run("triggers OnFocus when focused", func(t *testing.T) {
		var focusCalled bool
		config := &CheckboxConfig{
			ID: "test",
			OnFocus: func(id string) {
				focusCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderCheckbox(true, config)
		if !focusCalled {
			t.Error("OnFocus was not called when focused")
		}
	})

	t.Run("triggers OnBlur when not focused", func(t *testing.T) {
		var blurCalled bool
		config := &CheckboxConfig{
			ID: "test",
			OnBlur: func(id string) {
				blurCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderCheckbox(false, config)
		if !blurCalled {
			t.Error("OnBlur was not called when not focused")
		}
	})

	t.Run("does not trigger focus/blur without ID", func(t *testing.T) {
		var focusCalled, blurCalled bool
		config := &CheckboxConfig{
			ID:      "",
			OnFocus: func(id string) { focusCalled = true },
			OnBlur:  func(id string) { blurCalled = true },
		}

		renderCheckbox(true, config)
		if focusCalled {
			t.Error("OnFocus should not be called without ID")
		}

		renderCheckbox(false, config)
		if blurCalled {
			t.Error("OnBlur should not be called without ID")
		}
	})

	t.Run("handles OnKeyPress callback", func(t *testing.T) {
		config := &CheckboxConfig{
			ID: "test",
			OnKeyPress: func(id string, key retui.Key) bool {
				return true
			},
		}
		elem := renderCheckbox(true, config)
		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
	})
}

func TestCheckboxChaining(t *testing.T) {
	cb := Checkbox().
		ID("chain-test").
		Label("Chain").
		Width(15).
		Checked(true).
		Focused(true)

	if cb.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", cb.config.ID)
	}
	if cb.config.Label != "Chain" {
		t.Errorf("expected Label 'Chain', got %s", cb.config.Label)
	}
	if cb.config.Width != 15 {
		t.Errorf("expected Width 15, got %d", cb.config.Width)
	}
	if !cb.config.Checked {
		t.Errorf("expected Checked true, got %v", cb.config.Checked)
	}
	if !cb.focused {
		t.Errorf("expected focused true, got %v", cb.focused)
	}
}

func TestCheckboxDefaultStyles(t *testing.T) {
	t.Run("has default checked style", func(t *testing.T) {
		cb := Checkbox()
		expected := retui.NewStyle().Foreground(retui.Green).Bold(true)

		if !reflect.DeepEqual(cb.config.CheckedStyle, expected) {
			t.Errorf("expected checked style %+v, got %+v", expected, cb.config.CheckedStyle)
		}
	})

	t.Run("has default unchecked style", func(t *testing.T) {
		cb := Checkbox()
		expected := retui.NewStyle().Foreground(retui.BrightBlack)

		if !reflect.DeepEqual(cb.config.UncheckedStyle, expected) {
			t.Errorf("expected unchecked style %+v, got %+v", expected, cb.config.UncheckedStyle)
		}
	})
}

func TestCheckboxEdgeCases(t *testing.T) {
	t.Run("handles empty label gracefully", func(t *testing.T) {
		cb := Checkbox().Label("")
		elem := cb.Render()
		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		if elem.Text != "[ ]" {
			t.Errorf("expected '[ ]', got %q", elem.Text)
		}
	})

	t.Run("handles negative width gracefully", func(t *testing.T) {
		cb := Checkbox().Label("Test").Width(-5)
		elem := cb.Render()
		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
	})

	t.Run("handles very long labels", func(t *testing.T) {
		longLabel := "This is a very long checkbox label that should still render properly"
		cb := Checkbox().Label(longLabel)
		elem := cb.Render()
		if elem.Text != "[ ] "+longLabel {
			t.Errorf("expected label in text, got %q", elem.Text)
		}
	})

	t.Run("handles nil callbacks gracefully", func(t *testing.T) {
		cb := Checkbox().ID("test").OnChange(nil).OnKeyPress(nil).OnFocus(nil).OnBlur(nil)
		elem := cb.Render()
		if elem.Type != retui.ElementText {
			t.Error("Render returned wrong Element type with nil callbacks")
		}
	})

	t.Run("toggles checked state on render", func(t *testing.T) {
		config := &CheckboxConfig{
			ID:      "test",
			Checked: false,
			Label:   "Test",
		}
		elem := renderCheckbox(false, config)
		if elem.Text != "[ ] Test" {
			t.Errorf("expected '[ ] Test', got %q", elem.Text)
		}
	})
}

// Benchmark tests
func BenchmarkCheckboxRender(b *testing.B) {
	cb := Checkbox().Label("Benchmark Checkbox").Width(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Render()
	}
}

func BenchmarkRenderCheckbox(b *testing.B) {
	config := &CheckboxConfig{
		ID:      "bench",
		Checked: false,
		Label:   "Benchmark Checkbox",
		Width:   20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderCheckbox(false, config)
	}
}

// Example usage test
func ExampleCheckbox() {
	// Create a checkbox
	agreeCheckbox := Checkbox().
		ID("agree").
		Checked(false).
		Label("I agree to the terms").
		OnChange(func(id string, checked bool) {
			println("Checkbox", id, "changed to:", checked)
		}).
		Render()
	_ = agreeCheckbox
}
