package components

import (
	"reflect"
	"strings"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestPasswordBuilder_Defaults(t *testing.T) {
	p := Password()

	if p.config.ID != "" {
		t.Errorf("expected empty ID, got %q", p.config.ID)
	}
	if p.config.Value != "" {
		t.Errorf("expected empty Value, got %q", p.config.Value)
	}
	if p.config.Placeholder != "Enter password" {
		t.Errorf("expected Placeholder 'Enter password', got %q", p.config.Placeholder)
	}
	if p.config.Width != 30 {
		t.Errorf("expected Width 30, got %d", p.config.Width)
	}
	if p.config.Prefix != "[ " {
		t.Errorf("expected Prefix '[ ', got %q", p.config.Prefix)
	}
	if p.config.Suffix != " ]" {
		t.Errorf("expected Suffix ' ]', got %q", p.config.Suffix)
	}
	if p.config.MinLength != 0 {
		t.Errorf("expected MinLength 0, got %d", p.config.MinLength)
	}
	if p.config.MaxLength != 0 {
		t.Errorf("expected MaxLength 0, got %d", p.config.MaxLength)
	}
	if p.config.MaskChar != "•" {
		t.Errorf("expected MaskChar '•', got %q", p.config.MaskChar)
	}
	if !p.config.ShowLastChar {
		t.Errorf("expected ShowLastChar true, got %v", p.config.ShowLastChar)
	}
	if p.focused {
		t.Errorf("expected focused false, got %v", p.focused)
	}
}

func TestPasswordBuilder_StringFields(t *testing.T) {
	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{"ID", func() string { return Password().ID("test-id").config.ID }, "test-id"},
		{"Value", func() string { return Password().Value("secret").config.Value }, "secret"},
		{"Placeholder", func() string { return Password().Placeholder("Enter PIN").config.Placeholder }, "Enter PIN"},
		{"Prefix", func() string { return Password().Prefix("🔒 ").config.Prefix }, "🔒 "},
		{"Suffix", func() string { return Password().Suffix(" ✓").config.Suffix }, " ✓"},
		{"MaskChar", func() string { return Password().MaskChar("*").config.MaskChar }, "*"},
	}

	for _, tt := range tests {
		t.Run("sets "+tt.name+" correctly", func(t *testing.T) {
			if got := tt.got(); got != tt.want {
				t.Errorf("expected %s %q, got %q", tt.name, tt.want, got)
			}
		})
	}
}

func TestPasswordBuilder_Width(t *testing.T) {
	p := Password().Width(25)
	if p.config.Width != 25 {
		t.Errorf("expected Width 25, got %d", p.config.Width)
	}
}

func TestPasswordBuilder_Style(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Background(retui.Blue)
		p := Password().Style(style)
		if !reflect.DeepEqual(p.config.Style, style) {
			t.Errorf("Style not set correctly")
		}
	})
}

func TestPasswordBuilder_MinLength(t *testing.T) {
	t.Run("sets MinLength correctly", func(t *testing.T) {
		p := Password().MinLength(8)
		if p.config.MinLength != 8 {
			t.Errorf("expected MinLength 8, got %d", p.config.MinLength)
		}
	})

	t.Run("sets MinLength to 0", func(t *testing.T) {
		p := Password().MinLength(0)
		if p.config.MinLength != 0 {
			t.Errorf("expected MinLength 0, got %d", p.config.MinLength)
		}
	})
}

func TestPasswordBuilder_MaxLength(t *testing.T) {
	t.Run("sets MaxLength correctly", func(t *testing.T) {
		p := Password().MaxLength(20)
		if p.config.MaxLength != 20 {
			t.Errorf("expected MaxLength 20, got %d", p.config.MaxLength)
		}
	})

	t.Run("sets MaxLength to 0", func(t *testing.T) {
		p := Password().MaxLength(0)
		if p.config.MaxLength != 0 {
			t.Errorf("expected MaxLength 0, got %d", p.config.MaxLength)
		}
	})
}

