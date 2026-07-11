package retui

import (
	"strconv"
	"strings"
	"unicode"
)

type markdownBlockKind int

const (
	markdownParagraph markdownBlockKind = iota
	markdownHeading
	markdownQuote
	markdownListItem
	markdownRule
	markdownCode
	markdownTable
)

type markdownBlock struct {
	kind    markdownBlockKind
	text    string
	level   int
	depth   int
	ordered bool
	number  int
	task    int
	lines   []string
	headers []string
	rows    [][]string
}

type markdownSpan struct {
	text          string
	style         Style
	literal       bool
	strikethrough bool
}

type markdownCell struct {
	r     rune
	style Style
}

type markdownLine []markdownCell

func renderMarkdownLines(
	markdown string,
	width int,
	base Style,
) []markdownLine {
	if width <= 0 {
		return []markdownLine{{}}
	}

	blocks := parseMarkdownBlocks(markdown)
	lines := make([]markdownLine, 0)
	for i, block := range blocks {
		if i > 0 && needsMarkdownBlankLine(blocks[i-1], block) {
			lines = append(lines, markdownLine{})
		}
		lines = append(lines, renderMarkdownBlock(block, width, base)...)
	}
	if len(lines) == 0 {
		return []markdownLine{{}}
	}
	return lines
}

func needsMarkdownBlankLine(prev, next markdownBlock) bool {
	if prev.kind == markdownListItem && next.kind == markdownListItem {
		return false
	}
	if prev.kind == markdownQuote && next.kind == markdownQuote {
		return false
	}
	return true
}

func parseMarkdownBlocks(markdown string) []markdownBlock {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	markdown = strings.ReplaceAll(markdown, "\r", "\n")
	input := strings.Split(markdown, "\n")
	blocks := make([]markdownBlock, 0)
	lastWasList := false

	for i := 0; i < len(input); {
		line := input[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			lastWasList = false
			i++
			continue
		}

		if strings.HasPrefix(trimmed, "```") {
			code := make([]string, 0)
			fenceLen := markdownFenceLength(trimmed)
			i++
			foundClosing := false
			for i < len(input) {
				line := strings.TrimSpace(input[i])
				if isMarkdownClosingFence(line, fenceLen) {
					foundClosing = true
					i++
					break
				}
				code = append(code, input[i])
				i++
			}
			// If no closing fence was found, render what we have as code
			// but don't consume more input - treat remaining as paragraphs
			if !foundClosing {
				blocks = append(
					blocks,
					markdownBlock{kind: markdownCode, lines: code},
				)
				// Continue processing remaining lines normally
				continue
			}
			blocks = append(
				blocks,
				markdownBlock{kind: markdownCode, lines: code},
			)
			continue
		}

		if isMarkdownRule(trimmed) {
			blocks = append(blocks, markdownBlock{kind: markdownRule})
			i++
			continue
		}

		if level, text, ok := parseMarkdownHeading(trimmed); ok {
			blocks = append(
				blocks,
				markdownBlock{kind: markdownHeading, level: level, text: text},
			)
			i++
			continue
		}

		if isMarkdownTableStart(input, i) {
			headers := splitMarkdownTableRow(input[i])
			i += 2
			rows := make([][]string, 0)
			for i < len(input) && strings.Contains(input[i], "|") && strings.TrimSpace(input[i]) != "" {
				rows = append(rows, splitMarkdownTableRow(input[i]))
				i++
			}
			blocks = append(
				blocks,
				markdownBlock{
					kind:    markdownTable,
					headers: headers,
					rows:    rows,
				},
			)
			continue
		}

		if text, ok := parseMarkdownQuote(trimmed); ok {
			parts := []string{text}
			i++
			for i < len(input) {
				if next, ok := parseMarkdownQuote(strings.TrimSpace(input[i])); ok {
					parts = append(parts, next)
					i++
					continue
				}
				break
			}
			blocks = append(
				blocks,
				markdownBlock{
					kind: markdownQuote,
					text: strings.Join(parts, " "),
				},
			)
			continue
		}

		if item, ok := parseMarkdownListItem(line, lastWasList); ok {
			blocks = append(blocks, item)
			lastWasList = true
			i++
			continue
		}
		lastWasList = false

		parts := []string{trimmed}
		i++
		for i < len(input) {
			next := strings.TrimSpace(input[i])
			if next == "" || strings.HasPrefix(next, "```") ||
				isMarkdownRule(next) {
				break
			}
			if _, _, ok := parseMarkdownHeading(next); ok {
				break
			}
			if isMarkdownTableStart(input, i) {
				break
			}
			if _, ok := parseMarkdownQuote(next); ok {
				break
			}
			if _, ok := parseMarkdownListItem(input[i], false); ok {
				break
			}
			parts = append(parts, next)
			i++
		}
		blocks = append(
			blocks,
			markdownBlock{
				kind: markdownParagraph,
				text: strings.Join(parts, " "),
			},
		)
	}

	return blocks
}

