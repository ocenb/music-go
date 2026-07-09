package playlist

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/models"
	"github.com/ocenb/music-go/content-service/internal/storage/transactor"
)

type Repo struct {
	tm transactor.Querier
}

func New(tm transactor.Querier) *Repo {
	return &Repo{tm: tm}
}

func (r *Repo) GetByID(ctx context.Context, playlistID int64, currentUserID int64) (*models.PlaylistWithSavedModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT p.id, p.user_id, p.title, p.changeable_id, p.image, p.created_at, p.updated_at,
			CASE WHEN usp.user_id IS NOT NULL THEN true ELSE false END as is_saved,
			usp.added_at as saved_at
		FROM playlists p
		LEFT JOIN user_saved_playlists usp ON usp.playlist_id = p.id AND usp.user_id = $1
		WHERE p.id = $2
	`

	return scanPlaylistWithSaved(q.QueryRow(ctx, query, currentUserID, playlistID))
}

func (r *Repo) GetByChangeableID(ctx context.Context, username, changeableID string, currentUserID int64) (*models.PlaylistWithSavedModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT p.id, p.user_id, p.title, p.changeable_id, p.image, p.created_at, p.updated_at,
			CASE WHEN usp.user_id IS NOT NULL THEN true ELSE false END as is_saved,
			usp.added_at as saved_at
		FROM playlists p
		LEFT JOIN user_saved_playlists usp ON usp.playlist_id = p.id AND usp.user_id = $1
		WHERE p.changeable_id = $2 AND p.username = $3
	`

	return scanPlaylistWithSaved(q.QueryRow(ctx, query, currentUserID, changeableID, username))
}

func (r *Repo) GetMany(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*models.PlaylistWithSavedModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT p.id, p.user_id, p.title, p.changeable_id, p.image, p.created_at, p.updated_at,
			CASE WHEN usp.user_id IS NOT NULL THEN true ELSE false END as is_saved,
			usp.added_at as saved_at
		FROM playlists p
		LEFT JOIN user_saved_playlists usp ON usp.playlist_id = p.id AND usp.user_id = $1
		WHERE p.user_id = $2 AND ($3 = 0 OR p.id < $3)
		ORDER BY p.id DESC
		LIMIT $4
	`

	rows, err := q.Query(ctx, query, currentUserID, userID, lastID, take)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlists: %w", err)
	}
	defer rows.Close()

	return collectPlaylistWithSaved(rows)
}

func (r *Repo) GetManyWithSaved(ctx context.Context, userID int64, take int, lastID int64) ([]*models.PlaylistWithSavedModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		WITH my_playlists AS (
			SELECT p.id, p.user_id, p.title, p.changeable_id, p.image, p.created_at, p.updated_at,
				false as is_saved, NULL::timestamp as saved_at, p.created_at as sort_date
			FROM playlists p
			WHERE p.user_id = $1 AND ($2 = 0 OR p.id < $2)
		),
		saved_playlists AS (
			SELECT p.id, p.user_id, p.title, p.changeable_id, p.image, p.created_at, p.updated_at,
				true as is_saved, usp.added_at as saved_at, COALESCE(usp.added_at, p.created_at) as sort_date
			FROM playlists p
			JOIN user_saved_playlists usp ON p.id = usp.playlist_id
			WHERE usp.user_id = $1 AND ($2 = 0 OR p.id < $2)
		)
		SELECT id, user_id, title, changeable_id, image, created_at, updated_at, is_saved, saved_at FROM my_playlists
		UNION ALL
		SELECT id, user_id, title, changeable_id, image, created_at, updated_at, is_saved, saved_at FROM saved_playlists
		ORDER BY 9 DESC NULLS LAST, 6 DESC
		LIMIT $3
	`

	rows, err := q.Query(ctx, query, userID, lastID, take)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlists with saved: %w", err)
	}
	defer rows.Close()

	return collectPlaylistWithSaved(rows)
}

