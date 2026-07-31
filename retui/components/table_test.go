package components

import (
	"reflect"
	"strings"
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestTableBuilder_Defaults(t *testing.T) {
	tbl := Table()

	if tbl.config.ID != "" {
		t.Errorf("expected empty ID, got %q", tbl.config.ID)
	}
	if len(tbl.config.headers) > 0 {
		t.Errorf("expected empty headers, got %v", tbl.config.headers)
	}
	if len(tbl.config.rows) > 0 {
		t.Errorf("expected empty rows, got %v", tbl.config.rows)
	}
	if tbl.config.focused {
		t.Errorf("expected focused false, got %v", tbl.config.focused)
	}
	if tbl.config.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0, got %d", tbl.config.selectedIndex)
	}
	if tbl.config.minColumnWidth != 3 {
		t.Errorf("expected minColumnWidth 3, got %d", tbl.config.minColumnWidth)
	}
	if tbl.config.cellPadding != 1 {
		t.Errorf("expected cellPadding 1, got %d", tbl.config.cellPadding)
	}
	if !tbl.config.showHeaders {
		t.Errorf("expected showHeaders true, got %v", tbl.config.showHeaders)
	}
	if !tbl.config.showBorders {
		t.Errorf("expected showBorders true, got %v", tbl.config.showBorders)
	}
	if !tbl.config.showSelection {
		t.Errorf("expected showSelection true, got %v", tbl.config.showSelection)
	}
	if !tbl.config.selectable {
		t.Errorf("expected selectable true, got %v", tbl.config.selectable)
	}
	if tbl.config.width != 0 {
		t.Errorf("expected width 0, got %d", tbl.config.width)
	}
	if tbl.config.height != 0 {
		t.Errorf("expected height 0, got %d", tbl.config.height)
	}
	if tbl.config.explicitWidth {
		t.Errorf("expected explicitWidth false, got %v", tbl.config.explicitWidth)
	}
	if tbl.config.explicitHeight {
		t.Errorf("expected explicitHeight false, got %v", tbl.config.explicitHeight)
	}
}

func TestTableBuilder_ID(t *testing.T) {
	tbl := Table().ID("test-id")
	if tbl.config.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %s", tbl.config.ID)
	}
}

func TestTableBuilder_Headers(t *testing.T) {
	t.Run("sets Headers correctly", func(t *testing.T) {
		headers := []string{"Name", "Age", "City"}
		tbl := Table().Headers(headers)
		if !reflect.DeepEqual(tbl.config.headers, headers) {
			t.Errorf("expected headers %v, got %v", headers, tbl.config.headers)
		}
	})
}

func TestTableBuilder_Rows(t *testing.T) {
	t.Run("sets Rows correctly", func(t *testing.T) {
		rows := [][]string{
			{"Alice", "30", "NYC"},
			{"Bob", "25", "LA"},
		}
		tbl := Table().Rows(rows)
		if !reflect.DeepEqual(tbl.config.rows, rows) {
			t.Errorf("expected rows %v, got %v", rows, tbl.config.rows)
		}
	})
}

func TestTableBuilder_Focused(t *testing.T) {
	t.Run("sets Focused true", func(t *testing.T) {
		tbl := Table().Focused(true)
		if !tbl.config.focused {
			t.Errorf("expected focused true, got %v", tbl.config.focused)
		}
	})

	t.Run("sets Focused false", func(t *testing.T) {
		tbl := Table().Focused(false)
		if tbl.config.focused {
			t.Errorf("expected focused false, got %v", tbl.config.focused)
		}
	})
}

func TestTableBuilder_OnChange(t *testing.T) {
	var changed bool
	var captured int
	tbl := Table().OnChange(func(index int) {
		changed = true
		captured = index
	})

	if tbl.config.onChange == nil {
		t.Fatal("OnChange callback not set")
	}
	tbl.config.onChange(5)
	if !changed {
		t.Error("OnChange callback was not executed")
	}
	if captured != 5 {
		t.Errorf("expected index 5, got %d", captured)
	}
}

func TestTableBuilder_SelectedIndex(t *testing.T) {
	tbl := Table().SelectedIndex(3)
	if tbl.config.selectedIndex != 3 {
		t.Errorf("expected selectedIndex 3, got %d", tbl.config.selectedIndex)
	}
}

