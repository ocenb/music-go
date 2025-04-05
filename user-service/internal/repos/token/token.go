package token

import (
	"context"
	"database/sql"
	"time"

	"github.com/ocenb/music-go/user-service/internal/models"
	"github.com/ocenb/music-go/user-service/internal/utils"

	_ "github.com/lib/pq"
)

type TokenRepoInterface interface {
	GetTokenByID(ctx context.Context, tokenID string) (*models.TokenModel, error)
	CreateToken(ctx context.Context, tokenID string, userID int64, refreshToken string, expiresAt time.Time) error
	DeleteTokenByID(ctx context.Context, tokenID string) error
	DeleteAllUserTokens(ctx context.Context, userID int64) error
	DeleteExpiredTokens(ctx context.Context) error
}

type TokenRepo struct {
	postgres *sql.DB
}

func NewTokenRepo(postgres *sql.DB) TokenRepoInterface {
	return &TokenRepo{postgres: postgres}
}

func (r *TokenRepo) GetTokenByID(ctx context.Context, tokenID string) (*models.TokenModel, error) {
	query := `SELECT id, user_id, refresh_token, expires_at FROM tokens WHERE id = $1`

	tx, hasTx := utils.GetTxFromContext(ctx)
	if hasTx {
		row := tx.QueryRowContext(ctx, query, tokenID)

		var token models.TokenModel
		err := row.Scan(&token.ID, &token.UserId, &token.RefreshToken, &token.ExpiresAt)
		if err != nil {
			return nil, err
		}

		return &token, nil
	}

	row := r.postgres.QueryRowContext(ctx, query, tokenID)

	var token models.TokenModel
	err := row.Scan(&token.ID, &token.UserId, &token.RefreshToken, &token.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepo) CreateToken(ctx context.Context, tokenID string, userID int64, refreshToken string, expiresAt time.Time) error {
	query := `INSERT INTO tokens (id, user_id, refresh_token, expires_at) VALUES ($1, $2, $3, $4)`

	tx, hasTx := utils.GetTxFromContext(ctx)
	if hasTx {
		_, err := tx.ExecContext(ctx, query, tokenID, userID, refreshToken, expiresAt)
		return err
	}

	_, err := r.postgres.ExecContext(ctx, query, tokenID, userID, refreshToken, expiresAt)
	return err
}

func (r *TokenRepo) DeleteTokenByID(ctx context.Context, tokenID string) error {
	query := `DELETE FROM tokens WHERE id = $1`

	tx, hasTx := utils.GetTxFromContext(ctx)
	if hasTx {
		_, err := tx.ExecContext(ctx, query, tokenID)
		return err
	}

	_, err := r.postgres.ExecContext(ctx, query, tokenID)
	return err
}

func (r *TokenRepo) DeleteAllUserTokens(ctx context.Context, userID int64) error {
	query := `DELETE FROM tokens WHERE user_id = $1`

	tx, hasTx := utils.GetTxFromContext(ctx)
	if hasTx {
		_, err := tx.ExecContext(ctx, query, userID)
		return err
	}

	_, err := r.postgres.ExecContext(ctx, query, userID)
	return err
}

func (r *TokenRepo) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM tokens WHERE expires_at < $1`

	tx, hasTx := utils.GetTxFromContext(ctx)
	if hasTx {
		_, err := tx.ExecContext(ctx, query, time.Now())
		return err
	}

	_, err := r.postgres.ExecContext(ctx, query, time.Now())
	return err
}
