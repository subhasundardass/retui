package components

import (
	"reflect"
	"strings"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestProgressBar(t *testing.T) {
	t.Run("creates a progress bar with default colors", func(t *testing.T) {
		elem := ProgressBar(0.5, 10, retui.Green)

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}

		expectedStyle := retui.NewStyle().Foreground(retui.Green)
		if !reflect.DeepEqual(elem.Style, expectedStyle) {
			t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
		}
	})

	t.Run("renders full bar for value 1.0", func(t *testing.T) {
		width := 10
		elem := ProgressBar(1.0, width, retui.Green)

		expectedLen := width - 2
		if len([]rune(elem.Text)) != expectedLen {
			t.Errorf("expected text length %d, got %d", expectedLen, len([]rune(elem.Text)))
		}

		expected := strings.Repeat("█", expectedLen)
		if elem.Text != expected {
			t.Errorf("expected %q, got %q", expected, elem.Text)
		}
	})

	t.Run("renders empty bar for value 0.0", func(t *testing.T) {
		width := 10
		elem := ProgressBar(0.0, width, retui.Green)

		expectedLen := width - 2
		if len([]rune(elem.Text)) != expectedLen {
			t.Errorf("expected text length %d, got %d", expectedLen, len([]rune(elem.Text)))
		}

		expected := strings.Repeat("░", expectedLen)
		if elem.Text != expected {
			t.Errorf("expected %q, got %q", expected, elem.Text)
		}
	})

	t.Run("renders half-filled bar for value 0.5", func(t *testing.T) {
		width := 10
		elem := ProgressBar(0.5, width, retui.Green)

		expectedLen := width - 2
		if len([]rune(elem.Text)) != expectedLen {
			t.Errorf("expected text length %d, got %d", expectedLen, len([]rune(elem.Text)))
		}

		inner := width - 2
		filled := int(float64(inner) * 0.5) // floors to 4 for inner=8
		empty := inner - filled
		expected := strings.Repeat("█", filled) + strings.Repeat("░", empty)
		if elem.Text != expected {
			t.Errorf("expected %q, got %q", expected, elem.Text)
		}
	})

	t.Run("clamps value below 0 to 0", func(t *testing.T) {
		width := 10
		elem := ProgressBar(-0.5, width, retui.Green)

		expectedLen := width - 2
		expected := strings.Repeat("░", expectedLen)
		if elem.Text != expected {
			t.Errorf("expected %q, got %q", expected, elem.Text)
		}
	})

	t.Run("clamps value above 1 to 1", func(t *testing.T) {
		width := 10
		elem := ProgressBar(1.5, width, retui.Green)

		expectedLen := width - 2
		expected := strings.Repeat("█", expectedLen)
		if elem.Text != expected {
			t.Errorf("expected %q, got %q", expected, elem.Text)
		}
	})

	t.Run("handles minimum valid width of 2", func(t *testing.T) {
		width := 2
		elem := ProgressBar(0.5, width, retui.Green)

		// With width 2, inner = 0, so we get an empty string
		expectedLen := 0
		if len([]rune(elem.Text)) != expectedLen {
			t.Errorf("expected text length %d, got %d", expectedLen, len([]rune(elem.Text)))
		}
		if elem.Text != "" {
			t.Errorf("expected empty string, got %q", elem.Text)
		}
	})

	t.Run("handles different colors", func(t *testing.T) {
		colors := []struct {
			name  string
			color retui.Color
		}{
			{"Green", retui.Green},
			{"Red", retui.Red},
			{"Blue", retui.Blue},
			{"Cyan", retui.Cyan},
			{"Yellow", retui.Yellow},
			{"Magenta", retui.Magenta},
			{"White", retui.White},
			{"Black", retui.Black},
		}

		for _, tc := range colors {
			t.Run(tc.name, func(t *testing.T) {
				elem := ProgressBar(0.5, 10, tc.color)
				expectedStyle := retui.NewStyle().Foreground(tc.color)
				if !reflect.DeepEqual(elem.Style, expectedStyle) {
					t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
				}
			})
		}
	})

	t.Run("handles various values correctly", func(t *testing.T) {
		// For width 10, inner = 8
		// filled = int(8 * value) - floors the result
		tests := []struct {
			name     string
			value    float64
			width    int
			expected string
		}{
			{"0%", 0.0, 10, strings.Repeat("░", 8)},
			{"10%", 0.1, 10, strings.Repeat("░", 8)},  // int(8*0.1) = 0
			{"20%", 0.2, 10, "█░░░░░░░"},              // int(8*0.2) = 1
			{"30%", 0.3, 10, "██░░░░░░"},              // int(8*0.3) = 2
			{"40%", 0.4, 10, "███░░░░░"},              // int(8*0.4) = 3
			{"50%", 0.5, 10, "████░░░░"},              // int(8*0.5) = 4
			{"60%", 0.6, 10, "████░░░░"},              // int(8*0.6) = 4
			{"70%", 0.7, 10, "█████░░░"},              // int(8*0.7) = 5
			{"80%", 0.8, 10, "██████░░"},              // int(8*0.8) = 6
			{"90%", 0.9, 10, "███████░"},              // int(8*0.9) = 7
			{"100%", 1.0, 10, strings.Repeat("█", 8)}, // int(8*1.0) = 8
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				elem := ProgressBar(tt.value, tt.width, retui.Green)
				if elem.Text != tt.expected {
					t.Errorf("ProgressBar(%f, %d) = %q, want %q", tt.value, tt.width, elem.Text, tt.expected)
				}
			})
		}
	})

	t.Run("rounds correctly for fractional values", func(t *testing.T) {
		// For width 10, inner = 8
		// filled = int(8 * value) - floors the result
		tests := []struct {
			name     string
			value    float64
			width    int
			expected string
		}{
			{"33%", 0.33, 10, "██░░░░░░"}, // int(8*0.33) = 2
			{"34%", 0.34, 10, "██░░░░░░"}, // int(8*0.34) = 2
			{"35%", 0.35, 10, "██░░░░░░"}, // int(8*0.35) = 2
			{"66%", 0.66, 10, "█████░░░"}, // int(8*0.66) = 5
			{"67%", 0.67, 10, "█████░░░"}, // int(8*0.67) = 5
			{"68%", 0.68, 10, "█████░░░"}, // int(8*0.68) = 5
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				elem := ProgressBar(tt.value, tt.width, retui.Green)
				if elem.Text != tt.expected {
					t.Errorf("ProgressBar(%f, %d) = %q, want %q", tt.value, tt.width, elem.Text, tt.expected)
				}
			})
		}
	})

	t.Run("handles large width values", func(t *testing.T) {
		width := 100
		elem := ProgressBar(0.5, width, retui.Green)

		expectedLen := width - 2
		if len([]rune(elem.Text)) != expectedLen {
			t.Errorf("expected text length %d, got %d", expectedLen, len([]rune(elem.Text)))
		}

		inner := width - 2
		filled := int(float64(inner) * 0.5) // floors to 49 for inner=98
		empty := inner - filled
		expected := strings.Repeat("█", filled) + strings.Repeat("░", empty)
		if elem.Text != expected {
			t.Errorf("expected %q, got %q", expected, elem.Text)
		}
	})
}

