package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
	"github.com/subhasundardass/retui/retui/window"
)

// Table represents a table component with fluent API
type TableField struct {
	config tableConfig
}

const BORDER_COLOR = "#1d1d1d"

type tableConfig struct {
	ID            string
	headers       []string
	rows          [][]string
	focused       bool
	onChange      func(int)
	selectedIndex int

	columnWidths   []int
	minColumnWidth int
	maxColumnWidth int

	cellPadding int

	headerColor   retui.Color
	headerBg      retui.Color
	headerStyle   retui.Style
	selectedBg    retui.Color
	selectedFg    retui.Color
	selectedStyle retui.Style
	rowColor      retui.Color
	rowBg         retui.Color
	rowStyle      retui.Style
	borderColor   retui.Color
	borderStyle   retui.Style

	alignments []string

	showHeaders   bool
	showBorders   bool
	showSelection bool

	selectable bool

	// width/height are the table's resolved size. They can still be set
	// explicitly via .Width()/.Height() for a hard override, but normally
	// they're filled in automatically from the real space the layout
	// engine assigns this component — see Render()'s use of
	// retui.Element.ContentBuilder, which receives the resolved
	// width/height once the surrounding Box has been laid out.
	width  int
	height int

	// explicitWidth/explicitHeight track whether .Width()/.Height() were
	// called by the caller, so Render() knows whether to advertise a
	// Fixed sizing (respect the override) or Grow/Fit (fill/auto, and
	// accept whatever the parent assigns).
	explicitWidth  bool
	explicitHeight bool
}

// NewTable creates a new table with default configuration
func Table() *TableField {
	return &TableField{
		config: tableConfig{
			ID:             "",
			minColumnWidth: 3,
			cellPadding:    1,
			showHeaders:    true,
			showBorders:    true,
			showSelection:  true,
			selectable:     true,
			headerColor:    retui.Cyan,
			selectedBg:     retui.Blue,
			selectedFg:     retui.White,
			borderColor:    retui.Hex(BORDER_COLOR),
			rowColor:       retui.BrightBlack,
			alignments:     []string{},
		},
	}
}

// ID of table
func (i *TableField) ID(v string) *TableField {
	i.config.ID = v
	return i
}

// Headers sets the table headers
func (t *TableField) Headers(headers []string) *TableField {
	t.config.headers = headers
	return t
}

// Rows sets the table rows
func (t *TableField) Rows(rows [][]string) *TableField {
	t.config.rows = rows
	return t
}

// Focused sets whether the table is focused
func (t *TableField) Focused(focused bool) *TableField {
	t.config.focused = focused
	return t
}

// OnChange sets the callback for selection changes
func (t *TableField) OnChange(fn func(int)) *TableField {
	t.config.onChange = fn
	return t
}

// SelectedIndex sets the selected row index
func (t *TableField) SelectedIndex(index int) *TableField {
	t.config.selectedIndex = index
	return t
}

// ColumnWidths sets explicit column widths
func (t *TableField) ColumnWidths(widths []int) *TableField {
	t.config.columnWidths = widths
	return t
}

// MinColumnWidth sets the minimum column width
func (t *TableField) MinColumnWidth(width int) *TableField {
	t.config.minColumnWidth = width
	return t
}

// MaxColumnWidth sets the maximum column width
func (t *TableField) MaxColumnWidth(width int) *TableField {
	t.config.maxColumnWidth = width
	return t
}

// CellPadding sets the cell padding
func (t *TableField) CellPadding(padding int) *TableField {
	t.config.cellPadding = padding
	return t
}

// HeaderColor sets the header text color
func (t *TableField) HeaderColor(color retui.Color) *TableField {
	t.config.headerColor = color
	return t
}

// HeaderBackground sets the header background color
func (t *TableField) HeaderBackground(color retui.Color) *TableField {
	t.config.headerBg = color
	return t
}

// HeaderStyle sets the header text style
func (t *TableField) HeaderStyle(style retui.Style) *TableField {
	t.config.headerStyle = style
	return t
}

// SelectedBackground sets the selected row background color
func (t *TableField) SelectedBackground(color retui.Color) *TableField {
	t.config.selectedBg = color
	return t
}

// SelectedForeground sets the selected row text color
func (t *TableField) SelectedForeground(color retui.Color) *TableField {
	t.config.selectedFg = color
	return t
}

// SelectedStyle sets the selected row text style
func (t *TableField) SelectedStyle(style retui.Style) *TableField {
	t.config.selectedStyle = style
	return t
}

// RowColor sets the row text color
func (t *TableField) RowColor(color retui.Color) *TableField {
	t.config.rowColor = color
	return t
}

// RowBackground sets the row background color
func (t *TableField) RowBackground(color retui.Color) *TableField {
	t.config.rowBg = color
	return t
}

