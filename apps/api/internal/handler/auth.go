package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/qiuliw/discord-clone/apps/api/internal/model"
	"github.com/qiuliw/discord-clone/apps/api/internal/service"
)

type AuthHandler struct {
	auth *service.Auth
}

func NewAuthHandler(auth *service.Auth) *AuthHandler {
	return &AuthHandler{auth: auth}
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

func toPublic(u *model.User) publicUser {
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
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid email or password")
		case errors.Is(err, service.ErrEmailTaken):
			writeError(w, http.StatusConflict, "email already registered")
		default:
			writeError(w, http.StatusInternalServerError, "could not create account")
		}
		return
	}

	h.setSession(w, user.ID)
	writeJSON(w, http.StatusCreated, map[string]any{"user": toPublic(user)})
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

	h.setSession(w, user.ID)
	writeJSON(w, http.StatusOK, map[string]any{"user": toPublic(user)})
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
	c, err := r.Cookie(service.CookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not signed in")
		return
	}
	user, err := h.auth.UserFromToken(c.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not signed in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": toPublic(user)})
}

func (h *AuthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) setSession(w http.ResponseWriter, userID int64) {
	token, exp := h.auth.SignToken(userID)
	http.SetCookie(w, &http.Cookie{
		Name:     service.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  exp,
		MaxAge:   int(time.Until(exp).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
