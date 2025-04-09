package history

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

type HistoryRepoInterface interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Get(ctx context.Context, currentUserID int64, take int64) ([]*ListeningHistoryModel, error)
	Add(ctx context.Context, currentUserID, trackID int64) error
	Clear(ctx context.Context, currentUserID int64) error
}

type HistoryRepo struct {
	postgres *sql.DB
	log      *slog.Logger
}

func NewHistoryRepo(postgres *sql.DB, log *slog.Logger) HistoryRepoInterface {
	return &HistoryRepo{postgres: postgres, log: log}
}

func (r *HistoryRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.postgres.BeginTx(ctx, opts)
}

func (r *HistoryRepo) Get(ctx context.Context, currentUserID int64, take int64) ([]*ListeningHistoryModel, error) {
	query := `
		SELECT user_id, track_id, played_at
		FROM listening_history
		WHERE user_id = $1
		ORDER BY played_at DESC
		LIMIT $2
	`

	rows, err := r.postgres.QueryContext(ctx, query, currentUserID, take)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			r.log.Error("Failed to close rows", "error", err)
		}
	}()

	var history []*ListeningHistoryModel
	for rows.Next() {
		model := &ListeningHistoryModel{}
		err := rows.Scan(&model.UserID, &model.TrackID, &model.PlayedAt)
		if err != nil {
			return nil, err
		}
		history = append(history, model)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

func (r *HistoryRepo) Add(ctx context.Context, currentUserID, trackID int64) error {
	query := `
		INSERT INTO listening_history (user_id, track_id, played_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, track_id) DO UPDATE
		SET played_at = $3
	`

	_, err := r.postgres.ExecContext(ctx, query, currentUserID, trackID, time.Now())
	return err
}

func (r *HistoryRepo) Clear(ctx context.Context, currentUserID int64) error {
	query := `DELETE FROM listening_history WHERE user_id = $1`

	_, err := r.postgres.ExecContext(ctx, query, currentUserID)
	return err
}
