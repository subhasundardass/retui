package example

import (
	"strconv"

	"github.com/subhasundardass/retui/retui"
	"github.com/subhasundardass/retui/retui/components"
)

// CounterExample demonstrates basic state management: a value held in
// UseState, changed by a left (-) button and a right (+) button. Pressing
// the "+" or "-" key drives the same increment/decrement.
func CounterExample(props retui.Props) retui.Element {
	count, setCount := retui.UseState(0)

	switch retui.CurrentKey.Rune {
	case '-':
		setCount(count - 1)
	case '+':
		setCount(count + 1)
	}

	buttonStyle := retui.Style{}.Foreground(retui.Cyan).Bold(true)
	valueStyle := retui.Style{}.Bold(true)

	return retui.Box(
		props,
		retui.Style{},
		components.Panel(
			"State Example", 100,
			retui.Box(
				retui.Props{Direction: retui.Row, Gap: 1},
				retui.Style{},
				retui.Text("( - )", buttonStyle),
				retui.Text(strconv.Itoa(count), valueStyle),
				retui.Text("( + )", buttonStyle),
			),
		),
	)
}
