package components

import (
	"reflect"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestDateInputBuilder_Defaults(t *testing.T) {
	d := DateInput()

	if d.config.ID != "" {
		t.Errorf("expected empty ID, got %q", d.config.ID)
	}
	if d.config.Value != "" {
		t.Errorf("expected empty Value, got %q", d.config.Value)
	}
	if d.config.Format != "YYYY-MM-DD" {
		t.Errorf("expected Format 'YYYY-MM-DD', got %q", d.config.Format)
	}
	if d.config.Placeholder != "" {
		t.Errorf("expected empty Placeholder, got %q", d.config.Placeholder)
	}
	if d.config.Width != 20 {
		t.Errorf("expected Width 20, got %d", d.config.Width)
	}
	if d.config.Prefix != "" {
		t.Errorf("expected empty Prefix, got %q", d.config.Prefix)
	}
	if d.config.Suffix != "" {
		t.Errorf("expected empty Suffix, got %q", d.config.Suffix)
	}
	if d.config.Min != "" {
		t.Errorf("expected empty Min, got %q", d.config.Min)
	}
	if d.config.Max != "" {
		t.Errorf("expected empty Max, got %q", d.config.Max)
	}
	if d.focused {
		t.Errorf("expected focused false, got %v", d.focused)
	}
}

func TestDateInputBuilder_StringFields(t *testing.T) {
	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{"ID", func() string { return DateInput().ID("test-id").config.ID }, "test-id"},
		{"Value", func() string { return DateInput().Value("2024-01-15").config.Value }, "2024-01-15"},
		{"Format", func() string { return DateInput().Format("MM/DD/YYYY").config.Format }, "MM/DD/YYYY"},
		{"Placeholder", func() string { return DateInput().Placeholder("Enter date").config.Placeholder }, "Enter date"},
		{"Prefix", func() string { return DateInput().Prefix("📅 ").config.Prefix }, "📅 "},
		{"Suffix", func() string { return DateInput().Suffix(" ✓").config.Suffix }, " ✓"},
		{"Min", func() string { return DateInput().Min("1900-01-01").config.Min }, "1900-01-01"},
		{"Max", func() string { return DateInput().Max("2026-12-31").config.Max }, "2026-12-31"},
	}

	for _, tt := range tests {
		t.Run("sets "+tt.name+" correctly", func(t *testing.T) {
			if got := tt.got(); got != tt.want {
				t.Errorf("expected %s %q, got %q", tt.name, tt.want, got)
			}
		})
	}
}

func TestDateInputBuilder_Width(t *testing.T) {
	d := DateInput().Width(30)
	if d.config.Width != 30 {
		t.Errorf("expected Width 30, got %d", d.config.Width)
	}
}

func TestDateInputBuilder_Style(t *testing.T) {
	t.Run("sets Style correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Background(retui.Blue)
		d := DateInput().Style(style)
		if !reflect.DeepEqual(d.config.Style, style) {
			t.Errorf("Style not set correctly")
		}
	})
}

func TestDateInputBuilder_State(t *testing.T) {
	t.Run("sets Focused state correctly", func(t *testing.T) {
		d := DateInput().Focused(true)
		if !d.focused {
			t.Errorf("expected focused true, got %v", d.focused)
		}
	})

	t.Run("sets Focused state to false", func(t *testing.T) {
		d := DateInput().Focused(false)
		if d.focused {
			t.Errorf("expected focused false, got %v", d.focused)
		}
	})
}

