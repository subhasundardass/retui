package components

import "github.com/subhasundardass/retui/retui"

type ButtonConfig struct {
	ID       string
	Label    string
	Focused  bool
	Disabled bool
	Style    retui.Style
	OnPress  func(id string)
}

type ButtonOption func(*ButtonConfig)

func defaultButtonConfig() ButtonConfig {
	return ButtonConfig{
		Style: retui.NewStyle(),
	}
}

func WithButtonID(id string) ButtonOption {
	return func(c *ButtonConfig) {
		c.ID = id
	}
}

func WithLabel(label string) ButtonOption {
	return func(c *ButtonConfig) {
		c.Label = label
	}
}

func WithDisabled(disabled bool) ButtonOption {
	return func(c *ButtonConfig) {
		c.Disabled = disabled
	}
}

func WithButtonStyle(style retui.Style) ButtonOption {
	return func(c *ButtonConfig) {
		c.Style = style
	}
}

func WithOnPress(fn func(id string)) ButtonOption {
	return func(c *ButtonConfig) {
		c.OnPress = fn
	}
}

func Button(focused bool, opts ...ButtonOption) retui.Element {
	cfg := defaultButtonConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	cfg.Focused = focused

	style := cfg.Style

	if cfg.Disabled {
		style = style.Foreground(retui.Cyan)
	} else if cfg.Focused {
		style = style.
			Foreground(retui.Black).
			Background(retui.Cyan).
			Bold(true)
	} else {
		style = style.Foreground(retui.White)
	}

	if cfg.Focused &&
		retui.CurrentKey.Code == retui.KeyEnter &&
		cfg.OnPress != nil {
		cfg.OnPress(cfg.ID)
	}

	return retui.Text("[ "+cfg.Label+" ]", style)
}
