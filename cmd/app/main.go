package main

import (
	"github.com/subhasundardass/retui/example"
	"github.com/subhasundardass/retui/retui"
)

func main() {
	app := retui.NewApp(0, 0)

	renderFn := func(props retui.Props) retui.Element {
		return example.Example()
	}

	app.Run(renderFn, retui.Props{})
}
