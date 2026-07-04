package history

import (
	"context"
	"fmt"
	"time"

	"github.com/ocenb/music-go/content-service/internal/models"
	"github.com/ocenb/music-go/content-service/internal/storage/transactor"
)

type Repo struct {
	tm *transactor.Manager
}

func New(tm *transactor.Manager) *Repo {
	return &Repo{tm: tm}
}

func (r *Repo) Get(ctx context.Context, currentUserID int64, take int64) ([]*models.ListeningHistoryModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT user_id, track_id, played_at
		FROM listening_history
		WHERE user_id = $1
		ORDER BY played_at DESC
		LIMIT $2
	`

	rows, err := q.Query(ctx, query, currentUserID, take)
	if err != nil {
		return nil, fmt.Errorf("failed to get listening history: %w", err)
	}
	defer rows.Close()

	var history []*models.ListeningHistoryModel
	for rows.Next() {
		model := &models.ListeningHistoryModel{}
		if err := rows.Scan(&model.UserID, &model.TrackID, &model.PlayedAt); err != nil {
			return nil, fmt.Errorf("failed to scan listening history: %w", err)
		}
		history = append(history, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate listening history: %w", err)
	}

	return history, nil
}

func (r *Repo) Add(ctx context.Context, currentUserID, trackID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		INSERT INTO listening_history (user_id, track_id, played_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, track_id) DO UPDATE
		SET played_at = $3
	`

	if _, err := q.Exec(ctx, query, currentUserID, trackID, time.Now()); err != nil {
		return fmt.Errorf("failed to add listening history: %w", err)
	}

	return nil
}

func (r *Repo) Clear(ctx context.Context, currentUserID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `DELETE FROM listening_history WHERE user_id = $1`

	if _, err := q.Exec(ctx, query, currentUserID); err != nil {
		return fmt.Errorf("failed to clear listening history: %w", err)
	}

	return nil
}
