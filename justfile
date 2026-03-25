# templbook – component playground for templ
# Run with: just demo

# Generate templ components then start the server
demo:
    go mod tidy
    go run github.com/a-h/templ/cmd/templ@v0.3.857 generate
    go run ./example

# Watch for .templ changes and live-reload the playground (requires air)
# Install air once with: go install github.com/air-verse/air@latest
dev:
    go run github.com/a-h/templ/cmd/templ@v0.3.857 generate --watch &
    go run github.com/air-verse/air@latest

# Render all stories to static HTML in ./dist (for GitHub Pages etc.)
build:
    go run github.com/a-h/templ/cmd/templ@v0.3.857 generate
    go run ./example -build dist
