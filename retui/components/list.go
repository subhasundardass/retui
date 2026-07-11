package components

import "github.com/subhasundardass/retui/retui"

func List(items []string, focused bool) retui.Element {
	selected, setSelected := retui.UseState(0)

	if focused {
		if retui.CurrentKey.Code == retui.KeyDown && selected < len(items)-1 {
			setSelected(selected + 1)
		}
		if retui.CurrentKey.Code == retui.KeyUp && selected > 0 {
			setSelected(selected - 1)
		}
	}

	children := make([]retui.Element, len(items))
	for i, item := range items {
		prefix := "  "
		var style retui.Style
		if i == selected {
			prefix = "> "
			if focused {
				style = retui.NewStyle().
					Background(retui.Blue).
					Foreground(retui.Cyan).
					Bold(true)
			} else {
				style = retui.NewStyle().Foreground(retui.White).Bold(true)
			}
		} else {
			style = retui.NewStyle().Foreground(retui.BrightBlack)
		}
		children[i] = retui.Text(prefix+item, style)
	}
	return retui.Box(
		retui.Props{Direction: retui.Column},
		retui.NewStyle(),
		children...)
}