func TestTableBuilder_ColumnWidths(t *testing.T) {
	widths := []int{10, 20, 15}
	tbl := Table().ColumnWidths(widths)
	if !reflect.DeepEqual(tbl.config.columnWidths, widths) {
		t.Errorf("expected columnWidths %v, got %v", widths, tbl.config.columnWidths)
	}
}

func TestTableBuilder_MinColumnWidth(t *testing.T) {
	tbl := Table().MinColumnWidth(5)
	if tbl.config.minColumnWidth != 5 {
		t.Errorf("expected minColumnWidth 5, got %d", tbl.config.minColumnWidth)
	}
}

func TestTableBuilder_MaxColumnWidth(t *testing.T) {
	tbl := Table().MaxColumnWidth(50)
	if tbl.config.maxColumnWidth != 50 {
		t.Errorf("expected maxColumnWidth 50, got %d", tbl.config.maxColumnWidth)
	}
}

func TestTableBuilder_CellPadding(t *testing.T) {
	tbl := Table().CellPadding(2)
	if tbl.config.cellPadding != 2 {
		t.Errorf("expected cellPadding 2, got %d", tbl.config.cellPadding)
	}
}

func TestTableBuilder_Colors(t *testing.T) {
	t.Run("sets HeaderColor correctly", func(t *testing.T) {
		tbl := Table().HeaderColor(retui.Red)
		if tbl.config.headerColor != retui.Red {
			t.Errorf("expected headerColor Red, got %v", tbl.config.headerColor)
		}
	})

	t.Run("sets HeaderBackground correctly", func(t *testing.T) {
		tbl := Table().HeaderBackground(retui.Blue)
		if tbl.config.headerBg != retui.Blue {
			t.Errorf("expected headerBg Blue, got %v", tbl.config.headerBg)
		}
	})

	t.Run("sets SelectedBackground correctly", func(t *testing.T) {
		tbl := Table().SelectedBackground(retui.Green)
		if tbl.config.selectedBg != retui.Green {
			t.Errorf("expected selectedBg Green, got %v", tbl.config.selectedBg)
		}
	})

	t.Run("sets SelectedForeground correctly", func(t *testing.T) {
		tbl := Table().SelectedForeground(retui.Yellow)
		if tbl.config.selectedFg != retui.Yellow {
			t.Errorf("expected selectedFg Yellow, got %v", tbl.config.selectedFg)
		}
	})

	t.Run("sets RowColor correctly", func(t *testing.T) {
		tbl := Table().RowColor(retui.Cyan)
		if tbl.config.rowColor != retui.Cyan {
			t.Errorf("expected rowColor Cyan, got %v", tbl.config.rowColor)
		}
	})

	t.Run("sets RowBackground correctly", func(t *testing.T) {
		tbl := Table().RowBackground(retui.Navy)
		if tbl.config.rowBg != retui.Navy {
			t.Errorf("expected rowBg Navy, got %v", tbl.config.rowBg)
		}
	})

	t.Run("sets BorderColor correctly", func(t *testing.T) {
		tbl := Table().BorderColor(retui.White)
		if tbl.config.borderColor != retui.White {
			t.Errorf("expected borderColor White, got %v", tbl.config.borderColor)
		}
	})
}

func TestTableBuilder_Styles(t *testing.T) {
	t.Run("sets HeaderStyle correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Red).Bold(true)
		tbl := Table().HeaderStyle(style)
		if !reflect.DeepEqual(tbl.config.headerStyle, style) {
			t.Errorf("HeaderStyle not set correctly")
		}
	})

	t.Run("sets SelectedStyle correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.White).Background(retui.Blue)
		tbl := Table().SelectedStyle(style)
		if !reflect.DeepEqual(tbl.config.selectedStyle, style) {
			t.Errorf("SelectedStyle not set correctly")
		}
	})

	t.Run("sets RowStyle correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.BrightBlack)
		tbl := Table().RowStyle(style)
		if !reflect.DeepEqual(tbl.config.rowStyle, style) {
			t.Errorf("RowStyle not set correctly")
		}
	})

	t.Run("sets BorderStyle correctly", func(t *testing.T) {
		style := retui.NewStyle().Foreground(retui.Gray(1))
		tbl := Table().BorderStyle(style)
		if !reflect.DeepEqual(tbl.config.borderStyle, style) {
			t.Errorf("BorderStyle not set correctly")
		}
	})
}

