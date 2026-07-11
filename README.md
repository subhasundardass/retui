# retui

A Go framework for building interactive terminal UIs with React-style components and hooks.
Inspired by React and Flutter, Retui brings a component-based, reactive approach to building terminal applications.

## Installation

```bash
go get github.com/subhasundardass/tuix
```

Requires Go 1.21+.

---

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/subhasundardass/tuix/tuix"
)

func App(props tuix.Props) tuix.Element {
    count, setCount := tuix.UseState(0)

    if tuix.CurrentKey.Code == tuix.KeyEnter {
        setCount(count + 1)
    }

    label := tuix.NewStyle().Bold(true).Foreground(tuix.Cyan)

    return tuix.Box(
        tuix.Props{Direction: tuix.Column, Gap: 1, Padding: [4]int{1, 2, 1, 2}},
        tuix.NewStyle(),
        tuix.Text("Press Enter to count, Ctrl-C to quit", tuix.NewStyle()),
        tuix.Text(fmt.Sprintf("Count: %d", count), label),
    )
}

func main() {
    app := tuix.NewApp(60, 6)
    app.Run(App, tuix.Props{})
}
```

Run it:

```bash
go run .
```

Press **Ctrl-C** to exit (there is no `Exit()` function).

---

## Contributing

Contributions are welcome. Please follow these guidelines to keep the codebase consistent.

### Getting Started

```bash
git clone https://github.com/subhasundardass/retui
cd retui
go mod download
go test ./...
```

### Workflow

1. **Open an issue first** for non-trivial changes to align on the approach before writing code.
2. **Branch off `main`:** `git checkout -b feat/my-feature`
3. **Keep commits focused** — one logical change per commit with a clear message.
4. **Add tests** for new layout or rendering behaviour in `*_test.go` files.
5. **Run tests and vet before opening a PR:**
   ```bash
   go test ./...
   go vet ./...
   ```
6. **Open a pull request** against `main` with a description of what changed and why.

### Code Style

- Follow standard Go conventions (`gofmt`, `golint`)
- Keep component functions pure where possible; side effects belong in `UseEffect`
- Avoid adding dependencies; the stdlib + the two existing deps cover most needs

### Adding a Component

1. Write the component function in the appropriate file under `retui/components/`.
2. Use plain typed parameters where possible; reserve `props.Values` for genuinely dynamic data.
3. Add a runnable demo under `examples/<your-feature>/main.go`.
4. Document signature, keyboard contract, and a snippet in [`DOCS.md`](DOCS.md) under the relevant section.

### Reporting Bugs

Open a GitHub issue with:

- Go version (`go version`)
- Terminal emulator and OS
- Minimal reproduction case
- What you expected vs. what happened

---

## License

MIT — see [LICENSE.md](LICENSE).
