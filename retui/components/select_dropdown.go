package components

import (
	"strings"

	"github.com/subhasundardass/retui/retui"
)

type SelectOption struct {
	Label    string
	Value    string
	Disabled bool
}

type SelectConfig struct {
	ID           string
	Label        string
	Options      []SelectOption
	Value        string
	Placeholder  string
	Width        int
	Height       int
	Style        retui.Style
	Disabled     bool
	OnChange     func(id string, value string)
	OnKeyPress   func(id string, key retui.Key) bool
	OnFocus      func(id string)
	OnBlur       func(id string)
	OnOpenChange func(id string, isOpen bool)

	// OnFilter, when set, is the single source of truth for what options
	// are shown — including the empty query. This means:
	//
	//   - Options can be left empty entirely for a server/DB-backed select.
	//     Opening the dropdown calls OnFilter(id, "") to load the default
	//     page, exactly like typing calls it to load filtered results.
	//   - Static Options is ignored once OnFilter is set (it's never mixed
	//     with the local substring matcher).
	//
	// OnFilter is called exactly once per distinct query — on open, and
	// once per keystroke that actually changes the search text. It is
	// never called again just because the user pressed an arrow key, and
	// it is never called more than once for the same query. It's called
	// synchronously, no debounce — debounce it yourself if the source is
	// slow (e.g. wrap it and only call through after a timer).
	OnFilter func(id string, query string) []SelectOption

	OverlayAbsX int
	OverlayAbsY int
}

type SelectField struct {
	config  SelectConfig
	focused bool
}

// ─── Builder Methods ──────────────────────────────────────────────────────

func SelectDropdown() *SelectField {
	return &SelectField{
		config: SelectConfig{
			Width:       30,
			Height:      5,
			Style:       retui.NewStyle(),
			Placeholder: "Select...",
		},
	}
}

func (s *SelectField) ID(v string) *SelectField              { s.config.ID = v; return s }
func (s *SelectField) Label(v string) *SelectField           { s.config.Label = v; return s }
func (s *SelectField) Options(v []SelectOption) *SelectField { s.config.Options = v; return s }
func (s *SelectField) Value(v string) *SelectField           { s.config.Value = v; return s }
func (s *SelectField) Placeholder(v string) *SelectField     { s.config.Placeholder = v; return s }
func (s *SelectField) Width(w int) *SelectField              { s.config.Width = w; return s }
func (s *SelectField) Height(h int) *SelectField             { s.config.Height = h; return s }
func (s *SelectField) Style(st retui.Style) *SelectField     { s.config.Style = st; return s }
func (s *SelectField) Disabled(v bool) *SelectField          { s.config.Disabled = v; return s }
func (s *SelectField) Focused(v bool) *SelectField           { s.focused = v; return s }

func (s *SelectField) OnChange(fn func(string, string)) *SelectField {
	s.config.OnChange = fn
	return s
}

func (s *SelectField) OnKeyPress(fn func(string, retui.Key) bool) *SelectField {
	s.config.OnKeyPress = fn
	return s
}

func (s *SelectField) OnFocus(fn func(string)) *SelectField {
	s.config.OnFocus = fn
	return s
}

func (s *SelectField) OnBlur(fn func(string)) *SelectField {
	s.config.OnBlur = fn
	return s
}

func (s *SelectField) OnOpenChange(fn func(string, bool)) *SelectField {
	s.config.OnOpenChange = fn
	return s
}

// OnFilter registers a data source for the dropdown. Called with the
// current search text — including "" when the dropdown first opens, so
// it also doubles as "load default options". See SelectConfig.OnFilter
// for the full contract.
func (s *SelectField) OnFilter(fn func(string, string) []SelectOption) *SelectField {
	s.config.OnFilter = fn
	return s
}

func (s *SelectField) OverlayAbsPos(x, y int) *SelectField {
	s.config.OverlayAbsX = x
	s.config.OverlayAbsY = y
	return s
}

// ─── Render Method ──────────────────────────────────────────────────────

