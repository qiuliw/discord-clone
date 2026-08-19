package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qiuliw/discord-clone/pkg/middleware"
)

func TestAuthRoutesExposeOnlyMeAsProtected(t *testing.T) {
	router := New(NewAuthHandler(nil))

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   int
	}{
		{name: "me requires auth", method: http.MethodGet, path: "/api/auth/me", want: http.StatusUnauthorized},
		{name: "register remains public", method: http.MethodPost, path: "/api/auth/register", body: `{`, want: http.StatusBadRequest},
		{name: "login remains public", method: http.MethodPost, path: "/api/auth/login", body: `{`, want: http.StatusBadRequest},
		{name: "logout remains public", method: http.MethodPost, path: "/api/auth/logout", want: http.StatusOK},
		{name: "health remains public", method: http.MethodGet, path: "/api/health", want: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(res, req)

			if res.Code != tt.want {
				t.Fatalf("status = %d, want %d, body = %s", res.Code, tt.want, res.Body.String())
			}
		})
	}
}

func TestCORSPreflightRunsBeforeAuthentication(t *testing.T) {
	router := middleware.CORS([]string{"localhost"}, New(NewAuthHandler(nil)))
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/auth/me", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	req.Header.Set("Access-Control-Request-Headers", "Authorization")

	router.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Authorization") {
		t.Fatalf("allow headers = %q", got)
	}
}
