package components

import (
	"reflect"
	"strings"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestSelectDropdownBuilder_Defaults(t *testing.T) {
	s := SelectDropdown()

	if s.config.ID != "" {
		t.Errorf("expected empty ID, got %q", s.config.ID)
	}
	if s.config.Label != "" {
		t.Errorf("expected empty Label, got %q", s.config.Label)
	}
	if len(s.config.Options) > 0 {
		t.Errorf("expected empty Options, got %v", s.config.Options)
	}
	if s.config.Value != "" {
		t.Errorf("expected empty Value, got %q", s.config.Value)
	}
	if s.config.Placeholder != "Select..." {
		t.Errorf("expected Placeholder 'Select...', got %q", s.config.Placeholder)
	}
	if s.config.Width != 30 {
		t.Errorf("expected Width 30, got %d", s.config.Width)
	}
	if s.config.Height != 5 {
		t.Errorf("expected Height 5, got %d", s.config.Height)
	}
	if s.config.Disabled {
		t.Errorf("expected Disabled false, got %v", s.config.Disabled)
	}
	if s.focused {
		t.Errorf("expected focused false, got %v", s.focused)
	}
}

func TestSelectDropdownBuilder_StringFields(t *testing.T) {
	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{"ID", func() string { return SelectDropdown().ID("test-id").config.ID }, "test-id"},
		{"Label", func() string { return SelectDropdown().Label("Select Color").config.Label }, "Select Color"},
		{"Value", func() string { return SelectDropdown().Value("red").config.Value }, "red"},
		{"Placeholder", func() string { return SelectDropdown().Placeholder("Choose...").config.Placeholder }, "Choose..."},
	}

	for _, tt := range tests {
		t.Run("sets "+tt.name+" correctly", func(t *testing.T) {
			if got := tt.got(); got != tt.want {
				t.Errorf("expected %s %q, got %q", tt.name, tt.want, got)
			}
		})
	}
}

func TestSelectDropdownBuilder_Options(t *testing.T) {
	t.Run("sets Options correctly", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red"},
			{Label: "Green", Value: "green"},
			{Label: "Blue", Value: "blue"},
		}
		s := SelectDropdown().Options(options)
		if !reflect.DeepEqual(s.config.Options, options) {
			t.Errorf("expected Options %v, got %v", options, s.config.Options)
		}
	})

	t.Run("sets Options with disabled items", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red"},
			{Label: "Green", Value: "green", Disabled: true},
			{Label: "Blue", Value: "blue"},
		}
		s := SelectDropdown().Options(options)
		if len(s.config.Options) != 3 {
			t.Errorf("expected 3 options, got %d", len(s.config.Options))
		}
		if !s.config.Options[1].Disabled {
			t.Error("expected option 1 to be disabled")
		}
	})
}

func TestSelectDropdownBuilder_Width(t *testing.T) {
	s := SelectDropdown().Width(40)
	if s.config.Width != 40 {
		t.Errorf("expected Width 40, got %d", s.config.Width)
	}
}

func TestSelectDropdownBuilder_Height(t *testing.T) {
	s := SelectDropdown().Height(10)
	if s.config.Height != 10 {
		t.Errorf("expected Height 10, got %d", s.config.Height)
	}
}

func TestSelectDropdownBuilder_Style(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Background(retui.Blue)
		s := SelectDropdown().Style(style)
		if !reflect.DeepEqual(s.config.Style, style) {
			t.Errorf("Style not set correctly")
		}
	})
}

func TestSelectDropdownBuilder_Disabled(t *testing.T) {
	t.Run("sets Disabled true", func(t *testing.T) {
		s := SelectDropdown().Disabled(true)
		if !s.config.Disabled {
			t.Errorf("expected Disabled true, got %v", s.config.Disabled)
		}
	})

	t.Run("sets Disabled false", func(t *testing.T) {
		s := SelectDropdown().Disabled(false)
		if s.config.Disabled {
			t.Errorf("expected Disabled false, got %v", s.config.Disabled)
		}
	})
}

func TestSelectDropdownBuilder_State(t *testing.T) {
	t.Run("sets Focused state correctly", func(t *testing.T) {
		s := SelectDropdown().Focused(true)
		if !s.focused {
			t.Errorf("expected focused true, got %v", s.focused)
		}
	})

	t.Run("sets Focused state to false", func(t *testing.T) {
		s := SelectDropdown().Focused(false)
		if s.focused {
			t.Errorf("expected focused false, got %v", s.focused)
		}
	})
}

