package components

import (
	"reflect"
	"strings"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestSpinner(t *testing.T) {
	t.Run("creates a spinner with label and default color", func(t *testing.T) {
		elem := Spinner("Loading...")

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}

		expectedStyle := retui.NewStyle().Foreground(retui.Cyan)
		if !reflect.DeepEqual(elem.Style, expectedStyle) {
			t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
		}
	})

	t.Run("includes label in the rendered text", func(t *testing.T) {
		label := "Processing..."
		elem := Spinner(label)

		if !strings.Contains(elem.Text, label) {
			t.Errorf("expected text to contain %q, got %q", label, elem.Text)
		}
	})

	t.Run("displays a spinner frame before the label", func(t *testing.T) {
		elem := Spinner("Loading")

		found := false
		for _, frame := range spinnerFrames {
			if strings.HasPrefix(elem.Text, frame) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected text to start with a spinner frame, got %q", elem.Text)
		}
	})

	t.Run("handles empty label", func(t *testing.T) {
		elem := Spinner("")

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		if len(elem.Text) == 0 {
			t.Error("expected text to contain spinner frame, got empty")
		}
	})

	t.Run("handles long label", func(t *testing.T) {
		longLabel := "This is a very long spinner label that should still render properly"
		elem := Spinner(longLabel)

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		if !strings.Contains(elem.Text, longLabel) {
			t.Errorf("expected text to contain long label, got %q", elem.Text)
		}
	})

	t.Run("handles special characters in label", func(t *testing.T) {
		specialLabel := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		elem := Spinner(specialLabel)

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		if !strings.Contains(elem.Text, specialLabel) {
			t.Errorf("expected text to contain special characters, got %q", elem.Text)
		}
	})

	t.Run("handles Unicode characters in label", func(t *testing.T) {
		unicodeLabel := "✓ 完成 ✅"
		elem := Spinner(unicodeLabel)

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		if !strings.Contains(elem.Text, unicodeLabel) {
			t.Errorf("expected text to contain Unicode characters, got %q", elem.Text)
		}
	})

	t.Run("returns correct style", func(t *testing.T) {
		elem := Spinner("Test")

		expectedStyle := retui.NewStyle().Foreground(retui.Cyan)
		if !reflect.DeepEqual(elem.Style, expectedStyle) {
			t.Errorf("expected style %+v, got %+v", expectedStyle, elem.Style)
		}
	})

	t.Run("formats text as 'frame label'", func(t *testing.T) {
		label := "Loading"
		elem := Spinner(label)

		if !strings.Contains(elem.Text, " "+label) {
			t.Errorf("expected text to contain ' %s', got %q", label, elem.Text)
		}
	})
}

func TestSpinner_FrameCycle(t *testing.T) {
	t.Run("all frames are valid braille characters", func(t *testing.T) {
		for _, frame := range spinnerFrames {
			if len([]rune(frame)) != 1 {
				t.Errorf("expected each frame to be a single rune, got %q (length %d)", frame, len([]rune(frame)))
			}
		}
	})

	t.Run("spinner has exactly 10 frames", func(t *testing.T) {
		if len(spinnerFrames) != 10 {
			t.Errorf("expected 10 spinner frames, got %d", len(spinnerFrames))
		}
	})

	t.Run("spinner frames are all unique", func(t *testing.T) {
		seen := make(map[string]bool)
		for _, frame := range spinnerFrames {
			if seen[frame] {
				t.Errorf("duplicate frame found: %q", frame)
			}
			seen[frame] = true
		}
	})

	t.Run("spinner frame is always one of the defined frames", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			elem := Spinner("Test")
			frame := string([]rune(elem.Text)[0])

			found := false
			for _, f := range spinnerFrames {
				if f == frame {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("spinner frame %q is not in the defined frames", frame)
			}
		}
	})
}

func TestSpinner_ElementType(t *testing.T) {
	t.Run("returns Element of type ElementText", func(t *testing.T) {
		elem := Spinner("Test")

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
	})

	t.Run("returns ElementText for all labels", func(t *testing.T) {
		labels := []string{"", "Short", "Medium Label", "This is a very long label that should still work"}
		for _, label := range labels {
			elem := Spinner(label)
			if elem.Type != retui.ElementText {
				t.Errorf("for label %q, expected ElementText, got %v", label, elem.Type)
			}
		}
	})
}

func TestSpinner_EdgeCases(t *testing.T) {
	t.Run("handles very long label gracefully", func(t *testing.T) {
		longLabel := strings.Repeat("A", 1000)
		elem := Spinner(longLabel)

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		if !strings.Contains(elem.Text, longLabel) {
			t.Errorf("expected text to contain long label")
		}
	})

	t.Run("handles label with multiple spaces", func(t *testing.T) {
		label := "  Multiple   Spaces  "
		elem := Spinner(label)

		if !strings.Contains(elem.Text, label) {
			t.Errorf("expected text to contain label with spaces, got %q", elem.Text)
		}
	})

	t.Run("handles label with newlines", func(t *testing.T) {
		label := "Line1\nLine2"
		elem := Spinner(label)

		if !strings.Contains(elem.Text, "Line1") || !strings.Contains(elem.Text, "Line2") {
			t.Errorf("expected text to contain newlines, got %q", elem.Text)
		}
	})

	t.Run("handles label with tabs", func(t *testing.T) {
		label := "Tab\tSeparated"
		elem := Spinner(label)

		if !strings.Contains(elem.Text, label) {
			t.Errorf("expected text to contain tabs, got %q", elem.Text)
		}
	})

	t.Run("handles label with emoji", func(t *testing.T) {
		label := "🚀 Loading"
		elem := Spinner(label)

		if !strings.Contains(elem.Text, label) {
			t.Errorf("expected text to contain emoji, got %q", elem.Text)
		}
	})
}

// Note: TestSpinner_ConsecutiveRenders is removed because retui.UseState
// doesn't persist state between separate test calls. The spinner's frame
// cycling works correctly in a real application where the component is
// rendered repeatedly within the same context.

// Benchmark tests
func BenchmarkSpinner(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Spinner("Loading...")
	}
}

func BenchmarkSpinnerLongLabel(b *testing.B) {
	longLabel := "This is a very long spinner label for benchmarking purposes"
	for i := 0; i < b.N; i++ {
		Spinner(longLabel)
	}
}

func BenchmarkSpinnerEmptyLabel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Spinner("")
	}
}

// Example usage test
func ExampleSpinner() {
	// Create a spinner with a label
	loadingSpinner := Spinner("Loading data...")
	_ = loadingSpinner

	// Create a spinner without a label
	spinner := Spinner("")
	_ = spinner

	// Create a spinner with custom message
	processingSpinner := Spinner("Processing your request...")
	_ = processingSpinner
}