func (s *SelectField) Render() retui.Element {
	return renderSelect(s.focused, &s.config)
}

// ─── Core Rendering Function ────────────────────────────────────────────
//
// Focus model: `focused` is the single source of truth, same as InputField.
// The parent app decides which field is active and passes it in; this
// component trusts it directly and reads retui.CurrentKey when focused.
// No separate ID-based focus registry lookup is used for that decision.
//
// CaptureFocus/PushFocus around the overlay is used ONLY to ask retui to
// keep exclusive input focus while the dropdown is open, so no other
// component's focus changes underneath it. It does NOT change how this
// component itself decides whether it's focused.
//
// Search input: there is exactly one owner of the search text — the key
// switch below. It is rendered as plain text, not a nested focusable
// widget, specifically so there's never a second component independently
// reading the same keystroke.
//
// filteredOptions is a cached snapshot of "what filterOptions(config,
// filterText) returned last time filterText changed". It's recomputed
// (and OnFilter re-invoked) at exactly the points where filterText
// changes — open, character typed, backspace — and nowhere else. Arrow
// key navigation, Enter, Escape, Tab all read the cache, so OnFilter is
// never called more than once per distinct query.

func renderSelect(focused bool, config *SelectConfig) retui.Element {
	isOpen, setIsOpen := retui.UseState(false)
	highlighted, setHighlighted := retui.UseState(0)
	filterText, setFilterText := retui.UseState("")
	filteredOptions, setFilteredOptions := retui.UseState[[]SelectOption](nil)

	overlayID := config.ID + "__overlay"
	if config.ID == "" {
		overlayID = "__unnamed_select_overlay__"
	}

	if focused && config.OnFocus != nil && config.ID != "" {
		config.OnFocus(config.ID)
	}
	if !focused && config.OnBlur != nil && config.ID != "" {
		config.OnBlur(config.ID)
	}

	// openOverlay/closeOverlay mutate the local isOpen/filterText/
	// highlighted/filteredOptions variables directly, IN ADDITION TO
	// calling the state setters. The setters only take effect on the
	// next render; without the direct mutation, the rest of THIS render
	// (the defensive re-sync check below, and buildSelectElement) would
	// keep seeing pre-close values — which caused a stale "isOpen=true"
	// to survive into the next render, showing a spurious flash of the
	// dropdown and double-invoking closeOverlay (double PopFocus/
	// ReleaseCaptureFocus, corrupting focus handoff to the next field).
	canOpen := len(config.Options) > 0 || config.OnFilter != nil

	openOverlay := func(initialFilter string) {
		options := filterOptions(config, initialFilter)
		idx, ok := findValueIndex(options, config.Value)
		if !ok {
			idx = firstEnabledIndex(options)
		}

		isOpen = true
		filterText = initialFilter
		filteredOptions = options
		highlighted = idx
		setIsOpen(true)
		setFilterText(initialFilter)
		setFilteredOptions(options)
		setHighlighted(idx)

		retui.OpenOverlay(overlayID)
		retui.PushFocus(overlayID)
		retui.CaptureFocus(overlayID)

		if config.OnOpenChange != nil && config.ID != "" {
			config.OnOpenChange(config.ID, true)
		}
	}

	closeOverlay := func() {
		if !isOpen {
			// Already closed this render — avoid a second Pop/Release,
			// which would pop an entry that belongs to whatever field
			// focus moved to next.
			return
		}
		isOpen = false
		filterText = ""
		setIsOpen(false)
		setFilterText("")

		retui.CloseOverlay(overlayID)
		retui.ReleaseFocus()
		retui.PopFocus()

		if config.OnOpenChange != nil && config.ID != "" {
			config.OnOpenChange(config.ID, false)
		}
	}

	if !config.Disabled {

		keyID := config.ID
		active := focused

		// When the dropdown is open, the overlay owns the keyboard.
		if isOpen {
			keyID = overlayID
			active = retui.HasFocusCapture(overlayID)
		}

		key, ok := retui.UseFocusedKey(keyID, active)
		if ok {

			handled := false
			if config.OnKeyPress != nil {
				handled = config.OnKeyPress(config.ID, key)
			}

			if !handled && (key.Code != retui.KeyNone || key.Rune != 0) {
				if !isOpen {
					switch key.Code {
					case retui.KeyEnter, retui.KeySpace, retui.KeyDown:
						if canOpen {
							openOverlay("")
						}

					default:
						if key.Rune != 0 && key.Rune >= 32 && key.Rune <= 126 && canOpen {
							// Typing while closed opens the dropdown
							// and starts filtering immediately.
							openOverlay(string(key.Rune))
						}
					}
				} else {
					switch key.Code {

					case retui.KeyEscape:
						closeOverlay()

					case retui.KeyEnter:
						if highlighted >= 0 && highlighted < len(filteredOptions) {
							opt := filteredOptions[highlighted]
							if !opt.Disabled {
								config.Value = opt.Value
								if config.OnChange != nil && config.ID != "" {
									config.OnChange(config.ID, opt.Value)
								}
							}
						}
						closeOverlay()

					case retui.KeyUp:
						highlighted = nextEnabledIndex(filteredOptions, highlighted, -1)
						setHighlighted(highlighted)

					case retui.KeyDown:
						highlighted = nextEnabledIndex(filteredOptions, highlighted, 1)
						setHighlighted(highlighted)

					case retui.KeyHome:
						highlighted = firstEnabledIndex(filteredOptions)
						setHighlighted(highlighted)

					case retui.KeyEnd:
						highlighted = lastEnabledIndex(filteredOptions)
						setHighlighted(highlighted)

					case retui.KeyBackspace:
						if len(filterText) > 0 {
							filterText = filterText[:len(filterText)-1]
							filteredOptions = filterOptions(config, filterText)
							highlighted = firstEnabledIndex(filteredOptions)

							setFilterText(filterText)
							setFilteredOptions(filteredOptions)
							setHighlighted(highlighted)
						}

					case retui.KeyTab:
						closeOverlay()

					default:
						if key.Rune != 0 && key.Rune >= 32 && key.Rune <= 126 {
							filterText += string(key.Rune)
							filteredOptions = filterOptions(config, filterText)
							highlighted = firstEnabledIndex(filteredOptions)

							setFilterText(filterText)
							setFilteredOptions(filteredOptions)
							setHighlighted(highlighted)
						}
					}
				}
			}
		}
	}

	// If the field lost focus entirely while the overlay was open (e.g. the
	// app moved focus elsewhere), close it cleanly rather than leaving
	// capture/push state stuck. isOpen here is guaranteed up to date with
	// anything done above in this same render, so this can't double-fire.
	if isOpen && !focused {
		closeOverlay()
	}

	return buildSelectElement(config, focused, isOpen, filterText, filteredOptions, highlighted)
}

