package retui

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// Screen represents a double-buffered terminal canvas.
//
// Screen maintains two cell grids:
//
//   - cells: current frame
//   - prev: previously flushed frame
//
// Dirty region tracking allows only modified areas of the terminal
// to be repainted, minimizing ANSI output and significantly improving
// rendering performance.
//
// A Screen also manages:
//
//   - Terminal dimensions
//   - Raw mode lifecycle
//   - Cursor visibility
//   - Scrollback handling
//   - Incremental rendering
//
// Screen is not safe for concurrent use.

type Screen struct {
	height int
	width  int
	out    io.Writer

	cells [][]Cell
	prev  [][]Cell

	dirty bool

	oldState *term.State

	// Physical terminal viewport, queried at Start. Bounds what Flush
	// can address with absolute cursor moves.
	termRows int
	termCols int

	// Absolute terminal row where this Screen's row 0 lives. Starts at
	// 1; EnsureRoom decreases it (possibly past 1) as the program grows
	// and older rows are pushed into scrollback.
	anchorRow int

	// Dirty region tracking. dirtyRows[y] / dirtyCols[x] are set true
	// when any cell in that row/col changes. min/maxDirty* form a
	// bounding rectangle so Flush only scans the region that changed;
	// -1 means "nothing dirty yet this frame".
	dirtyRows   []bool
	dirtyCols   []bool
	minDirtyRow int
	maxDirtyRow int
	minDirtyCol int
	maxDirtyCol int
}

// Cell represents a single terminal cell.
//
// A cell contains:
//
//   - Rune: character to render
//   - Style: foreground/background attributes
//   - Wide: whether the rune occupies two columns
//
// Wide is used to properly render East Asian and emoji characters.
type Cell struct {
	Rune  rune
	Style Style
	Wide  bool
}

// GetCell returns the cell at the specified coordinates.
//
// Parameters:
//
//   - x: column index
//   - y: row index
//
// Example:
//
//	cell := screen.GetCell(10, 5)
//	fmt.Println(string(cell.Rune))
func (s *Screen) GetCell(x int, y int) Cell {
	return s.cells[x][y]
}

func makeCellGrid(width, height int) [][]Cell {
	grid := make([][]Cell, width)
	for i := range grid {
		grid[i] = make([]Cell, height)
	}
	return grid
}

// NewScreenWriter creates a new Screen that renders to the provided
// io.Writer.
//
// Parameters:
//
//   - width: screen width in cells
//   - height: screen height in cells
//   - out: destination writer
//
// Example:
//
//	screen := retui.NewScreenWriter(
//	    80,
//	    24,
//	    os.Stdout,
//	)
func NewScreenWriter(width int, height int, out io.Writer) *Screen {
	s := &Screen{
		height: height,
		width:  width,
		out:    out,
		cells:  makeCellGrid(width, height),
		prev:   makeCellGrid(width, height),
	}
	s.resetDirtyTracking()
	return s
}

// resetDirtyTracking (re)allocates dirty slices for the current
// width/height and resets the bounding rectangle to "nothing dirty".
// Must be called whenever width/height changes and at construction.
func (s *Screen) resetDirtyTracking() {
	s.dirtyRows = make([]bool, s.height)
	s.dirtyCols = make([]bool, s.width)
	s.minDirtyRow = -1
	s.maxDirtyRow = -1
	s.minDirtyCol = -1
	s.maxDirtyCol = -1
}

// Width returns the current screen width in cells.
//
// Example:
//
//	w := screen.Width()
func (s Screen) Width() int { return s.width }

// Height returns the current screen height in cells.
//
// Example:
//
//	h := screen.Height()
func (s Screen) Height() int { return s.height }

// Resize changes the screen dimensions and recreates all cell buffers.
//
// Existing contents are discarded.
//
// Example:
//
//	screen.Resize(120, 40)
func (s *Screen) Resize(width int, height int) {
	s.width = width
	s.height = height
	s.cells = makeCellGrid(width, height)
	s.prev = makeCellGrid(width, height)
	s.resetDirtyTracking()
	s.dirty = true
}

