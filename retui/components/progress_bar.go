package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

// ProgressBar renders a filled/empty block bar representing value (0-1)
// across width characters, including the two end caps.
func ProgressBar(value float64, width int, color retui.Color) retui.Element {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	inner := width - 2
	filled := int(float64(inner) * value)
	empty := inner - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return retui.Text(bar, retui.Style{}.Foreground(color))
}