// ─── Render Helpers ────────────────────────────────────────────────────────

func buildSelectElement(
	config *SelectConfig,
	focused bool,
	isOpen bool,
	filterText string,
	options []SelectOption,
	highlighted int,
) retui.Element {

	if highlighted >= len(options) {
		highlighted = len(options) - 1
	}
	if highlighted < 0 {
		highlighted = 0
	}

	// Resolve the label for the currently-selected value. Check the
	// static Options first, then fall back to whatever's currently
	// loaded (covers the fully server-backed case, where the selected
	// value's label may only exist in a previously-fetched page).
	selectedLabel := config.Placeholder
	target := strings.TrimSpace(config.Value)
	if target != "" {
		for _, opt := range config.Options {
			if strings.TrimSpace(opt.Value) == target {
				selectedLabel = opt.Label
				break
			}
		}
		if selectedLabel == config.Placeholder {
			for _, opt := range options {
				if strings.TrimSpace(opt.Value) == target {
					selectedLabel = opt.Label
					break
				}
			}
		}
	}

	displayText := selectedLabel
	if isOpen {
		if filterText != "" {
			displayText = filterText
		} else {
			displayText = config.Placeholder
		}
	}

	arrow := "▼"
	if isOpen {
		arrow = "▲"
	}

	// Focused-background applies whenever the field is the active one,
	// same condition used for keyboard handling above — kept in sync so
	// the highlight always matches whether keys are actually going here.
	inputHighlighted := focused && !config.Disabled

	textStyle := config.Style
	if inputHighlighted {
		textStyle = textStyle.Foreground(retui.BrightWhite).Background(retui.Blue).Bold(true)
	} else {
		textStyle = textStyle.Foreground(retui.BrightBlack).Background(retui.Hex("#0c0c0c")).Bold(true)
	}

	arrowStyle := retui.NewStyle().Foreground(retui.BrightBlack).Bold(true)
	if inputHighlighted {
		arrowStyle = retui.NewStyle().Foreground(retui.Cyan).Bold(true)
	}

	paddedDisplay := truncateText(displayText, config.Width-2)
	displayLen := len([]rune(paddedDisplay))
	if displayLen < config.Width-2 {
		paddedDisplay = paddedDisplay + strings.Repeat(" ", config.Width-2-displayLen)
	}

	inputBox := retui.Box(
		retui.Props{Direction: retui.Row, Width: retui.Fixed(config.Width)},
		retui.NewStyle(),
		retui.Text(paddedDisplay, textStyle),
		retui.Text(" ", retui.NewStyle()),
		retui.Text(arrow, arrowStyle),
	)

	// Local offset within the row: how far inputBox sits to the right of
	// the row's own left edge (0 with no label; label width + gap otherwise).
	localXOffset := 0
	inputRow := retui.Element(inputBox)
	if config.Label != "" {
		labelText := config.Label + ":"
		localXOffset = len([]rune(labelText)) + 1 // label width + Gap:1
		labelStyle := retui.NewStyle().Foreground(retui.White)
		inputRow = retui.Box(
			retui.Props{Direction: retui.Row, Gap: 1},
			retui.NewStyle(),
			retui.Text(labelText, labelStyle),
			inputBox,
		)
	}

	if !isOpen {
		return retui.Box(retui.Props{Direction: retui.Column}, retui.NewStyle(), inputRow)
	}

	searchInput := retui.Box(
		retui.Props{Padding: [4]int{0, 1, 0, 1}},
		retui.NewStyle().Border(retui.Border{Bottom: true, Color: retui.Cyan}),
		retui.Box(
			retui.Props{Direction: retui.Row, Gap: 1},
			retui.NewStyle(),
			retui.Text("Search:", retui.NewStyle().Foreground(retui.BrightBlack).Bold(true)),
			retui.Text(
				padSearchText(filterText, config.Width-8),
				retui.NewStyle().Foreground(retui.White).Background(retui.Hex("#1a1a1a")),
			),
		),
	)

	optionList := buildOptionsList(config, options, highlighted)

	dropdownBox := retui.Box(
		retui.Props{Direction: retui.Column, Gap: 0},
		retui.NewStyle().Background(retui.Black).Border(retui.Border{
			Top: true, Right: true, Bottom: true, Left: true,
			Chars: retui.BorderRounded, Color: retui.Cyan,
		}),
		searchInput,
		optionList,
	)

	// Overlay coords are absolute screen position: the caller-supplied
	// top-left of this select (OverlayAbsX/Y), plus the local offset to
	// line up under inputBox rather than under the label, plus one row
	// down to sit directly below the input.
	overlayX := config.OverlayAbsX + localXOffset
	overlayY := config.OverlayAbsY + 1

	return retui.Box(
		retui.Props{Direction: retui.Column},
		retui.NewStyle(),
		inputRow,
		retui.Overlay(overlayX, overlayY, dropdownBox),
	)
}

