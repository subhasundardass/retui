# retui — Complete Reference

A Go framework for building interactive terminal UIs with a React-style mental
model. Components are plain functions; layout is flexbox; rendering is
cell-diffed.

This document is the deep technical reference — precise signatures, gotchas,
and source pointers. For a one-screen pitch, see [README.md](README.md). For
a friendlier, guided walkthrough (better if you're new to retui), see the
[wiki](https://github.com/subhasundardass/retui/wiki) — start with **Core
Concepts**, then **Layout System** and **Hooks**.

This document covers the core rendering primitives: elements, layout,
styling, and hooks. It does **not** cover the built-in component library,
screen navigation, focus management, or the window/modal system — those are
documented in the wiki's **Components**, **Navigation & Focus**, and
**Window System** pages, since they're large enough to deserve their own
space rather than being duplicated here.

---

## Table of contents

- [Quick start](#quick-start)
- [Mental model](#mental-model)
- [Easy](#easy)
  - [Text](#text)
  - [Box](#box)
  - [Styling](#styling)
  - [Borders](#borders)
- [Layout](#layout)
  - [Direction, Gap, Padding](#direction-gap-padding)
  - [Sizing: Fit, Fixed, Grow](#sizing-fit-fixed-grow)
  - [Align & Justify](#align--justify)
- [Hooks](#hooks)
  - [UseState](#usestate)
  - [UseStateKeyed](#usestatekeyed)
  - [UseEffect](#useeffect)
  - [UseContext](#usecontext)
- [Advanced](#advanced)
- [Recipes](#recipes)
- [API reference index](#api-reference-index)

---

## Quick start

```bash
go get github.com/subhasundardass/retui
```

Requires Go 1.26+.

```go
package main

import "github.com/subhasundardass/retui/retui"

func App(props retui.Props) retui.Element {
    return retui.Box(
        retui.Props{Padding: [4]int{1, 2, 1, 2}},
        retui.NewStyle(),
        retui.Text("hello, retui", retui.NewStyle().Bold(true).Foreground(retui.Cyan)),
    )
}

func main() {
    app := retui.NewApp(60, 6)
    app.Run(App, retui.Props{})
}
```

Press **Ctrl-C** to exit. There's also a `retui.Exit()` function you can call
programmatically — e.g. from a keybinding or a "Quit" menu item — to request
a graceful shutdown without waiting for Ctrl-C.

→ The repo's runnable demo lives at [`cmd/app`](cmd/app/main.go) — a single
program exercising all built-in components: `go run ./cmd/app`. There's also
a minimal counter at the repo root: `go run .`

---

## Mental model

Three ideas you need before anything else makes sense:

1. **Components are functions.** A component takes `retui.Props` and returns a
   `retui.Element` tree. There is no class, no lifecycle object — just a
   function that gets called on every render.
2. **The call tree is the component tree.** When `App` calls `Header()` and
   `Footer()`, those calls happen _during_ `App`'s execution, so any hooks
   they call (and any context they read) are scoped to the current render.
3. **Hooks are positional.** `UseState` and `UseEffect` identify their state
   slot by _call order within a render_, not by name. Never call them inside
   `if`/`for` — the slot index would shift between renders and you'd silently
   read another component's state.

The runtime's job:

```
keyboard / ticker / resize event
        │
        ▼
   re-render the whole tree (up to 3x — see "The two-pass render")
        │
        ▼
   measure → layout (2-pass flexbox)
        │
        ▼
   paint into a cell grid
        │
        ▼
   diff against previous frame
        │
        ▼
   write only changed cells to the terminal
```

Source files referenced throughout: [`retui/runtime.go`](retui/runtime.go),
[`retui/renderer.go`](retui/renderer.go),
[`retui/layout.go`](retui/layout.go),
[`retui/hooks.go`](retui/hooks.go).

---

## Easy

### Text

```go
retui.Text("hello", retui.NewStyle().Bold(true))
```

`Text` renders a single line. Newlines in the string are **not** treated as
line breaks — use [`MultilineText`](#wrappedtext-vs-multilinetext) for that.

### Box

`Box` is the only container. It lays out children using a flexbox-like
algorithm.

```go
retui.Box(
    retui.Props{
        Direction: retui.Column,         // Row or Column (default: Row)
        Gap:       1,                   // empty cells between children
        Padding:   [4]int{1, 2, 1, 2},  // top, right, bottom, left
        Align:     retui.AlignCenter,    // cross-axis alignment
        Justify:   retui.JustifyStart,   // main-axis distribution
        Width:     retui.Grow(1),        // optional; default is Fit()
        Height:    retui.Fit(),          // optional; default is Fit()
    },
    retui.NewStyle(),  // background, foreground, border (no padding)
    childA,
    childB,
)
```

### Styling

`Style` is immutable and chainable.

```go
s := retui.NewStyle().
    Bold(true).
    Italic(true).
    Underline(true).
    Foreground(retui.Hex("#ff6b6b")).
    Background(retui.ANSI256(236))
```

Colors come in three flavours:

| Constructor                  | Range                     | Example                |
| ---------------------------- | ------------------------- | ---------------------- |
| `retui.Red`, `retui.Cyan`, … | ANSI 16 (named)           | `retui.BrightMagenta`  |
| `retui.ANSI256(n)`           | 256-color palette (0–255) | `retui.ANSI256(214)`   |
| `retui.Hex("#rrggbb")`       | 24-bit truecolor          | `retui.Hex("#ffd93d")` |

Named colors: `Black`, `Red`, `Green`, `Yellow`, `Blue`, `Magenta`, `Cyan`,
`White`, and their `Bright*` variants — see
[`retui/style.go`](retui/style.go).

**Inheritance.** Styles flow from parent to child: a child whose foreground is
`ColorNone` inherits the parent's foreground. The same applies to background.
Bold is _promoted_ (a bold parent makes children bold); italic and underline
are not currently part of the inheritance merge — they're per-element only.

### Borders

Borders are part of `Style`, not `Props`:

```go
retui.NewStyle().Border(retui.Border{
    Top: true, Right: true, Bottom: true, Left: true,
    Chars: retui.BorderRounded,  // or BorderSharp, BorderDouble, BorderThick
    Color: retui.Cyan,
})
```

You can toggle individual sides — `Left: true` alone draws a single-side
accent rail. A border can also carry a `Title string`, rendered embedded in
the top edge.

When a border is active, the layout engine automatically inflates the box's
padding by 1 cell on the bordered side so children don't get clipped. You do
not need to manually account for the border in your own padding.

---

## Layout

### Direction, Gap, Padding

`Direction` is the main axis. `Row` stacks children left-to-right;
`Column` stacks them top-to-bottom.

```go
retui.Box(
    retui.Props{Direction: retui.Row, Gap: 2, Padding: [4]int{0, 1, 0, 1}},
    retui.NewStyle(),
    childA, childB, childC,
)
```

`Padding` order is `{top, right, bottom, left}` — same as CSS shorthand.

### Sizing: Fit, Fixed, Grow

Each axis (`Width` and `Height`) can use one of three modes:

| Mode             | Behaviour                                         |
| ---------------- | ------------------------------------------------- |
| `retui.Fit()`    | Hugs the content (default for `Box`).             |
| `retui.Fixed(n)` | Exactly `n` cells.                                |
| `retui.Grow(n)`  | Flex-grow with weight `n`; shares leftover space. |

`Grow(1)` on the root makes your app fill the terminal width. Multiple
siblings with `Grow` divide leftover space by weight: `Grow(2)` + `Grow(1)`
splits 2:1.

### Align & Justify

`Justify` distributes children along the **main axis**:

| Value                       | Effect                                            |
| --------------------------- | ------------------------------------------------- |
| `retui.JustifyStart`        | Pack at start (default)                           |
| `retui.JustifyEnd`          | Pack at end                                       |
| `retui.JustifyCenter`       | Center as a group                                 |
| `retui.JustifySpaceBetween` | First/last hug edges; equal gaps between siblings |
| `retui.JustifySpaceAround`  | Equal gaps including half-gaps at the edges       |

`Align` controls the **cross axis** for each child:

| Value                | Effect                                        |
| -------------------- | --------------------------------------------- |
| `retui.AlignStretch` | Children stretch to fill cross axis (default) |
| `retui.AlignStart`   | Pack at start                                 |
| `retui.AlignCenter`  | Centered                                      |
| `retui.AlignEnd`     | Pack at end                                   |

---

## Hooks

Hooks live in [`retui/hooks.go`](retui/hooks.go) and follow the same rules as
React: call them at the top of a component, in the same order, every render.

### UseState

```go
value, setValue := retui.UseState(0)
setValue(value + 1)  // schedules a re-render
```

The initial value is used only on the first call to that slot. The setter is a
closure over the slot index, so it's safe to capture in goroutines and
`UseEffect` callbacks — it always writes to the same slot.

⚠ **Don't read state and write it back unconditionally in a component body.**
The runtime re-renders the component tree **twice per event** (see
[The two-pass render](#the-two-pass-render)); a bare `setValue(value+1)` in
the body increments by 2, not 1. Always gate on a condition (`if key == ...`).

### UseStateKeyed

`UseState` identifies its slot by call position, which breaks if the number of
hook calls can vary between renders — e.g. a tree component where nodes
expand and collapse, changing how many rows exist. `UseStateKeyed` fixes this
by keying state to a stable string instead:

```go
expanded, setExpanded := retui.UseStateKeyed("node-"+nodeID, false)
```

Use `UseState` by default; reach for `UseStateKeyed` specifically when you're
rendering a variable number of items and each needs independent state.

### UseEffect

```go
retui.UseEffect(func() func() {
    ticker := time.NewTicker(time.Second)
    go func() {
        for range ticker.C {
            setNow(time.Now())
        }
    }()
    return func() { ticker.Stop() }  // cleanup
}, []any{someDep})
```

- The effect runs after the render commits.
- If any element in `deps` differs from the previous render, the previous
  cleanup runs and the effect re-runs.
- Return `nil` if you don't need cleanup.
- An empty `deps` (`[]any{}`) runs the effect exactly once, on mount.

Caveat: state written from a goroutine doesn't trigger a render directly —
the next event (key, internal 500ms tick, or resize) will pick it up.

### UseContext

Share a value across a subtree without prop-drilling.

```go
type Theme struct { Fg, Bg retui.Color }

var ThemeContext = retui.CreateContext(Theme{Fg: retui.White, Bg: retui.Black})

func Header() retui.Element {
    t := retui.UseContext(ThemeContext)
    return retui.Text("◆ hi", retui.NewStyle().Foreground(t.Fg))
}

func App(props retui.Props) retui.Element {
    return ThemeContext.Provide(Theme{Fg: retui.BrightCyan, Bg: retui.Black}, func() retui.Element {
        return Header()
    })
}
```

**Crucial gotcha:** `Provide` takes a **render thunk** (`func() Element`),
not pre-built children. Children must be created _inside_ the thunk so they
run while the value is on the context stack. Children built outside the thunk
have already executed and `UseContext` inside them sees the default value.

See [Context API in depth](#context-api-in-depth) for the why.

---

## Advanced

### Conditional rendering with `If`

```go
retui.If(loggedIn, dashboard, loginPrompt)
```

`If` returns one of two pre-built elements. Because it's a regular function
call, **both branches are evaluated** before `If` runs. Use it for cheap
elements you've already built; don't try to guard expensive work behind one
branch.

Source: [`retui/elements.go`](retui/elements.go).

### `WrappedText` vs `MultilineText`

| Constructor                     | Splits on `\n`?  | Word-wraps?           |
| ------------------------------- | ---------------- | --------------------- |
| `retui.Text(s, style)`          | no (single line) | no                    |
| `retui.MultilineText(s, style)` | yes              | no                    |
| `retui.WrappedText(s, style)`   | yes              | yes (to parent width) |

`WrappedText` sets `Width: Grow(1)` internally, so it expands to fill its
parent's cross-axis space and breaks lines to fit. It registers a `reflow`
callback with the layout engine, which is why
[`retui/layout.go`](retui/layout.go) runs a second measure pass when a
wrapped (or `Markdown`) element is present in the tree.

### Context API in depth

The Context API is in [`retui/hooks.go`](retui/hooks.go). Three exports:

- `retui.CreateContext[T](defaultValue T) *Context[T]` — construct
- `(*Context[T]).Provide(value T, render func() Element) Element` — scope
- `retui.UseContext[T](*Context[T]) T` — read

Under the hood, each `Context` owns its own `[]T` stack. `Provide` appends a
value, runs the render thunk, and pops via `defer`. `UseContext` returns the
top of the stack, or the context's `defaultValue` if the stack is empty.

**Why a thunk?** Children in retui are eager Go function arguments. If
`Provide` took children directly — `Provide(value, child1, child2)` — Go
would evaluate `child1`/`child2` _before_ `Provide` ran, so any
`UseContext` inside them would see the empty stack. The thunk defers
descendant evaluation until _after_ the push.

**Stack identity vs cursor identity.** Unlike `UseState` (slot-by-call-order),
context is keyed by `*Context[T]` pointer. There's no cursor to reset between
renders, and stacks survive across renders because they live on the Context
object, not in a global slab.

### Bracketed paste

The runtime enables bracketed paste mode on startup (terminal emits
`\x1b[200~`…`\x1b[201~` around clipboard content). [`KeyScanner`](retui/key.go)
reassembles paste fragments across multiple `stdin.Read` calls and delivers
them as a single `Key{Code: KeyPaste, Paste: "…"}` event.

At the moment, no built-in component consumes `KeyPaste` automatically — the
event is parsed and delivered by the runtime, but inserting pasted text into
a field is left to you. If you're building a custom text input, check for
`key.Code == retui.KeyPaste` in your key-handling code and use `key.Paste` as
the string to insert (sanitizing it yourself — stripping control characters,
normalizing line endings — if your field needs that).

### Resize handling

The runtime listens for `SIGWINCH` and re-queries the terminal size via
`golang.org/x/term`, then re-renders. Before the re-query, the screen is
cleared (`\033[H\033[2J\033[3J`) so leftover glyphs from a smaller-resize
don't linger. This is handled automatically — your code doesn't need to do
anything special.

### The two-pass render

Each event triggers up to three render passes of your component tree (see the
wiki's **Advanced: Runtime, Renderer & Screen** page for the full breakdown):

1. **Pass 1** — `retui.CurrentKey` is set; state setters mutate state.
2. **Pass 2** — `retui.CurrentKey` is zeroed; the tree renders with updated
   state. This is what gets painted.
3. **Pass 3** (conditional) — if a `UseEffect` callback (which runs after
   Pass 2 paints) itself calls a state setter, the tree renders once more
   immediately, without waiting for another external event.

This is why unconditional `setValue(value+1)` in a component body double-
increments per event. Two practical rules:

- **Gate setters on a condition** — usually a key check (`if Code ==
KeyEnter { ... }`). Pass 1 fires the handler; pass 2's condition is false
  because `CurrentKey` was zeroed.
- **Side effects belong in `UseEffect`**, never in the component body —
  otherwise they fire twice.

Source: [`retui/runtime.go`](retui/runtime.go) `App.Render`.

---

## Recipes

Beyond the primitives above, here are short patterns for common needs, using
the real built-in component builders from `retui/components`.

### Focus cycling

Track which interactive element is focused with a single `UseState`, then
feed each field's `Focused(...)` from it:

```go
focus, setFocus := retui.UseState(0)
if retui.CurrentKey.Code == retui.KeyTab {
    setFocus((focus + 1) % 3)
}

return retui.Box(
    retui.Props{Direction: retui.Column, Gap: 1},
    retui.NewStyle(),
    components.TextInput().ID("name").Focused(focus == 0).
        Value(name).OnChange(func(id, v string) { setName(v) }).Render(),
    components.TextInput().ID("email").Focused(focus == 1).
        Value(email).OnChange(func(id, v string) { setEmail(v) }).Render(),
    components.Button().ID("submit").Label("Submit").Focused(focus == 2).
        OnClick(func(id string) { /* submit */ }).Render(),
)
```

For most apps, prefer `retui.SetFocusOrder` + `retui.IsFocused` (covered in
the wiki's **Navigation & Focus** page) over hand-rolled `UseState` cycling
like this — it's the same idea, but shared framework-wide instead of
reimplemented per screen.

### Polling external data

```go
data, setData := retui.UseState[*Result](nil)
retui.UseEffect(func() func() {
    done := make(chan struct{})
    go func() {
        t := time.NewTicker(5 * time.Second)
        defer t.Stop()
        for {
            select {
            case <-done: return
            case <-t.C:
                if r, err := fetch(); err == nil { setData(r) }
            }
        }
    }()
    return func() { close(done) }
}, []any{})
```

### Toast notifications

A short-lived banner that auto-dismisses:

```go
toast, setToast := retui.UseState("")
retui.UseEffect(func() func() {
    if toast == "" { return nil }
    timer := time.AfterFunc(3*time.Second, func() { setToast("") })
    return func() { timer.Stop() }
}, []any{toast})
```

---

## API reference index

Quick jump-to-source for everything covered in this document. For the
built-in component library, navigation/focus, and the window system, see the
wiki instead of this index.

### Core types

- [`Props`](retui/node.go), [`Element`](retui/node.go), [`LayoutProps`](retui/node.go)
- [`Box`](retui/elements.go), [`Text`](retui/elements.go),
  [`MultilineText`](retui/elements.go), [`WrappedText`](retui/elements.go),
  [`If`](retui/elements.go)

### Layout primitives

- [`Direction`](retui/layout.go) — `Row`, `Column`
- [`Sizing`](retui/layout.go) — `Fit()`, `Fixed(n)`, `Grow(n)`
- [`Alignment`](retui/layout.go) — `AlignStretch`, `AlignStart`,
  `AlignCenter`, `AlignEnd`
- [`Justify`](retui/layout.go) — `JustifyStart`, `JustifyEnd`,
  `JustifyCenter`, `JustifySpaceBetween`, `JustifySpaceAround`

### Style

- [`Style`](retui/style.go) — `NewStyle()`, `.Bold()`, `.Italic()`,
  `.Underline()`, `.Foreground()`, `.Background()`, `.Border()`
- [`Color`](retui/style.go) — `Black`…`BrightWhite`, `Hex(s)`, `ANSI256(n)`
- [`Border`](retui/style.go) — sides + `Chars` + `Color` + `Title`
- Presets: `BorderSharp`, `BorderRounded`, `BorderDouble`, `BorderThick`

### Hooks

- [`UseState[T](initial T) (T, func(T))`](retui/hooks.go)
- [`UseStateKeyed[T](key string, initial T) (T, func(T))`](retui/hooks.go)
- [`UseEffect(fn func() func(), deps []any)`](retui/hooks.go)
- [`CreateContext[T](defaultValue T) *Context[T]`](retui/hooks.go)
- [`UseContext[T](c *Context[T]) T`](retui/hooks.go)
- [`(*Context[T]).Provide(v T, render func() Element) Element`](retui/hooks.go)

### Keyboard

- [`Key`](retui/key.go), [`KeyCode`](retui/key.go), [`CurrentKey`](retui/key.go)
- [`ParseKey`](retui/key.go), [`KeyScanner`](retui/key.go)

### Runtime

- [`NewApp(width, height int) *App`](retui/runtime.go)
- [`(*App).Run(fn func(Props) Element, props Props)`](retui/runtime.go)
- [`Exit()`](retui/runtime.go) — requests a graceful shutdown

### Elsewhere in the wiki

- **Components** — `Button`, `TextInput`, `Password`, `NumberInput`,
  `DateInput`, `Checkbox`, `SelectPicker`, `List`, `Tree`, `Panel`
- **Navigation & Focus** — `PushScreen`/`PopScreen`, `SetFocus`/`IsFocused`,
  `SetFocusOrder`, focus capture, `UseFocusedKey`/`UseFocusedKeySimple`
- **Window System** — `window.NewWindow`, modal windows, Z-order
- **Advanced: Runtime, Renderer & Screen** — the full render-loop and
  terminal-diffing internals, in depth
