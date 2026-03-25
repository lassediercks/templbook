package templbook

import (
	"encoding/json"
	"strconv"

	"github.com/a-h/templ"
)

func itoa(n int) string { return strconv.Itoa(n) }

// PropDef describes a single editable prop in the playground's controls panel.
type PropDef struct {
	Name    string   `json:"name"`
	Label   string   `json:"label,omitempty"`
	Type    string   `json:"type"`    // "text" | "select" | "bool" | "number"
	Default string   `json:"default"`
	Options []string `json:"options,omitempty"`
}

// propsJSON serialises props for embedding in a data attribute.
func propsJSON(props []PropDef) string {
	if len(props) == 0 {
		return ""
	}
	b, _ := json.Marshal(props)
	return string(b)
}

// Story is a named example of a component.
type Story struct {
	Name      string
	Props     []PropDef
	Build     func(props map[string]string) templ.Component // dynamic; takes precedence over Component
	Component templ.Component                               // static fallback
	Code      string                                        // optional source snippet shown below the preview
}

// Group is a named collection of stories for a single component.
type Group struct {
	Category string
	Name     string
	Stories  []Story
}

// IndexedGroup pairs a Group with its flat index in Book.Groups (used for story URL generation).
type IndexedGroup struct {
	Index int
	Group Group
}

// CategoryView is a category and all groups that belong to it.
type CategoryView struct {
	Name   string
	Groups []IndexedGroup
}

// Book is a collection of component groups that can be browsed in a web UI.
type Book struct {
	Title     string
	Groups    []Group
	Decorator func(templ.Component) templ.Component
}

// New creates a new Book with the given title.
func New(title string) *Book {
	return &Book{Title: title}
}

// Add appends a group of stories under the given category and component name.
func (b *Book) Add(category, component string, stories ...Story) *Book {
	b.Groups = append(b.Groups, Group{Category: category, Name: component, Stories: stories})
	return b
}

// GroupedByCategory returns groups organised by category, preserving insertion order.
func (b *Book) GroupedByCategory() []CategoryView {
	seen := map[string]int{}
	var cats []CategoryView
	for i, g := range b.Groups {
		idx, ok := seen[g.Category]
		if !ok {
			idx = len(cats)
			seen[g.Category] = idx
			cats = append(cats, CategoryView{Name: g.Category})
		}
		cats[idx].Groups = append(cats[idx].Groups, IndexedGroup{Index: i, Group: g})
	}
	return cats
}

// WithDecorator sets a function that wraps every story's component before rendering.
// Use this to inject global styles or layout wrappers.
// If not called, stories are automatically wrapped in a minimal HTML document.
func (b *Book) WithDecorator(fn func(templ.Component) templ.Component) *Book {
	b.Decorator = fn
	return b
}

// decorate applies the decorator if set, otherwise wraps the component in a default HTML page with an optional code block.
func (b *Book) decorate(c templ.Component, code string) templ.Component {
	if b.Decorator != nil {
		return b.Decorator(c)
	}
	return defaultPage(c, code)
}