func parseMarkdownHeading(line string) (int, string, bool) {
	level := 0
	for level < len(line) && level < 6 && line[level] == '#' {
		level++
	}
	if level == 0 || level >= len(line) || line[level] != ' ' {
		return 0, "", false
	}
	return level, strings.TrimSpace(line[level+1:]), true
}

func markdownFenceLength(line string) int {
	count := 0
	for count < len(line) && line[count] == '`' {
		count++
	}
	return count
}

func isMarkdownClosingFence(line string, openingLen int) bool {
	if openingLen < 3 || len(line) < openingLen {
		return false
	}
	for i := 0; i < openingLen; i++ {
		if line[i] != '`' {
			return false
		}
	}
	return strings.TrimSpace(line[openingLen:]) == ""
}

func isMarkdownRule(line string) bool {
	if len(line) < 3 {
		return false
	}
	var marker rune
	count := 0
	for _, r := range line {
		if unicode.IsSpace(r) {
			continue
		}
		if marker == 0 {
			if r != '-' && r != '*' && r != '_' {
				return false
			}
			marker = r
		}
		if r != marker {
			return false
		}
		count++
	}
	return count >= 3
}

func parseMarkdownQuote(line string) (string, bool) {
	if !strings.HasPrefix(line, ">") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(line, ">")), true
}

func parseMarkdownListItem(line string, inList bool) (markdownBlock, bool) {
	indent := 0
	for indent < len(line) && line[indent] == ' ' {
		indent++
	}
	if !inList && indent > 3 {
		return markdownBlock{}, false
	}
	line = line[indent:]

	if len(line) >= 2 && (line[0] == '-' || line[0] == '*' || line[0] == '+') &&
		line[1] == ' ' {
		text := strings.TrimSpace(line[2:])
		task, rest := parseMarkdownTask(text)
		return markdownBlock{
			kind:  markdownListItem,
			text:  rest,
			task:  task,
			depth: indent,
		}, true
	}

	dot := 0
	for dot < len(line) && line[dot] >= '0' && line[dot] <= '9' {
		dot++
	}
	if dot == 0 || dot > 9 || dot+1 >= len(line) || line[dot] != '.' ||
		line[dot+1] != ' ' {
		return markdownBlock{}, false
	}
	n, err := strconv.Atoi(line[:dot])
	if err != nil || n < 0 {
		return markdownBlock{}, false
	}
	text := strings.TrimSpace(line[dot+2:])
	task, rest := parseMarkdownTask(text)
	return markdownBlock{
		kind:    markdownListItem,
		text:    rest,
		ordered: true,
		number:  n,
		task:    task,
		depth:   indent,
	}, true
}

func parseMarkdownTask(text string) (int, string) {
	if len(text) >= 4 && text[0] == '[' && text[2] == ']' && text[3] == ' ' {
		switch text[1] {
		case ' ':
			return 0, strings.TrimSpace(text[4:])
		case 'x', 'X':
			return 1, strings.TrimSpace(text[4:])
		}
	}
	return -1, text
}

