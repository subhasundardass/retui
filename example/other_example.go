package example

import (
	"github.com/subhasundardass/retui/retui"
	"github.com/subhasundardass/retui/retui/components"
)

// OtherExample demonstrates Badge, Spinner, and ProgressBar rendered
// together: a row of status badges, a loading spinner, and a stack of
// progress bars at different completion levels and colors.
func OtherExample(props retui.Props) retui.Element {
	return retui.Box(
		props,
		retui.NewStyle(),
		components.Panel().
			Header(retui.Text("Other Example", retui.NewStyle())).
			Width(retui.Fixed(100)).
			Children(
				// Status badges
				retui.Box(
					retui.Props{Direction: retui.Row, Gap: 1},
					retui.NewStyle(),
					components.Badge("ONLINE", retui.White, retui.Green),
					components.Badge("WARNING", retui.Black, retui.Yellow),
					components.Badge("ERROR", retui.White, retui.Red),
				),

				retui.Text(" ", retui.NewStyle()),

				// Loading spinner
				components.Spinner("Loading data..."),

				retui.Text(" ", retui.NewStyle()),

				// Progress bars at a few different completion levels
				retui.Box(
					retui.Props{Direction: retui.Column, Gap: 1},
					retui.NewStyle(),
					components.ProgressBar(0.35, 30, retui.Cyan),
					components.ProgressBar(0.72, 30, retui.Green),
					components.ProgressBar(1.0, 30, retui.Yellow),
				),
			).
			Render(),
	)
}