// Start initializes terminal rendering.
//
// It:
//
//   - Clears the terminal
//   - Enables raw mode
//   - Queries terminal dimensions
//   - Hides the cursor
//   - Enables bracketed paste mode
//
// Example:
//
//	screen.Start()
//	defer screen.Stop()
func (s *Screen) Start() {
	fmt.Print("\033[H\033[2J\033[3J")

	oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
	s.oldState = oldState

	if cols, rows, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		s.termCols = cols
		s.termRows = rows
	}
	s.anchorRow = 1

	fmt.Fprintf(s.out, "\033[?25l")   // hide cursor
	fmt.Fprintf(s.out, "\033[?2004h") // bracketed paste on
}

// Stop restores the terminal to its previous state.
//
// It:
//
//   - Shows the cursor
//   - Disables bracketed paste
//   - Restores terminal settings
//
// Example:
//
//	defer screen.Stop()
func (s Screen) Stop() {

	//Add nil check for oldState
	if s.oldState == nil {
		return
	}

	fmt.Fprintf(s.out, "\033[?2004l") // bracketed paste off
	fmt.Fprintf(s.out, "\033[?25h")   // show cursor
	term.Restore(int(os.Stdin.Fd()), s.oldState)

}

// SetCell writes value/style into cell (x, y). It diffs against the
// current cell value — if nothing changed, it returns immediately
// without marking dirty. This is the primary optimization: since
// renderer.go calls SetCell for every visible cell every frame, the
// diff here keeps the dirty rectangle tight (only genuinely changed
// cells) so Flush emits minimal ANSI output.
func (s *Screen) SetCell(x int, y int, value rune, style Style) {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return
	}

	wide := runewidth.RuneWidth(value) == 2
	newCell := Cell{Rune: value, Style: style, Wide: wide}

	if s.cells[x][y] == newCell {
		return // nothing changed — skip dirty mark entirely
	}

	s.cells[x][y] = newCell
	s.markDirty(x, y)

	if wide && x+1 < s.width {
		// Right half of a wide glyph: blank it so a stale narrow
		// character from a previous frame cannot show through.
		neighbor := Cell{Rune: 0, Style: style, Wide: true}
		if s.cells[x+1][y] != neighbor {
			s.cells[x+1][y] = neighbor
			s.markDirty(x+1, y)
		}
	}
}

// markDirty flags (x, y)'s row and column and expands the bounding
// rectangle. Callers must have already bounds-checked x and y.
func (s *Screen) markDirty(x, y int) {
	s.dirty = true
	s.dirtyRows[y] = true
	s.dirtyCols[x] = true

	if s.minDirtyRow == -1 || y < s.minDirtyRow {
		s.minDirtyRow = y
	}
	if y > s.maxDirtyRow {
		s.maxDirtyRow = y
	}
	if s.minDirtyCol == -1 || x < s.minDirtyCol {
		s.minDirtyCol = x
	}
	if x > s.maxDirtyCol {
		s.maxDirtyCol = x
	}
}

