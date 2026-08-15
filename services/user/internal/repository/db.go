package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"

	"github.com/qiuliw/discord-clone/services/user/migrations"
)

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	database, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(1)

	if err := migrateUp(database); err != nil {
		_ = database.Close()
		return nil, err
	}

	return database, nil
}

func migrateUp(database *sql.DB) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}
	driver, err := sqlite.WithInstance(database, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("migration driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		if !alreadyExists(err) {
			return fmt.Errorf("migrate up: %w", err)
		}
		if _, _, vErr := m.Version(); errors.Is(vErr, migrate.ErrNilVersion) {
			if err := m.Force(1); err != nil {
				return fmt.Errorf("migrate baseline: %w", err)
			}
			return nil
		}
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func alreadyExists(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "already exists")
}
