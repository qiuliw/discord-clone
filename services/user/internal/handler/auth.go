package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/qiuliw/discord-clone/services/user/internal/domain"
	"github.com/qiuliw/discord-clone/services/user/internal/service"
)

type AuthHandler struct {
	auth *service.Auth
}

func NewAuthHandler(auth *service.Auth) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/auth/me", h.Me)
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type publicUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func toPublic(u domain.User) publicUser {
	return publicUser{ID: u.ID, Email: u.Email, Name: u.Name}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	user, err := h.auth.Register(service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid email or password")
		case errors.Is(err, domain.ErrEmailTaken):
			writeError(w, http.StatusConflict, "email already registered")
		default:
			writeError(w, http.StatusInternalServerError, "could not create account")
		}
		return
	}

	token, err := h.issueToken(w, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create token")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"user":  toPublic(user),
		"token": token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	user, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := h.issueToken(w, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user":  toPublic(user),
		"token": token,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     service.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	raw, ok := bearerOrCookie(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "not signed in")
		return
	}
	user, err := h.auth.UserFromToken(raw)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not signed in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": toPublic(user)})
}

func (h *AuthHandler) issueToken(w http.ResponseWriter, userID int64) (string, error) {
	token, exp, err := h.auth.SignToken(userID)
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     service.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  exp,
		MaxAge:   int(time.Until(exp).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return token, nil
}

func bearerOrCookie(r *http.Request) (string, bool) {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if token != "" {
			return token, true
		}
	}
	if c, err := r.Cookie(service.CookieName); err == nil && c.Value != "" {
		return c.Value, true
	}
	return "", false
}
