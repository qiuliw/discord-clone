package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/qiuliw/discord-clone/services/user/internal/domain"
)

// Authenticator resolves a user from an access token.
type Authenticator interface {
	UserFromToken(string) (domain.User, error)
}

type userContextKey struct{}

// UserFromContext returns the authenticated user attached to the request.
func UserFromContext(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(userContextKey{}).(domain.User)
	return user, ok
}

// RequireAuth authenticates a request before invoking the next handler.
func RequireAuth(auth Authenticator, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerOrCookie(r, cookieName)
			if !ok {
				writeUnauthorized(w)
				return
			}

			user, err := auth.UserFromToken(token)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerOrCookie(r *http.Request, cookieName string) (string, bool) {
	values := r.Header.Values("Authorization")
	if len(values) > 0 {
		if len(values) != 1 {
			return "", false
		}
		parts := strings.Fields(values[0])
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			return "", false
		}
		return parts[1], true
	}

	cookie, err := r.Cookie(cookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return "", false
	}
	return cookie.Value, true
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "not signed in"})
}
