package components

import (
	"reflect"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestBadge(t *testing.T) {
	t.Run("creates a badge with label and colors", func(t *testing.T) {
		elem := Badge("Test", retui.White, retui.Blue)

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}

		// Style has no exported getters, so compare the whole built Style
		// rather than reaching into individual color fields.
		expectedStyle := retui.NewStyle().Foreground(retui.White).Background(retui.Blue).Bold(true)
		if !reflect.DeepEqual(elem.Style, expectedStyle) {
			t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
		}
	})

	t.Run("renders with spaces around label", func(t *testing.T) {
		label := "Status"
		elem := Badge(label, retui.Green, retui.Black)

		if elem.Text != " "+label+" " {
			t.Errorf("expected text ' %s ', got %q", label, elem.Text)
		}
	})

	t.Run("handles empty label", func(t *testing.T) {
		elem := Badge("", retui.White, retui.Red)

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		if elem.Text != "  " {
			t.Errorf("expected text '  ', got %q", elem.Text)
		}
	})

	t.Run("handles single character label", func(t *testing.T) {
		elem := Badge("A", retui.Black, retui.Yellow)

		if elem.Text != " A " {
			t.Errorf("expected text ' A ', got %q", elem.Text)
		}
	})

	t.Run("handles long label", func(t *testing.T) {
		longLabel := "This is a very long badge label that should still render properly"
		elem := Badge(longLabel, retui.White, retui.Blue)

		if elem.Text != " "+longLabel+" " {
			t.Errorf("expected text with spaces, got %q", elem.Text)
		}
	})

	t.Run("handles special characters in label", func(t *testing.T) {
		specialLabel := "!@#$%^&*()"
		elem := Badge(specialLabel, retui.Cyan, retui.Navy)

		if elem.Text != " "+specialLabel+" " {
			t.Errorf("expected text with special characters, got %q", elem.Text)
		}
	})

	t.Run("handles Unicode characters", func(t *testing.T) {
		unicodeLabel := "✓ 完成"
		elem := Badge(unicodeLabel, retui.Green, retui.Black)

		if elem.Text != " "+unicodeLabel+" " {
			t.Errorf("expected text with Unicode, got %q", elem.Text)
		}
	})

	t.Run("handles different color combinations", func(t *testing.T) {
		tests := []struct {
			name string
			fg   retui.Color
			bg   retui.Color
		}{
			{"White on Navy", retui.White, retui.Navy},
			{"Black on Yellow", retui.Black, retui.Yellow},
			{"Green on Red", retui.Green, retui.Red},
			{"Cyan on Navy", retui.Cyan, retui.Navy},
			{"Magenta on White", retui.Magenta, retui.White},
			{"Yellow on Blue", retui.Yellow, retui.Blue},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				elem := Badge("Test", tt.fg, tt.bg)

				expectedStyle := retui.NewStyle().Foreground(tt.fg).Background(tt.bg).Bold(true)
				if !reflect.DeepEqual(elem.Style, expectedStyle) {
					t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
				}
			})
		}
	})

	t.Run("always returns bold style", func(t *testing.T) {
		elem := Badge("Test", retui.White, retui.Blue)

		expectedStyle := retui.NewStyle().Foreground(retui.White).Background(retui.Blue).Bold(true)
		if !reflect.DeepEqual(elem.Style, expectedStyle) {
			t.Errorf("expected bold style %+v, got %+v", expectedStyle, elem.Style)
		}
	})

	t.Run("creates badge with default colors", func(t *testing.T) {
		// White on Navy is a common default combination for badges.
		elem := Badge("Default", retui.White, retui.Navy)

		expectedStyle := retui.NewStyle().Foreground(retui.White).Background(retui.Navy).Bold(true)
		if !reflect.DeepEqual(elem.Style, expectedStyle) {
			t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
		}
	})

	t.Run("multiple badges can be created independently", func(t *testing.T) {
		badge1 := Badge("Success", retui.Green, retui.Black)
		badge2 := Badge("Error", retui.Red, retui.White)

		// Verify they have different styles overall, since individual
		// color fields aren't directly accessible for comparison.
		if reflect.DeepEqual(badge1.Style, badge2.Style) {
			t.Error("badges should have different styles")
		}

		// Verify they have different text
		if badge1.Text == badge2.Text {
			t.Error("badges should have different text")
		}
	})

	t.Run("badge text preserves case", func(t *testing.T) {
		cases := []string{"lowercase", "UPPERCASE", "MixedCase", "cAmElCaSe"}

		for _, c := range cases {
			t.Run(c, func(t *testing.T) {
				elem := Badge(c, retui.White, retui.Blue)
				if elem.Text != " "+c+" " {
					t.Errorf("expected preserved case %q, got %q", " "+c+" ", elem.Text)
				}
			})
		}
	})

	t.Run("badge with numbers in label", func(t *testing.T) {
		numberLabel := "12345"
		elem := Badge(numberLabel, retui.White, retui.Navy)

		if elem.Text != " "+numberLabel+" " {
			t.Errorf("expected text with numbers, got %q", elem.Text)
		}
	})

	t.Run("badge with mixed alphanumeric", func(t *testing.T) {
		mixedLabel := "ABC123"
		elem := Badge(mixedLabel, retui.Cyan, retui.Black)

		if elem.Text != " "+mixedLabel+" " {
			t.Errorf("expected text with mixed alphanumeric, got %q", elem.Text)
		}
	})
}

// Benchmark tests
func BenchmarkBadge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Badge("Benchmark", retui.White, retui.Blue)
	}
}

func BenchmarkBadgeLongLabel(b *testing.B) {
	longLabel := "This is a very long badge label for benchmarking"
	for i := 0; i < b.N; i++ {
		Badge(longLabel, retui.White, retui.Blue)
	}
}

// Example usage test
func ExampleBadge() {
	// Create a success badge
	successBadge := Badge("Success", retui.Green, retui.Black)
	_ = successBadge

	// Create an error badge
	errorBadge := Badge("Error", retui.Red, retui.White)
	_ = errorBadge

	// Create a warning badge
	warningBadge := Badge("Warning", retui.Yellow, retui.Navy)
	_ = warningBadge
}
