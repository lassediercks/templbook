package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/demo/twittempl/components"
)

func findAddr(startPort int) string {
	for port := startPort; port < startPort+100; port++ {
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return addr
		}
	}
	return fmt.Sprintf(":%d", startPort)
}

func main() {
	tweets := []components.Tweet{
		{
			ID:          "1",
			DisplayName: "Rob Pike",
			Username:    "rob_pike",
			TimeAgo:     "2h",
			Body:        "Simplicity is complicated.\n\nThe most powerful ideas in computer science are the ones that let you build complex systems out of simple, orthogonal components. Go was designed around exactly this principle.",
			Replies:     312,
			Retweets:    1400,
			Likes:       8700,
			Views:       142000,
			AvatarColor: "#0ea5e9",
		},
		{
			ID:          "2",
			DisplayName: "Russ Cox",
			Username:    "_rsc",
			TimeAgo:     "4h",
			Body:        "Go 1.24 is out! Highlights include generic type aliases, improved finalization semantics, and a new benchmark framework in testing/synctest. The toolchain keeps getting faster too. 🎉\n\ngolang.org/doc/go1.24",
			Replies:     89,
			Retweets:    2100,
			Likes:       12300,
			Views:       287000,
			AvatarColor: "#8b5cf6",
		},
		{
			ID:          "3",
			DisplayName: "Adrian Hesketh",
			Username:    "adrianhesketh",
			TimeAgo:     "6h",
			Body:        "templ v0.3 just landed with significant compile-time improvements and a redesigned LSP. Type-safe HTML templating in Go has never been this smooth.\n\nIf you're building Go web apps without templ, you're missing out.",
			Replies:     54,
			Retweets:    430,
			Likes:       2900,
			Views:       38000,
			AvatarColor: "#f59e0b",
		},
		{
			ID:          "4",
			DisplayName: "Kelsey Hightower",
			Username:    "kelseyhightower",
			TimeAgo:     "8h",
			Body:        "Kubernetes is a platform for building platforms. The real power isn't in running containers — it's in the abstractions it provides for teams to build reliable systems on top of.\n\nMost companies should be building on the platform, not operating it.",
			Replies:     201,
			Retweets:    3800,
			Likes:       19400,
			Views:       512000,
			AvatarColor: "#ef4444",
		},
		{
			ID:          "5",
			DisplayName: "Xe Iaso",
			Username:    "theprincessxena",
			TimeAgo:     "10h",
			Body:        "Hot take: the reason Go error handling feels verbose is because errors actually *are* that important, and languages that hide them behind exceptions are hiding the wrong thing from you.\n\nExplicit is better than implicit. Every. Single. Time.",
			Replies:     178,
			Retweets:    920,
			Likes:       7600,
			Views:       94000,
			AvatarColor: "#06b6d4",
		},
		{
			ID:          "6",
			DisplayName: "Filippo Valsorda",
			Username:    "filosottile",
			TimeAgo:     "13h",
			Body:        "Reminder that memory safety bugs are still the #1 source of critical CVEs in systems software. If you're writing new networked code in C or C++ in 2025, you need a very strong justification.\n\nGo, Rust, and even Zig have all matured to the point where the tradeoffs are real.",
			Replies:     93,
			Retweets:    1200,
			Likes:       8100,
			Views:       110000,
			AvatarColor: "#10b981",
		},
		{
			ID:          "7",
			DisplayName: "Mara Bos",
			Username:    "m_ou_se",
			TimeAgo:     "1d",
			Body:        "Just merged a change that makes the Rust standard library's Mutex 40% faster on highly contended workloads by switching to a hybrid spin/park strategy.\n\nSometimes the best performance work is the kind that users never have to think about.",
			Replies:     67,
			Retweets:    780,
			Likes:       5400,
			Views:       72000,
			AvatarColor: "#ec4899",
		},
		{
			ID:          "8",
			DisplayName: "Charity Majors",
			Username:    "mipsytipsy",
			TimeAgo:     "1d",
			Body:        "Observability is not about dashboards.\n\nIt's about the ability to ask arbitrary questions about your system's behaviour in production without having to predict those questions in advance.\n\nIf you can't do that, you don't have observability. You have monitoring.",
			Replies:     144,
			Retweets:    2400,
			Likes:       14700,
			Views:       198000,
			AvatarColor: "#f97316",
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := components.Feed(tweets).Render(r.Context(), w); err != nil {
			http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
			log.Printf("render error: %v", err)
		}
	})

	addr := findAddr(8080)
	log.Printf("Twittempl server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
