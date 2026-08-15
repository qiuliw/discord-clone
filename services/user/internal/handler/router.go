package handler

import (
	"net/http"

	"github.com/qiuliw/discord-clone/pkg/middleware"
)

// Mounter 将一组路由注册到 mux 上。
type Mounter interface {
	Mount(mux *http.ServeMux)
}

func New(mounts ...Mounter) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", Health)
	for _, m := range mounts {
		m.Mount(mux)
	}
	return middleware.CORS(mux)
}
