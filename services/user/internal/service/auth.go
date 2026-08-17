package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/qiuliw/discord-clone/services/user/internal/domain"
)

// 常量配置
const (
	CookieName = "token"             // HTTP Cookie 存储令牌的键名
	TTL        = 30 * 24 * time.Hour // JWT 有效期 30天
	bcryptCost = 12                  // bcrypt 哈希计算开销，越高越安全、加密越慢
	issuer     = "discord-clone"     // JWT签发者标识
)

// Auth 用户认证业务服务
// 负责注册、登录、密码哈希、JWT签发与解析
type Auth struct {
	users  domain.UserRepository // 用户仓储层，持久化读写用户数据
	secret []byte                // JWT HMAC-SHA256 签名密钥
}

// NewAuth 构造认证服务实例
func NewAuth(users domain.UserRepository, secret []byte) *Auth {
	return &Auth{users: users, secret: secret}
}

// LoadOrCreateSecret 加载JWT签名密钥
// 优先级：环境变量 JWT_SECRET > 本地密钥文件 > 自动生成并持久化密钥文件
// 目的：服务重启后仍然可以正常校验历史签发的token，避免用户全部掉线
func LoadOrCreateSecret(path string) ([]byte, error) {
	// 优先读取环境变量密钥
	if env := os.Getenv("JWT_SECRET"); env != "" {
		return []byte(env), nil
	}
	// 读取本地已有密钥文件，要求密钥长度至少32字节
	if data, err := os.ReadFile(path); err == nil && len(data) >= 32 {
		return data, nil
	}
	// 生成32字节安全随机密钥
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	// 创建密钥存放目录
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	// 写入密钥文件，权限0600：仅当前用户可读可写
	if err := os.WriteFile(path, secret, 0o600); err != nil {
		return nil, err
	}
	return secret, nil
}

// RegisterInput 用户注册入参
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// Claims JWT自定义载荷
type Claims struct {
	UserID               int64 `json:"uid"` // 用户唯一ID
	jwt.RegisteredClaims       // JWT标准声明(签发时间、过期时间、签发者等)
}

// Register 用户注册
// 完成邮箱格式校验、密码合法性校验、密码bcrypt哈希，创建用户记录
func (a *Auth) Register(in RegisterInput) (domain.User, error) {
	// 统一小写并去除首尾空格
	email := strings.ToLower(strings.TrimSpace(in.Email))
	// 校验邮箱格式
	if _, err := mail.ParseAddress(email); err != nil {
		return domain.User{}, domain.ErrInvalidInput
	}

	// bcrypt算法限制最多处理前72字节密码
	// 限制长度防止超长密码产生预期外行为，同时规定最小长度
	if utf8.RuneCountInString(in.Password) < 8 || len(in.Password) > 72 {
		return domain.User{}, domain.ErrInvalidInput
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return domain.User{}, domain.ErrInvalidInput
	}

	// 生成密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return domain.User{}, err
	}

	// 仓储层新建用户
	user, err := a.users.Create(email, name, string(hash))
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

// Login 用户登录校验
// 根据邮箱查询用户，比对bcrypt密码哈希
// 安全策略：邮箱不存在 / 密码错误返回相同错误，防止攻击者枚举有效邮箱
func (a *Auth) Login(email, password string) (domain.User, error) {
	user, err := a.users.GetByEmail(strings.ToLower(strings.TrimSpace(email)))
	// 用户不存在 或 密码不匹配，统一返回凭证错误
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	return user, nil
}

// UserFromToken 通过JWT令牌获取完整用户信息
// 1. 解析token拿到userID
// 2. 使用userID查询数据库，防止token有效但账号已被删除
func (a *Auth) UserFromToken(token string) (domain.User, error) {
	userID, err := a.parseJWT(token)
	if err != nil {
		return domain.User{}, err
	}
	user, err := a.users.GetByID(userID)
	if err != nil {
		// 用户已删除，视为令牌无效
		if errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, domain.ErrInvalidToken
		}
		return domain.User{}, err
	}
	return user, nil
}

// SignToken 签发JWT令牌
// 根据用户ID生成带过期时间的HS256签名token
// 返回签名字符串与过期时间，供上层设置Cookie使用
func (a *Auth) SignToken(userID int64) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(TTL)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	// 使用HS256算法签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(a.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// parseJWT 解析并校验JWT，提取用户ID
// 校验签名算法、签名合法性、载荷有效性
func (a *Auth) parseJWT(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		// 防御算法切换攻击：强制只允许HS256
		if t.Method != jwt.SigningMethodHS256 {
			return nil, domain.ErrInvalidToken
		}
		return a.secret, nil
	})
	if err != nil {
		return 0, domain.ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	// 校验token有效、载荷合法、用户ID大于0
	if !ok || !token.Valid || claims.UserID <= 0 {
		return 0, domain.ErrInvalidToken
	}
	return claims.UserID, nil
}
