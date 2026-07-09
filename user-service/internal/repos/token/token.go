package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/models"
	"github.com/ocenb/music-go/user-service/internal/storage/transactor"
)

type Repo struct {
	tm transactor.Querier
}

func New(tm transactor.Querier) *Repo {
	return &Repo{tm: tm}
}

func (r *Repo) GetTokenByID(ctx context.Context, tokenID string) (*models.TokenModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `SELECT id, user_id, refresh_token, expires_at::text, created_at::text FROM tokens WHERE id = $1`

	var token models.TokenModel
	err := q.QueryRow(ctx, query, tokenID).Scan(&token.ID, &token.UserID, &token.RefreshToken, &token.ExpiresAt, &token.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get token by id: %w", err)
	}

	return &token, nil
}

func (r *Repo) CreateToken(ctx context.Context, tokenID string, userID int64, refreshToken string, expiresAt time.Time) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `INSERT INTO tokens (id, user_id, refresh_token, expires_at) VALUES ($1, $2, $3, $4)`
	if _, err := q.Exec(ctx, query, tokenID, userID, refreshToken, expiresAt); err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	return nil
}

func (r *Repo) DeleteTokenByID(ctx context.Context, tokenID string) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `DELETE FROM tokens WHERE id = $1`
	cmdTag, err := q.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errs.ErrTokenNotFound
	}

	return nil
}

func (r *Repo) DeleteAllUserTokens(ctx context.Context, userID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `DELETE FROM tokens WHERE user_id = $1`
	if _, err := q.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("failed to delete user tokens: %w", err)
	}

	return nil
}

func (r *Repo) DeleteExpiredTokens(ctx context.Context) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `DELETE FROM tokens WHERE expires_at < $1`
	if _, err := q.Exec(ctx, query, time.Now()); err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return nil
}