func isMarkdownTableStart(lines []string, i int) bool {
	if i+1 >= len(lines) || !strings.Contains(lines[i], "|") {
		return false
	}
	cells := splitMarkdownTableRow(lines[i])
	if len(cells) == 0 {
		return false
	}
	separator := splitMarkdownTableRow(lines[i+1])
	if len(separator) == 0 {
		return false
	}
	for _, cell := range separator {
		cell = strings.TrimSpace(cell)
		cell = strings.Trim(cell, ":")
		if len(cell) < 3 {
			return false
		}
		for _, r := range cell {
			if r != '-' {
				return false
			}
		}
	}
	return true
}

func splitMarkdownTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, len(parts))
	for i, part := range parts {
		cells[i] = strings.TrimSpace(part)
	}
	return cells
}

func renderMarkdownBlock(
	block markdownBlock,
	width int,
	base Style,
) []markdownLine {
	switch block.kind {
	case markdownHeading:
		style := base.Bold(true).Foreground(Cyan)
		return wrapMarkdownSpans(
			parseMarkdownInline(block.text, style),
			"",
			"",
			width,
		)
	case markdownParagraph:
		return wrapMarkdownSpans(
			parseMarkdownInline(block.text, base),
			"",
			"",
			width,
		)
	case markdownQuote:
		prefix := "│ "
		style := base.Foreground(BrightBlack).Italic(true)
		return wrapMarkdownSpans(
			parseMarkdownInline(block.text, style),
			prefix,
			"│ ",
			width,
		)
	case markdownListItem:
		prefix := strings.Repeat(" ", block.depth) + markdownListPrefix(block)
		return wrapMarkdownSpans(
			parseMarkdownInline(block.text, base),
			prefix,
			strings.Repeat(" ", runeWidthString(prefix)),
			width,
		)
	case markdownRule:
		return []markdownLine{
			markdownTextLine(
				strings.Repeat("─", max(width, 1)),
				base.Foreground(BrightBlack),
			),
		}
	case markdownCode:
		return renderMarkdownCode(block.lines, width, base)
	case markdownTable:
		return renderMarkdownTable(block, width, base)
	default:
		return []markdownLine{{}}
	}
}

func markdownListPrefix(block markdownBlock) string {
	prefix := "• "
	if block.ordered {
		prefix = strconv.Itoa(block.number) + ". "
	}
	if block.task == 0 {
		prefix += "[ ] "
	} else if block.task == 1 {
		prefix += "[x] "
	}
	return prefix
}

func renderMarkdownCode(lines []string, width int, base Style) []markdownLine {
	style := base.Foreground(Blue)
	out := make([]markdownLine, 0, max(len(lines), 1))
	if len(lines) == 0 {
		return []markdownLine{{}}
	}
	for _, line := range lines {
		out = append(
			out,
			markdownTextLine(truncateMarkdownText("  "+line, width), style),
		)
	}
	return out
}

func renderMarkdownTable(
	block markdownBlock,
	width int,
	base Style,
) []markdownLine {
	cols := len(block.headers)
	for _, row := range block.rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return []markdownLine{{}}
	}

	widths := make([]int, cols)
	for i, cell := range block.headers {
		widths[i] = max(widths[i], runeWidthString(stripMarkdownInline(cell)))
	}
	for _, row := range block.rows {
		for i, cell := range row {
			widths[i] = max(
				widths[i],
				runeWidthString(stripMarkdownInline(cell)),
			)
		}
	}

	out := make([]markdownLine, 0, len(block.rows)+2)
	out = append(
		out,
		renderMarkdownTableRow(
			block.headers,
			widths,
			width,
			base.Bold(true).Foreground(Cyan),
		),
	)
	out = append(
		out,
		markdownTextLine(
			truncateMarkdownText(markdownTableSeparator(widths), width),
			base.Foreground(BrightBlack),
		),
	)
	for _, row := range block.rows {
		out = append(out, renderMarkdownTableRow(row, widths, width, base))
	}
	return out
}

func renderMarkdownTableRow(
	cells []string,
	widths []int,
	width int,
	style Style,
) markdownLine {
	var b strings.Builder
	b.WriteString("│")
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = stripMarkdownInline(cells[i])
		}
		b.WriteString(" ")
		b.WriteString(cell)
		b.WriteString(strings.Repeat(" ", max(w-runeWidthString(cell), 0)))
		b.WriteString(" │")
	}
	return markdownTextLine(truncateMarkdownText(b.String(), width), style)
}

