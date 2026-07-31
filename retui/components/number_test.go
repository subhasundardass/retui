package components

import (
	"reflect"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestNumberInputBuilder_Defaults(t *testing.T) {
	n := NumberInput()

	if n.config.ID != "" {
		t.Errorf("expected empty ID, got %q", n.config.ID)
	}
	if n.config.Value != 0 {
		t.Errorf("expected Value 0, got %f", n.config.Value)
	}
	if !n.config.Empty {
		t.Errorf("expected Empty true, got %v", n.config.Empty)
	}
	if n.config.Placeholder != "0" {
		t.Errorf("expected Placeholder '0', got %q", n.config.Placeholder)
	}
	if n.config.Width != 30 {
		t.Errorf("expected Width 30, got %d", n.config.Width)
	}
	if n.config.Prefix != "" {
		t.Errorf("expected empty Prefix, got %q", n.config.Prefix)
	}
	if n.config.Suffix != "" {
		t.Errorf("expected empty Suffix, got %q", n.config.Suffix)
	}
	if n.config.HasMin {
		t.Errorf("expected HasMin false, got %v", n.config.HasMin)
	}
	if n.config.HasMax {
		t.Errorf("expected HasMax false, got %v", n.config.HasMax)
	}
	if n.config.Step != 1 {
		t.Errorf("expected Step 1, got %f", n.config.Step)
	}
	if n.config.Decimals != 0 {
		t.Errorf("expected Decimals 0, got %d", n.config.Decimals)
	}
	if n.config.ArrowStep {
		t.Errorf("expected ArrowStep false, got %v", n.config.ArrowStep)
	}
	if !n.config.SelectAllOnFocus {
		t.Errorf("expected SelectAllOnFocus true, got %v", n.config.SelectAllOnFocus)
	}
	if n.focused {
		t.Errorf("expected focused false, got %v", n.focused)
	}
}

func TestNumberInputBuilder_StringFields(t *testing.T) {
	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{"ID", func() string { return NumberInput().ID("test-id").config.ID }, "test-id"},
		{"Placeholder", func() string { return NumberInput().Placeholder("Enter number").config.Placeholder }, "Enter number"},
		{"Prefix", func() string { return NumberInput().Prefix("$ ").config.Prefix }, "$ "},
		{"Suffix", func() string { return NumberInput().Suffix(" USD").config.Suffix }, " USD"},
	}

	for _, tt := range tests {
		t.Run("sets "+tt.name+" correctly", func(t *testing.T) {
			if got := tt.got(); got != tt.want {
				t.Errorf("expected %s %q, got %q", tt.name, tt.want, got)
			}
		})
	}
}

func TestNumberInputBuilder_Value(t *testing.T) {
	t.Run("sets Value and marks Empty false", func(t *testing.T) {
		n := NumberInput().Value(42.5)
		if n.config.Value != 42.5 {
			t.Errorf("expected Value 42.5, got %f", n.config.Value)
		}
		if n.config.Empty {
			t.Errorf("expected Empty false, got %v", n.config.Empty)
		}
	})

	t.Run("sets Value to 0 and marks Empty false", func(t *testing.T) {
		n := NumberInput().Value(0)
		if n.config.Value != 0 {
			t.Errorf("expected Value 0, got %f", n.config.Value)
		}
		if n.config.Empty {
			t.Errorf("expected Empty false, got %v", n.config.Empty)
		}
	})
}

func TestNumberInputBuilder_Empty(t *testing.T) {
	t.Run("sets Empty true", func(t *testing.T) {
		n := NumberInput().Empty(true)
		if !n.config.Empty {
			t.Errorf("expected Empty true, got %v", n.config.Empty)
		}
	})

	t.Run("sets Empty false", func(t *testing.T) {
		n := NumberInput().Empty(false)
		if n.config.Empty {
			t.Errorf("expected Empty false, got %v", n.config.Empty)
		}
	})
}

func TestNumberInputBuilder_Width(t *testing.T) {
	n := NumberInput().Width(25)
	if n.config.Width != 25 {
		t.Errorf("expected Width 25, got %d", n.config.Width)
	}
}

