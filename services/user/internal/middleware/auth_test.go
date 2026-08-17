package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qiuliw/discord-clone/services/user/internal/domain"
)

type fakeAuthenticator struct {
	user   domain.User
	err    error
	tokens []string
}

func (f *fakeAuthenticator) UserFromToken(token string) (domain.User, error) {
	f.tokens = append(f.tokens, token)
	return f.user, f.err
}

func TestRequireAuthRejectsInvalidCredentials(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*http.Request)
	}{
		{name: "missing credentials"},
		{name: "empty cookie", setup: func(r *http.Request) {
			r.Header.Add("Cookie", "token=")
		}},
		{name: "non bearer authorization", setup: func(r *http.Request) {
			r.Header.Set("Authorization", "Basic abc")
		}},
		{name: "missing bearer token", setup: func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer")
		}},
		{name: "extra bearer fields", setup: func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer one two")
		}},
		{name: "multiple authorization values", setup: func(r *http.Request) {
			r.Header.Add("Authorization", "Bearer one")
			r.Header.Add("Authorization", "Bearer two")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &fakeAuthenticator{}
			called := false
			handler := RequireAuth(auth, "token")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				called = true
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.setup != nil {
				tt.setup(req)
			}
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			if res.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
			}
			if got := res.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("content type = %q", got)
			}
			if got := res.Body.String(); got != "{\"error\":\"not signed in\"}\n" {
				t.Fatalf("body = %q", got)
			}
			if called {
				t.Fatal("next handler was called")
			}
			if len(auth.tokens) != 0 {
				t.Fatalf("authenticator called with %v", auth.tokens)
			}
		})
	}
}

func TestRequireAuthRejectsAuthenticationError(t *testing.T) {
	auth := &fakeAuthenticator{err: errors.New("secret verification detail")}
	called := false
	handler := RequireAuth(auth, "token")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized || called {
		t.Fatalf("status = %d, called = %v", res.Code, called)
	}
	if strings.Contains(res.Body.String(), "verification") {
		t.Fatalf("response leaked authentication error: %s", res.Body.String())
	}
	if len(auth.tokens) != 1 || auth.tokens[0] != "invalid" {
		t.Fatalf("tokens = %v", auth.tokens)
	}
}

func TestRequireAuthPassesAuthenticatedUser(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*http.Request)
		wantToken string
	}{
		{
			name: "bearer",
			setup: func(r *http.Request) {
				r.Header.Set("Authorization", "bEaReR bearer-token")
			},
			wantToken: "bearer-token",
		},
		{
			name: "cookie",
			setup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: "token", Value: "cookie-token"})
			},
			wantToken: "cookie-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantUser := domain.User{ID: 7, Email: "user@example.com", Name: "User"}
			auth := &fakeAuthenticator{user: wantUser}
			handler := RequireAuth(auth, "token")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, ok := UserFromContext(r.Context())
				if !ok || user != wantUser {
					t.Fatalf("context user = %#v, ok = %v", user, ok)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			tt.setup(req)
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			if res.Code != http.StatusNoContent {
				t.Fatalf("status = %d", res.Code)
			}
			if len(auth.tokens) != 1 || auth.tokens[0] != tt.wantToken {
				t.Fatalf("tokens = %v, want [%s]", auth.tokens, tt.wantToken)
			}
		})
	}
}

func TestRequireAuthDoesNotFallbackFromBearerToCookie(t *testing.T) {
	auth := &fakeAuthenticator{err: domain.ErrInvalidToken}
	handler := RequireAuth(auth, "token")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler was called")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-bearer")
	req.AddCookie(&http.Cookie{Name: "token", Value: "valid-cookie"})
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", res.Code)
	}
	if len(auth.tokens) != 1 || auth.tokens[0] != "invalid-bearer" {
		t.Fatalf("tokens = %v", auth.tokens)
	}
}