func TestTableBuilder_Alignments(t *testing.T) {
	alignments := []string{"left", "center", "right"}
	tbl := Table().Alignments(alignments)
	if !reflect.DeepEqual(tbl.config.alignments, alignments) {
		t.Errorf("expected alignments %v, got %v", alignments, tbl.config.alignments)
	}
}

func TestTableBuilder_ShowFlags(t *testing.T) {
	t.Run("sets ShowHeaders correctly", func(t *testing.T) {
		tbl := Table().ShowHeaders(false)
		if tbl.config.showHeaders {
			t.Errorf("expected showHeaders false, got %v", tbl.config.showHeaders)
		}
	})

	t.Run("sets ShowBorders correctly", func(t *testing.T) {
		tbl := Table().ShowBorders(false)
		if tbl.config.showBorders {
			t.Errorf("expected showBorders false, got %v", tbl.config.showBorders)
		}
	})

	t.Run("sets ShowSelection correctly", func(t *testing.T) {
		tbl := Table().ShowSelection(false)
		if tbl.config.showSelection {
			t.Errorf("expected showSelection false, got %v", tbl.config.showSelection)
		}
	})
}

func TestTableBuilder_Selectable(t *testing.T) {
	t.Run("sets Selectable false", func(t *testing.T) {
		tbl := Table().Selectable(false)
		if tbl.config.selectable {
			t.Errorf("expected selectable false, got %v", tbl.config.selectable)
		}
	})
}

func TestTableBuilder_Width(t *testing.T) {
	t.Run("sets Width correctly", func(t *testing.T) {
		tbl := Table().Width(100)
		if tbl.config.width != 100 {
			t.Errorf("expected width 100, got %d", tbl.config.width)
		}
		if !tbl.config.explicitWidth {
			t.Errorf("expected explicitWidth true, got %v", tbl.config.explicitWidth)
		}
	})
}

func TestTableBuilder_Height(t *testing.T) {
	t.Run("sets Height correctly", func(t *testing.T) {
		tbl := Table().Height(20)
		if tbl.config.height != 20 {
			t.Errorf("expected height 20, got %d", tbl.config.height)
		}
		if !tbl.config.explicitHeight {
			t.Errorf("expected explicitHeight true, got %v", tbl.config.explicitHeight)
		}
	})
}

func TestTableChaining(t *testing.T) {
	headers := []string{"Name", "Age"}
	rows := [][]string{{"Alice", "30"}, {"Bob", "25"}}

	tbl := Table().
		ID("chain-test").
		Headers(headers).
		Rows(rows).
		Focused(true).
		SelectedIndex(1).
		ColumnWidths([]int{20, 10}).
		MinColumnWidth(5).
		MaxColumnWidth(30).
		CellPadding(2).
		HeaderColor(retui.Red).
		HeaderBackground(retui.Blue).
		SelectedBackground(retui.Green).
		SelectedForeground(retui.Yellow).
		RowColor(retui.Cyan).
		RowBackground(retui.Navy).
		BorderColor(retui.White).
		Alignments([]string{"left", "right"}).
		ShowHeaders(false).
		ShowBorders(false).
		ShowSelection(false).
		Selectable(false).
		Width(100).
		Height(20)

	if tbl.config.ID != "chain-test" {
		t.Errorf("expected ID 'chain-test', got %s", tbl.config.ID)
	}
	if !reflect.DeepEqual(tbl.config.headers, headers) {
		t.Errorf("expected headers %v, got %v", headers, tbl.config.headers)
	}
	if !reflect.DeepEqual(tbl.config.rows, rows) {
		t.Errorf("expected rows %v, got %v", rows, tbl.config.rows)
	}
	if !tbl.config.focused {
		t.Errorf("expected focused true, got %v", tbl.config.focused)
	}
	if tbl.config.selectedIndex != 1 {
		t.Errorf("expected selectedIndex 1, got %d", tbl.config.selectedIndex)
	}
	if tbl.config.minColumnWidth != 5 {
		t.Errorf("expected minColumnWidth 5, got %d", tbl.config.minColumnWidth)
	}
	if tbl.config.cellPadding != 2 {
		t.Errorf("expected cellPadding 2, got %d", tbl.config.cellPadding)
	}
	if tbl.config.width != 100 {
		t.Errorf("expected width 100, got %d", tbl.config.width)
	}
	if tbl.config.height != 20 {
		t.Errorf("expected height 20, got %d", tbl.config.height)
	}
}

