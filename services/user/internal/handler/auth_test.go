package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/qiuliw/discord-clone/services/user/internal/domain"
	usermiddleware "github.com/qiuliw/discord-clone/services/user/internal/middleware"
)

type handlerAuthenticator struct {
	user domain.User
}

func (a handlerAuthenticator) UserFromToken(string) (domain.User, error) {
	return a.user, nil
}

func TestMeReturnsPublicUserFromContext(t *testing.T) {
	user := domain.User{
		ID:           42,
		Email:        "member@example.com",
		Name:         "Member",
		PasswordHash: "must-not-leak",
		CreatedAt:    "2026-08-17",
	}
	h := &AuthHandler{}
	handler := usermiddleware.RequireAuth(handlerAuthenticator{user: user}, "token")(http.HandlerFunc(h.Me))
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "valid"})
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	if strings.Contains(res.Body.String(), "PasswordHash") || strings.Contains(res.Body.String(), "must-not-leak") {
		t.Fatalf("response leaked password hash: %s", res.Body.String())
	}
	var body struct {
		User publicUser `json:"user"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.User != (publicUser{ID: user.ID, Email: user.Email, Name: user.Name}) {
		t.Fatalf("user = %#v", body.User)
	}
}

func TestMeRejectsMissingContext(t *testing.T) {
	h := &AuthHandler{}
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)

	h.Me(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", res.Code)
	}
}

func TestLogoutIsIdempotentAndExpiresCookie(t *testing.T) {
	for _, withCookie := range []bool{false, true} {
		t.Run(map[bool]string{false: "without cookie", true: "with invalid cookie"}[withCookie], func(t *testing.T) {
			h := &AuthHandler{}
			res := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
			if withCookie {
				req.AddCookie(&http.Cookie{Name: "token", Value: "invalid"})
			}

			h.Logout(res, req)

			if res.Code != http.StatusOK {
				t.Fatalf("status = %d", res.Code)
			}
			cookies := res.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("cookies = %v", cookies)
			}
			cookie := cookies[0]
			if cookie.Name != "token" || cookie.Value != "" || cookie.Path != "/" {
				t.Fatalf("cookie identity = %#v", cookie)
			}
			if cookie.MaxAge >= 0 || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
				t.Fatalf("cookie attributes = %#v", cookie)
			}
			if !cookie.Expires.Before(time.Now()) {
				t.Fatalf("expires = %v", cookie.Expires)
			}
		})
	}
}