// RowStyle sets the row text style
func (t *TableField) RowStyle(style retui.Style) *TableField {
	t.config.rowStyle = style
	return t
}

// BorderColor sets the border color
func (t *TableField) BorderColor(color retui.Color) *TableField {
	t.config.borderColor = color
	return t
}

// BorderStyle sets the border style
func (t *TableField) BorderStyle(style retui.Style) *TableField {
	t.config.borderStyle = style
	return t
}

// Alignments sets column alignments (left, center, right)
func (t *TableField) Alignments(alignments []string) *TableField {
	t.config.alignments = alignments
	return t
}

// ShowHeaders sets whether to show headers
func (t *TableField) ShowHeaders(show bool) *TableField {
	t.config.showHeaders = show
	return t
}

// ShowBorders sets whether to show borders
func (t *TableField) ShowBorders(show bool) *TableField {
	t.config.showBorders = show
	return t
}

// ShowSelection sets whether to show selection
func (t *TableField) ShowSelection(show bool) *TableField {
	t.config.showSelection = show
	return t
}

// Selectable sets whether the table is selectable
func (t *TableField) Selectable(selectable bool) *TableField {
	t.config.selectable = selectable
	return t
}

// Width sets an explicit, hard-coded table width. Optional — if you don't
// call this, the table automatically fills whatever width its parent
// (e.g. a Box) assigns it.
func (t *TableField) Width(width int) *TableField {
	t.config.width = width
	t.config.explicitWidth = true
	return t
}

// Height sets an explicit, hard-coded table height (enables row
// scrolling/truncation to fit). Optional — if you don't call this, the
// table shows all rows and sizes itself naturally to fit its content.
func (t *TableField) Height(height int) *TableField {
	t.config.height = height
	t.config.explicitHeight = true
	return t
}

// Render renders the table as a retui.Element.
//
// It does NOT build the table's rows/columns immediately. Column widths
// depend on the space actually available, which isn't known yet at this
// point — it depends on whatever parent (e.g. Box) this element ends up
// inside, and that's only resolved during the layout pass. So Render
// returns a placeholder Element carrying a ContentBuilder: the renderer
// calls it back with the real resolved width/height once layout knows
// them, and *that's* when t.build() actually runs.
func (t *TableField) Render() retui.Element {
	cfg := &t.config

	selected, setSelected := retui.UseState(cfg.selectedIndex)

	if cfg.focused && cfg.selectable && !window.IsAnyModalOpen() {
		switch retui.CurrentKey.Code {
		case retui.KeyDown:
			if selected < len(cfg.rows)-1 {
				selected++
				setSelected(selected)
				if cfg.onChange != nil {
					cfg.onChange(selected)
				}
			}
		case retui.KeyUp:
			if selected > 0 {
				selected--
				setSelected(selected)
				if cfg.onChange != nil {
					cfg.onChange(selected)
				}
			}
		case retui.KeyEnter:
			if cfg.onChange != nil {
				cfg.onChange(selected)
			}
		}
	}

	cfg.selectedIndex = selected

	widthSizing := retui.Grow(1)
	if cfg.explicitWidth && cfg.width > 0 {
		widthSizing = retui.Fixed(cfg.width)
	}

	heightSizing := retui.Grow(1)
	if cfg.explicitHeight && cfg.height > 0 {
		heightSizing = retui.Fixed(cfg.height)
	}

	return retui.Element{
		Type: retui.ElementBox,
		Layout: retui.LayoutProps{
			Direction:    retui.Column,
			WidthSizing:  widthSizing,
			HeightSizing: heightSizing,
		},
		ContentBuilder: func(width, height int) retui.Element {
			// Only adopt the resolved size when the caller didn't pin an
			// explicit one — an explicit .Width()/.Height() always wins.
			if !cfg.explicitWidth {
				cfg.width = width
			}
			if !cfg.explicitHeight {
				cfg.height = height
			}
			return t.build()
		},
	}
}

// build constructs the table element
func (t *TableField) build() retui.Element {
	cfg := &t.config
	colCount := len(cfg.headers)
	if len(cfg.rows) > 0 && len(cfg.rows[0]) > colCount {
		colCount = len(cfg.rows[0])
	}

	colWidths := t.calculateColumnWidths(colCount)

	var rows []retui.Element

	if cfg.showBorders {
		rows = append(rows, t.buildBorderLine(colWidths, "┌", "┬", "┐"))
	}

	if cfg.showHeaders {
		rows = append(rows, t.buildHeaderRow(colWidths))
		if cfg.showBorders {
			rows = append(rows, t.buildBorderLine(colWidths, "├", "┼", "┤"))
		}
	}

	start, end := t.visibleRowRange()
	for i := start; i < end && i < len(cfg.rows); i++ {
		row := cfg.rows[i]
		rowData := make([]string, colCount)
		for j := 0; j < colCount; j++ {
			if j < len(row) {
				rowData[j] = row[j]
			}
		}
		isSelected := cfg.showSelection && i == cfg.selectedIndex
		rows = append(rows, t.buildDataRow(rowData, colWidths, isSelected))
	}

	if cfg.showBorders {
		rows = append(rows, t.buildBorderLine(colWidths, "└", "┴", "┘"))
	}

	props := retui.Props{
		Direction: retui.Column,
	}

	if cfg.width > 0 {
		props.Width = retui.Fixed(cfg.width)
	}
	if cfg.height > 0 {
		props.Height = retui.Fixed(cfg.height)
	}

	return retui.Box(
		props,
		retui.NewStyle(),
		rows...,
	)
}