func (r *Repo) Create(ctx context.Context, userID int64, username, title, changeableID, image string) (*models.PlaylistModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		INSERT INTO playlists (user_id, username, title, changeable_id, image)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, username, title, changeable_id, image, created_at, updated_at
	`

	var playlist models.PlaylistModel
	var createdAt, updatedAt time.Time

	err := q.QueryRow(ctx, query, userID, username, title, changeableID, image).Scan(
		&playlist.ID,
		&playlist.UserID,
		&playlist.Username,
		&playlist.Title,
		&playlist.ChangeableID,
		&playlist.Image,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}

	playlist.CreatedAt = createdAt
	playlist.UpdatedAt = updatedAt

	return &playlist, nil
}

func (r *Repo) CheckPermission(ctx context.Context, userID, playlistID int64) (bool, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM playlists
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	if err := q.QueryRow(ctx, query, playlistID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check playlist permission: %w", err)
	}

	return exists, nil
}

func (r *Repo) Delete(ctx context.Context, playlistID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		DELETE FROM playlists
		WHERE id = $1
	`

	if _, err := q.Exec(ctx, query, playlistID); err != nil {
		return fmt.Errorf("failed to delete playlist: %w", err)
	}

	return nil
}

func (r *Repo) ChangeTitle(ctx context.Context, playlistID int64, title string) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE playlists
		SET title = $1
		WHERE id = $2
	`

	if _, err := q.Exec(ctx, query, title, playlistID); err != nil {
		return fmt.Errorf("failed to change playlist title: %w", err)
	}

	return nil
}

func (r *Repo) ChangeChangeableID(ctx context.Context, playlistID int64, changeableID string) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE playlists
		SET changeable_id = $1
		WHERE id = $2
	`

	if _, err := q.Exec(ctx, query, changeableID, playlistID); err != nil {
		return fmt.Errorf("failed to change playlist changeable id: %w", err)
	}

	return nil
}

func (r *Repo) ChangeImage(ctx context.Context, playlistID int64, image string) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE playlists
		SET image = $1
		WHERE id = $2
	`

	if _, err := q.Exec(ctx, query, image, playlistID); err != nil {
		return fmt.Errorf("failed to change playlist image: %w", err)
	}

	return nil
}

func (r *Repo) CheckTitle(ctx context.Context, userID int64, title string) (bool, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM playlists
			WHERE user_id = $1 AND title = $2
		)
	`

	var exists bool
	if err := q.QueryRow(ctx, query, userID, title).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check playlist title: %w", err)
	}

	return exists, nil
}

func (r *Repo) CheckChangeableID(ctx context.Context, userID int64, changeableID string) (bool, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM playlists
			WHERE user_id = $1 AND changeable_id = $2
		)
	`

	var exists bool
	if err := q.QueryRow(ctx, query, userID, changeableID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check playlist changeable id: %w", err)
	}

	return exists, nil
}

func (r *Repo) SavePlaylist(ctx context.Context, userID, playlistID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		INSERT INTO user_saved_playlists (user_id, playlist_id, added_at)
		VALUES ($1, $2, $3)
	`

	if _, err := q.Exec(ctx, query, userID, playlistID, time.Now()); err != nil {
		return fmt.Errorf("failed to save playlist: %w", err)
	}

	return nil
}

func (r *Repo) RemoveFromSaved(ctx context.Context, userID, playlistID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		DELETE FROM user_saved_playlists
		WHERE user_id = $1 AND playlist_id = $2
	`

	if _, err := q.Exec(ctx, query, userID, playlistID); err != nil {
		return fmt.Errorf("failed to remove playlist from saved: %w", err)
	}

	return nil
}

func scanPlaylistWithSaved(row pgx.Row) (*models.PlaylistWithSavedModel, error) {
	var playlist models.PlaylistWithSavedModel
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&playlist.ID,
		&playlist.UserID,
		&playlist.Title,
		&playlist.ChangeableID,
		&playlist.Image,
		&createdAt,
		&updatedAt,
		&playlist.IsSaved,
		&playlist.SavedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrPlaylistNotFound
		}
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}

	playlist.CreatedAt = createdAt
	playlist.UpdatedAt = updatedAt

	return &playlist, nil
}

func collectPlaylistWithSaved(rows pgx.Rows) ([]*models.PlaylistWithSavedModel, error) {
	var playlists []*models.PlaylistWithSavedModel

	for rows.Next() {
		var playlist models.PlaylistWithSavedModel
		var createdAt, updatedAt time.Time

		if err := rows.Scan(
			&playlist.ID,
			&playlist.UserID,
			&playlist.Title,
			&playlist.ChangeableID,
			&playlist.Image,
			&createdAt,
			&updatedAt,
			&playlist.IsSaved,
			&playlist.SavedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan playlist: %w", err)
		}

		playlist.CreatedAt = createdAt
		playlist.UpdatedAt = updatedAt

		playlists = append(playlists, &playlist)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate playlists: %w", err)
	}

	return playlists, nil
}
