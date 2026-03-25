# Twittempl – Twitter-like demo in Go + templ
# Run with: just demo

# Generate templ components then start the server
demo:
    go mod tidy
    go run github.com/a-h/templ/cmd/templ@v0.3.857 generate
    go run ./...
