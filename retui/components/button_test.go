package components

import (
	"reflect"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestButtonBuilder_Defaults(t *testing.T) {
	btn := Button()

	if btn.config.ID != "" {
		t.Errorf("expected empty ID, got %q", btn.config.ID)
	}
	if btn.config.Label != "" {
		t.Errorf("expected empty Label, got %q", btn.config.Label)
	}
	if btn.config.Width != 0 {
		t.Errorf("expected Width 0, got %d", btn.config.Width)
	}
	if btn.focused {
		t.Errorf("expected focused false, got %v", btn.focused)
	}
	if btn.active {
		t.Errorf("expected active false, got %v", btn.active)
	}
}

func TestButtonBuilder_StringFields(t *testing.T) {
	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{"ID", func() string { return Button().ID("test-id").config.ID }, "test-id"},
		{"Label", func() string { return Button().Label("Click Me").config.Label }, "Click Me"},
		{"Prefix", func() string { return Button().Prefix("[").config.Prefix }, "["},
		{"Suffix", func() string { return Button().Suffix("]").config.Suffix }, "]"},
	}

	for _, tt := range tests {
		t.Run("sets "+tt.name+" correctly", func(t *testing.T) {
			if got := tt.got(); got != tt.want {
				t.Errorf("expected %s %q, got %q", tt.name, tt.want, got)
			}
		})
	}
}

func TestButtonBuilder_Width(t *testing.T) {
	btn := Button().Width(20)
	if btn.config.Width != 20 {
		t.Errorf("expected Width 20, got %d", btn.config.Width)
	}
}

func TestButtonBuilder_Styles(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Background(retui.Blue)
		btn := Button().Style(style)
		if !reflect.DeepEqual(btn.config.Style, style) {
			t.Errorf("Style not set correctly")
		}
	})

	t.Run("sets HoverStyle correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Yellow).Background(retui.Green)
		btn := Button().HoverStyle(style)
		if !reflect.DeepEqual(btn.config.HoverStyle, style) {
			t.Errorf("HoverStyle not set correctly")
		}
	})

	t.Run("sets ActiveStyle correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.White).Background(retui.Red)
		btn := Button().ActiveStyle(style)
		if !reflect.DeepEqual(btn.config.ActiveStyle, style) {
			t.Errorf("ActiveStyle not set correctly")
		}
	})
}

func TestButtonBuilder_State(t *testing.T) {
	t.Run("sets Focused state correctly", func(t *testing.T) {
		btn := Button().Focused(true)
		if !btn.focused {
			t.Errorf("expected focused true, got %v", btn.focused)
		}
	})

	t.Run("sets Active state correctly", func(t *testing.T) {
		btn := Button().Active(true)
		if !btn.active {
			t.Errorf("expected active true, got %v", btn.active)
		}
	})
}

func TestButtonCallbacks(t *testing.T) {
	t.Run("OnClick callback is set and invoked correctly", func(t *testing.T) {
		var clicked bool
		btn := Button().ID("test-btn").OnClick(func(id string) {
			clicked = true
			if id != "test-btn" {
				t.Errorf("expected id 'test-btn', got %s", id)
			}
		})

		if btn.config.OnClick == nil {
			t.Fatal("OnClick callback not set")
		}
		btn.config.OnClick("test-btn")
		if !clicked {
			t.Error("OnClick callback was not executed")
		}
	})

	t.Run("OnKeyPress callback is set and invoked correctly", func(t *testing.T) {
		var keyPressed bool
		btn := Button().ID("test-btn").OnKeyPress(func(id string, key retui.Key) bool {
			keyPressed = true
			if id != "test-btn" {
				t.Errorf("expected id 'test-btn', got %s", id)
			}
			return true
		})

		if btn.config.OnKeyPress == nil {
			t.Fatal("OnKeyPress callback not set")
		}
		if result := btn.config.OnKeyPress("test-btn", retui.Key{Code: retui.KeyEnter}); !result {
			t.Error("OnKeyPress should return true")
		}
		if !keyPressed {
			t.Error("OnKeyPress callback was not executed")
		}
	})

	t.Run("OnFocus callback is set and invoked correctly", func(t *testing.T) {
		var focused bool
		btn := Button().ID("test-btn").OnFocus(func(id string) {
			focused = true
			if id != "test-btn" {
				t.Errorf("expected id 'test-btn', got %s", id)
			}
		})

		if btn.config.OnFocus == nil {
			t.Fatal("OnFocus callback not set")
		}
		btn.config.OnFocus("test-btn")
		if !focused {
			t.Error("OnFocus callback was not executed")
		}
	})

	t.Run("OnBlur callback is set and invoked correctly", func(t *testing.T) {
		var blurred bool
		btn := Button().ID("test-btn").OnBlur(func(id string) {
			blurred = true
			if id != "test-btn" {
				t.Errorf("expected id 'test-btn', got %s", id)
			}
		})

		if btn.config.OnBlur == nil {
			t.Fatal("OnBlur callback not set")
		}
		btn.config.OnBlur("test-btn")
		if !blurred {
			t.Error("OnBlur callback was not executed")
		}
	})
}

