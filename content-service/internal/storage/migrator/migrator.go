package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver for goose migrations
	"github.com/pressly/goose/v3"
)

type Migrator struct {
	db *sql.DB
}

func New(ctx context.Context, dsn string, fsys fs.FS, log goose.Logger) (*Migrator, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres for migrations: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres for migrations: %w", err)
	}

	goose.SetLogger(log)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}
	goose.SetBaseFS(fsys)

	return &Migrator{db}, nil
}

func (m *Migrator) Up() error {
	if err := goose.Up(m.db, "."); err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}
	return nil
}

func (m *Migrator) Close() error {
	return m.db.Close()
}
