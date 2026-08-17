package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/qiuliw/discord-clone/services/user/internal/domain"
	usermiddleware "github.com/qiuliw/discord-clone/services/user/internal/middleware"
	"github.com/qiuliw/discord-clone/services/user/internal/service"
)

// AuthHandler 认证相关HTTP处理器
// 处理注册、登录、登出、获取当前用户等鉴权接口
type AuthHandler struct {
	auth *service.Auth // 认证业务服务层实例
}

// NewAuthHandler 创建认证处理器构造函数
func NewAuthHandler(auth *service.Auth) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Mount 将所有认证路由注册到mux路由器
func (h *AuthHandler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.Register) // 用户注册
	mux.HandleFunc("POST /api/auth/login", h.Login)       // 用户登录
	mux.HandleFunc("POST /api/auth/logout", h.Logout)     // 用户登出
	mux.Handle(
		"GET /api/auth/me",
		usermiddleware.RequireAuth(h.auth, service.CookieName)(http.HandlerFunc(h.Me)),
	) // 获取当前登录用户信息
}

// authRequest 注册/登录统一请求结构体
// 登录场景仅使用 Email、Password；注册场景全部字段生效
type authRequest struct {
	Email    string `json:"email"`    // 用户邮箱
	Password string `json:"password"` // 用户明文密码
	Name     string `json:"name"`     // 用户昵称，注册必填
}

// publicUser 对外返回的用户脱敏模型
// 剔除密码、内部敏感字段，仅暴露公开信息
type publicUser struct {
	ID    int64  `json:"id"`    // 用户ID
	Email string `json:"email"` // 用户邮箱
	Name  string `json:"name"`  // 用户昵称
}

// toPublic 将领域用户实体转换为对外脱敏用户模型
func toPublic(u domain.User) publicUser {
	return publicUser{ID: u.ID, Email: u.Email, Name: u.Name}
}

// Register 用户注册接口
// 请求：POST /api/auth/register
// 接收邮箱、密码、昵称，创建账号并签发身份令牌
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
			writeError(w, http.StatusBadRequest, "invalid name, email, or password")
		case errors.Is(err, domain.ErrEmailTaken):
			writeError(w, http.StatusConflict, "email already registered")
		default:
			writeError(w, http.StatusInternalServerError, "could not create account")
		}
		return
	}

	// 签发身份令牌并写入cookie
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

// Login 用户登录接口
// 请求：POST /api/auth/login
// 校验邮箱密码，验证通过后返回用户信息与访问令牌
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

// Logout 用户登出接口
// 请求：POST /api/auth/logout
// 清除浏览器HttpOnly会话Cookie，前端仍需自行销毁本地存储token
func (h *AuthHandler) Logout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     service.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1, // MaxAge=-1 立即删除cookie
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// Me 获取当前登录用户信息接口
// 请求：GET /api/auth/me
// 用户由 RequireAuth 中间件验证后写入请求上下文
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := usermiddleware.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not signed in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": toPublic(user)})
}

// issueToken 生成身份令牌
// 1. 调用service签发token
// 2. 设置HttpOnly Cookie供浏览器SPA自动携带
// 返回token字符串，方便前端通过Authorization请求头手动使用
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
		HttpOnly: true, // 禁止JS读取，防范XSS窃取令牌
		SameSite: http.SameSiteLaxMode,
	})
	return token, nil
}