func TestButtonRender(t *testing.T) {
	t.Run("returns an Element with text type", func(t *testing.T) {
		elem := Button().Label("Test").Render()
		if elem.Type != retui.ElementBox {
			t.Error("expected Render() to return an Element of type ElementBox")
		}
	})

	t.Run("renders with default label when empty", func(t *testing.T) {
		if elem := Button().Render(); elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("includes prefix in rendered output", func(t *testing.T) {
		if elem := Button().Label("Test").Prefix("[").Render(); elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("includes suffix in rendered output", func(t *testing.T) {
		if elem := Button().Label("Test").Suffix("]").Render(); elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestRenderButtonFunction(t *testing.T) {
	t.Run("renders with default style when not focused or active", func(t *testing.T) {
		config := &ButtonConfig{
			ID:    "test",
			Label: "Test",
			Style: retui.NewStyle().Foreground(retui.White),
		}
		if elem := renderButton(false, false, config); elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with hover style when focused", func(t *testing.T) {
		config := &ButtonConfig{
			ID:         "test",
			Label:      "Test",
			HoverStyle: retui.NewStyle().Foreground(retui.Yellow),
		}
		if elem := renderButton(true, false, config); elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with active style when active", func(t *testing.T) {
		config := &ButtonConfig{
			ID:          "test",
			Label:       "Test",
			ActiveStyle: retui.NewStyle().Foreground(retui.Green),
		}
		if elem := renderButton(false, true, config); elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles width padding", func(t *testing.T) {
		config := &ButtonConfig{
			ID:    "test",
			Label: "Test",
			Width: 20,
		}
		if elem := renderButton(false, false, config); elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("triggers OnFocus when focused", func(t *testing.T) {
		var focusCalled bool
		config := &ButtonConfig{
			ID: "test",
			OnFocus: func(id string) {
				focusCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderButton(true, false, config)
		if !focusCalled {
			t.Error("OnFocus was not called when focused")
		}
	})

	t.Run("triggers OnBlur when not focused", func(t *testing.T) {
		var blurCalled bool
		config := &ButtonConfig{
			ID: "test",
			OnBlur: func(id string) {
				blurCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderButton(false, false, config)
		if !blurCalled {
			t.Error("OnBlur was not called when not focused")
		}
	})

	t.Run("does not trigger focus/blur without ID", func(t *testing.T) {
		var focusCalled, blurCalled bool
		config := &ButtonConfig{
			ID:      "",
			OnFocus: func(id string) { focusCalled = true },
			OnBlur:  func(id string) { blurCalled = true },
		}

		renderButton(true, false, config)
		if focusCalled {
			t.Error("OnFocus should not be called without ID")
		}

		renderButton(false, false, config)
		if blurCalled {
			t.Error("OnBlur should not be called without ID")
		}
	})
}

func TestButtonChaining(t *testing.T) {
	btn := Button().
		ID("chain-test").
		Label("Chain").
		Width(15).
		Prefix("[").
		Suffix("]").
		Focused(true).
		Active(false)

	if btn.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", btn.config.ID)
	}
	if btn.config.Label != "Chain" {
		t.Errorf("expected Label 'Chain', got %s", btn.config.Label)
	}
	if btn.config.Width != 15 {
		t.Errorf("expected Width 15, got %d", btn.config.Width)
	}
	if btn.config.Prefix != "[" {
		t.Errorf("expected Prefix '[', got %s", btn.config.Prefix)
	}
	if btn.config.Suffix != "]" {
		t.Errorf("expected Suffix ']', got %s", btn.config.Suffix)
	}
	if !btn.focused {
		t.Errorf("expected focused true, got %v", btn.focused)
	}
	if btn.active {
		t.Errorf("expected active false, got %v", btn.active)
	}
}

func TestButtonDefaultStyles(t *testing.T) {
	expected := retui.NewStyle().Foreground(retui.White).Background(retui.Navy).Bold(true)

	t.Run("has default hover style", func(t *testing.T) {
		btn := Button()
		if !reflect.DeepEqual(btn.config.HoverStyle, expected) {
			t.Errorf("expected hover style %+v, got %+v", expected, btn.config.HoverStyle)
		}
	})

	t.Run("has default active style", func(t *testing.T) {
		btn := Button()
		if !reflect.DeepEqual(btn.config.ActiveStyle, expected) {
			t.Errorf("expected active style %+v, got %+v", expected, btn.config.ActiveStyle)
		}
	})
}
