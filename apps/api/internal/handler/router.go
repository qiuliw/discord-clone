package handler

import (
	"net/http"

	"github.com/qiuliw/discord-clone/apps/api/internal/middleware"
	"github.com/qiuliw/discord-clone/apps/api/internal/service"
)

func New(auth *service.Auth) http.Handler {
	h := NewAuthHandler(auth)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/auth/me", h.Me)
	return middleware.CORS(mux)
}
