package example

import (
	"github.com/subhasundardass/retui/retui"
	"github.com/subhasundardass/retui/retui/components"
)

func BasicInputExample(props retui.Props) retui.Element {
	return retui.Box(
		props,
		retui.NewStyle(),
		components.Panel(
			"Basic Inputs", 100,

			retui.Text("Hee", retui.NewStyle().Foreground(retui.BrightCyan).Bold(true)),
			retui.Text("Hee", retui.NewStyle().Foreground(retui.BrightCyan).Bold(true)),
			retui.Text("Hee", retui.NewStyle().Foreground(retui.BrightCyan).Bold(true)),
			retui.Text("Hee", retui.NewStyle().Foreground(retui.BrightCyan).Bold(true)),
		),
	)
}