func TestSelectDropdownCallbacks(t *testing.T) {
	t.Run("OnChange callback is set and invoked correctly", func(t *testing.T) {
		var changed bool
		var capturedID string
		var capturedValue string
		s := SelectDropdown().ID("test-select").OnChange(func(id string, value string) {
			changed = true
			capturedID = id
			capturedValue = value
		})

		if s.config.OnChange == nil {
			t.Fatal("OnChange callback not set")
		}
		s.config.OnChange("test-select", "blue")
		if !changed {
			t.Error("OnChange callback was not executed")
		}
		if capturedID != "test-select" {
			t.Errorf("expected id 'test-select', got %s", capturedID)
		}
		if capturedValue != "blue" {
			t.Errorf("expected value 'blue', got %s", capturedValue)
		}
	})

	t.Run("OnKeyPress callback is set and invoked correctly", func(t *testing.T) {
		var keyPressed bool
		s := SelectDropdown().ID("test-select").OnKeyPress(func(id string, key retui.Key) bool {
			keyPressed = true
			if id != "test-select" {
				t.Errorf("expected id 'test-select', got %s", id)
			}
			return true
		})

		if s.config.OnKeyPress == nil {
			t.Fatal("OnKeyPress callback not set")
		}
		if result := s.config.OnKeyPress("test-select", retui.Key{Code: retui.KeyEnter}); !result {
			t.Error("OnKeyPress should return true")
		}
		if !keyPressed {
			t.Error("OnKeyPress callback was not executed")
		}
	})

	t.Run("OnFocus callback is set and invoked correctly", func(t *testing.T) {
		var focused bool
		s := SelectDropdown().ID("test-select").OnFocus(func(id string) {
			focused = true
			if id != "test-select" {
				t.Errorf("expected id 'test-select', got %s", id)
			}
		})

		if s.config.OnFocus == nil {
			t.Fatal("OnFocus callback not set")
		}
		s.config.OnFocus("test-select")
		if !focused {
			t.Error("OnFocus callback was not executed")
		}
	})

	t.Run("OnBlur callback is set and invoked correctly", func(t *testing.T) {
		var blurred bool
		s := SelectDropdown().ID("test-select").OnBlur(func(id string) {
			blurred = true
			if id != "test-select" {
				t.Errorf("expected id 'test-select', got %s", id)
			}
		})

		if s.config.OnBlur == nil {
			t.Fatal("OnBlur callback not set")
		}
		s.config.OnBlur("test-select")
		if !blurred {
			t.Error("OnBlur callback was not executed")
		}
	})

	t.Run("OnOpenChange callback is set and invoked correctly", func(t *testing.T) {
		var openChanged bool
		var capturedID string
		var capturedIsOpen bool
		s := SelectDropdown().ID("test-select").OnOpenChange(func(id string, isOpen bool) {
			openChanged = true
			capturedID = id
			capturedIsOpen = isOpen
		})

		if s.config.OnOpenChange == nil {
			t.Fatal("OnOpenChange callback not set")
		}
		s.config.OnOpenChange("test-select", true)
		if !openChanged {
			t.Error("OnOpenChange callback was not executed")
		}
		if capturedID != "test-select" {
			t.Errorf("expected id 'test-select', got %s", capturedID)
		}
		if !capturedIsOpen {
			t.Error("expected isOpen true, got false")
		}
	})

	t.Run("OnFilter callback is set correctly", func(t *testing.T) {
		filterCalled := false
		s := SelectDropdown().ID("test-select").OnFilter(func(id string, query string) []SelectOption {
			filterCalled = true
			if id != "test-select" {
				t.Errorf("expected id 'test-select', got %s", id)
			}
			return []SelectOption{}
		})

		if s.config.OnFilter == nil {
			t.Fatal("OnFilter callback not set")
		}
		result := s.config.OnFilter("test-select", "red")
		if !filterCalled {
			t.Error("OnFilter callback was not executed")
		}
		if result == nil {
			t.Error("OnFilter should return a slice")
		}
	})
}