// Helper methods (same as before, just receiver methods now)

func (t *TableField) visibleRowRange() (int, int) {
	cfg := &t.config
	total := len(cfg.rows)
	if cfg.height <= 0 || total == 0 {
		return 0, total
	}

	overhead := 0
	if cfg.showBorders {
		overhead += 2
	}
	if cfg.showHeaders {
		overhead++
		if cfg.showBorders {
			overhead++
		}
	}

	visible := cfg.height - overhead
	if visible < 1 {
		visible = 1
	}
	if visible >= total {
		return 0, total
	}

	start := cfg.selectedIndex - visible + 1
	if start < 0 {
		start = 0
	}
	if maxStart := total - visible; start > maxStart {
		start = maxStart
	}

	return start, start + visible
}

func (t *TableField) calculateColumnWidths(colCount int) []int {
	cfg := &t.config
	widths := make([]int, colCount)
	if colCount == 0 {
		return widths
	}

	// cfg.width is now always populated before build() runs — either the
	// caller's explicit .Width(n), or the real width the layout engine
	// assigned this table's parent Box (see Render()'s ContentBuilder).
	// The terminal-width fallback below only exists as a last-resort
	// safety net (e.g. build() invoked outside the normal render path)
	// and should not be relied on in ordinary usage.
	availableWidth := cfg.width
	if availableWidth <= 0 {
		availableWidth = retui.StdOutScreen.Width()
	}

	if cfg.showBorders {
		availableWidth -= 2
		availableWidth -= colCount - 1
	}
	if availableWidth <= 0 {
		availableWidth = 10
	}

	totalNaturalWidth := 0
	for i := 0; i < colCount; i++ {
		if i < len(cfg.columnWidths) && cfg.columnWidths[i] > 0 {
			widths[i] = cfg.columnWidths[i]
			totalNaturalWidth += widths[i]
			continue
		}

		maxWidth := 0
		if cfg.showHeaders && i < len(cfg.headers) {
			if l := t.displayWidth(cfg.headers[i]); l > maxWidth {
				maxWidth = l
			}
		}
		for _, row := range cfg.rows {
			if i < len(row) {
				if l := t.displayWidth(row[i]); l > maxWidth {
					maxWidth = l
				}
			}
		}
		maxWidth += cfg.cellPadding * 2
		if maxWidth < cfg.minColumnWidth {
			maxWidth = cfg.minColumnWidth
		}
		if cfg.maxColumnWidth > 0 && maxWidth > cfg.maxColumnWidth {
			maxWidth = cfg.maxColumnWidth
		}
		widths[i] = maxWidth
		totalNaturalWidth += maxWidth
	}

	if totalNaturalWidth < availableWidth {
		extraSpace := availableWidth - totalNaturalWidth
		extraPerColumn := extraSpace / colCount
		remainder := extraSpace % colCount
		for i := 0; i < colCount; i++ {
			widths[i] += extraPerColumn
			if i < remainder {
				widths[i]++
			}
			if cfg.maxColumnWidth > 0 && widths[i] > cfg.maxColumnWidth {
				widths[i] = cfg.maxColumnWidth
			}
		}
	}

	if totalNaturalWidth > availableWidth {
		overflow := totalNaturalWidth - availableWidth
		t.shrinkColumnsToFit(widths, overflow)
	}

	return widths
}

func (t *TableField) shrinkColumnsToFit(widths []int, overflow int) {
	cfg := &t.config
	minWidth := cfg.minColumnWidth
	if minWidth < 1 {
		minWidth = 1
	}

	shrinkPass := func(floor int) {
		for overflow > 0 {
			shrankAny := false
			for i := range widths {
				if overflow == 0 {
					break
				}
				if widths[i] > floor {
					widths[i]--
					overflow--
					shrankAny = true
				}
			}
			if !shrankAny {
				break
			}
		}
	}

	shrinkPass(minWidth)
	if overflow > 0 {
		shrinkPass(1)
	}
}

