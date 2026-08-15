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

const (
	CookieName = "token"
	TTL        = 30 * 24 * time.Hour
	bcryptCost = 12
	issuer     = "discord-clone"
)

type Auth struct {
	users  domain.UserRepository
	secret []byte
}

func NewAuth(users domain.UserRepository, secret []byte) *Auth {
	return &Auth{users: users, secret: secret}
}

func LoadOrCreateSecret(path string) ([]byte, error) {
	if env := os.Getenv("JWT_SECRET"); env != "" {
		return []byte(env), nil
	}
	if data, err := os.ReadFile(path); err == nil && len(data) >= 32 {
		return data, nil
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, secret, 0o600); err != nil {
		return nil, err
	}
	return secret, nil
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type Claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

func (a *Auth) Register(in RegisterInput) (domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return domain.User{}, domain.ErrInvalidInput
	}
	if utf8.RuneCountInString(in.Password) < 8 || len(in.Password) > 72 {
		return domain.User{}, domain.ErrInvalidInput
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = strings.SplitN(email, "@", 2)[0]
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return domain.User{}, err
	}

	user, err := a.users.Create(email, name, string(hash))
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (a *Auth) Login(email, password string) (domain.User, error) {
	user, err := a.users.GetByEmail(strings.ToLower(strings.TrimSpace(email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	return user, nil
}

func (a *Auth) UserFromToken(token string) (domain.User, error) {
	userID, err := a.parseJWT(token)
	if err != nil {
		return domain.User{}, err
	}
	user, err := a.users.GetByID(userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, domain.ErrInvalidToken
		}
		return domain.User{}, err
	}
	return user, nil
}

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
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(a.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

func (a *Auth) parseJWT(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, domain.ErrInvalidToken
		}
		return a.secret, nil
	})
	if err != nil {
		return 0, domain.ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.UserID <= 0 {
		return 0, domain.ErrInvalidToken
	}
	return claims.UserID, nil
}