func TestProgressBar_ElementType(t *testing.T) {
	t.Run("returns Element of type ElementText", func(t *testing.T) {
		elem := ProgressBar(0.5, 10, retui.Green)

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
	})

	t.Run("returns ElementText for all values", func(t *testing.T) {
		values := []struct {
			name  string
			value float64
		}{
			{"zero", 0.0},
			{"quarter", 0.25},
			{"half", 0.5},
			{"three_quarter", 0.75},
			{"full", 1.0},
		}
		for _, v := range values {
			t.Run(v.name, func(t *testing.T) {
				elem := ProgressBar(v.value, 10, retui.Green)
				if elem.Type != retui.ElementText {
					t.Errorf("for value %f, expected ElementText, got %v", v.value, elem.Type)
				}
			})
		}
	})
}

func TestProgressBar_EdgeCases(t *testing.T) {
	t.Run("handles zero value with custom color", func(t *testing.T) {
		elem := ProgressBar(0.0, 10, retui.Red)
		expectedStyle := retui.NewStyle().Foreground(retui.Red)
		if !reflect.DeepEqual(elem.Style, expectedStyle) {
			t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
		}
		// For width 10, inner = 8
		expected := strings.Repeat("░", 8)
		if elem.Text != expected {
			t.Errorf("expected empty bar %q, got %q", expected, elem.Text)
		}
	})

	t.Run("handles full value with custom color", func(t *testing.T) {
		elem := ProgressBar(1.0, 10, retui.Blue)
		expectedStyle := retui.NewStyle().Foreground(retui.Blue)
		if !reflect.DeepEqual(elem.Style, expectedStyle) {
			t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
		}
		// For width 10, inner = 8
		expected := strings.Repeat("█", 8)
		if elem.Text != expected {
			t.Errorf("expected full bar %q, got %q", expected, elem.Text)
		}
	})

	t.Run("handles very large width", func(t *testing.T) {
		width := 100000
		elem := ProgressBar(0.5, width, retui.Green)
		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		expectedLen := width - 2
		if len([]rune(elem.Text)) != expectedLen {
			t.Errorf("expected text length %d, got %d", expectedLen, len([]rune(elem.Text)))
		}
	})
}

func TestProgressBar_InvalidWidth(t *testing.T) {
	t.Run("handles width 1 gracefully", func(t *testing.T) {
		// Width 1 causes negative inner value
		// The function will panic, so we recover
		defer func() {
			if r := recover(); r != nil {
				// Expected behavior for width 1
				t.Log("ProgressBar panicked with width 1 as expected")
			}
		}()
		_ = ProgressBar(0.5, 1, retui.Green)
		t.Error("Expected panic but didn't get one")
	})

	t.Run("handles width 0 gracefully", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Log("ProgressBar panicked with width 0 as expected")
			}
		}()
		_ = ProgressBar(0.5, 0, retui.Green)
		t.Error("Expected panic but didn't get one")
	})

	t.Run("handles negative width gracefully", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Log("ProgressBar panicked with negative width as expected")
			}
		}()
		_ = ProgressBar(0.5, -5, retui.Green)
		t.Error("Expected panic but didn't get one")
	})
}

// Benchmark tests
func BenchmarkProgressBar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProgressBar(0.5, 50, retui.Green)
	}
}

func BenchmarkProgressBarLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProgressBar(0.75, 200, retui.Blue)
	}
}

func BenchmarkProgressBarVeryLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProgressBar(0.33, 1000, retui.Cyan)
	}
}

// Example usage test
func ExampleProgressBar() {
	// Create progress bars with different values
	progress1 := ProgressBar(0.25, 20, retui.Green)
	_ = progress1

	progress2 := ProgressBar(0.50, 20, retui.Cyan)
	_ = progress2

	progress3 := ProgressBar(0.75, 20, retui.Yellow)
	_ = progress3

	progress4 := ProgressBar(1.00, 20, retui.Green)
	_ = progress4

	// Output:
}
