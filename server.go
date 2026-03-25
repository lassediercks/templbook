package templbook

import (
	"bufio"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/a-h/templ"
)

//go:embed ui.css
var uiCSS []byte

// startupToken is set once when the server starts and sent to SSE clients so
// they can detect a server restart (new token = reload the preview).
var startupToken = strconv.FormatInt(time.Now().UnixNano(), 36)

func handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Send the server startup token. On reconnect (server restart) the client
	// compares this to the previous value and reloads the iframe if different.
	fmt.Fprintf(w, "data: %s\n\n", startupToken)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	tick := time.NewTicker(15 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-tick.C:
			fmt.Fprintf(w, ": ping\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// Run starts the playground on the given address (default ":6006") and calls
// log.Fatal on error. It is the simplest way to launch the playground:
//
//	book.Run()         // → http://localhost:6006
//	book.Run(":8080")  // → http://localhost:8080
func (b *Book) Run(addr ...string) {
	a := ":6006"
	if len(addr) > 0 {
		a = addr[0]
	}
	if err := b.ListenAndServe(a); err != nil {
		log.Fatal(err)
	}
}

// ListenAndServe starts the templbook playground UI on addr (e.g. ":6006").
// If the port is already in use it asks the user interactively whether to try
// the next available port.
func (b *Book) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", b.handleShell)
	mux.HandleFunc("GET /render", b.handleRender)
	mux.HandleFunc("GET /sse", handleSSE)
	mux.HandleFunc("GET /static/ui.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write(uiCSS)
	})

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		if !isAddrInUse(err) {
			return err
		}
		// Port taken — ask user.
		alt, askErr := askAlternatePort(addr)
		if askErr != nil {
			return err // user declined or stdin closed
		}
		ln, err = net.Listen("tcp", alt)
		if err != nil {
			return err
		}
		addr = alt
	}

	log.Printf("templbook: playground running at http://localhost%s", addr)
	return http.Serve(ln, mux)
}

func isAddrInUse(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var sysErr *os.SyscallError
		if errors.As(opErr, &sysErr) {
			return errors.Is(sysErr.Err, syscall.EADDRINUSE)
		}
	}
	return false
}

func askAlternatePort(addr string) (string, error) {
	_, port, _ := net.SplitHostPort(addr)
	p, _ := strconv.Atoi(port)

	// Find next free port.
	next := ""
	for candidate := p + 1; candidate < p+100; candidate++ {
		a := fmt.Sprintf(":%d", candidate)
		if ln, err := net.Listen("tcp", a); err == nil {
			ln.Close()
			next = a
			break
		}
	}

	prompt := fmt.Sprintf("templbook: port %s is already in use", port)
	if next != "" {
		fmt.Fprintf(os.Stderr, "%s. Use %s instead? [Y/n] ", prompt, next)
	} else {
		fmt.Fprintf(os.Stderr, "%s. Enter an alternate port: ", prompt)
	}

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)

	// Empty or "y" → accept the suggested port.
	if next != "" && (line == "" || strings.EqualFold(line, "y")) {
		return next, nil
	}
	// User typed a port number.
	if line != "" {
		if !strings.HasPrefix(line, ":") {
			line = ":" + line
		}
		return line, nil
	}
	return "", fmt.Errorf("no alternate port chosen")
}

func (b *Book) handleShell(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	storyPath := func(g, s int) string {
		return fmt.Sprintf("/render?g=%d&s=%d", g, s)
	}
	if err := shell(b, storyPath, "/static/ui.css").Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (b *Book) handleRender(w http.ResponseWriter, r *http.Request) {
	g, err1 := strconv.Atoi(r.URL.Query().Get("g"))
	s, err2 := strconv.Atoi(r.URL.Query().Get("s"))
	if err1 != nil || err2 != nil || g < 0 || g >= len(b.Groups) {
		http.NotFound(w, r)
		return
	}
	group := b.Groups[g]
	if s < 0 || s >= len(group.Stories) {
		http.NotFound(w, r)
		return
	}

	story := group.Stories[s]

	// Build dynamic props map from query params, filling in defaults.
	props := map[string]string{}
	for _, def := range story.Props {
		props[def.Name] = def.Default
	}
	for k, vs := range r.URL.Query() {
		if k != "g" && k != "s" && len(vs) > 0 {
			props[k] = vs[0]
		}
	}

	var c templ.Component
	if story.Build != nil {
		c = story.Build(props)
	} else {
		c = story.Component
	}
	component := b.decorate(c, story.Code)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		fmt.Fprintf(w, `<p style="color:red;font-family:monospace;padding:16px">render error: %s</p>`, err)
	}
}

// Build renders all stories to static HTML files in outDir, suitable for
// deployment to GitHub Pages or any static web server.
//
// Output structure:
//
//	outDir/
//	  index.html          — the playground shell
//	  static/ui.css       — playground styles
//	  stories/G-S.html    — pre-rendered story for group G, story S
func (b *Book) Build(outDir string) error {
	storiesDir := filepath.Join(outDir, "stories")
	staticDir := filepath.Join(outDir, "static")

	for _, dir := range []string{storiesDir, staticDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Write CSS.
	if err := os.WriteFile(filepath.Join(staticDir, "ui.css"), uiCSS, 0644); err != nil {
		return err
	}

	// Render each story to its own file.
	ctx := context.Background()
	for gi, group := range b.Groups {
		for si, story := range group.Stories {
			component := b.decorate(story.Component, story.Code)
			path := filepath.Join(storiesDir, fmt.Sprintf("%d-%d.html", gi, si))
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			err = component.Render(ctx, f)
			f.Close()
			if err != nil {
				return fmt.Errorf("render %s/%s: %w", group.Name, story.Name, err)
			}
		}
	}

	// Render the shell with paths relative to the root.
	storyPath := func(g, s int) string {
		return fmt.Sprintf("stories/%d-%d.html", g, s)
	}
	index, err := os.Create(filepath.Join(outDir, "index.html"))
	if err != nil {
		return err
	}
	defer index.Close()
	if err := shell(b, storyPath, "static/ui.css").Render(ctx, index); err != nil {
		return err
	}

	log.Printf("templbook: built %d stories to %s", b.storyCount(), outDir)
	return nil
}

func (b *Book) storyCount() int {
	n := 0
	for _, g := range b.Groups {
		n += len(g.Stories)
	}
	return n
}