func TestPasswordBuilder_ShowLastChar(t *testing.T) {
	t.Run("sets ShowLastChar true", func(t *testing.T) {
		p := Password().ShowLastChar(true)
		if !p.config.ShowLastChar {
			t.Errorf("expected ShowLastChar true, got %v", p.config.ShowLastChar)
		}
	})

	t.Run("sets ShowLastChar false", func(t *testing.T) {
		p := Password().ShowLastChar(false)
		if p.config.ShowLastChar {
			t.Errorf("expected ShowLastChar false, got %v", p.config.ShowLastChar)
		}
	})
}

func TestPasswordBuilder_State(t *testing.T) {
	t.Run("sets Focused state correctly", func(t *testing.T) {
		p := Password().Focused(true)
		if !p.focused {
			t.Errorf("expected focused true, got %v", p.focused)
		}
	})

	t.Run("sets Focused state to false", func(t *testing.T) {
		p := Password().Focused(false)
		if p.focused {
			t.Errorf("expected focused false, got %v", p.focused)
		}
	})
}

func TestPasswordCallbacks(t *testing.T) {
	t.Run("OnChange callback is set and invoked correctly", func(t *testing.T) {
		var changed bool
		var capturedID string
		var capturedValue string
		p := Password().ID("test-pass").OnChange(func(id string, value string) {
			changed = true
			capturedID = id
			capturedValue = value
		})

		if p.config.OnChange == nil {
			t.Fatal("OnChange callback not set")
		}
		p.config.OnChange("test-pass", "newpass")
		if !changed {
			t.Error("OnChange callback was not executed")
		}
		if capturedID != "test-pass" {
			t.Errorf("expected id 'test-pass', got %s", capturedID)
		}
		if capturedValue != "newpass" {
			t.Errorf("expected value 'newpass', got %s", capturedValue)
		}
	})

	t.Run("OnKeyPress callback is set and invoked correctly", func(t *testing.T) {
		var keyPressed bool
		p := Password().ID("test-pass").OnKeyPress(func(id string, key retui.Key) bool {
			keyPressed = true
			if id != "test-pass" {
				t.Errorf("expected id 'test-pass', got %s", id)
			}
			return true
		})

		if p.config.OnKeyPress == nil {
			t.Fatal("OnKeyPress callback not set")
		}
		if result := p.config.OnKeyPress("test-pass", retui.Key{Code: retui.KeyEnter}); !result {
			t.Error("OnKeyPress should return true")
		}
		if !keyPressed {
			t.Error("OnKeyPress callback was not executed")
		}
	})

	t.Run("OnFocus callback is set and invoked correctly", func(t *testing.T) {
		var focused bool
		p := Password().ID("test-pass").OnFocus(func(id string) {
			focused = true
			if id != "test-pass" {
				t.Errorf("expected id 'test-pass', got %s", id)
			}
		})

		if p.config.OnFocus == nil {
			t.Fatal("OnFocus callback not set")
		}
		p.config.OnFocus("test-pass")
		if !focused {
			t.Error("OnFocus callback was not executed")
		}
	})

	t.Run("OnBlur callback is set and invoked correctly", func(t *testing.T) {
		var blurred bool
		p := Password().ID("test-pass").OnBlur(func(id string) {
			blurred = true
			if id != "test-pass" {
				t.Errorf("expected id 'test-pass', got %s", id)
			}
		})

		if p.config.OnBlur == nil {
			t.Fatal("OnBlur callback not set")
		}
		p.config.OnBlur("test-pass")
		if !blurred {
			t.Error("OnBlur callback was not executed")
		}
	})

	t.Run("OnSubmit callback is set and invoked correctly", func(t *testing.T) {
		var submitted bool
		var capturedID string
		var capturedValue string
		p := Password().ID("test-pass").OnSubmit(func(id string, value string) {
			submitted = true
			capturedID = id
			capturedValue = value
		})

		if p.config.OnSubmit == nil {
			t.Fatal("OnSubmit callback not set")
		}
		p.config.OnSubmit("test-pass", "secret")
		if !submitted {
			t.Error("OnSubmit callback was not executed")
		}
		if capturedID != "test-pass" {
			t.Errorf("expected id 'test-pass', got %s", capturedID)
		}
		if capturedValue != "secret" {
			t.Errorf("expected value 'secret', got %s", capturedValue)
		}
	})
}

