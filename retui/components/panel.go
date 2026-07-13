package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

func Panel(title string, width int, children ...retui.Element) retui.Element {
	borderColor := retui.Hex("#535353")
	borderStyle := retui.NewStyle().Foreground(borderColor)
	titleStyle := retui.NewStyle().Foreground(retui.Hex("#c7c7c7")).Bold(true)

	innerWidth := width - 2

	// Top border (without title)
	topBorder := "┌" + strings.Repeat("─", innerWidth) + "┐"

	// Separator with proper junctions
	separator := "├" + strings.Repeat("─", innerWidth) + "┤"

	// Bottom border
	bottomBorder := "└" + strings.Repeat("─", innerWidth) + "┘"

	// Build content with side borders
	contentRows := []retui.Element{}
	for _, child := range children {
		lineCount := measureHeight(child)

		contentRows = append(contentRows, retui.Box(
			retui.Props{
				Direction: retui.Row,
				Width:     retui.Fixed(width),
			},
			retui.NewStyle(),
			verticalBorder(lineCount, borderStyle),
			retui.Box(
				retui.Props{
					Width:   retui.Grow(1),
					Padding: [4]int{0, 1, 0, 1},
				},
				retui.NewStyle(),
				child,
			),
			verticalBorder(lineCount, borderStyle),
		))
	}

	return retui.Box(
		retui.Props{
			Direction: retui.Column,
			Width:     retui.Fixed(width),
			Gap:       0,
		},
		retui.NewStyle(),
		// Top border
		retui.Text(topBorder, borderStyle),
		// Title section
		retui.Box(
			retui.Props{
				Direction: retui.Row,
				Width:     retui.Fixed(width),
				Padding:   [4]int{0, 0, 0, 0},
			},
			retui.NewStyle(),
			retui.Text("│ ", borderStyle),
			retui.Box(
				retui.Props{
					Width: retui.Grow(1),
				},
				retui.NewStyle(),
				retui.Text(title, titleStyle),
			),
			retui.Text(" │", borderStyle),
		),
		// Separator line under title
		retui.Text(separator, borderStyle),
		// Content with side borders
		retui.Box(
			retui.Props{
				Direction: retui.Column,
				Width:     retui.Fixed(width),
				Gap:       0,
			},
			retui.NewStyle(),
			contentRows...,
		),
		// Bottom border
		retui.Text(bottomBorder, borderStyle),
	)
}

// verticalBorder returns a Column of n stacked single-line "│" Text
// elements — the same stacking pattern List uses for its items — instead
// of one Text with embedded newlines, since a single Text element does
// not render as multiple independent lines in this framework.
func verticalBorder(n int, style retui.Style) retui.Element {
	if n < 1 {
		n = 1
	}
	pipes := make([]retui.Element, n)
	for i := range pipes {
		pipes[i] = retui.Text("│", style)
	}
	return retui.Box(
		retui.Props{
			Direction: retui.Column,
		},
		retui.NewStyle(),
		pipes...,
	)
}

// measureHeight returns how many lines an Element will render as.
func measureHeight(el retui.Element) int {
	switch el.Type {
	case retui.ElementText:
		if el.Text == "" {
			return 1
		}
		return strings.Count(el.Text, "\n") + 1
	case retui.ElementBox:
		return measureBoxHeight(el)
	default:
		return 1
	}
}

func measureBoxHeight(el retui.Element) int {
	pad := el.Layout.PaddingTop + el.Layout.PaddingBottom
	if len(el.Children) == 0 {
		return 1 + pad
	}
	if el.Layout.Direction == retui.Row {
		max := 0
		for _, c := range el.Children {
			if h := measureHeight(c); h > max {
				max = h
			}
		}
		return max + pad
	}
	total := el.Layout.Gap * (len(el.Children) - 1)
	for _, c := range el.Children {
		total += measureHeight(c)
	}
	return total + pad
}
