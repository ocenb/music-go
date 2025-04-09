package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ocenb/music-go/content-service/internal/utils"
)

type BeginTx interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

func WithTransaction(ctx context.Context, repo BeginTx, fn func(txCtx context.Context) error) error {
	tx, err := repo.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var txErr error
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && txErr == nil {
			txErr = fmt.Errorf("tx err: %v, rb err: %v", txErr, rbErr)
		}
	}()

	txCtx := utils.SetTxToContext(ctx, tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
