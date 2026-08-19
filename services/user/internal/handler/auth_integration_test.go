package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/qiuliw/discord-clone/services/user/internal/repository"
	"github.com/qiuliw/discord-clone/services/user/internal/service"
)

// newIntegrationRouter 用真实 SQLite 数据库搭建完整的注册/登录处理链。
func newIntegrationRouter(t *testing.T) http.Handler {
	t.Helper()
	db, err := repository.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	// 32 字节密钥，满足 HMAC-SHA256 签名要求。
	secret := []byte("0123456789abcdef0123456789abcdef")
	auth := service.NewAuth(repository.NewUserRepository(db), secret)
	return New(NewAuthHandler(auth))
}

func postJSON(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}

// TestRegisterLoginRoundTrip 覆盖注册、重复注册、登录、错误密码的完整链路，
// 确认请求路径本身不会崩溃（早期问题表现为代理层 ECONNRESET）。
func TestRegisterLoginRoundTrip(t *testing.T) {
	h := newIntegrationRouter(t)

	res := postJSON(t, h, "/api/auth/register", `{"email":"a@b.com","password":"password123","name":"Alice"}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", res.Code, res.Body.String())
	}

	res = postJSON(t, h, "/api/auth/register", `{"email":"a@b.com","password":"password123","name":"Alice"}`)
	if res.Code != http.StatusConflict {
		t.Fatalf("duplicate register status = %d, body = %s", res.Code, res.Body.String())
	}

	res = postJSON(t, h, "/api/auth/login", `{"email":"A@B.COM","password":"password123"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", res.Code, res.Body.String())
	}

	res = postJSON(t, h, "/api/auth/login", `{"email":"a@b.com","password":"wrong-password"}`)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("bad login status = %d, body = %s", res.Code, res.Body.String())
	}
}
