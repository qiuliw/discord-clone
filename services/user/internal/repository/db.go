package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

	// _pragma=foreign_keys(1) 开启外键约束
	// _pragma=busy_timeout(5000) 设置超时
	database, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	// sqlite 天然单写，串行。并行会无用竞争
	database.SetMaxOpenConns(1)

	if err := migrateUp(database); err != nil {
		_ = database.Close()
		return nil, err
	}

	return database, nil
}

// migrateUp 执行数据库迁移。迁移语句使用 IF NOT EXISTS，
// 全新库与已存在 users 表的旧库都能幂等通过。
func migrateUp(database *sql.DB) error {
	// 来自嵌入的文件系统
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}
	// 注册 migrate 的 sqlite 驱动
	driver, err := sqlite.WithInstance(database, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("migration driver: %w", err)
	}
	// 创建 migrate 实例
	// sqlc.yaml 的文件系统源与 sqlite 的驱动实例
	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	// 执行迁移
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
