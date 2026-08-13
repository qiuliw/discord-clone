package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/qiuliw/discord-clone/apps/api/internal/model"
	"github.com/qiuliw/discord-clone/apps/api/internal/repository"
)

const (
	CookieName = "session"
	TTL        = 30 * 24 * time.Hour
	bcryptCost = 12
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidSession     = errors.New("invalid session")
)

type Auth struct {
	users  *repository.UserRepository
	secret []byte
}

func NewAuth(users *repository.UserRepository, secret []byte) *Auth {
	return &Auth{users: users, secret: secret}
}

func LoadOrCreateSecret(path string) ([]byte, error) {
	if env := os.Getenv("SESSION_SECRET"); env != "" {
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

func (a *Auth) Register(in RegisterInput) (*model.User, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidInput
	}
	if utf8.RuneCountInString(in.Password) < 8 || len(in.Password) > 72 {
		return nil, ErrInvalidInput
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = strings.SplitN(email, "@", 2)[0]
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return nil, err
	}

	user, err := a.users.Create(email, name, string(hash))
	if err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return user, nil
}

func (a *Auth) Login(email, password string) (*model.User, error) {
	user, err := a.users.GetByEmail(strings.ToLower(strings.TrimSpace(email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (a *Auth) UserFromToken(token string) (*model.User, error) {
	userID, err := parseToken(a.secret, token)
	if err != nil {
		return nil, err
	}
	user, err := a.users.GetByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidSession
		}
		return nil, err
	}
	return user, nil
}

func (a *Auth) SignToken(userID int64) (string, time.Time) {
	exp := time.Now().Add(TTL)
	return signToken(a.secret, userID, exp), exp
}

func signToken(secret []byte, userID int64, exp time.Time) string {
	payload := fmt.Sprintf("%d.%d", userID, exp.Unix())
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig
}

func parseToken(secret []byte, token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, ErrInvalidSession
	}
	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return 0, ErrInvalidSession
	}
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, ErrInvalidSession
	}
	expUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, ErrInvalidSession
	}
	if time.Now().Unix() > expUnix {
		return 0, ErrInvalidSession
	}
	return userID, nil
}