func markdownTableSeparator(widths []int) string {
	var b strings.Builder
	b.WriteString("├")
	for i, w := range widths {
		if i > 0 {
			b.WriteString("┼")
		}
		b.WriteString(strings.Repeat("─", w+2))
	}
	b.WriteString("┤")
	return b.String()
}

func parseMarkdownInline(text string, base Style) []markdownSpan {
	spans := make([]markdownSpan, 0)
	for len(text) > 0 {
		if strings.HasPrefix(text, "`") {
			if end := strings.Index(text[1:], "`"); end >= 0 {
				content := text[1 : end+1]
				spans = append(
					spans,
					markdownSpan{
						text:    content,
						style:   base.Foreground(Blue),
						literal: true,
					},
				)
				text = text[end+2:]
				continue
			}
		}
		if strings.HasPrefix(text, "~~") {
			if end := strings.Index(text[2:], "~~"); end >= 0 {
				content := text[2 : end+2]
				spans = append(
					spans,
					markdownSpan{
						text:          content,
						style:         base.Foreground(BrightBlack),
						strikethrough: true,
					},
				)
				text = text[end+4:]
				continue
			}
		}
		if strings.HasPrefix(text, "**") {
			if end := strings.Index(text[2:], "**"); end >= 0 {
				content := text[2 : end+2]
				spans = append(
					spans,
					markdownSpan{text: content, style: base.Bold(true)},
				)
				text = text[end+4:]
				continue
			}
		}
		if strings.HasPrefix(text, "__") {
			if end := strings.Index(text[2:], "__"); end >= 0 {
				content := text[2 : end+2]
				spans = append(
					spans,
					markdownSpan{text: content, style: base.Bold(true)},
				)
				text = text[end+4:]
				continue
			}
		}
		if strings.HasPrefix(text, "*") {
			if end := strings.Index(text[1:], "*"); end >= 0 {
				content := text[1 : end+1]
				spans = append(
					spans,
					markdownSpan{text: content, style: base.Italic(true)},
				)
				text = text[end+2:]
				continue
			}
		}
		if strings.HasPrefix(text, "_") {
			if end := strings.Index(text[1:], "_"); end >= 0 {
				content := text[1 : end+1]
				spans = append(
					spans,
					markdownSpan{text: content, style: base.Italic(true)},
				)
				text = text[end+2:]
				continue
			}
		}
		if strings.HasPrefix(text, "[") {
			if span, rest, ok := parseMarkdownLink(text, base); ok {
				spans = append(spans, span)
				text = rest
				continue
			}
		}

		next := nextMarkdownDelimiter(text)
		spans = append(spans, markdownSpan{text: text[:next], style: base})
		text = text[next:]
	}
	return spans
}

func parseMarkdownLink(text string, base Style) (markdownSpan, string, bool) {
	closeLabel := strings.Index(text, "]")
	if closeLabel <= 1 || closeLabel+1 >= len(text) ||
		text[closeLabel+1] != '(' {
		return markdownSpan{}, text, false
	}
	closeURL := strings.Index(text[closeLabel+2:], ")")
	if closeURL < 0 {
		return markdownSpan{}, text, false
	}
	label := text[1:closeLabel]
	url := text[closeLabel+2 : closeLabel+2+closeURL]
	rendered := label
	if strings.TrimSpace(url) != "" {
		rendered += " (" + url + ")"
	}
	return markdownSpan{
		text:  rendered,
		style: base.Foreground(Cyan).Underline(true),
	}, text[closeLabel+3+closeURL:], true
}

func nextMarkdownDelimiter(text string) int {
	best := len(text)
	for _, marker := range []string{"`", "~~", "**", "__", "*", "_", "["} {
		if idx := strings.Index(text[1:], marker); idx >= 0 && idx+1 < best {
			best = idx + 1
		}
	}
	if best == 0 {
		return len(text)
	}
	return best
}