func (t *TableField) buildHeaderRow(colWidths []int) retui.Element {
	cfg := &t.config
	var cells []retui.Element
	borderStyle := cfg.borderStyle
	borderStyle = borderStyle.Foreground(cfg.borderColor)

	if cfg.showBorders {
		cells = append(cells, retui.Text("│", borderStyle))
	}

	for i, header := range cfg.headers {
		width := colWidths[i]
		alignment := t.getAlignment(i)

		header = t.truncateToWidth(header, width, cfg.cellPadding)
		paddedText := t.padText(header, width, alignment, cfg.cellPadding)

		style := cfg.headerStyle
		style = style.Foreground(cfg.headerColor).Bold(true)
		if cfg.headerBg != (retui.Color{}) {
			style = style.Background(cfg.headerBg)
		}

		cells = append(cells, retui.Text(paddedText, style))

		if cfg.showBorders {
			cells = append(cells, retui.Text("│", borderStyle))
		}
	}

	return retui.Box(
		retui.Props{Direction: retui.Row},
		retui.NewStyle(),
		cells...,
	)
}

func (t *TableField) buildDataRow(rowData []string, colWidths []int, selected bool) retui.Element {
	cfg := &t.config
	var cells []retui.Element
	borderStyle := cfg.borderStyle
	borderStyle = borderStyle.Foreground(cfg.borderColor)

	if cfg.showBorders {
		cells = append(cells, retui.Text("│", borderStyle))
	}

	for i, cellText := range rowData {
		width := colWidths[i]
		alignment := t.getAlignment(i)

		cellText = t.truncateToWidth(cellText, width, cfg.cellPadding)
		paddedText := t.padText(cellText, width, alignment, cfg.cellPadding)

		var style retui.Style
		if selected && cfg.showSelection {
			style = cfg.selectedStyle
			style = style.Background(cfg.selectedBg).Foreground(cfg.selectedFg)
		} else {
			style = cfg.rowStyle
			style = style.Foreground(cfg.rowColor)
			if cfg.rowBg != (retui.Color{}) {
				style = style.Background(cfg.rowBg)
			}
		}

		cells = append(cells, retui.Text(paddedText, style))

		if cfg.showBorders {
			cells = append(cells, retui.Text("│", borderStyle))
		}
	}

	return retui.Box(
		retui.Props{Direction: retui.Row},
		retui.NewStyle(),
		cells...,
	)
}

func (t *TableField) buildBorderLine(colWidths []int, left, mid, right string) retui.Element {
	cfg := &t.config
	var parts []string
	for _, w := range colWidths {
		parts = append(parts, strings.Repeat("─", w))
	}
	line := left + strings.Join(parts, mid) + right

	style := cfg.borderStyle
	style = style.Foreground(cfg.borderColor)

	return retui.Text(line, style)
}

func (t *TableField) getAlignment(colIndex int) string {
	cfg := &t.config
	if colIndex < len(cfg.alignments) {
		return cfg.alignments[colIndex]
	}
	return "left"
}

func (t *TableField) displayWidth(s string) int {
	width := 0
	for _, r := range s {
		width += retui.RuneWidth(r)
	}
	return width
}

func (t *TableField) truncateToWidth(text string, width, padding int) string {
	maxTextWidth := width - padding*2
	if maxTextWidth <= 0 {
		return ""
	}
	if t.displayWidth(text) <= maxTextWidth {
		return text
	}

	budget := maxTextWidth
	if maxTextWidth > 3 {
		budget = maxTextWidth - 1
	}

	runes := []rune(text)
	w, cut := 0, 0
	for _, r := range runes {
		rw := retui.RuneWidth(r)
		if w+rw > budget {
			break
		}
		w += rw
		cut++
	}
	truncated := string(runes[:cut])

	if maxTextWidth > 3 {
		return truncated + "…"
	}
	return truncated
}

func (t *TableField) padText(text string, width int, alignment string, padding int) string {
	availableWidth := width - padding*2
	if availableWidth < 0 {
		availableWidth = 0
	}
	textWidth := t.displayWidth(text)
	if textWidth > availableWidth {
		text = t.truncateToWidth(text, width, padding)
		textWidth = t.displayWidth(text)
	}
	pad := padding
	if pad < 0 {
		pad = 0
	}

	switch alignment {
	case "center":
		leftPad := (availableWidth - textWidth) / 2
		rightPad := availableWidth - textWidth - leftPad
		return strings.Repeat(" ", pad+leftPad) + text + strings.Repeat(" ", pad+rightPad)
	case "right":
		leftPad := availableWidth - textWidth
		return strings.Repeat(" ", pad+leftPad) + text + strings.Repeat(" ", pad)
	default:
		return strings.Repeat(" ", pad) + text + strings.Repeat(" ", availableWidth-textWidth+pad)
	}
}