func buildOptionsList(config *SelectConfig, options []SelectOption, highlighted int) retui.Element {
	if len(options) == 0 {
		return retui.Box(
			retui.Props{
				Direction: retui.Column,
				Width:     retui.Fixed(config.Width + 2),
				Padding:   [4]int{0, 1, 0, 1},
			},
			retui.NewStyle(),
			retui.Text(" No results found", retui.NewStyle().Foreground(retui.BrightBlack)),
		)
	}

	start, end := 0, len(options)
	if len(options) > config.Height {
		start = highlighted - config.Height/2
		if start < 0 {
			start = 0
		}
		end = start + config.Height
		if end > len(options) {
			end = len(options)
			start = end - config.Height
		}
	}

	rows := []retui.Element{}
	for i := start; i < end; i++ {
		opt := options[i]
		style := retui.NewStyle().Foreground(retui.White)
		prefix := "  "

		switch {
		case opt.Disabled:
			style = retui.NewStyle().Foreground(retui.BrightBlack)
			if i == highlighted {
				// still show it's the cursor position, but visibly disabled
				style = retui.NewStyle().Foreground(retui.BrightBlack).Background(retui.Hex("#1a1a1a")).Italic(true)
			}
		case i == highlighted:
			style = retui.NewStyle().Background(retui.Blue).Foreground(retui.White).Bold(true)
		case opt.Value == config.Value:
			prefix = "✓ "
		}

		label := truncateText(opt.Label, config.Width-2)
		rows = append(rows, retui.Text(prefix+label, style))
	}

	return retui.Box(
		retui.Props{
			Direction: retui.Column,
			Width:     retui.Fixed(config.Width + 2),
			Padding:   [4]int{0, 1, 0, 1},
		},
		retui.NewStyle(),
		rows...,
	)
}