func TestNumberInputBuilder_Style(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Background(retui.Blue)
		n := NumberInput().Style(style)
		if !reflect.DeepEqual(n.config.Style, style) {
			t.Errorf("Style not set correctly")
		}
	})
}

func TestNumberInputBuilder_Bounds(t *testing.T) {
	t.Run("sets Min and enables HasMin", func(t *testing.T) {
		n := NumberInput().Min(10.5)
		if n.config.Min != 10.5 {
			t.Errorf("expected Min 10.5, got %f", n.config.Min)
		}
		if !n.config.HasMin {
			t.Errorf("expected HasMin true, got %v", n.config.HasMin)
		}
	})

	t.Run("sets Max and enables HasMax", func(t *testing.T) {
		n := NumberInput().Max(100.0)
		if n.config.Max != 100.0 {
			t.Errorf("expected Max 100.0, got %f", n.config.Max)
		}
		if !n.config.HasMax {
			t.Errorf("expected HasMax true, got %v", n.config.HasMax)
		}
	})

	t.Run("sets both Min and Max", func(t *testing.T) {
		n := NumberInput().Min(0).Max(100)
		if n.config.Min != 0 {
			t.Errorf("expected Min 0, got %f", n.config.Min)
		}
		if !n.config.HasMin {
			t.Errorf("expected HasMin true, got %v", n.config.HasMin)
		}
		if n.config.Max != 100 {
			t.Errorf("expected Max 100, got %f", n.config.Max)
		}
		if !n.config.HasMax {
			t.Errorf("expected HasMax true, got %v", n.config.HasMax)
		}
	})
}

func TestNumberInputBuilder_Step(t *testing.T) {
	t.Run("sets Step correctly", func(t *testing.T) {
		n := NumberInput().Step(0.5)
		if n.config.Step != 0.5 {
			t.Errorf("expected Step 0.5, got %f", n.config.Step)
		}
	})

	t.Run("ignores non-positive Step and uses default 1", func(t *testing.T) {
		n := NumberInput().Step(-5)
		if n.config.Step != 1 {
			t.Errorf("expected Step 1, got %f", n.config.Step)
		}
	})

	t.Run("ignores zero Step and uses default 1", func(t *testing.T) {
		n := NumberInput().Step(0)
		if n.config.Step != 1 {
			t.Errorf("expected Step 1, got %f", n.config.Step)
		}
	})
}

func TestNumberInputBuilder_Decimals(t *testing.T) {
	t.Run("sets Decimals correctly", func(t *testing.T) {
		n := NumberInput().Decimals(2)
		if n.config.Decimals != 2 {
			t.Errorf("expected Decimals 2, got %d", n.config.Decimals)
		}
	})

	t.Run("ignores negative Decimals and uses 0", func(t *testing.T) {
		n := NumberInput().Decimals(-1)
		if n.config.Decimals != 0 {
			t.Errorf("expected Decimals 0, got %d", n.config.Decimals)
		}
	})
}

func TestNumberInputBuilder_ArrowStep(t *testing.T) {
	t.Run("sets ArrowStep true", func(t *testing.T) {
		n := NumberInput().ArrowStep(true)
		if !n.config.ArrowStep {
			t.Errorf("expected ArrowStep true, got %v", n.config.ArrowStep)
		}
	})

	t.Run("sets ArrowStep false", func(t *testing.T) {
		n := NumberInput().ArrowStep(false)
		if n.config.ArrowStep {
			t.Errorf("expected ArrowStep false, got %v", n.config.ArrowStep)
		}
	})
}

func TestNumberInputBuilder_SelectAllOnFocus(t *testing.T) {
	t.Run("sets SelectAllOnFocus true", func(t *testing.T) {
		n := NumberInput().SelectAllOnFocus(true)
		if !n.config.SelectAllOnFocus {
			t.Errorf("expected SelectAllOnFocus true, got %v", n.config.SelectAllOnFocus)
		}
	})

	t.Run("sets SelectAllOnFocus false", func(t *testing.T) {
		n := NumberInput().SelectAllOnFocus(false)
		if n.config.SelectAllOnFocus {
			t.Errorf("expected SelectAllOnFocus false, got %v", n.config.SelectAllOnFocus)
		}
	})
}

