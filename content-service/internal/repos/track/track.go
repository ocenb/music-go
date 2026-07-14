package track

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

func (r *Repo) GetByID(ctx context.Context, trackID, currentUserID int64) (*models.TrackWithLikedModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT t.id, t.user_id, t.username, t.title, t.changeable_id, t.audio, t.image, t.duration, t.plays, t.created_at, t.updated_at,
			CASE WHEN ult.user_id IS NOT NULL THEN true ELSE false END as is_liked
		FROM tracks t
		LEFT JOIN user_liked_tracks ult ON ult.track_id = t.id AND ult.user_id = $1
		WHERE t.id = $2
	`

	return scanTrackWithLiked(q.QueryRow(ctx, query, currentUserID, trackID))
}

func (r *Repo) GetByChangeableID(ctx context.Context, username, changeableID string, currentUserID int64) (*models.TrackWithLikedModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT t.id, t.user_id, t.username, t.title, t.changeable_id, t.audio, t.image, t.duration, t.plays, t.created_at, t.updated_at,
			CASE WHEN ult.user_id IS NOT NULL THEN true ELSE false END as is_liked
		FROM tracks t
		LEFT JOIN user_liked_tracks ult ON ult.track_id = t.id AND ult.user_id = $1
		WHERE t.changeable_id = $2 AND t.username = $3
	`

	return scanTrackWithLiked(q.QueryRow(ctx, query, currentUserID, changeableID, username))
}

func (r *Repo) GetMany(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*models.TrackWithLikedModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT t.id, t.user_id, t.username, t.title, t.changeable_id, t.audio, t.image, t.duration, t.plays, t.created_at, t.updated_at,
			CASE WHEN ult.user_id IS NOT NULL THEN true ELSE false END as is_liked
		FROM tracks t
		LEFT JOIN user_liked_tracks ult ON ult.track_id = t.id AND ult.user_id = $1
		WHERE t.user_id = $2 AND ($3 = 0 OR t.id < $3)
		ORDER BY t.id DESC
		LIMIT $4
	`

	rows, err := q.Query(ctx, query, currentUserID, userID, lastID, take)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracks: %w", err)
	}
	defer rows.Close()

	return collectTrackWithLiked(rows)
}

func (r *Repo) GetManyPopular(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*models.TrackWithLikedModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT t.id, t.user_id, t.username, t.title, t.changeable_id, t.audio, t.image, t.duration, t.plays, t.created_at, t.updated_at,
			CASE WHEN ult.user_id IS NOT NULL THEN true ELSE false END as is_liked
		FROM tracks t
		LEFT JOIN user_liked_tracks ult ON ult.track_id = t.id AND ult.user_id = $1
		WHERE t.user_id = $2 AND ($3 = 0 OR t.id < $3)
		ORDER BY t.plays DESC, t.id DESC
		LIMIT $4
	`

	rows, err := q.Query(ctx, query, currentUserID, userID, lastID, take)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular tracks: %w", err)
	}
	defer rows.Close()

	return collectTrackWithLiked(rows)
}

func (r *Repo) Create(ctx context.Context, userID int64, username, title, changeableID, audio, image string, duration int64) (*models.TrackModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		INSERT INTO tracks (user_id, username, title, changeable_id, audio, image, duration)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, username, title, changeable_id, audio, image, duration, plays, created_at, updated_at
	`

	var track models.TrackModel
	var createdAt, updatedAt time.Time

	err := q.QueryRow(ctx, query, userID, username, title, changeableID, audio, image, duration).Scan(
		&track.ID,
		&track.UserID,
		&track.Username,
		&track.Title,
		&track.ChangeableID,
		&track.Audio,
		&track.Image,
		&track.Duration,
		&track.Plays,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create track: %w", err)
	}

	track.CreatedAt = createdAt
	track.UpdatedAt = updatedAt

	return &track, nil
}

func (r *Repo) AddPlay(ctx context.Context, trackID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE tracks
		SET plays = plays + 1
		WHERE id = $1
	`

	if _, err := q.Exec(ctx, query, trackID); err != nil {
		return fmt.Errorf("failed to add play: %w", err)
	}

	return nil
}

func (r *Repo) CheckPermission(ctx context.Context, userID, trackID int64) (bool, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM tracks
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	if err := q.QueryRow(ctx, query, trackID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check track permission: %w", err)
	}

	return exists, nil
}

