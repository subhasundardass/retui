package components

import "github.com/subhasundardass/retui/retui"

// spinnerFrames are the braille glyphs cycled through by Spinner.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner renders an animated braille spinner. Advances one frame per render.
func Spinner(label string) retui.Element {
	frame, setFrame := retui.UseState(0)
	setFrame((frame + 1) % len(spinnerFrames))
	return retui.Text(
		spinnerFrames[frame]+" "+label,
		retui.Style{}.Foreground(retui.Cyan),
	)
}