// Flush writes only the cells that changed since the last Flush to the
// terminal. Three levels of skipping keep output minimal:
//
//  1. Dirty bounding rectangle — only rows/cols within min/maxDirty*
//     are scanned at all.
//  2. dirtyRows[y] / dirtyCols[x] — whole rows and columns with no
//     changes are skipped without examining individual cells.
//  3. curr == prev — cells within the dirty region that were written
//     but ended up with the same value are skipped (e.g. a component
//     re-rendered an unchanged character).
//
// All output is buffered into a single 16 KB bufio.Writer and flushed
// once at the end, reducing terminal writes to a single syscall per
// frame regardless of how many cells changed.
func (s *Screen) Flush() {
	if !s.dirty {
		return
	}

	startRow := s.minDirtyRow
	endRow := s.maxDirtyRow
	startCol := s.minDirtyCol
	endCol := s.maxDirtyCol

	// If dirty but no region tracked, flush the whole screen
	if startRow == -1 && s.dirty {
		startRow = 0
		endRow = s.height - 1
		startCol = 0
		endCol = s.width - 1
	}

	if startRow < 0 {
		startRow = 0
	}
	if endRow >= s.height {
		endRow = s.height - 1
	}
	if startCol < 0 {
		startCol = 0
	}
	if endCol >= s.width {
		endCol = s.width - 1
	}

	buf := bufio.NewWriterSize(s.out, 16384)

	cursorRow, cursorCol := -1, -1
	currentStyle := Style{}
	styleActive := false

	for y := startRow; y <= endRow; y++ {
		if !s.dirtyRows[y] {
			continue
		}

		absRow := s.anchorRow + y
		if absRow < 1 || absRow > s.termRows {
			continue
		}

		for x := startCol; x <= endCol; x++ {
			if !s.dirtyCols[x] {
				continue
			}

			curr := s.cells[x][y]
			prev := s.prev[x][y]

			if curr == prev {
				continue
			}

			if cursorRow != absRow || cursorCol != x {
				fmt.Fprintf(buf, "\033[%d;%dH", absRow, x+1)
				cursorRow, cursorCol = absRow, x
			}

			if !styleActive || curr.Style != currentStyle {
				fmt.Fprint(buf, curr.Style.ANSIPrefix())
				currentStyle = curr.Style
				styleActive = true
			}

			char := curr.Rune
			if char == 0 {
				char = ' '
			}
			fmt.Fprint(buf, string(char))
			cursorCol += runewidth.RuneWidth(char)
			if cursorCol == x {
				// Zero-width rune: advance so the next cell's
				// reposition check doesn't stall.
				cursorCol = x + 1
			}

			s.prev[x][y] = curr
		}
	}

	if styleActive {
		fmt.Fprint(buf, "\033[0m")
	}

	buf.Flush() // single syscall for the entire frame
	s.clearDirty()
}

// clearDirty resets all dirty state after a Flush without reallocating
// the tracking slices.
func (s *Screen) clearDirty() {
	s.dirty = false
	s.minDirtyRow = -1
	s.maxDirtyRow = -1
	s.minDirtyCol = -1
	s.maxDirtyCol = -1
	for i := range s.dirtyRows {
		s.dirtyRows[i] = false
	}
	for i := range s.dirtyCols {
		s.dirtyCols[i] = false
	}
}

// ForceMarkAllDirty marks the entire screen as dirty, forcing a full redraw.
// This is useful after resize operations or when the screen content needs
// a complete refresh.
func (s *Screen) ForceMarkAllDirty() {
	if s == nil {
		return
	}
	s.dirty = true
	for y := 0; y < s.height; y++ {
		s.dirtyRows[y] = true
	}
	for x := 0; x < s.width; x++ {
		s.dirtyCols[x] = true
	}
	s.minDirtyRow = 0
	s.maxDirtyRow = s.height - 1
	s.minDirtyCol = 0
	s.maxDirtyCol = s.width - 1
}

