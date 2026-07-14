package components

import "github.com/subhasundardass/retui/retui"

// Badge renders a short colored label with padding.
func Badge(label string, fg retui.Color, bg retui.Color) retui.Element {
	return retui.Text(
		" "+label+" ",
		retui.Style{}.Foreground(fg).Background(bg).Bold(true),
	)
}