func TestNumberInputBuilder_State(t *testing.T) {
	t.Run("sets Focused state correctly", func(t *testing.T) {
		n := NumberInput().Focused(true)
		if !n.focused {
			t.Errorf("expected focused true, got %v", n.focused)
		}
	})

	t.Run("sets Focused state to false", func(t *testing.T) {
		n := NumberInput().Focused(false)
		if n.focused {
			t.Errorf("expected focused false, got %v", n.focused)
		}
	})
}

func TestNumberInputCallbacks(t *testing.T) {
	t.Run("OnChange callback is set and invoked correctly", func(t *testing.T) {
		var changed bool
		var capturedID string
		var capturedValue float64
		n := NumberInput().ID("test-num").OnChange(func(id string, value float64) {
			changed = true
			capturedID = id
			capturedValue = value
		})

		if n.config.OnChange == nil {
			t.Fatal("OnChange callback not set")
		}
		n.config.OnChange("test-num", 42.5)
		if !changed {
			t.Error("OnChange callback was not executed")
		}
		if capturedID != "test-num" {
			t.Errorf("expected id 'test-num', got %s", capturedID)
		}
		if capturedValue != 42.5 {
			t.Errorf("expected value 42.5, got %f", capturedValue)
		}
	})

	t.Run("OnKeyPress callback is set and invoked correctly", func(t *testing.T) {
		var keyPressed bool
		n := NumberInput().ID("test-num").OnKeyPress(func(id string, key retui.Key) bool {
			keyPressed = true
			if id != "test-num" {
				t.Errorf("expected id 'test-num', got %s", id)
			}
			return true
		})

		if n.config.OnKeyPress == nil {
			t.Fatal("OnKeyPress callback not set")
		}
		if result := n.config.OnKeyPress("test-num", retui.Key{Code: retui.KeyEnter}); !result {
			t.Error("OnKeyPress should return true")
		}
		if !keyPressed {
			t.Error("OnKeyPress callback was not executed")
		}
	})

	t.Run("OnFocus callback is set and invoked correctly", func(t *testing.T) {
		var focused bool
		n := NumberInput().ID("test-num").OnFocus(func(id string) {
			focused = true
			if id != "test-num" {
				t.Errorf("expected id 'test-num', got %s", id)
			}
		})

		if n.config.OnFocus == nil {
			t.Fatal("OnFocus callback not set")
		}
		n.config.OnFocus("test-num")
		if !focused {
			t.Error("OnFocus callback was not executed")
		}
	})

	t.Run("OnBlur callback is set and invoked correctly", func(t *testing.T) {
		var blurred bool
		n := NumberInput().ID("test-num").OnBlur(func(id string) {
			blurred = true
			if id != "test-num" {
				t.Errorf("expected id 'test-num', got %s", id)
			}
		})

		if n.config.OnBlur == nil {
			t.Fatal("OnBlur callback not set")
		}
		n.config.OnBlur("test-num")
		if !blurred {
			t.Error("OnBlur callback was not executed")
		}
	})

	t.Run("OnSubmit callback is set and invoked correctly", func(t *testing.T) {
		var submitted bool
		var capturedID string
		var capturedValue float64
		n := NumberInput().ID("test-num").OnSubmit(func(id string, value float64) {
			submitted = true
			capturedID = id
			capturedValue = value
		})

		if n.config.OnSubmit == nil {
			t.Fatal("OnSubmit callback not set")
		}
		n.config.OnSubmit("test-num", 100.0)
		if !submitted {
			t.Error("OnSubmit callback was not executed")
		}
		if capturedID != "test-num" {
			t.Errorf("expected id 'test-num', got %s", capturedID)
		}
		if capturedValue != 100.0 {
			t.Errorf("expected value 100.0, got %f", capturedValue)
		}
	})
}

