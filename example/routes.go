package example

import "github.com/subhasundardass/retui/retui"

// ─── Screen Definition ────────────────────────────────────────────────────
type Screen struct {
	ID     string
	Title  string
	Render func(props retui.Props) retui.Element
}

// ─── Helper Functions ─────────────────────────────────────────────────────

func GetScreen(id string) (Screen, bool) {
	screen, ok := Registry[id]
	return screen, ok
}

// Registry holds all available screens
var Registry = map[string]Screen{
	"basic-inputs": {
		ID:     "basic-inputs",
		Title:  "Basic Input",
		Render: BasicInputExample,
	},
	"windows": {
		ID:     "windows",
		Title:  "Windows",
		Render: WindowsExample,
	},
}