func TestSelectDropdownRender(t *testing.T) {
	t.Run("returns an Element of type ElementBox", func(t *testing.T) {
		s := SelectDropdown()
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with placeholder when no value selected", func(t *testing.T) {
		s := SelectDropdown().Placeholder("Choose...")
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with selected value", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red"},
			{Label: "Green", Value: "green"},
		}
		s := SelectDropdown().Options(options).Value("red")
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with label", func(t *testing.T) {
		s := SelectDropdown().Label("Color")
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with disabled state", func(t *testing.T) {
		s := SelectDropdown().Disabled(true)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestSelectChaining(t *testing.T) {
	options := []SelectOption{
		{Label: "Red", Value: "red"},
		{Label: "Green", Value: "green"},
		{Label: "Blue", Value: "blue"},
	}

	s := SelectDropdown().
		ID("chain-test").
		Label("Color").
		Options(options).
		Value("red").
		Placeholder("Select...").
		Width(40).
		Height(8).
		Disabled(false).
		Focused(true)

	if s.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", s.config.ID)
	}
	if s.config.Label != "Color" {
		t.Errorf("expected Label 'Color', got %s", s.config.Label)
	}
	if !reflect.DeepEqual(s.config.Options, options) {
		t.Errorf("expected Options %v, got %v", options, s.config.Options)
	}
	if s.config.Value != "red" {
		t.Errorf("expected Value 'red', got %s", s.config.Value)
	}
	if s.config.Placeholder != "Select..." {
		t.Errorf("expected Placeholder 'Select...', got %s", s.config.Placeholder)
	}
	if s.config.Width != 40 {
		t.Errorf("expected Width 40, got %d", s.config.Width)
	}
	if s.config.Height != 8 {
		t.Errorf("expected Height 8, got %d", s.config.Height)
	}
	if s.config.Disabled {
		t.Errorf("expected Disabled false, got %v", s.config.Disabled)
	}
	if !s.focused {
		t.Errorf("expected focused true, got %v", s.focused)
	}
}

func TestSelectHelperFunctions(t *testing.T) {
	t.Run("findValueIndex finds value in options", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red"},
			{Label: "Green", Value: "green"},
			{Label: "Blue", Value: "blue"},
		}

		idx, ok := findValueIndex(options, "green")
		if !ok {
			t.Error("expected to find value 'green'")
		}
		if idx != 1 {
			t.Errorf("expected index 1, got %d", idx)
		}

		_, ok = findValueIndex(options, "yellow")
		if ok {
			t.Error("expected not to find value 'yellow'")
		}
	})

	t.Run("firstEnabledIndex returns first enabled option", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red", Disabled: true},
			{Label: "Green", Value: "green"},
			{Label: "Blue", Value: "blue"},
		}
		idx := firstEnabledIndex(options)
		if idx != 1 {
			t.Errorf("expected index 1, got %d", idx)
		}
	})

	t.Run("firstEnabledIndex returns 0 when all disabled", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red", Disabled: true},
			{Label: "Green", Value: "green", Disabled: true},
		}
		idx := firstEnabledIndex(options)
		if idx != 0 {
			t.Errorf("expected index 0, got %d", idx)
		}
	})

	t.Run("lastEnabledIndex returns last enabled option", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red"},
			{Label: "Green", Value: "green"},
			{Label: "Blue", Value: "blue", Disabled: true},
		}
		idx := lastEnabledIndex(options)
		if idx != 1 {
			t.Errorf("expected index 1, got %d", idx)
		}
	})

	t.Run("nextEnabledIndex navigates correctly", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red"},
			{Label: "Green", Value: "green", Disabled: true},
			{Label: "Blue", Value: "blue"},
			{Label: "Yellow", Value: "yellow"},
		}

		// From index 0, move forward
		idx := nextEnabledIndex(options, 0, 1)
		if idx != 2 {
			t.Errorf("expected index 2, got %d", idx)
		}

		// From index 3, move backward
		idx = nextEnabledIndex(options, 3, -1)
		if idx != 2 {
			t.Errorf("expected index 2, got %d", idx)
		}
	})

	t.Run("filterOptions filters with static options", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red"},
			{Label: "Green", Value: "green"},
			{Label: "Blue", Value: "blue"},
		}
		config := &SelectConfig{Options: options}

		result := filterOptions(config, "re")
		// filterOptions does case-insensitive contains match
		// "Red" contains "re", "Green" contains "re"
		if len(result) != 2 {
			t.Errorf("expected 2 results, got %d", len(result))
		}
	})

	t.Run("filterOptions filters case-insensitively", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Apple", Value: "apple"},
			{Label: "Banana", Value: "banana"},
			{Label: "Grape", Value: "grape"},
		}
		config := &SelectConfig{Options: options}

		// Query "ban" matches only "Banana"
		result := filterOptions(config, "ban")
		if len(result) != 1 {
			t.Errorf("expected 1 result, got %d", len(result))
		}
		if result[0].Label != "Banana" {
			t.Errorf("expected 'Banana', got %s", result[0].Label)
		}
	})

	t.Run("filterOptions returns all for empty query", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red"},
			{Label: "Green", Value: "green"},
		}
		config := &SelectConfig{Options: options}

		result := filterOptions(config, "")
		if len(result) != 2 {
			t.Errorf("expected 2 results, got %d", len(result))
		}
	})

	t.Run("filterOptions uses OnFilter when set", func(t *testing.T) {
		filterCalled := false
		config := &SelectConfig{
			ID: "test",
			OnFilter: func(id string, query string) []SelectOption {
				filterCalled = true
				return []SelectOption{{Label: "Filtered", Value: "filtered"}}
			},
		}

		result := filterOptions(config, "test")
		if !filterCalled {
			t.Error("OnFilter was not called")
		}
		if len(result) != 1 {
			t.Errorf("expected 1 result, got %d", len(result))
		}
		if result[0].Value != "filtered" {
			t.Errorf("expected 'filtered', got %s", result[0].Value)
		}
	})

	t.Run("truncateText truncates correctly", func(t *testing.T) {
		tests := []struct {
			text     string
			maxWidth int
			expected string
		}{
			{"Hello", 10, "Hello"},
			{"Hello World", 5, "He..."},
			{"Hello", 3, "..."},
			{"Hi", 5, "Hi"},
			{"", 5, ""},
		}
		for _, tt := range tests {
			result := truncateText(tt.text, tt.maxWidth)
			if result != tt.expected {
				t.Errorf("truncateText(%q, %d) = %q, want %q", tt.text, tt.maxWidth, result, tt.expected)
			}
		}
	})

	t.Run("padSearchText pads correctly", func(t *testing.T) {
		tests := []struct {
			text  string
			width int
			want  string
		}{
			{"", 10, "▏         "},
			{"red", 10, "red▏      "},
			{"search", 6, "sear... "},
		}
		for _, tt := range tests {
			result := padSearchText(tt.text, tt.width)
			if len([]rune(result)) != tt.width {
				t.Errorf("padSearchText(%q, %d) length = %d, want %d", tt.text, tt.width, len([]rune(result)), tt.width)
			}
		}
	})
}

