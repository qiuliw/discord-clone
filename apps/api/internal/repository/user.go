package repository

import (
	"database/sql"
	"time"

	"github.com/qiuliw/discord-clone/apps/api/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(email, name, passwordHash string) (*model.User, error) {
	res, err := r.db.Exec(
		`INSERT INTO users (email, name, password_hash) VALUES (?, ?, ?)`,
		email, name, passwordHash,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	return scanUser(r.db.QueryRow(
		`SELECT id, email, name, password_hash, created_at FROM users WHERE email = ?`,
		email,
	))
}

func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	return scanUser(r.db.QueryRow(
		`SELECT id, email, name, password_hash, created_at FROM users WHERE id = ?`,
		id,
	))
}

func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var created string
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &created); err != nil {
		return nil, err
	}
	if t, err := time.Parse("2006-01-02 15:04:05", created); err == nil {
		u.CreatedAt = t
	}
	return &u, nil
}
