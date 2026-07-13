package example

import (
	"github.com/subhasundardass/retui/retui"
	"github.com/subhasundardass/retui/retui/window"
)

type LayoutProps struct {
	Title   string
	Content retui.Element
}

func Example() retui.Element {
	// ── Focus ────────────────────────────────────────────────────────────────
	retui.SetFocusOrder([]string{"sidebar", "content"})

	if retui.CurrentFocus() == "" && !window.IsAnyModalOpen() {
		retui.SetFocus("sidebar")
	}

	if retui.CurrentKey.Code == retui.KeyTab {
		retui.FocusNext()
	}

	//--Current Screen
	currentScreenID := retui.CurrentScreen()

	var content retui.Element
	screen, ok := GetScreen(currentScreenID)
	if !ok {
		retui.Debug("❌ Screen not found:", currentScreenID)
		content = retui.Text("404 - Page Not Found\n\nScreen: "+currentScreenID,
			retui.NewStyle().Foreground(retui.Red))
	} else {
		retui.Debug("✅ Screen found:", screen.ID, screen.Title)
		content = screen.Render(retui.Props{})
	}

	header := retui.Box(
		retui.Props{
			Direction: retui.Row,
			Padding:   [4]int{0, 1, 0, 1},
			Width:     retui.Grow(1),
			Height:    retui.Fit(),
			Justify:   retui.JustifySpaceBetween,
		},
		retui.NewStyle().Foreground(retui.Hex("#535353")).Border(retui.Border{Left: true, Right: true, Bottom: true, Top: true}),
		retui.Text("Example", retui.NewStyle().Foreground(retui.BrightWhite).Bold(true)),
		retui.Text("Version: 1.0.0", retui.NewStyle().Foreground(retui.BrightWhite)),
	)

	mainContent := retui.Box(
		retui.Props{
			Direction: retui.Column,
			Padding:   [4]int{0, 0, 0, 0},
			Width:     retui.Grow(1),
			Gap:       1,
		},
		retui.NewStyle(),
		content,
	)

	body := retui.Box(
		retui.Props{
			Direction: retui.Column,
			Height:    retui.Grow(1),
			Width:     retui.Grow(1),
			Justify:   retui.JustifySpaceBetween,
			Gap:       0,
		},
		retui.NewStyle(),
		retui.Box(
			retui.Props{},
			retui.NewStyle(),
			retui.Box(
				retui.Props{},
				retui.NewStyle(),
				Sidebar(),
				mainContent,
			),
		),
	)

	return retui.Box(
		retui.Props{
			Direction: retui.Column,
			Gap:       0,
			Width:     retui.Grow(1),
			Height:    retui.Grow(1),
		},
		retui.NewStyle().Background(retui.Black),
		header,
		body,
		retui.Box(
			retui.Props{
				Height: retui.Grow(1),
			},
			retui.NewStyle(),
		),
	)
}