func TestDateInputCallbacks(t *testing.T) {
	t.Run("OnChange callback is set and invoked correctly", func(t *testing.T) {
		var changed bool
		var capturedID string
		var capturedValue string
		d := DateInput().ID("test-date").OnChange(func(id string, value string) {
			changed = true
			capturedID = id
			capturedValue = value
		})

		if d.config.OnChange == nil {
			t.Fatal("OnChange callback not set")
		}
		d.config.OnChange("test-date", "2024-01-15")
		if !changed {
			t.Error("OnChange callback was not executed")
		}
		if capturedID != "test-date" {
			t.Errorf("expected id 'test-date', got %s", capturedID)
		}
		if capturedValue != "2024-01-15" {
			t.Errorf("expected value '2024-01-15', got %s", capturedValue)
		}
	})

	t.Run("OnKeyPress callback is set and invoked correctly", func(t *testing.T) {
		var keyPressed bool
		d := DateInput().ID("test-date").OnKeyPress(func(id string, key retui.Key) bool {
			keyPressed = true
			if id != "test-date" {
				t.Errorf("expected id 'test-date', got %s", id)
			}
			return true
		})

		if d.config.OnKeyPress == nil {
			t.Fatal("OnKeyPress callback not set")
		}
		if result := d.config.OnKeyPress("test-date", retui.Key{Code: retui.KeyEnter}); !result {
			t.Error("OnKeyPress should return true")
		}
		if !keyPressed {
			t.Error("OnKeyPress callback was not executed")
		}
	})

	t.Run("OnFocus callback is set and invoked correctly", func(t *testing.T) {
		var focused bool
		d := DateInput().ID("test-date").OnFocus(func(id string) {
			focused = true
			if id != "test-date" {
				t.Errorf("expected id 'test-date', got %s", id)
			}
		})

		if d.config.OnFocus == nil {
			t.Fatal("OnFocus callback not set")
		}
		d.config.OnFocus("test-date")
		if !focused {
			t.Error("OnFocus callback was not executed")
		}
	})

	t.Run("OnBlur callback is set and invoked correctly", func(t *testing.T) {
		var blurred bool
		d := DateInput().ID("test-date").OnBlur(func(id string) {
			blurred = true
			if id != "test-date" {
				t.Errorf("expected id 'test-date', got %s", id)
			}
		})

		if d.config.OnBlur == nil {
			t.Fatal("OnBlur callback not set")
		}
		d.config.OnBlur("test-date")
		if !blurred {
			t.Error("OnBlur callback was not executed")
		}
	})

	t.Run("OnSubmit callback is set and invoked correctly", func(t *testing.T) {
		var submitted bool
		var capturedID string
		var capturedValue string
		d := DateInput().ID("test-date").OnSubmit(func(id string, value string) {
			submitted = true
			capturedID = id
			capturedValue = value
		})

		if d.config.OnSubmit == nil {
			t.Fatal("OnSubmit callback not set")
		}
		d.config.OnSubmit("test-date", "2024-01-15")
		if !submitted {
			t.Error("OnSubmit callback was not executed")
		}
		if capturedID != "test-date" {
			t.Errorf("expected id 'test-date', got %s", capturedID)
		}
		if capturedValue != "2024-01-15" {
			t.Errorf("expected value '2024-01-15', got %s", capturedValue)
		}
	})
}

