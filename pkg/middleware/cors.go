package middleware

import (
	"net/http"
	"net/url"
)

func CORS(allowedHosts []string, next http.Handler) http.Handler {
	hosts := make(map[string]bool, len(allowedHosts))
	for _, h := range allowedHosts {
		hosts[h] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// 无 Origin：同源请求或非浏览器请求
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		// 不允许的 Origin
		u, err := url.Parse(origin)
		if err != nil || !hosts[u.Hostname()] {
			http.Error(w, "CORS origin denied", http.StatusForbidden)
			return
		}

		// CORS 响应头
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Vary", "Origin")

		// Preflight 预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
