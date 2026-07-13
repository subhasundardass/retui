package example

import (
	"github.com/subhasundardass/retui/retui"
	"github.com/subhasundardass/retui/retui/components"
)

func WindowsExample(props retui.Props) retui.Element {
	return retui.Box(
		props,
		retui.NewStyle(),
		components.Panel(
			"Windows", 100,

			retui.Text("Window 1", retui.NewStyle().Foreground(retui.BrightCyan).Bold(true)),
			retui.Text("Window 2", retui.NewStyle().Foreground(retui.BrightCyan).Bold(true)),
			retui.Text("Window 3", retui.NewStyle().Foreground(retui.BrightCyan).Bold(true)),
			retui.Text("Modal Window", retui.NewStyle().Foreground(retui.BrightCyan).Bold(true)),
		),
	)
}