func TestTableRender(t *testing.T) {
	t.Run("returns an Element with ContentBuilder", func(t *testing.T) {
		tbl := Table().Headers([]string{"A", "B"}).Rows([][]string{{"1", "2"}})
		elem := tbl.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
		if elem.ContentBuilder == nil {
			t.Error("expected ContentBuilder to be set")
		}
	})

	t.Run("builds table with headers", func(t *testing.T) {
		tbl := Table().
			Headers([]string{"Name", "Age"}).
			Rows([][]string{{"Alice", "30"}, {"Bob", "25"}}).
			Width(30)

		elem := tbl.Render()
		if elem.ContentBuilder == nil {
			t.Error("expected ContentBuilder to be set")
		}

		// Build the table
		built := elem.ContentBuilder(30, 10)
		if built.Type != retui.ElementBox {
			t.Errorf("expected built element of type ElementBox, got %v", built.Type)
		}
	})

	t.Run("builds table without headers", func(t *testing.T) {
		tbl := Table().
			Rows([][]string{{"Alice", "30"}, {"Bob", "25"}}).
			ShowHeaders(false).
			Width(30)

		elem := tbl.Render()
		built := elem.ContentBuilder(30, 10)
		if built.Type != retui.ElementBox {
			t.Errorf("expected built element of type ElementBox, got %v", built.Type)
		}
	})

	t.Run("builds table without borders", func(t *testing.T) {
		tbl := Table().
			Headers([]string{"Name", "Age"}).
			Rows([][]string{{"Alice", "30"}, {"Bob", "25"}}).
			ShowBorders(false).
			Width(30)

		elem := tbl.Render()
		built := elem.ContentBuilder(30, 10)
		if built.Type != retui.ElementBox {
			t.Errorf("expected built element of type ElementBox, got %v", built.Type)
		}
	})
}

func TestTableHelperFunctions(t *testing.T) {
	tbl := Table()

	t.Run("displayWidth calculates width correctly", func(t *testing.T) {
		tests := []struct {
			text string
			want int
		}{
			{"Hello", 5},
			{"你好", 4}, // Chinese characters are 2 cells wide each
			{"", 0},
		}
		for _, tt := range tests {
			if got := tbl.displayWidth(tt.text); got != tt.want {
				t.Errorf("displayWidth(%q) = %d, want %d", tt.text, got, tt.want)
			}
		}
	})

	t.Run("truncateToWidth truncates correctly", func(t *testing.T) {
		tests := []struct {
			text    string
			width   int
			padding int
			want    string
		}{
			{"Hello", 10, 1, "Hello"},
			{"Hello World", 8, 1, "Hello…"}, // Actual behavior: keeps "Hello" + "…"
			{"Hello", 3, 1, "H"},            // Actual behavior: maxTextWidth = 1, keeps first char
			{"Hi", 5, 1, "Hi"},
		}
		for _, tt := range tests {
			if got := tbl.truncateToWidth(tt.text, tt.width, tt.padding); got != tt.want {
				t.Errorf("truncateToWidth(%q, %d, %d) = %q, want %q", tt.text, tt.width, tt.padding, got, tt.want)
			}
		}
	})

	t.Run("padText pads correctly with left alignment", func(t *testing.T) {
		result := tbl.padText("Test", 10, "left", 1)
		// width 10, padding 1 on each side, text "Test" (4 chars)
		// Total: 1 + 4 + 5 = 10
		expected := " Test     " // 1 padding + 4 text + 5 spaces = 10
		if result != expected {
			t.Errorf("padText() = %q, want %q", result, expected)
		}
	})

	t.Run("padText pads correctly with center alignment", func(t *testing.T) {
		result := tbl.padText("Test", 10, "center", 1)
		// Should have 1 padding on each side plus centering
		if len(result) != 10 {
			t.Errorf("expected length 10, got %d", len(result))
		}
	})

	t.Run("padText pads correctly with right alignment", func(t *testing.T) {
		result := tbl.padText("Test", 10, "right", 1)
		// width 10, padding 1 on each side, text "Test" (4 chars)
		// Total: 5 spaces + 4 text + 1 padding = 10
		expected := "     Test " // 5 spaces + 4 text + 1 padding = 10
		if result != expected {
			t.Errorf("padText() = %q, want %q", result, expected)
		}
	})

	t.Run("getAlignment returns default alignment", func(t *testing.T) {
		alignment := tbl.getAlignment(0)
		if alignment != "left" {
			t.Errorf("expected 'left', got %q", alignment)
		}
	})

	t.Run("getAlignment returns custom alignment", func(t *testing.T) {
		tbl.config.alignments = []string{"center", "right"}
		if got := tbl.getAlignment(0); got != "center" {
			t.Errorf("expected 'center', got %q", got)
		}
		if got := tbl.getAlignment(1); got != "right" {
			t.Errorf("expected 'right', got %q", got)
		}
		if got := tbl.getAlignment(2); got != "left" {
			t.Errorf("expected 'left', got %q", got)
		}
	})
}

