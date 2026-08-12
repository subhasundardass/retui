<div align="center">

# retui

**A component-driven framework for building fast, modern Terminal User Interfaces in Go —**
**components, hooks, and flexbox-style layout, instead of manually painting characters to a screen.**

[![Go Reference](https://pkg.go.dev/badge/github.com/subhasundardass/retui.svg)](https://pkg.go.dev/github.com/subhasundardass/retui)
[![CI](https://github.com/subhasundardass/retui/actions/workflows/ci.yml/badge.svg)](https://github.com/subhasundardass/retui/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/subhasundardass/retui?include_prereleases)](https://github.com/subhasundardass/retui/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)
[![Go Version](https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go)](go.mod)

[Quick Start](#quick-start) · [Why retui?](#why-retui) · [Features](#features) · [Components](#components) · [Wiki](https://github.com/subhasundardass/retui/wiki) · [Contributing](#contributing)

<img src="https://github.com/subhasundardass/retui/raw/main/retui_banner.png" alt="retui banner" width="720">

</div>

---

If you've used React (or something like it), retui will feel familiar — except it renders to your terminal instead of the DOM.

- Your UI is built from small **functions that return `Element`** — components.
- **Hooks** (`UseState`, `UseEffect`, `UseContext`) let components remember things between renders.
- A **flexbox-style layout system** (`Box`, `Row`, `Column`) handles positioning, so you never calculate coordinates by hand.
- retui diffs the screen at the cell level and only repaints what changed — your app stays fast as it grows.

No prior terminal-UI experience needed. If you can write a Go function, you can build with retui.

## Why retui?

|                                             |                                                                                                                                                            |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 🧩 **Component model, not message-passing** | Build UIs the way you already think about them — small functions and props — instead of wiring an Elm-style `Update`/`Msg` state machine for every screen. |
| 📐 **Real flexbox layout**                  | `Box`, `Row`, `Column`, gap, padding, align, justify. No manual `x, y` coordinate math, ever.                                                              |
| ⚡ **Cell-level diffing**                   | Only the terminal cells that actually changed are repainted — smooth, flicker-free updates even on large screens.                                          |
| 🔋 **Batteries included**                   | Focus management, modals/windows, forms, and a virtualized `Table` ship in the core — not bolted on as third-party plugins.                                |
| 🎓 **Gentle learning curve**                | If you know React's mental model, you already know 80% of retui.                                                                                           |

## Quick Start

```bash
go get github.com/subhasundardass/retui
```

Requires **Go 1.26+**.

Here's a complete, working app — a counter you control with the Enter key. This is the entire app: no boilerplate, no manual main-loop plumbing.

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

### Try the example app

The repo ships with a demo that exercises the built-in components — a good way to see what's possible before building your own:

```bash
go run ./cmd/app
```

## Features

- **Components** — see the [full table below](#components) for what ships out of the box
- **Hooks** — `UseState`, `UseStateKeyed`, `UseReducer`, `UseMemo`, `UseRef`, `UseEffect`, `UseContext`
- **Layout** — flexbox-style `Box`/`Row`/`Column` with sizing, gap, padding, margin, align, and justify
- **Styling** — colors, borders (sharp, rounded, double, thick), text attributes, style inheritance down the tree
- **Focus & navigation** — keyboard-driven focus traversal between components and screens
- **Windows & modals** — floating, overlaid dialogs and popups via the `window` package
- **Forms** — typed form state with `UseForm[T]`
- **Markdown rendering** — headers, bold, italic, code, links, lists, blockquotes, and tables, straight from a string

## Components

Everything below ships in `retui/components` — no third-party plugins required for a typical app.

| Component        | What it does                                                                                            |
| ---------------- | ------------------------------------------------------------------------------------------------------- |
| `Button`         | Clickable/focusable action trigger with an `OnPress` callback                                           |
| `TextInput`      | Single-line text field with placeholder, focus, and submit handling                                     |
| `Password`       | Masked single-line input for secrets                                                                    |
| `TextArea`       | Multi-line text field                                                                                   |
| `Number`         | Numeric input with an empty/unset state distinct from zero                                              |
| `Date`           | Date input with configurable min/max bounds                                                             |
| `Checkbox`       | Boolean toggle                                                                                          |
| `List`           | Keyboard-navigable selectable list (Up/Down/Home/End/Enter)                                             |
| `Tree`           | Expandable/collapsible hierarchical list                                                                |
| `Table`          | Tabular data grid with row virtualization — only visible rows are rendered, so large datasets stay fast |
| `SelectDropdown` | Dropdown single-select                                                                                  |
| `SelectPicker`   | Inline single-select picker                                                                             |
| `ProgressBar`    | Filled/empty block bar for progress (0–1)                                                               |
| `Spinner`        | Animated braille loading indicator                                                                      |
| `Badge`          | Short colored status label                                                                              |
| `Panel`          | Bordered, titled content container                                                                      |
| `Toast`          | Transient notification message                                                                          |

> **Honest note:** `List` currently renders every item in the underlying slice each frame — it does not row-virtualize the way `Table` does. Fine for lists of a few hundred items; if you're paging through thousands of rows, reach for `Table` instead, or [open an issue](https://github.com/subhasundardass/retui/issues/new) if `List` virtualization would help your use case.

## retui vs. other Go TUI libraries

|               | retui                           | bubbletea                             | tview                                          |
| ------------- | ------------------------------- | ------------------------------------- | ---------------------------------------------- |
| Mental model  | Components + hooks              | Elm architecture (`Update`/`Msg`)     | Widget tree, imperative                        |
| Layout        | Flexbox-style, built-in         | Manual / third-party (`lipgloss`)     | `Flex`/`Grid` primitives, built-in             |
| State updates | `UseState`, re-render on change | Explicit `Msg` dispatch               | Direct mutation + `Draw()`                     |
| Best fit      | Devs who know React             | Devs who like explicit state machines | Devs who want widget-first, imperative control |

Each is a solid choice — this is about fit, not a ranking. If React's component model is where your instincts already are, retui will feel like home.

## Learn retui — the Wiki

The README is just the "hello world." Everything else lives in the [wiki](https://github.com/subhasundardass/retui/wiki), written to be read in order if you're new, or jumped into if you already know what you're looking for:

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
