# templbook

A component playground for [templ](https://templ.guide/) — like Storybook, but just Go.

Define stories in pure Go, browse them in an isolated iframe UI, edit props live from the controls panel.

## Adding templbook to an existing project

If you already have a Go project with templ components, the setup is two files and one command.

**1. Get the package**

```sh
go get github.com/lassediercks/templbook
```

**2. Create `stories/main.go`**

This file lives alongside your app but runs as a separate binary. It imports your existing components — no changes to them required.

```
yourproject/
├── components/
│   ├── button.templ
│   └── button_templ.go   ← your existing components, untouched
├── stories/
│   └── main.go           ← new file, only used for the playground
└── go.mod
```

```go
// stories/main.go
package main

import (
    "github.com/lassediercks/templbook"
    "yourmodule/components"
)

func main() {
    templbook.New("My App").
        Add("Atoms", "Button",
            templbook.Story{
                Name:      "Primary",
                Component: components.Button("Click me", "#"),
            },
            templbook.Story{
                Name:      "Outline",
                Component: components.ButtonOutline("Click me", "#"),
            },
        ).
        Run()
}
```

**3. Run it**

```sh
go run ./stories
# → http://localhost:6006
```

Stories are automatically wrapped in a minimal HTML document with CSS resets — no layout wrapper needed.

---

## Live reload during development

Install [air](https://github.com/air-verse/air) for hot-reload, then run:

```sh
go run github.com/a-h/templ/cmd/templ@latest generate --watch &
go run github.com/air-verse/air@latest --build.cmd "go build -o /tmp/stories ./stories" --build.bin "/tmp/stories"
```

Or with [just](https://github.com/casey/just), add a `dev` recipe:

```justfile
dev:
    go run github.com/a-h/templ/cmd/templ@latest generate --watch &
    air
```

When a `.templ` file changes, templ regenerates it, air restarts the server, and the preview iframe reloads automatically via SSE.

---

## Props and the controls panel

Stories that take varying input can declare props. The playground renders a **Controls** panel at the bottom with form inputs, and a **JSON** tab for direct editing — both update the preview live.

```go
templbook.Story{
    Name: "Button",
    Props: []templbook.PropDef{
        {Name: "label",   Label: "Label",   Type: "text",   Default: "Click me"},
        {Name: "variant", Label: "Variant", Type: "select", Default: "primary",
         Options: []string{"primary", "outline"}},
    },
    Build: func(p map[string]string) templ.Component {
        return components.Button(p["label"], "#", components.ButtonVariant(p["variant"]))
    },
}
```

`PropDef.Type` supports: `text`, `select`, `bool`, `number`.

When `Build` is set it takes precedence over `Component`. Props are passed as `map[string]string` with defaults pre-filled.

---

## Static build

Export all stories to plain HTML for GitHub Pages or any static host:

```sh
go run ./stories -build dist
```

```
dist/
├── index.html
├── static/ui.css
└── stories/0-0.html, 0-1.html, …
```

---

## API

### `templbook.New(title string) *Book`

Creates a new book. The title appears in the sidebar header.

### `(*Book).Add(category, component string, stories ...Story) *Book`

Registers stories under a two-level hierarchy (`category > component`). Chainable. Categories are collapsible in the sidebar.

### `Story`

```go
type Story struct {
    Name      string
    Props     []PropDef                              // optional — enables the controls panel
    Build     func(map[string]string) templ.Component // dynamic; receives resolved props
    Component templ.Component                        // static fallback when Build is nil
    Code      string                                 // optional source snippet shown below the preview
}
```

### `PropDef`

```go
type PropDef struct {
    Name    string
    Label   string   // display label (falls back to Name)
    Type    string   // "text" | "select" | "bool" | "number"
    Default string
    Options []string // for Type "select"
}
```

### `(*Book).Run(addr ...string)`

Starts the playground server and calls `log.Fatal` on error. The address defaults to `:6006` if omitted. Prompts for an alternate port if the requested one is in use.

```go
book.Run()          // → :6006
book.Run(":8080")   // → :8080
```

### `(*Book).ListenAndServe(addr string) error`

Lower-level alternative to `Run` — returns the error instead of calling `log.Fatal`. Useful if you need custom error handling or want to start the server in a goroutine.

### `(*Book).Build(outDir string) error`

Renders all stories to static HTML in `outDir`.

### `(*Book).WithDecorator(fn func(templ.Component) templ.Component) *Book`

Wraps every story with a custom layout — useful for injecting global stylesheets or a shell component. Without it, a built-in HTML document with CSS resets is used.

---

## Inspiration

Heavily inspired by [Storybook](https://storybook.js.org/) — the goal was to see how close you could get with just Go and templ, no JS build tooling.

## Disclaimer

This is a heavily vibe-coded project. This is an experiment if something that could potentially be re-created in a few hours still can provide value and worked at together.
