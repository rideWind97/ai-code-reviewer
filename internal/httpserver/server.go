package httpserver

import (
	"fmt"
	"net/http"
)

// Options 是 HTTP 路由的依赖集合。
//
// 这里把 GitHub webhook handler 从外部传入，是为了让 httpserver 只负责“挂路由”，
// 不关心 GitHub handler 里面怎么验签、怎么入队。
type Options struct {
	GitHubWebhook http.Handler
}

// New 创建整个 HTTP 服务的根 Handler。
//
// Go 标准库的 http.Handler 是一个很常见的接口：
// 只要实现 ServeHTTP，就可以被 http.ListenAndServe 使用。
func New(opts Options) http.Handler {
	mux := http.NewServeMux()

	// /healthz 通常给负载均衡、Kubernetes 或本地调试使用。
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintln(w, "ok")
	})

	// GitHub webhook 是可选依赖，方便测试时只验证 health check。
	if opts.GitHubWebhook != nil {
		mux.Handle("/webhooks/github", opts.GitHubWebhook)
	}
	return mux
}