// ─── Helpers ──────────────────────────────────────────────────────────────

func findValueIndex(options []SelectOption, value string) (int, bool) {
	value = strings.TrimSpace(value)
	for i, opt := range options {
		if strings.TrimSpace(opt.Value) == value {
			return i, true
		}
	}
	return -1, false
}

func firstEnabledIndex(options []SelectOption) int {
	for i, opt := range options {
		if !opt.Disabled {
			return i
		}
	}
	return 0
}

func lastEnabledIndex(options []SelectOption) int {
	for i := len(options) - 1; i >= 0; i-- {
		if !options[i].Disabled {
			return i
		}
	}
	return 0
}

func nextEnabledIndex(options []SelectOption, current int, direction int) int {
	if len(options) == 0 {
		return current
	}
	i := current + direction
	for i >= 0 && i < len(options) {
		if !options[i].Disabled {
			return i
		}
		i += direction
	}
	return current
}

// filterOptions returns the option list for a given query.
//
// If OnFilter is set, it's the single source of truth for every query —
// including "" — so a select can be entirely server-backed with no
// static Options: opening it calls OnFilter(id, "") to load the default
// page, and further calls happen as the user types.
//
// If OnFilter is nil, Options is used as-is for an empty query, and a
// case-insensitive local substring match over Options is used otherwise.
//
// Callers must only invoke this when the query actually changes (open,
// character typed, backspace) — never on navigation or on every render —
// so a database-backed OnFilter isn't hit repeatedly for the same query.
func filterOptions(config *SelectConfig, query string) []SelectOption {
	if config.OnFilter != nil {
		return config.OnFilter(config.ID, query)
	}
	if query == "" {
		return config.Options
	}
	var out []SelectOption
	for _, opt := range config.Options {
		if strings.Contains(strings.ToLower(opt.Label), strings.ToLower(query)) {
			out = append(out, opt)
		}
	}
	return out
}

func truncateText(text string, maxWidth int) string {
	runes := []rune(text)
	if len(runes) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}
	return string(runes[:maxWidth-3]) + "..."
}

// padSearchText renders the search box's text with a trailing caret,
// truncated/padded to width. There is no separate cursor state — the
// caret always sits at the end because filterText is only ever appended
// to or trimmed from the end.
func padSearchText(text string, width int) string {
	if width < 1 {
		width = 1
	}
	display := text + "▏"
	truncated := truncateText(display, width)
	runeLen := len([]rune(truncated))
	if runeLen < width {
		truncated += strings.Repeat(" ", width-runeLen)
	}
	return truncated
}
