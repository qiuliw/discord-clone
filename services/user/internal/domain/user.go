package domain

import "errors"

type User struct {
	ID           int64
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    string
}

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidToken       = errors.New("invalid token")
	ErrNotFound           = errors.New("user not found")
)

type UserRepository interface {
	Create(email, name, passwordHash string) (User, error)
	GetByEmail(email string) (User, error)
	GetByID(id int64) (User, error)
}