func stripMarkdownInline(text string) string {
	spans := parseMarkdownInline(text, Style{})
	var b strings.Builder
	for _, span := range spans {
		b.WriteString(span.text)
	}
	return b.String()
}

func wrapMarkdownSpans(
	spans []markdownSpan,
	prefix, continuation string,
	width int,
) []markdownLine {
	if width <= 0 {
		return []markdownLine{{}}
	}

	prefixCells := markdownTextLine(prefix, Style{})
	continuationCells := markdownTextLine(continuation, Style{})
	line := append(markdownLine{}, prefixCells...)
	lineWidth := runeWidthMarkdownLine(line)
	limit := max(width, 1)
	out := make([]markdownLine, 0)
	hadContent := false

	for _, span := range spans {
		words := splitMarkdownWords(span)
		for _, word := range words {
			wordWidth := runeWidthMarkdownLine(word)
			if hadContent && lineWidth+wordWidth > limit &&
				lineWidth > runeWidthMarkdownLine(prefixCells) {
				out = append(out, trimMarkdownLineRight(line))
				line = append(markdownLine{}, continuationCells...)
				lineWidth = runeWidthMarkdownLine(line)
			}
			for wordWidth > limit-lineWidth && len(word) > 0 {
				remaining := limit - lineWidth
				if remaining <= 0 {
					out = append(out, trimMarkdownLineRight(line))
					line = append(markdownLine{}, continuationCells...)
					lineWidth = runeWidthMarkdownLine(line)
					remaining = limit - lineWidth
				}
				part, rest := splitMarkdownLineAtWidth(word, remaining)
				line = append(line, part...)
				out = append(out, trimMarkdownLineRight(line))
				line = append(markdownLine{}, continuationCells...)
				lineWidth = runeWidthMarkdownLine(line)
				word = rest
				wordWidth = runeWidthMarkdownLine(word)
				hadContent = true
			}
			line = append(line, word...)
			lineWidth += wordWidth
			if len(word) > 0 {
				hadContent = true
			}
		}
	}

	out = append(out, trimMarkdownLineRight(line))
	return out
}

func splitMarkdownWords(span markdownSpan) []markdownLine {
	words := make([]markdownLine, 0)
	var current markdownLine
	for _, r := range span.text {
		cell := markdownCell{r: r, style: span.style}
		if unicode.IsSpace(r) {
			if len(current) > 0 {
				words = append(words, current)
				current = nil
			}
			if len(words) == 0 || len(words[len(words)-1]) != 1 ||
				words[len(words)-1][0].r != ' ' {
				words = append(words, markdownLine{{r: ' ', style: span.style}})
			}
			continue
		}
		current = append(current, cell)
	}
	if len(current) > 0 {
		words = append(words, current)
	}
	return words
}

func splitMarkdownLineAtWidth(
	line markdownLine,
	width int,
) (markdownLine, markdownLine) {
	if width <= 0 {
		return nil, line
	}
	used := 0
	for i, cell := range line {
		w := RuneWidth(cell.r)
		if used+w > width {
			return line[:i], line[i:]
		}
		used += w
	}
	return line, nil
}

func trimMarkdownLineRight(line markdownLine) markdownLine {
	end := len(line)
	for end > 0 && line[end-1].r == ' ' {
		end--
	}
	return line[:end]
}

func markdownTextLine(text string, style Style) markdownLine {
	line := make(markdownLine, 0, len(text))
	for _, r := range text {
		line = append(line, markdownCell{r: r, style: style})
	}
	return line
}

func runeWidthMarkdownLine(line markdownLine) int {
	width := 0
	for _, cell := range line {
		width += RuneWidth(cell.r)
	}
	return width
}

func runeWidthString(text string) int {
	width := 0
	for _, r := range text {
		width += RuneWidth(r)
	}
	return width
}

func truncateMarkdownText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	used := 0
	var b strings.Builder
	for _, r := range text {
		w := RuneWidth(r)
		if used+w > width {
			break
		}
		b.WriteRune(r)
		used += w
	}
	return b.String()
}