func TestNumberInputRender(t *testing.T) {
	t.Run("returns an Element of type ElementBox", func(t *testing.T) {
		n := NumberInput()
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with placeholder when empty", func(t *testing.T) {
		n := NumberInput().Placeholder("Enter number")
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with value when not empty", func(t *testing.T) {
		n := NumberInput().Value(42.5)
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with prefix", func(t *testing.T) {
		n := NumberInput().Prefix("$ ").Value(42.5)
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with suffix", func(t *testing.T) {
		n := NumberInput().Suffix(" USD").Value(42.5)
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with width", func(t *testing.T) {
		n := NumberInput().Value(42.5).Width(20)
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestRenderNumberInputFunction(t *testing.T) {
	t.Run("renders with default style when not focused", func(t *testing.T) {
		config := &NumberInputConfig{
			ID:    "test",
			Value: 42.5,
			Width: 20,
		}
		elem := renderNumberInput(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with focus style when focused", func(t *testing.T) {
		config := &NumberInputConfig{
			ID:    "test",
			Value: 42.5,
			Width: 20,
		}
		elem := renderNumberInput(true, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with placeholder when empty", func(t *testing.T) {
		config := &NumberInputConfig{
			ID:          "test",
			Empty:       true,
			Placeholder: "Enter number",
			Width:       20,
		}
		elem := renderNumberInput(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("triggers OnFocus when focused", func(t *testing.T) {
		var focusCalled bool
		config := &NumberInputConfig{
			ID:    "test",
			Value: 42.5,
			OnFocus: func(id string) {
				focusCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderNumberInput(true, config)
		if !focusCalled {
			t.Error("OnFocus was not called when focused")
		}
	})

	t.Run("triggers OnBlur when not focused", func(t *testing.T) {
		var blurCalled bool
		config := &NumberInputConfig{
			ID:    "test",
			Value: 42.5,
			OnBlur: func(id string) {
				blurCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderNumberInput(false, config)
		if !blurCalled {
			t.Error("OnBlur was not called when not focused")
		}
	})

	t.Run("does not trigger focus/blur without ID", func(t *testing.T) {
		var focusCalled, blurCalled bool
		config := &NumberInputConfig{
			ID:      "",
			Value:   42.5,
			OnFocus: func(id string) { focusCalled = true },
			OnBlur:  func(id string) { blurCalled = true },
		}

		renderNumberInput(true, config)
		if focusCalled {
			t.Error("OnFocus should not be called without ID")
		}

		renderNumberInput(false, config)
		if blurCalled {
			t.Error("OnBlur should not be called without ID")
		}
	})
}

func TestNumberInputChaining(t *testing.T) {
	n := NumberInput().
		ID("chain-test").
		Value(42.5).
		Placeholder("Enter number").
		Width(25).
		Prefix("$ ").
		Suffix(" USD").
		Min(0).
		Max(100).
		Step(0.5).
		Decimals(2).
		ArrowStep(true).
		SelectAllOnFocus(true).
		Focused(true)

	if n.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", n.config.ID)
	}
	if n.config.Value != 42.5 {
		t.Errorf("expected Value 42.5, got %f", n.config.Value)
	}
	if n.config.Placeholder != "Enter number" {
		t.Errorf("expected Placeholder 'Enter number', got %s", n.config.Placeholder)
	}
	if n.config.Width != 25 {
		t.Errorf("expected Width 25, got %d", n.config.Width)
	}
	if n.config.Prefix != "$ " {
		t.Errorf("expected Prefix '$ ', got %s", n.config.Prefix)
	}
	if n.config.Suffix != " USD" {
		t.Errorf("expected Suffix ' USD', got %s", n.config.Suffix)
	}
	if n.config.Min != 0 {
		t.Errorf("expected Min 0, got %f", n.config.Min)
	}
	if !n.config.HasMin {
		t.Errorf("expected HasMin true, got %v", n.config.HasMin)
	}
	if n.config.Max != 100 {
		t.Errorf("expected Max 100, got %f", n.config.Max)
	}
	if !n.config.HasMax {
		t.Errorf("expected HasMax true, got %v", n.config.HasMax)
	}
	if n.config.Step != 0.5 {
		t.Errorf("expected Step 0.5, got %f", n.config.Step)
	}
	if n.config.Decimals != 2 {
		t.Errorf("expected Decimals 2, got %d", n.config.Decimals)
	}
	if !n.config.ArrowStep {
		t.Errorf("expected ArrowStep true, got %v", n.config.ArrowStep)
	}
	if !n.config.SelectAllOnFocus {
		t.Errorf("expected SelectAllOnFocus true, got %v", n.config.SelectAllOnFocus)
	}
	if !n.focused {
		t.Errorf("expected focused true, got %v", n.focused)
	}
}

func TestNumberInputHelperFunctions(t *testing.T) {
	t.Run("formatNumber formats numbers correctly", func(t *testing.T) {
		tests := []struct {
			value    float64
			decimals int
			want     string
		}{
			{42.5, 1, "42.5"},
			{42.5, 2, "42.50"},
			{42, 0, "42"},
			{0.0, 0, "0"},
			{-0.0001, 3, "-0.000"},
			{0.000, 2, "0.00"},
		}
		for _, tt := range tests {
			if got := formatNumber(tt.value, tt.decimals); got != tt.want {
				t.Errorf("formatNumber(%f, %d) = %q, want %q", tt.value, tt.decimals, got, tt.want)
			}
		}
	})

	t.Run("roundToDecimals rounds correctly", func(t *testing.T) {
		tests := []struct {
			value    float64
			decimals int
			want     float64
		}{
			{1.2345, 2, 1.23},
			{1.235, 2, 1.24},
			{1.2, 0, 1},
			{1.5, 0, 2},
			{1.2345, 3, 1.235},
			{1.999, 2, 2.00},
		}
		for _, tt := range tests {
			if got := roundToDecimals(tt.value, tt.decimals); got != tt.want {
				t.Errorf("roundToDecimals(%f, %d) = %f, want %f", tt.value, tt.decimals, got, tt.want)
			}
		}
	})

	t.Run("parseNumber parses numbers correctly", func(t *testing.T) {
		tests := []struct {
			s       string
			wantVal float64
			wantOk  bool
		}{
			{"42.5", 42.5, true},
			{"-42.5", -42.5, true},
			{"0", 0, true},
			{"", 0, false},
			{"-", 0, false},
			{".", 0, false},
			{"-.", 0, false},
			{"abc", 0, false},
			{"1e10", 1e10, true},
			{"-1e10", -1e10, true},
		}
		for _, tt := range tests {
			val, ok := parseNumber(tt.s)
			if val != tt.wantVal || ok != tt.wantOk {
				t.Errorf("parseNumber(%q) = (%f, %v), want (%f, %v)", tt.s, val, ok, tt.wantVal, tt.wantOk)
			}
		}
	})

	t.Run("clampValue clamps correctly", func(t *testing.T) {
		tests := []struct {
			value  float64
			hasMin bool
			min    float64
			hasMax bool
			max    float64
			want   float64
		}{
			{5, true, 0, false, 0, 5},
			{-5, true, 0, false, 0, 0},
			{5, false, 0, true, 10, 5},
			{15, false, 0, true, 10, 10},
			{5, true, 0, true, 10, 5},
			{-5, true, 0, true, 10, 0},
			{15, true, 0, true, 10, 10},
			{0, true, 0, true, 0, 0}, // Bound of exactly 0
			{5, true, 5, true, 5, 5}, // Min = Max = 5
		}
		for _, tt := range tests {
			if got := clampValue(tt.value, tt.hasMin, tt.min, tt.hasMax, tt.max); got != tt.want {
				t.Errorf("clampValue(%f, %v, %f, %v, %f) = %f, want %f",
					tt.value, tt.hasMin, tt.min, tt.hasMax, tt.max, got, tt.want)
			}
		}
	})

	t.Run("isValidNumber validates numbers correctly", func(t *testing.T) {
		tests := []struct {
			s        string
			decimals int
			want     bool
		}{
			{"42.5", 2, true},
			{"42.5", 1, true},
			{"42.5", 0, false},
			{"42", 0, true},
			{"42.", 2, true},
			{"42.", 0, true}, // "42." parses as 42, valid partial input
			{"42.", 1, true},
			{"-42.5", 2, true},
			{"", 2, true},
			{"-", 2, true},
			{".", 2, true},
			{"-.", 2, true},
			{"abc", 2, false},
			{"12.34.56", 2, false}, // Multiple decimal points
			{"-", 0, true},
			{"42.123", 2, false}, // Too many decimal places
			{"42.12", 2, true},
		}
		for _, tt := range tests {
			if got := isValidNumber(tt.s, tt.decimals); got != tt.want {
				t.Errorf("isValidNumber(%q, %d) = %v, want %v", tt.s, tt.decimals, got, tt.want)
			}
		}
	})

	t.Run("applyParsedValue applies parsed value correctly", func(t *testing.T) {
		tests := []struct {
			config    *NumberInputConfig
			text      string
			wantVal   float64
			wantEmpty bool
		}{
			{
				&NumberInputConfig{ID: "test", Value: 0, Empty: true},
				"42.5",
				42.5,
				false,
			},
			{
				&NumberInputConfig{ID: "test", Value: 0, Empty: true},
				"",
				0,
				true,
			},
			{
				&NumberInputConfig{ID: "test", Value: 10, Empty: false},
				"",
				0,
				true,
			},
			{
				&NumberInputConfig{ID: "test", Value: 10, Empty: false, HasMin: true, Min: 0, HasMax: true, Max: 100},
				"150",
				100,
				false,
			},
			{
				&NumberInputConfig{ID: "test", Value: 10, Empty: false, HasMin: true, Min: 0, HasMax: true, Max: 100},
				"-5",
				0,
				false,
			},
			{
				&NumberInputConfig{ID: "test", Value: 0, Empty: true},
				"-",
				0,
				true, // Partial input doesn't commit
			},
			{
				&NumberInputConfig{ID: "test", Value: 0, Empty: true},
				".",
				0,
				true, // Partial input doesn't commit
			},
		}
		for _, tt := range tests {
			applyParsedValue(tt.config, tt.text)
			if tt.config.Value != tt.wantVal {
				t.Errorf("applyParsedValue() Value = %f, want %f", tt.config.Value, tt.wantVal)
			}
			if tt.config.Empty != tt.wantEmpty {
				t.Errorf("applyParsedValue() Empty = %v, want %v", tt.config.Empty, tt.wantEmpty)
			}
		}
	})
}
func TestNumberInputEdgeCases(t *testing.T) {
	t.Run("handles negative width gracefully", func(t *testing.T) {
		n := NumberInput().Value(42.5).Width(-5)
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very large numbers", func(t *testing.T) {
		n := NumberInput().Value(1e308)
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very small numbers", func(t *testing.T) {
		n := NumberInput().Value(1e-308)
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nil callbacks gracefully", func(t *testing.T) {
		n := NumberInput().ID("test").Value(42.5).
			OnChange(nil).OnKeyPress(nil).OnFocus(nil).OnBlur(nil).OnSubmit(nil)
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Error("Render returned wrong Element type with nil callbacks")
		}
	})

	t.Run("handles empty prefix and suffix", func(t *testing.T) {
		n := NumberInput().Value(42.5).Prefix("").Suffix("")
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("validates bounds correctly", func(t *testing.T) {
		n := NumberInput().Value(150).Min(0).Max(100)
		elem := n.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

// Benchmark tests
func BenchmarkNumberInputRender(b *testing.B) {
	n := NumberInput().Value(42.5).Width(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n.Render()
	}
}

func BenchmarkRenderNumberInput(b *testing.B) {
	config := &NumberInputConfig{
		ID:    "bench",
		Value: 42.5,
		Width: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderNumberInput(false, config)
	}
}

// Example usage test
func ExampleNumberInput() {
	// Integer input
	ageInput := NumberInput().
		ID("age").
		Placeholder("Enter age").
		Width(20).
		Min(0).
		Max(150).
		Step(1).
		Decimals(0).
		Prefix("👤 ").
		Suffix(" years").
		OnChange(func(id string, value float64) {
			// Handle change
		}).
		OnSubmit(func(id string, value float64) {
			// Handle submit
		}).
		Render()
	_ = ageInput

	// Decimal input
	priceInput := NumberInput().
		ID("price").
		Placeholder("0.00").
		Width(25).
		Min(0).
		Max(9999.99).
		Step(0.01).
		Decimals(2).
		Prefix("$ ").
		Suffix(" USD").
		Render()
	_ = priceInput
}
