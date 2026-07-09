package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver for goose migrations
	"github.com/pressly/goose/v3"
)

type Migrator interface {
	Up(ctx context.Context) error
	Close() error
}

type migrator struct {
	db       *sql.DB
	provider *goose.Provider
}

func New(ctx context.Context, dsn string, fsys fs.FS, log goose.Logger) (Migrator, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres for migrations: %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping postgres for migrations: %w", err)
	}

	opts := []goose.ProviderOption{}
	if log != nil {
		opts = append(opts, goose.WithLogger(log))
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys, opts...)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create migration provider: %w", err)
	}

	return &migrator{db: db, provider: provider}, nil
}

func (m *migrator) Up(ctx context.Context) error {
	if _, err := m.provider.Up(ctx); err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}
	return nil
}

func (m *migrator) Close() error {
	return m.db.Close()
}
