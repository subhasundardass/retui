package example

import (
	"github.com/subhasundardass/retui/retui"
	"github.com/subhasundardass/retui/retui/components"
	"github.com/subhasundardass/retui/retui/window"
)

func WindowsExample(props retui.Props) retui.Element {

	// Handle keyboard shortcuts
	handleKeyboard()

	return retui.Box(
		props,
		retui.NewStyle(),
		components.Panel(
			"Windows", 100,

			retui.Text("1 - Simple window 1", retui.NewStyle()),
			retui.Text("2 - Simple Window 2", retui.NewStyle()),
			retui.Text("3 - Open modal Window", retui.NewStyle()),
			retui.Text("c - Close all windows", retui.NewStyle()),
			retui.Text("ESC - Close focused window", retui.NewStyle()),
		),
	)
}

func handleKeyboard() {

	switch retui.CurrentKey.Rune {
	case '1':
		w := window1()
		w.Show()
		// w.Focus()

	case '2':
		w := window2()
		w.Show()
		// w.Focus()

	case '3':
		w := modalWindow()
		w.Show()
		// w.Focus()

	case 'c':
		window.CloseAll()
	case 'q':
		retui.Exit()
	}

}

func window1() *window.Window {

	content := retui.Box(
		retui.Props{
			Height: retui.Fixed(5),
			Width:  retui.Grow(1),
		},
		retui.NewStyle(),
		retui.Text("Window 1", retui.NewStyle()),
	)

	win := window.NewWindow().
		SetTitle("Simple Window").
		SetModal(false).
		SetContent(content).
		SetPosition(20, 10)

	win.OnKeyPress(func(key retui.Key) {
		if key.Code == retui.KeyEscape {
			retui.Infof("Modal received key: %+v", key)
			win.Close()
		}
	})

	return win
}

func window2() *window.Window {

	content := components.Panel(
		"Window 1", 40,
		retui.Box(
			retui.Props{
				Height: retui.Fixed(4),
				Width:  retui.Grow(1),
			},
			retui.NewStyle(),
			retui.Text("Window 2", retui.NewStyle()),
		),
	)

	win := window.NewWindow().
		SetTitle("Simple Window").
		SetModal(false).
		SetContent(content).
		SetPosition(30, 20)

	win.OnKeyPress(func(key retui.Key) {
		if key.Code == retui.KeyEscape {
			retui.Infof("Modal received key: %+v", key)
			win.Close()
		}
	})

	return win
}

func modalWindow() *window.Window {

	content := retui.Box(
		retui.Props{
			Height: retui.Fixed(5),
			Width:  retui.Grow(1),
		},
		retui.NewStyle(),
		retui.Text("Modal Window", retui.NewStyle()),
	)

	win := window.NewWindow().
		SetTitle("Simple Window").
		SetModal(true).
		SetContent(content).
		SetPosition(10, 20)

	win.OnKeyPress(func(key retui.Key) {
		if key.Code == retui.KeyEscape {
			retui.Infof("Modal received key: %+v", key)
			win.Close()
		}
	})

	return win
}