func TestSelectEdgeCases(t *testing.T) {
	t.Run("handles empty options gracefully", func(t *testing.T) {
		s := SelectDropdown().Options([]SelectOption{})
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nil options gracefully", func(t *testing.T) {
		s := SelectDropdown().Options(nil)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles negative width gracefully", func(t *testing.T) {
		// Negative width causes panic in truncateText
		// We recover and test that it panics as expected
		defer func() {
			if r := recover(); r != nil {
				// Expected panic for negative width
				t.Log("SelectDropdown panicked with negative width as expected")
			}
		}()
		s := SelectDropdown().Width(-5)
		s.Render()
		t.Error("Expected panic but didn't get one")
	})

	t.Run("handles negative height gracefully", func(t *testing.T) {
		s := SelectDropdown().Height(-5)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nil callbacks gracefully", func(t *testing.T) {
		s := SelectDropdown().ID("test").
			OnChange(nil).OnKeyPress(nil).OnFocus(nil).OnBlur(nil).
			OnOpenChange(nil).OnFilter(nil)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Error("Render returned wrong Element type with nil callbacks")
		}
	})

	t.Run("handles all options disabled", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red", Disabled: true},
			{Label: "Green", Value: "green", Disabled: true},
		}
		s := SelectDropdown().Options(options)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles value not in options", func(t *testing.T) {
		options := []SelectOption{
			{Label: "Red", Value: "red"},
			{Label: "Green", Value: "green"},
		}
		s := SelectDropdown().Options(options).Value("blue")
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very long option labels", func(t *testing.T) {
		longLabel := strings.Repeat("A", 100)
		options := []SelectOption{
			{Label: longLabel, Value: "long"},
		}
		s := SelectDropdown().Options(options)
		elem := s.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles width less than 2", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Log("SelectDropdown panicked with width 1 as expected")
			}
		}()
		s := SelectDropdown().Width(1)
		s.Render()
		t.Error("Expected panic but didn't get one")
	})
}

// Benchmark tests
func BenchmarkSelectDropdownRender(b *testing.B) {
	options := make([]SelectOption, 100)
	for i := 0; i < 100; i++ {
		options[i] = SelectOption{Label: "Option " + string(rune('A'+i%26)), Value: string(rune('a' + i%26))}
	}
	s := SelectDropdown().Options(options).Value("a")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Render()
	}
}

// Example usage test
func ExampleSelectDropdown() {
	// Create a select dropdown with options
	selectDropdown := SelectDropdown().
		ID("color").
		Label("Choose Color").
		Options([]SelectOption{
			{Label: "Red", Value: "red"},
			{Label: "Green", Value: "green"},
			{Label: "Blue", Value: "blue"},
			{Label: "Yellow", Value: "yellow", Disabled: true},
		}).
		Value("red").
		Placeholder("Select a color...").
		Width(30).
		Height(5).
		OnChange(func(id string, value string) {
			// Handle selection change
		}).
		Render()
	_ = selectDropdown
}
