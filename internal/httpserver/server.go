package httpserver

import (
	"fmt"
	"net/http"
)

type Options struct {
	GitHubWebhook http.Handler
}

func New(opts Options) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintln(w, "ok")
	})
	if opts.GitHubWebhook != nil {
		mux.Handle("/webhooks/github", opts.GitHubWebhook)
	}
	return mux
}
