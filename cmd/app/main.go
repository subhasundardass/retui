package main

import (
	examples "github.com/subhasundardass/retui/example"
	"github.com/subhasundardass/retui/retui"
)

func main() {
	app := retui.NewApp(0, 0)
	app.Run(examples.Example, retui.Props{})
}
