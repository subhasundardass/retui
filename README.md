# retui

Build terminal apps in Go the same way you'd build a modern web app — with components, hooks, and a flexbox-style layout system, instead of manually painting characters to a screen.

<img src="retui_banner.png" alt="Retui Framework" width="700"/>

## What is this, really?

If you've ever used React (or something like it), retui will feel familiar:

- Your UI is built from small **functions that return `Element`** — components.
- **Hooks** (`UseState`, `UseEffect`, `UseContext`) let components remember things between renders.
- A **flexbox-style layout system** (`Box`, `Row`, `Column`) handles positioning, so you never calculate coordinates by hand.
- RetUI automatically figures out what changed and only redraws that — your terminal app stays fast even as it grows.

No prior terminal-UI experience needed. If you can write a Go function, you can build with retui.

## Installation

```bash
go get github.com/subhasundardass/retui
```

Requires **Go 1.26+**.

## Quick Start

Here's a complete, working app — a counter you control with the Enter key:

```go
package main

import (
    "fmt"

    "github.com/subhasundardass/retui/retui"
)

func App(props retui.Props) retui.Element {
    count, setCount := retui.UseState(0)

    if retui.CurrentKey.Code == retui.KeyEnter {
        setCount(count + 1)
    }

    return retui.Box(
        retui.Props{Direction: retui.Column, Gap: 1, Padding: [4]int{1, 2, 1, 2}},
        retui.NewStyle(),
        retui.Text("Press Enter to count, Ctrl-C to quit", retui.NewStyle()),
        retui.Text(fmt.Sprintf("Count: %d", count), retui.NewStyle().Bold(true).Foreground(retui.Cyan)),
    )
}

func main() {
    app := retui.NewApp(0, 0)
    app.Run(App, retui.Props{})
}
```

Run it:

```bash
go run .
```

That's genuinely it. One function, one call to `app.Run`, and you have a working terminal app.

### Try the example app

The repo ships with a demo that exercises the built-in components — a good way to see what's possible before building your own:

```bash
go run ./cmd/app
```

## Learn retui — the Wiki

The README is just the "hello world." Everything else lives in the wiki, written to be read in order if you're new, or jumped into if you already know what you're looking for:

| Page                                     | What it covers                                                                                   |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------ |
| **Core Concepts**                        | The 8 ideas behind retui — start here if you're brand new                                        |
| **Layout System**                        | `Box`, `Row`/`Column`, sizing, gap, padding, align, justify                                      |
| **Hooks**                                | `UseState`, `UseStateKeyed`, `UseEffect`, `UseContext`, and the focus-aware key hooks            |
| **Components**                           | Every built-in component (`Button`, `TextInput`, `List`, `Tree`, etc.) and how to build your own |
| **Styling**                              | Colors, borders, text attributes, and how styles inherit down the tree                           |
| **Navigation & Focus**                   | Moving between screens, and controlling which component has keyboard focus                       |
| **Window System**                        | Floating, overlaid windows — dialogs, popups, and modals                                         |
| **Advanced: Runtime, Renderer & Screen** | How retui actually works internally — the render loop, layout engine, and terminal diffing       |

👉 New here? Read **Core Concepts** first, then **Layout System** and **Hooks** — those three alone are enough to build most simple apps. Come back for **Components**, **Styling**, **Navigation & Focus**, and **Window System** as you need them.

## Contributing

Contributions are welcome! See [`CONTRIBUTING.md`](CONTRIBUTING.md) for how to get set up, our branch/PR workflow, and code style expectations.

Quick version:

```bash
git clone https://github.com/subhasundardass/retui
cd retui
go mod download
go test ./...
```

Open an issue first for anything non-trivial, so we can align on the approach before you write code.

## License

MIT — see [LICENSE.md](LICENSE.md).
