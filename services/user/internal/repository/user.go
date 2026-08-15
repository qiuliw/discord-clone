package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/qiuliw/discord-clone/services/user/internal/domain"
	"github.com/qiuliw/discord-clone/services/user/internal/repository/db"
)

type UserRepository struct {
	q *db.Queries
}

func NewUserRepository(database *sql.DB) *UserRepository {
	return &UserRepository{q: db.New(database)}
}

func (r *UserRepository) Create(email, name, passwordHash string) (domain.User, error) {
	u, err := r.q.CreateUser(context.Background(), db.CreateUserParams{
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, err
	}
	return toDomain(u), nil
}

func (r *UserRepository) GetByEmail(email string) (domain.User, error) {
	u, err := r.q.GetUserByEmail(context.Background(), email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return toDomain(u), nil
}

func (r *UserRepository) GetByID(id int64) (domain.User, error) {
	u, err := r.q.GetUserByID(context.Background(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return toDomain(u), nil
}

func toDomain(u db.User) domain.User {
	return domain.User{
		ID:           u.ID,
		Email:        u.Email,
		Name:         u.Name,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
}