func TestPasswordRender(t *testing.T) {
	t.Run("returns an Element of type ElementBox", func(t *testing.T) {
		p := Password()
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with placeholder when empty", func(t *testing.T) {
		p := Password().Placeholder("Enter password")
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with masked value", func(t *testing.T) {
		p := Password().Value("secret")
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with prefix", func(t *testing.T) {
		p := Password().Prefix("🔒 ").Value("secret")
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with suffix", func(t *testing.T) {
		p := Password().Suffix(" ✓").Value("secret")
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with width", func(t *testing.T) {
		p := Password().Value("secret").Width(20)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestRenderPasswordFunction(t *testing.T) {
	t.Run("renders with default style when not focused", func(t *testing.T) {
		config := &PasswordConfig{
			ID:    "test",
			Value: "secret",
			Width: 20,
		}
		elem := renderPassword(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with focus style when focused", func(t *testing.T) {
		config := &PasswordConfig{
			ID:    "test",
			Value: "secret",
			Width: 20,
		}
		elem := renderPassword(true, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with placeholder when empty", func(t *testing.T) {
		config := &PasswordConfig{
			ID:          "test",
			Value:       "",
			Placeholder: "Enter password",
			Width:       20,
		}
		elem := renderPassword(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with custom mask character", func(t *testing.T) {
		config := &PasswordConfig{
			ID:       "test",
			Value:    "secret",
			MaskChar: "*",
			Width:    20,
		}
		elem := renderPassword(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("triggers OnFocus when focused", func(t *testing.T) {
		var focusCalled bool
		config := &PasswordConfig{
			ID:    "test",
			Value: "secret",
			OnFocus: func(id string) {
				focusCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderPassword(true, config)
		if !focusCalled {
			t.Error("OnFocus was not called when focused")
		}
	})

	t.Run("triggers OnBlur when not focused", func(t *testing.T) {
		var blurCalled bool
		config := &PasswordConfig{
			ID:    "test",
			Value: "secret",
			OnBlur: func(id string) {
				blurCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderPassword(false, config)
		if !blurCalled {
			t.Error("OnBlur was not called when not focused")
		}
	})

	t.Run("does not trigger focus/blur without ID", func(t *testing.T) {
		var focusCalled, blurCalled bool
		config := &PasswordConfig{
			ID:      "",
			Value:   "secret",
			OnFocus: func(id string) { focusCalled = true },
			OnBlur:  func(id string) { blurCalled = true },
		}

		renderPassword(true, config)
		if focusCalled {
			t.Error("OnFocus should not be called without ID")
		}

		renderPassword(false, config)
		if blurCalled {
			t.Error("OnBlur should not be called without ID")
		}
	})
}

func TestPasswordChaining(t *testing.T) {
	p := Password().
		ID("chain-test").
		Value("secret").
		Placeholder("Enter password").
		Width(25).
		Prefix("🔒 ").
		Suffix(" ✓").
		MinLength(8).
		MaxLength(20).
		MaskChar("*").
		ShowLastChar(false).
		Focused(true)

	if p.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", p.config.ID)
	}
	if p.config.Value != "secret" {
		t.Errorf("expected Value 'secret', got %s", p.config.Value)
	}
	if p.config.Placeholder != "Enter password" {
		t.Errorf("expected Placeholder 'Enter password', got %s", p.config.Placeholder)
	}
	if p.config.Width != 25 {
		t.Errorf("expected Width 25, got %d", p.config.Width)
	}
	if p.config.Prefix != "🔒 " {
		t.Errorf("expected Prefix '🔒 ', got %s", p.config.Prefix)
	}
	if p.config.Suffix != " ✓" {
		t.Errorf("expected Suffix ' ✓', got %s", p.config.Suffix)
	}
	if p.config.MinLength != 8 {
		t.Errorf("expected MinLength 8, got %d", p.config.MinLength)
	}
	if p.config.MaxLength != 20 {
		t.Errorf("expected MaxLength 20, got %d", p.config.MaxLength)
	}
	if p.config.MaskChar != "*" {
		t.Errorf("expected MaskChar '*', got %s", p.config.MaskChar)
	}
	if p.config.ShowLastChar {
		t.Errorf("expected ShowLastChar false, got %v", p.config.ShowLastChar)
	}
	if !p.focused {
		t.Errorf("expected focused true, got %v", p.focused)
	}
}

func TestPasswordHelperFunctions(t *testing.T) {
	t.Run("maskPassword masks password correctly", func(t *testing.T) {
		tests := []struct {
			value    string
			maskChar string
			want     string
		}{
			{"secret", "•", "••••••"},
			{"secret", "*", "******"},
			{"", "•", ""},
			{"hello", "●", "●●●●●"},
			{"1234", "x", "xxxx"},
			{"password123", "•", "•••••••••••"},
		}
		for _, tt := range tests {
			if got := maskPassword(tt.value, tt.maskChar); got != tt.want {
				t.Errorf("maskPassword(%q, %q) = %q, want %q", tt.value, tt.maskChar, got, tt.want)
			}
		}
	})

	t.Run("maskPassword handles Unicode characters", func(t *testing.T) {
		value := "密码"
		maskChar := "•"
		expected := "••"
		if got := maskPassword(value, maskChar); got != expected {
			t.Errorf("maskPassword(%q, %q) = %q, want %q", value, maskChar, got, expected)
		}
	})
}

func TestPasswordEdgeCases(t *testing.T) {
	t.Run("handles negative width gracefully", func(t *testing.T) {
		p := Password().Value("secret").Width(-5)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles empty mask character gracefully", func(t *testing.T) {
		p := Password().Value("secret").MaskChar("")
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very long password", func(t *testing.T) {
		longPassword := strings.Repeat("a", 1000)
		p := Password().Value(longPassword)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nil callbacks gracefully", func(t *testing.T) {
		p := Password().ID("test").Value("secret").
			OnChange(nil).OnKeyPress(nil).OnFocus(nil).OnBlur(nil).OnSubmit(nil)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Error("Render returned wrong Element type with nil callbacks")
		}
	})

	t.Run("handles empty prefix and suffix", func(t *testing.T) {
		p := Password().Value("secret").Prefix("").Suffix("")
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("validates min length correctly", func(t *testing.T) {
		p := Password().Value("short").MinLength(8)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("validates max length correctly", func(t *testing.T) {
		p := Password().Value("too long password").MaxLength(5)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles ShowLastChar false", func(t *testing.T) {
		p := Password().Value("secret").ShowLastChar(false)
		elem := p.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

// Benchmark tests
func BenchmarkPasswordRender(b *testing.B) {
	p := Password().Value("secret").Width(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Render()
	}
}

func BenchmarkRenderPassword(b *testing.B) {
	config := &PasswordConfig{
		ID:    "bench",
		Value: "secret",
		Width: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderPassword(false, config)
	}
}

// Example usage test
func ExamplePassword() {
	// Simple password input
	passwordInput := Password().
		ID("password").
		Placeholder("Enter password").
		Width(30).
		MinLength(8).
		MaxLength(20).
		Prefix("🔒 ").
		Suffix(" ✓").
		OnChange(func(id string, value string) {
			// Handle change
		}).
		OnSubmit(func(id string, value string) {
			// Handle submit
		}).
		Render()
	_ = passwordInput

	// Password with custom mask
	customMask := Password().
		ID("pin").
		Placeholder("Enter PIN").
		Width(20).
		MaskChar("*").
		ShowLastChar(false).
		MaxLength(4).
		Prefix("PIN: ").
		Render()
	_ = customMask
}
