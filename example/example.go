package examples

import (
	"fmt"
	"time"

	"github.com/subhasundardass/retui/retui"
)

func Example(props retui.Props) retui.Element {

	header := retui.Box(
		retui.Props{
			Direction: retui.Row,
			Padding:   [4]int{0, 1, 0, 1},
			Width:     retui.Grow(1),
			Height:    retui.Fit(),
			Justify:   retui.JustifySpaceBetween,
		},
		retui.NewStyle().Foreground(retui.Hex("#ffffff")),
		retui.Text("Example", retui.NewStyle().Bold(true)),
		retui.Text("Version: 1.0.0", retui.NewStyle()),
	)

	body := retui.Box(
		retui.Props{
			Direction: retui.Row,
			Height:    retui.Grow(1),
			Width:     retui.Grow(1),
			Justify:   retui.JustifyCenter,
		},
		retui.NewStyle(),
		retui.Box(
			retui.Props{},
			retui.NewStyle(),

			// retui.Text("Welcome to retui", retui.NewStyle()),
			retui.Box(
				retui.Props{
					Direction: retui.Column,
					Justify:   retui.JustifyCenter,
				},
				retui.NewStyle(),
				retui.Text("Welcome to Retui", retui.NewStyle().Bold(true)),
				retui.Box(
					retui.Props{
						Width:   retui.Fixed(40),
						Padding: [4]int{1, 0, 0, 0},
					},
					retui.NewStyle(),
					retui.WrappedText(
						"A Go framework for building interactive terminal UIs with React-style components and hooks.",
						retui.NewStyle().Italic(true),
					),
				),

				retui.Box(
					retui.Props{
						Width:   retui.Fixed(40),
						Padding: [4]int{1, 0, 0, 0},
					},
					retui.NewStyle(),
					Counter(),
				),
			),
		),
	)

	return retui.Box(
		retui.Props{
			Direction: retui.Column,
			Gap:       0, Width: retui.Grow(1), Height: retui.Grow(1),
		},
		retui.NewStyle().
			Background(retui.Black),
		header,
		body,

		retui.Box(
			retui.Props{
				Height: retui.Grow(1),
			}, retui.NewStyle()),

		Footer(),
	)
}

func Footer() retui.Element {
	now, setNow := retui.UseState(time.Now())

	// Update once every second
	if retui.CurrentTick {
		current := time.Now()
		if current.Second() != now.Second() {
			setNow(current)
		}
	}

	return retui.Box(
		retui.Props{
			Direction: retui.Row,
			Padding:   [4]int{0, 1, 0, 1},
			Width:     retui.Grow(1),
			Height:    retui.Fit(),
			Justify:   retui.JustifySpaceBetween,
		},
		retui.NewStyle().Foreground(retui.Hex("#ffffff")),
		retui.Text(
			"Now : "+now.Format("02-Jan-2006 03:04:05 PM"),
			retui.NewStyle(),
		),
		retui.Text("Status: Connected", retui.NewStyle()),
	)
}

func Counter() retui.Element {
	count, setCount := retui.UseState(0)

	if retui.CurrentKey.Rune == '+' {
		setCount(count + 1)
	}

	if retui.CurrentKey.Rune == '-' {
		setCount(count - 1)
	}

	return retui.Box(
		retui.Props{
			Direction: retui.Row,
			Gap:       1,
		},
		retui.NewStyle(),
		retui.Text("Counter", retui.NewStyle()),
		retui.Text("[-]", retui.NewStyle().Foreground(retui.BrightRed).Bold(true)),
		retui.Text(
			fmt.Sprintf("%3d", count),
			retui.NewStyle().Foreground(retui.BrightYellow).Bold(true),
		),
		retui.Text("  [+]", retui.NewStyle().Foreground(retui.BrightGreen).Bold(true)),
	)
}