// EnsureRoom guarantees that contentH rows fit within the physical
// terminal viewport. When the content would overflow, it writes the
// rows inline so the terminal scrolls older content into scrollback.
// All output is buffered and flushed in one syscall with style batching
// (one ANSI prefix per style run, not per cell).
func (s *Screen) EnsureRoom(contentH int) {
	if s.termRows == 0 {
		return
	}
	bottom := s.anchorRow + contentH - 1
	if bottom <= s.termRows {
		return
	}
	delta := bottom - s.termRows

	topRow := s.anchorRow
	if topRow < 1 {
		topRow = 1
	}

	buf := bufio.NewWriterSize(s.out, 16384)
	fmt.Fprintf(buf, "\033[%d;1H", topRow)

	startY := 0
	if s.anchorRow < 1 {
		startY = 1 - s.anchorRow
	}

	for y := startY; y < contentH; y++ {
		currentStyle := Style{}
		styleActive := false

		for x := 0; x < s.width; x++ {
			cell := s.cells[x][y]

			if !styleActive || cell.Style != currentStyle {
				if styleActive {
					fmt.Fprint(buf, "\033[0m")
				}
				fmt.Fprint(buf, cell.Style.ANSIPrefix())
				currentStyle = cell.Style
				styleActive = true
			}

			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			fmt.Fprint(buf, string(r))
		}

		if styleActive {
			fmt.Fprint(buf, "\033[0m")
		}

		if y < contentH-1 {
			fmt.Fprint(buf, "\r\n")
		}
	}

	buf.Flush()

	// Sync prev so Flush won't re-emit these rows.
	for y := startY; y < contentH; y++ {
		for x := 0; x < s.width; x++ {
			s.prev[x][y] = s.cells[x][y]
		}
	}

	s.anchorRow -= delta

	// Clear dirty only for rows we just wrote; rows outside contentH
	// may still have pending dirty state from this frame.
	for y := startY; y < contentH && y < len(s.dirtyRows); y++ {
		s.dirtyRows[y] = false
	}
	s.recomputeDirtyBounds()
}

// recomputeDirtyBounds rescans dirtyRows/dirtyCols to rebuild the
// bounding rectangle after a partial row clear (used by EnsureRoom).
func (s *Screen) recomputeDirtyBounds() {
	s.minDirtyRow, s.maxDirtyRow = -1, -1
	for y, d := range s.dirtyRows {
		if !d {
			continue
		}
		if s.minDirtyRow == -1 {
			s.minDirtyRow = y
		}
		s.maxDirtyRow = y
	}

	s.minDirtyCol, s.maxDirtyCol = -1, -1
	for x, d := range s.dirtyCols {
		if !d {
			continue
		}
		if s.minDirtyCol == -1 {
			s.minDirtyCol = x
		}
		s.maxDirtyCol = x
	}

	s.dirty = s.minDirtyRow != -1 && s.minDirtyCol != -1
}

// Clear resets every cell to a blank space. Only cells whose content
// actually differs from blank are marked dirty — since Render calls
// Clear unconditionally every frame, marking every cell dirty
// regardless would defeat dirty-region tracking entirely.
func (s *Screen) Clear() {
	blank := Cell{Rune: ' '}
	for x := range s.width {
		for y := range s.height {
			if s.cells[x][y] == blank {
				continue
			}
			s.cells[x][y] = blank
			s.markDirty(x, y)
		}
	}
}

func RuneWidth(value rune) int {
	return runewidth.RuneWidth(value)
}

// HandleResize is called on SIGWINCH. Updates termRows/termCols and
// resets the cell grids so the next paint targets the new dimensions.
func (s *Screen) HandleResize() {
	fmt.Print("\033[H\033[2J\033[3J")
	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || cols <= 0 || rows <= 0 {
		return
	}
	s.termCols = cols
	s.termRows = rows
	s.SetDimensions(cols, rows) // reallocates cells+prev, resets dirty tracking
	s.dirty = true
	s.anchorRow = 1
	fmt.Fprint(s.out, "\033[H\033[2J")
}

func (s *Screen) SetDimensions(width, height int) {
	s.width = width
	s.height = height
	s.cells = makeCellGrid(width, height)
	s.prev = makeCellGrid(width, height)
	s.resetDirtyTracking()
}

// StdOutScreen is a package-level Screen writing to os.Stdout.
// Always constructed via NewScreenWriter — never use a bare &Screen{}
// literal, as that leaves dirtyRows/dirtyCols nil and SetCell will panic.
var StdOutScreen *Screen = func() *Screen {
	width, height := 80, 24
	if cols, rows, err := term.GetSize(int(os.Stdout.Fd())); err == nil && cols > 0 && rows > 0 {
		width, height = cols, rows
	}
	return NewScreenWriter(width, height, os.Stdout)
}()

// --Overlay Modal
func (s *Screen) PaintOverlay() {}