func TestDateInputRender(t *testing.T) {
	t.Run("returns an Element of type ElementBox", func(t *testing.T) {
		elem := DateInput().Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with default format", func(t *testing.T) {
		elem := DateInput().Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with placeholder", func(t *testing.T) {
		elem := DateInput().Placeholder("Enter date").Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with prefix", func(t *testing.T) {
		elem := DateInput().Prefix("📅 ").Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with suffix", func(t *testing.T) {
		elem := DateInput().Suffix(" ✓").Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with value", func(t *testing.T) {
		elem := DateInput().Value("2024-01-15").Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestRenderDateInputFunction(t *testing.T) {
	t.Run("renders with default style when not focused", func(t *testing.T) {
		config := &DateConfig{
			ID:     "test",
			Format: "YYYY-MM-DD",
			Width:  20,
		}
		elem := renderDateInput(false, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders with focus style when focused", func(t *testing.T) {
		config := &DateConfig{
			ID:     "test",
			Format: "YYYY-MM-DD",
			Width:  20,
		}
		elem := renderDateInput(true, config)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("triggers OnFocus when focused", func(t *testing.T) {
		var focusCalled bool
		config := &DateConfig{
			ID: "test",
			OnFocus: func(id string) {
				focusCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderDateInput(true, config)
		if !focusCalled {
			t.Error("OnFocus was not called when focused")
		}
	})

	t.Run("triggers OnBlur when not focused", func(t *testing.T) {
		var blurCalled bool
		config := &DateConfig{
			ID: "test",
			OnBlur: func(id string) {
				blurCalled = true
				if id != "test" {
					t.Errorf("expected id 'test', got %s", id)
				}
			},
		}
		renderDateInput(false, config)
		if !blurCalled {
			t.Error("OnBlur was not called when not focused")
		}
	})

	t.Run("does not trigger focus/blur without ID", func(t *testing.T) {
		var focusCalled, blurCalled bool
		config := &DateConfig{
			ID:      "",
			OnFocus: func(id string) { focusCalled = true },
			OnBlur:  func(id string) { blurCalled = true },
		}

		renderDateInput(true, config)
		if focusCalled {
			t.Error("OnFocus should not be called without ID")
		}

		renderDateInput(false, config)
		if blurCalled {
			t.Error("OnBlur should not be called without ID")
		}
	})
}

func TestDateInputChaining(t *testing.T) {
	d := DateInput().
		ID("chain-test").
		Value("2024-01-15").
		Format("MM/DD/YYYY").
		Placeholder("Enter date").
		Width(25).
		Prefix("📅 ").
		Suffix(" ✓").
		Min("1900-01-01").
		Max("2026-12-31").
		Focused(true)

	if d.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", d.config.ID)
	}
	if d.config.Value != "2024-01-15" {
		t.Errorf("expected Value '2024-01-15', got %s", d.config.Value)
	}
	if d.config.Format != "MM/DD/YYYY" {
		t.Errorf("expected Format 'MM/DD/YYYY', got %s", d.config.Format)
	}
	if d.config.Placeholder != "Enter date" {
		t.Errorf("expected Placeholder 'Enter date', got %s", d.config.Placeholder)
	}
	if d.config.Width != 25 {
		t.Errorf("expected Width 25, got %d", d.config.Width)
	}
	if d.config.Prefix != "📅 " {
		t.Errorf("expected Prefix '📅 ', got %s", d.config.Prefix)
	}
	if d.config.Suffix != " ✓" {
		t.Errorf("expected Suffix ' ✓', got %s", d.config.Suffix)
	}
	if d.config.Min != "1900-01-01" {
		t.Errorf("expected Min '1900-01-01', got %s", d.config.Min)
	}
	if d.config.Max != "2026-12-31" {
		t.Errorf("expected Max '2026-12-31', got %s", d.config.Max)
	}
	if !d.focused {
		t.Errorf("expected focused true, got %v", d.focused)
	}
}

func TestDateInputEdgeCases(t *testing.T) {
	t.Run("handles empty format gracefully", func(t *testing.T) {
		d := DateInput().Format("")
		elem := d.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles negative width gracefully", func(t *testing.T) {
		d := DateInput().Width(-5)
		elem := d.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very long placeholder", func(t *testing.T) {
		longPlaceholder := "This is a very long placeholder text that should still render properly"
		d := DateInput().Placeholder(longPlaceholder)
		elem := d.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nil callbacks gracefully", func(t *testing.T) {
		d := DateInput().ID("test").OnChange(nil).OnKeyPress(nil).OnFocus(nil).OnBlur(nil).OnSubmit(nil)
		elem := d.Render()
		if elem.Type != retui.ElementBox {
			t.Error("Render returned wrong Element type with nil callbacks")
		}
	})

	t.Run("handles different date formats", func(t *testing.T) {
		formats := []string{
			"YYYY-MM-DD",
			"MM/DD/YYYY",
			"DD/MM/YYYY",
			"YYYY.MM.DD",
		}
		for _, format := range formats {
			t.Run(format, func(t *testing.T) {
				d := DateInput().Format(format)
				elem := d.Render()
				if elem.Type != retui.ElementBox {
					t.Errorf("expected Element of type ElementBox for format %s, got %v", format, elem.Type)
				}
			})
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("extractDigitsOnly extracts only digits", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{"2024-01-15", "20240115"},
			{"MM/DD/YYYY", ""},
			{"12/31/2024", "12312024"},
			{"", ""},
			{"abc123def456", "123456"},
		}
		for _, tt := range tests {
			if got := extractDigitsOnly(tt.input); got != tt.want {
				t.Errorf("extractDigitsOnly(%q) = %q, want %q", tt.input, got, tt.want)
			}
		}
	})

	t.Run("countDigitsInMask counts digit placeholders", func(t *testing.T) {
		tests := []struct {
			mask string
			want int
		}{
			{"YYYY-MM-DD", 8},
			{"MM/DD/YYYY", 8},
			{"DD/MM/YYYY", 8},
			{"", 0},
			{"YY-MM-DD", 6},
		}
		for _, tt := range tests {
			if got := countDigitsInMask(tt.mask); got != tt.want {
				t.Errorf("countDigitsInMask(%q) = %d, want %d", tt.mask, got, tt.want)
			}
		}
	})

	t.Run("applyMaskFormat applies digits to mask", func(t *testing.T) {
		tests := []struct {
			rawDigits string
			mask      string
			want      string
		}{
			{"2024", "YYYY-MM-DD", "2024-MM-DD"},
			{"20240115", "YYYY-MM-DD", "2024-01-15"},
			{"1231", "MM/DD/YYYY", "12/31/YYYY"},
			{"12312024", "MM/DD/YYYY", "12/31/2024"},
			{"", "YYYY-MM-DD", ""},
			{"15", "DD/MM/YYYY", "15/MM/YYYY"},
			{"15012024", "DD/MM/YYYY", "15/01/2024"},
		}
		for _, tt := range tests {
			if got := applyMaskFormat(tt.rawDigits, tt.mask); got != tt.want {
				t.Errorf("applyMaskFormat(%q, %q) = %q, want %q", tt.rawDigits, tt.mask, got, tt.want)
			}
		}
	})

	t.Run("isValidCalendarDate validates dates correctly", func(t *testing.T) {
		tests := []struct {
			year, month, day int
			want             bool
		}{
			{2024, 1, 15, true},
			{2024, 2, 29, true},  // Leap year
			{2023, 2, 29, false}, // Not leap year
			{2024, 4, 31, false}, // April has 30 days
			{2024, 13, 1, false}, // Invalid month
			{2024, 1, 32, false}, // Invalid day
		}
		for _, tt := range tests {
			if got := isValidCalendarDate(tt.year, tt.month, tt.day); got != tt.want {
				t.Errorf("isValidCalendarDate(%d, %d, %d) = %v, want %v", tt.year, tt.month, tt.day, got, tt.want)
			}
		}
	})

	t.Run("dateSortKey returns correct sort key", func(t *testing.T) {
		tests := []struct {
			year, month, day int
			want             int
		}{
			{2024, 1, 15, 20240115},
			{2023, 12, 31, 20231231},
			{1900, 1, 1, 19000101},
		}
		for _, tt := range tests {
			if got := dateSortKey(tt.year, tt.month, tt.day); got != tt.want {
				t.Errorf("dateSortKey(%d, %d, %d) = %d, want %d", tt.year, tt.month, tt.day, got, tt.want)
			}
		}
	})

	t.Run("maskFieldSpecs identifies field positions", func(t *testing.T) {
		y, m, d := maskFieldSpecs("YYYY-MM-DD")
		if y.start != 0 || y.length != 4 {
			t.Errorf("expected year at 0,4 got %d,%d", y.start, y.length)
		}
		if m.start != 4 || m.length != 2 {
			t.Errorf("expected month at 4,2 got %d,%d", m.start, m.length)
		}
		if d.start != 6 || d.length != 2 {
			t.Errorf("expected day at 6,2 got %d,%d", d.start, d.length)
		}

		y, m, d = maskFieldSpecs("MM/DD/YYYY")
		if m.start != 0 || m.length != 2 {
			t.Errorf("expected month at 0,2 got %d,%d", m.start, m.length)
		}
		if d.start != 2 || d.length != 2 {
			t.Errorf("expected day at 2,2 got %d,%d", d.start, d.length)
		}
		if y.start != 4 || y.length != 4 {
			t.Errorf("expected year at 4,4 got %d,%d", y.start, y.length)
		}
	})

	t.Run("extractDateParts extracts date parts correctly", func(t *testing.T) {
		tests := []struct {
			rawDigits    string
			mask         string
			wantYear     int
			wantMonth    int
			wantDay      int
			wantComplete bool
		}{
			{"20240115", "YYYY-MM-DD", 2024, 1, 15, true},
			{"2024", "YYYY-MM-DD", 2024, 0, 0, false},
			{"12312024", "MM/DD/YYYY", 2024, 12, 31, true},
			{"12", "MM/DD/YYYY", 0, 12, 0, false},
			{"15012024", "DD/MM/YYYY", 2024, 1, 15, true},
			{"15", "DD/MM/YYYY", 0, 0, 15, false},
		}
		for _, tt := range tests {
			year, month, day, complete := extractDateParts(tt.rawDigits, tt.mask)
			if year != tt.wantYear || month != tt.wantMonth || day != tt.wantDay || complete != tt.wantComplete {
				t.Errorf("extractDateParts(%q, %q) = (%d, %d, %d, %v), want (%d, %d, %d, %v)",
					tt.rawDigits, tt.mask, year, month, day, complete,
					tt.wantYear, tt.wantMonth, tt.wantDay, tt.wantComplete)
			}
		}
	})

	t.Run("parseFormattedDate parses formatted dates", func(t *testing.T) {
		tests := []struct {
			value     string
			mask      string
			wantYear  int
			wantMonth int
			wantDay   int
			wantOk    bool
		}{
			{"2024-01-15", "YYYY-MM-DD", 2024, 1, 15, true},
			{"12/31/2024", "MM/DD/YYYY", 2024, 12, 31, true},
			{"15/01/2024", "DD/MM/YYYY", 2024, 1, 15, true},
			{"2024-02-29", "YYYY-MM-DD", 2024, 2, 29, true}, // Leap year
			{"2023-02-29", "YYYY-MM-DD", 0, 0, 0, false},    // Invalid date
			{"2024-13-01", "YYYY-MM-DD", 0, 0, 0, false},    // Invalid month
			{"", "YYYY-MM-DD", 0, 0, 0, false},              // Empty
		}
		for _, tt := range tests {
			year, month, day, ok := parseFormattedDate(tt.value, tt.mask)
			if year != tt.wantYear || month != tt.wantMonth || day != tt.wantDay || ok != tt.wantOk {
				t.Errorf("parseFormattedDate(%q, %q) = (%d, %d, %d, %v), want (%d, %d, %d, %v)",
					tt.value, tt.mask, year, month, day, ok,
					tt.wantYear, tt.wantMonth, tt.wantDay, tt.wantOk)
			}
		}
	})

	t.Run("evaluateDate evaluates date validity", func(t *testing.T) {
		tests := []struct {
			config    *DateConfig
			rawDigits string
			mask      string
			wantOk    bool
			wantMsg   string
		}{
			{
				&DateConfig{},
				"20240115",
				"YYYY-MM-DD",
				true,
				"",
			},
			{
				&DateConfig{},
				"2024",
				"YYYY-MM-DD",
				false,
				"",
			},
			{
				&DateConfig{},
				"20240229",
				"YYYY-MM-DD",
				true,
				"",
			},
			{
				&DateConfig{},
				"20240230",
				"YYYY-MM-DD",
				false,
				"Enter a valid date",
			},
			{
				&DateConfig{Min: "2024-01-01"},
				"20231231",
				"YYYY-MM-DD",
				false,
				"Date is before the minimum allowed",
			},
			{
				&DateConfig{Max: "2024-12-31"},
				"20250101",
				"YYYY-MM-DD",
				false,
				"Date is after the maximum allowed",
			},
			{
				&DateConfig{Min: "2024-01-01", Max: "2024-12-31"},
				"20240615",
				"YYYY-MM-DD",
				true,
				"",
			},
		}
		for _, tt := range tests {
			ok, msg := evaluateDate(tt.config, tt.rawDigits, tt.mask)
			if ok != tt.wantOk || msg != tt.wantMsg {
				t.Errorf("evaluateDate(config, %q, %q) = (%v, %q), want (%v, %q)",
					tt.rawDigits, tt.mask, ok, msg, tt.wantOk, tt.wantMsg)
			}
		}
	})

	t.Run("calculateCursorDisplayPos calculates correct position", func(t *testing.T) {
		tests := []struct {
			mask           string
			digitCursorPos int
			want           int
		}{
			{"YYYY-MM-DD", 0, 0},
			{"YYYY-MM-DD", 4, 5}, // After YYYY and the dash
			{"YYYY-MM-DD", 6, 8}, // After YYYY-MM and the dash
			{"MM/DD/YYYY", 0, 0},
			{"MM/DD/YYYY", 2, 3}, // After MM and the slash
			{"MM/DD/YYYY", 4, 6}, // After MM/DD and the slash
		}
		for _, tt := range tests {
			if got := calculateCursorDisplayPos(tt.mask, tt.digitCursorPos); got != tt.want {
				t.Errorf("calculateCursorDisplayPos(%q, %d) = %d, want %d",
					tt.mask, tt.digitCursorPos, got, tt.want)
			}
		}
	})
}

// Benchmark tests
func BenchmarkDateInputRender(b *testing.B) {
	d := DateInput().Value("2024-01-15").Width(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Render()
	}
}

func BenchmarkRenderDateInput(b *testing.B) {
	config := &DateConfig{
		ID:     "bench",
		Value:  "2024-01-15",
		Format: "YYYY-MM-DD",
		Width:  20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderDateInput(false, config)
	}
}

// Example usage test
func ExampleDateInput() {
	// Create a date input with format and range constraint
	dateInput := DateInput().
		ID("birthday").
		Format("YYYY-MM-DD").
		Placeholder("Enter date").
		Width(25).
		Suffix(" ✓").
		Min("1900-01-01").
		Max("2026-12-31").
		OnChange(func(id string, value string) {
			// Handle change
		}).
		OnSubmit(func(id string, value string) {
			// Handle submit
		}).
		Render()
	_ = dateInput
}
