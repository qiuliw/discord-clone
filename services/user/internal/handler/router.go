package handler

import "net/http"

// Mounter 将一组路由注册到 mux 上。
type Mounter interface {
	Mount(mux *http.ServeMux)
}

// New 用 Mounter 组合业务路由，避免本函数与具体 handler 耦合。
func New(mounts ...Mounter) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", Health)
	for _, m := range mounts {
		m.Mount(mux)
	}
	return mux
}
