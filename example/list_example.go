package example

import (
	"github.com/subhasundardass/retui/retui"
	"github.com/subhasundardass/retui/retui/components"
)

func ListExample(props retui.Props) retui.Element {
	selectedIndex, setSelectedIndex := retui.UseState(0)
	// retui.SetFocus("content")

	list := components.List().
		ID("fruit-list").
		Items([]string{"🍎 Apple", "🍌 Banana", "🍊 Orange", "🍇 Grape", "🥭 Mango"}).
		Selected(selectedIndex).
		Width(30).
		OnSelect(func(id string, index int, value string) {
			// println("Selected:", value)
			setSelectedIndex(index)
		}).
		Focused(true).
		Render()

	return retui.Box(
		retui.Props{},
		retui.NewStyle(),
		components.Panel().
			Header(retui.Text("List Example", retui.NewStyle())).
			Width(retui.Fixed(100)).
			Children(
				retui.Box(
					retui.Props{},
					retui.NewStyle(),
					list,
				),
			).
			Render(),
	)
}