func (r *Repo) Delete(ctx context.Context, trackID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		DELETE FROM tracks
		WHERE id = $1
	`

	if _, err := q.Exec(ctx, query, trackID); err != nil {
		return fmt.Errorf("failed to delete track: %w", err)
	}

	return nil
}

func (r *Repo) ChangeTitle(ctx context.Context, trackID int64, title string) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE tracks
		SET title = $1
		WHERE id = $2
	`

	if _, err := q.Exec(ctx, query, title, trackID); err != nil {
		return fmt.Errorf("failed to change track title: %w", err)
	}

	return nil
}

func (r *Repo) ChangeChangeableID(ctx context.Context, trackID int64, changeableID string) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE tracks
		SET changeable_id = $1
		WHERE id = $2
	`

	if _, err := q.Exec(ctx, query, changeableID, trackID); err != nil {
		return fmt.Errorf("failed to change track changeable id: %w", err)
	}

	return nil
}

func (r *Repo) ChangeImage(ctx context.Context, trackID int64, image string) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE tracks
		SET image = $1
		WHERE id = $2
	`

	if _, err := q.Exec(ctx, query, image, trackID); err != nil {
		return fmt.Errorf("failed to change track image: %w", err)
	}

	return nil
}

func (r *Repo) CheckTitle(ctx context.Context, userID int64, title string) (bool, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM tracks
			WHERE user_id = $1 AND title = $2
		)
	`

	var exists bool
	if err := q.QueryRow(ctx, query, userID, title).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check track title: %w", err)
	}

	return exists, nil
}

func (r *Repo) CheckChangeableID(ctx context.Context, userID int64, changeableID string) (bool, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM tracks
			WHERE user_id = $1 AND changeable_id = $2
		)
	`

	var exists bool
	if err := q.QueryRow(ctx, query, userID, changeableID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check track changeable id: %w", err)
	}

	return exists, nil
}

func (r *Repo) GetManyLiked(ctx context.Context, currentUserID int64) ([]*models.UserLikedTrackModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT user_id, track_id, added_at
		FROM user_liked_tracks
		WHERE user_id = $1
		ORDER BY added_at DESC
	`

	rows, err := q.Query(ctx, query, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get liked tracks: %w", err)
	}
	defer rows.Close()

	var likedTracks []*models.UserLikedTrackModel
	for rows.Next() {
		model := &models.UserLikedTrackModel{}
		if err := rows.Scan(&model.UserID, &model.TrackID, &model.AddedAt); err != nil {
			return nil, fmt.Errorf("failed to scan liked track: %w", err)
		}
		likedTracks = append(likedTracks, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate liked tracks: %w", err)
	}

	return likedTracks, nil
}

func (r *Repo) AddToLiked(ctx context.Context, currentUserID, trackID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		INSERT INTO user_liked_tracks (user_id, track_id, added_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, track_id) DO UPDATE
		SET added_at = $3
	`

	if _, err := q.Exec(ctx, query, currentUserID, trackID, time.Now()); err != nil {
		return fmt.Errorf("failed to add track to liked: %w", err)
	}

	return nil
}

func (r *Repo) RemoveFromLiked(ctx context.Context, currentUserID, trackID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `DELETE FROM user_liked_tracks WHERE user_id = $1 AND track_id = $2`

	if _, err := q.Exec(ctx, query, currentUserID, trackID); err != nil {
		return fmt.Errorf("failed to remove track from liked: %w", err)
	}

	return nil
}

func scanTrackWithLiked(row pgx.Row) (*models.TrackWithLikedModel, error) {
	var track models.TrackWithLikedModel
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&track.ID,
		&track.UserID,
		&track.Username,
		&track.Title,
		&track.ChangeableID,
		&track.Audio,
		&track.Image,
		&track.Duration,
		&track.Plays,
		&createdAt,
		&updatedAt,
		&track.IsLiked,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrTrackNotFound
		}
		return nil, fmt.Errorf("failed to get track: %w", err)
	}

	track.CreatedAt = createdAt
	track.UpdatedAt = updatedAt

	return &track, nil
}

func collectTrackWithLiked(rows pgx.Rows) ([]*models.TrackWithLikedModel, error) {
	var tracks []*models.TrackWithLikedModel

	for rows.Next() {
		var track models.TrackWithLikedModel
		var createdAt, updatedAt time.Time

		if err := rows.Scan(
			&track.ID,
			&track.UserID,
			&track.Username,
			&track.Title,
			&track.ChangeableID,
			&track.Audio,
			&track.Image,
			&track.Duration,
			&track.Plays,
			&createdAt,
			&updatedAt,
			&track.IsLiked,
		); err != nil {
			return nil, fmt.Errorf("failed to scan track: %w", err)
		}

		track.CreatedAt = createdAt
		track.UpdatedAt = updatedAt

		tracks = append(tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tracks: %w", err)
	}

	return tracks, nil
}