func TestTableEdgeCases(t *testing.T) {
	t.Run("handles empty headers", func(t *testing.T) {
		tbl := Table().Headers([]string{}).Rows([][]string{{"A", "B"}})
		elem := tbl.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles empty rows", func(t *testing.T) {
		tbl := Table().Headers([]string{"A", "B"}).Rows([][]string{})
		elem := tbl.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles rows with different column counts", func(t *testing.T) {
		tbl := Table().
			Headers([]string{"A", "B", "C"}).
			Rows([][]string{
				{"1", "2"},
				{"3", "4", "5", "6"},
			})
		elem := tbl.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles negative width", func(t *testing.T) {
		tbl := Table().Width(-5)
		elem := tbl.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles negative height", func(t *testing.T) {
		tbl := Table().Height(-5)
		elem := tbl.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles selected index out of bounds", func(t *testing.T) {
		tbl := Table().
			Rows([][]string{{"A"}, {"B"}}).
			SelectedIndex(10)
		elem := tbl.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very long text in cells", func(t *testing.T) {
		longText := strings.Repeat("A", 100)
		tbl := Table().
			Headers([]string{"Long"}).
			Rows([][]string{{longText}}).
			Width(30)
		elem := tbl.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles many rows", func(t *testing.T) {
		rows := make([][]string, 100)
		for i := 0; i < 100; i++ {
			rows[i] = []string{string(rune('A' + i%26)), string(rune('a' + i%26))}
		}
		tbl := Table().Headers([]string{"A", "B"}).Rows(rows).Height(10)
		elem := tbl.Render()
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

// Benchmark tests
func BenchmarkTableRender(b *testing.B) {
	rows := make([][]string, 100)
	for i := 0; i < 100; i++ {
		rows[i] = []string{
			"Name " + string(rune('A'+i%26)),
			string(rune('0' + i%10)),
			"City " + string(rune('A'+i%26)),
		}
	}
	tbl := Table().
		Headers([]string{"Name", "Age", "City"}).
		Rows(rows).
		Width(80).
		Height(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tbl.Render()
	}
}

func BenchmarkTableBuild(b *testing.B) {
	rows := make([][]string, 50)
	for i := 0; i < 50; i++ {
		rows[i] = []string{"Name", "25", "City"}
	}
	tbl := Table().
		Headers([]string{"Name", "Age", "City"}).
		Rows(rows).
		Width(60)

	elem := tbl.Render()
	contentBuilder := elem.ContentBuilder

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contentBuilder(60, 20)
	}
}

// Example usage test
func ExampleTable() {
	// Create a table
	table := Table().
		ID("users").
		Headers([]string{"Name", "Age", "City"}).
		Rows([][]string{
			{"Alice", "30", "NYC"},
			{"Bob", "25", "LA"},
			{"Charlie", "35", "Chicago"},
		}).
		Width(50).
		Height(10).
		SelectedIndex(0).
		OnChange(func(index int) {
			// Handle selection change
		}).
		Render()
	_ = table
}
