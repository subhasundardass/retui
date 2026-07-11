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
		contentRows = append(contentRows, retui.Box(
			retui.Props{
				Direction: retui.Row,
				Width:     retui.Fixed(width),
			},
			retui.NewStyle(),
			retui.Text("│", borderStyle),
			retui.Box(
				retui.Props{
					Width:   retui.Grow(1),
					Padding: [4]int{0, 1, 0, 1},
				},
				retui.NewStyle(),
				child,
			),
			retui.Text("│", borderStyle),
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
