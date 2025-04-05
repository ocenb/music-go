package auth

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type AuthRepoInterface interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type AuthRepo struct {
	postgres *sql.DB
}

func NewAuthRepo(postgres *sql.DB) AuthRepoInterface {
	return &AuthRepo{postgres: postgres}
}

func (r *AuthRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.postgres.BeginTx(ctx, opts)
}
