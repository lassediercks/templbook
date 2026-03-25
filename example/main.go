package main

import (
	"context"
	"flag"
	"io"
	"log"

	"github.com/a-h/templ"
	"github.com/lassediercks/templbook/example/components"
	"github.com/lassediercks/templbook"
)

func main() {
	outDir := flag.String("build", "", "render stories to static HTML in this directory")
	flag.Parse()

	book := templbook.New("templbook").
		Add("Design Tokens", "Colors & Gradients",
			templbook.Story{Name: "All", Component: components.TokensColors()},
		).
		Add("Design Tokens", "Typography",
			templbook.Story{Name: "All", Component: components.TokensTypography()},
		).
		Add("Atoms", "CodeBlock",
			templbook.Story{
				Name: "Default",
				Component: templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
					return components.CodeBlock().Render(templ.WithChildren(ctx, templ.Raw("go get github.com/lassediercks/templbook")), w)
				}),
				Code: `@components.CodeBlock() {\n\tgo get github.com/lassediercks/templbook\n}`,
			},
		).
		Add("Atoms", "Button",
			templbook.Story{
				Name: "Primary",
				Props: []templbook.PropDef{
					{Name: "label", Label: "Label", Type: "text", Default: "Get Started"},
					{Name: "variant", Label: "Variant", Type: "select", Default: "primary", Options: []string{"primary", "outline"}},
				},
				Build: func(p map[string]string) templ.Component {
					return components.Button(p["label"], "#", components.ButtonVariant(p["variant"]))
				},
				Code: `@components.Button("Get Started", "#", components.ButtonVariantPrimary)`,
			},
			templbook.Story{
				Name: "Outline",
				Props: []templbook.PropDef{
					{Name: "label", Label: "Label", Type: "text", Default: "View on GitHub"},
					{Name: "variant", Label: "Variant", Type: "select", Default: "outline", Options: []string{"primary", "outline"}},
				},
				Build: func(p map[string]string) templ.Component {
					return components.Button(p["label"], "#", components.ButtonVariant(p["variant"]))
				},
				Code: `@components.Button("View on GitHub", "#", components.ButtonVariantOutline)`,
			},
		).
		Add("Sections", "Hero",
			templbook.Story{
				Name:      "Default",
				Component: components.Hero(),
				Code:      `@components.Hero()`,
			},
		).
		Add("Sections", "Features",
			templbook.Story{
				Name:      "Default",
				Component: components.Features(components.DefaultFeatures),
				Code:      `@components.Features(components.DefaultFeatures)`,
			},
		).
		Add("Sections", "GettingStarted",
			templbook.Story{
				Name:      "Default",
				Component: components.GettingStarted(),
				Code:      `@components.GettingStarted()`,
			},
		).
		Add("Pages", "LandingPage",
			templbook.Story{
				Name:      "Full page",
				Component: components.LandingPage(),
				Code:      `@components.LandingPage()`,
			},
		)

	if *outDir != "" {
		if err := book.Build(*outDir); err != nil {
			log.Fatalf("templbook: build failed: %v", err)
		}
		return
	}

	book.Run()
}
