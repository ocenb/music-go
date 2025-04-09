package track

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

type TrackRepoInterface interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	GetByID(ctx context.Context, trackID int64, currentUserID int64) (*TrackWithLikedModel, error)
	GetByChangeableID(ctx context.Context, username, changeableID string, currentUserID int64) (*TrackWithLikedModel, error)
	GetMany(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*TrackWithLikedModel, error)
	GetManyPopular(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*TrackWithLikedModel, error)
	Create(ctx context.Context, userID int64, username, title, changeableID, audio, image string, duration int64) (*TrackModel, error)
	AddPlay(ctx context.Context, trackID int64) error
	CheckPermission(ctx context.Context, userID, trackID int64) (bool, error)
	Delete(ctx context.Context, trackID int64) error
	ChangeTitle(ctx context.Context, trackID int64, title string) error
	ChangeChangeableID(ctx context.Context, trackID int64, changeableID string) error
	ChangeImage(ctx context.Context, trackID int64, image string) error
	CheckTitle(ctx context.Context, userID int64, title string) (bool, error)
	CheckChangeableID(ctx context.Context, userID int64, changeableID string) (bool, error)
	GetManyLiked(ctx context.Context, currentUserID int64) ([]*UserLikedTrackModel, error)
	AddToLiked(ctx context.Context, currentUserID, trackID int64) error
	RemoveFromLiked(ctx context.Context, currentUserID, trackID int64) error
}

type TrackRepo struct {
	postgres *sql.DB
	log      *slog.Logger
}

func NewTrackRepo(postgres *sql.DB, log *slog.Logger) TrackRepoInterface {
	return &TrackRepo{postgres: postgres, log: log}
}

func (r *TrackRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.postgres.BeginTx(ctx, opts)
}

func (r *TrackRepo) GetByID(ctx context.Context, trackID int64, currentUserID int64) (*TrackWithLikedModel, error) {
	query := `
		SELECT t.id, t.user_id, t.username, t.title, t.changeable_id, t.audio, t.image, t.duration, t.plays, t.created_at, t.updated_at,
			CASE WHEN ult.user_id IS NOT NULL THEN true ELSE false END as is_liked
		FROM tracks t
		LEFT JOIN user_liked_tracks ult ON ult.track_id = t.id AND ult.user_id = $1
		WHERE t.id = $2
	`

	var track TrackWithLikedModel
	var createdAt, updatedAt time.Time

	err := r.postgres.QueryRowContext(ctx, query, currentUserID, trackID).Scan(
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
		return nil, err
	}

	track.CreatedAt = createdAt
	track.UpdatedAt = updatedAt

	return &track, nil
}

func (r *TrackRepo) GetByChangeableID(ctx context.Context, username, changeableID string, currentUserID int64) (*TrackWithLikedModel, error) {
	query := `
		SELECT t.id, t.user_id, t.username, t.title, t.changeable_id, t.audio, t.image, t.duration, t.plays, t.created_at, t.updated_at,
			CASE WHEN ult.user_id IS NOT NULL THEN true ELSE false END as is_liked
		FROM tracks t
		LEFT JOIN user_liked_tracks ult ON ult.track_id = t.id AND ult.user_id = $1
		WHERE t.changeable_id = $2 AND t.username = $3
	`

	var track TrackWithLikedModel
	var createdAt, updatedAt time.Time

	err := r.postgres.QueryRowContext(ctx, query, currentUserID, changeableID, username).Scan(
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
		return nil, err
	}

	track.CreatedAt = createdAt
	track.UpdatedAt = updatedAt

	return &track, nil
}

func (r *TrackRepo) GetMany(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*TrackWithLikedModel, error) {
	query := `
		SELECT t.id, t.user_id, t.username, t.title, t.changeable_id, t.audio, t.image, t.duration, t.plays, t.created_at, t.updated_at,
			CASE WHEN ult.user_id IS NOT NULL THEN true ELSE false END as is_liked
		FROM tracks t
		LEFT JOIN user_liked_tracks ult ON ult.track_id = t.id AND ult.user_id = $1
		WHERE t.user_id = $2 AND ($3 = 0 OR t.id < $3)
		ORDER BY t.id DESC
		LIMIT $4
	`

	rows, err := r.postgres.QueryContext(ctx, query, currentUserID, userID, lastID, take)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			r.log.Error("Failed to close rows", "error", err)
		}
	}()

	var tracks []*TrackWithLikedModel

	for rows.Next() {
		var track TrackWithLikedModel
		var createdAt, updatedAt time.Time

		err := rows.Scan(
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
			return nil, err
		}

		track.CreatedAt = createdAt
		track.UpdatedAt = updatedAt

		tracks = append(tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tracks, nil
}

func (r *TrackRepo) GetManyPopular(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*TrackWithLikedModel, error) {
	query := `
		SELECT t.id, t.user_id, t.username, t.title, t.changeable_id, t.audio, t.image, t.duration, t.plays, t.created_at, t.updated_at,
			CASE WHEN ult.user_id IS NOT NULL THEN true ELSE false END as is_liked
		FROM tracks t
		LEFT JOIN user_liked_tracks ult ON ult.track_id = t.id AND ult.user_id = $1
		WHERE t.user_id = $2 AND ($3 = 0 OR t.id < $3)
		ORDER BY t.plays DESC, t.id DESC
		LIMIT $4
	`

	rows, err := r.postgres.QueryContext(ctx, query, currentUserID, userID, lastID, take)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			r.log.Error("Failed to close rows", "error", err)
		}
	}()

	var tracks []*TrackWithLikedModel

	for rows.Next() {
		var track TrackWithLikedModel
		var createdAt, updatedAt time.Time

		err := rows.Scan(
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
			return nil, err
		}

		track.CreatedAt = createdAt
		track.UpdatedAt = updatedAt

		tracks = append(tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tracks, nil
}

func (r *TrackRepo) Create(ctx context.Context, userID int64, username, title, changeableID, audio, image string, duration int64) (*TrackModel, error) {
	query := `
		INSERT INTO tracks (user_id, username, title, changeable_id, audio, image, duration)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, username, title, changeable_id, audio, image, duration, plays, created_at, updated_at
	`

	var track TrackModel
	var createdAt, updatedAt time.Time

	err := r.postgres.QueryRowContext(
		ctx, query, userID, username, title, changeableID, audio, image, duration,
	).Scan(
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
		return nil, err
	}

	track.CreatedAt = createdAt
	track.UpdatedAt = updatedAt

	return &track, nil
}

func (r *TrackRepo) AddPlay(ctx context.Context, trackID int64) error {
	query := `
		UPDATE tracks
		SET plays = plays + 1
		WHERE id = $1
	`

	_, err := r.postgres.ExecContext(ctx, query, trackID)
	return err
}

func (r *TrackRepo) CheckPermission(ctx context.Context, userID, trackID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM tracks
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.postgres.QueryRowContext(ctx, query, trackID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *TrackRepo) Delete(ctx context.Context, trackID int64) error {
	query := `
		DELETE FROM tracks
		WHERE id = $1
	`

	_, err := r.postgres.ExecContext(ctx, query, trackID)
	return err
}

func (r *TrackRepo) ChangeTitle(ctx context.Context, trackID int64, title string) error {
	query := `
		UPDATE tracks
		SET title = $1
		WHERE id = $2
	`

	_, err := r.postgres.ExecContext(ctx, query, title, trackID)
	return err
}

func (r *TrackRepo) ChangeChangeableID(ctx context.Context, trackID int64, changeableID string) error {
	query := `
		UPDATE tracks
		SET changeable_id = $1
		WHERE id = $2
	`

	_, err := r.postgres.ExecContext(ctx, query, changeableID, trackID)
	return err
}

func (r *TrackRepo) ChangeImage(ctx context.Context, trackID int64, image string) error {
	query := `
		UPDATE tracks
		SET image = $1
		WHERE id = $2
	`

	_, err := r.postgres.ExecContext(ctx, query, image, trackID)
	return err
}

func (r *TrackRepo) CheckTitle(ctx context.Context, userID int64, title string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM tracks
			WHERE user_id = $1 AND title = $2
		)
	`

	var exists bool
	err := r.postgres.QueryRowContext(ctx, query, userID, title).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *TrackRepo) CheckChangeableID(ctx context.Context, userID int64, changeableID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM tracks
			WHERE user_id = $1 AND changeable_id = $2
		)
	`

	var exists bool
	err := r.postgres.QueryRowContext(ctx, query, userID, changeableID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *TrackRepo) GetManyLiked(ctx context.Context, currentUserID int64) ([]*UserLikedTrackModel, error) {
	query := `
		SELECT user_id, track_id, added_at
		FROM user_liked_tracks
		WHERE user_id = $1
		ORDER BY added_at DESC
	`

	rows, err := r.postgres.QueryContext(ctx, query, currentUserID)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			r.log.Error("Failed to close rows", "error", err)
		}
	}()

	var likedTracks []*UserLikedTrackModel
	for rows.Next() {
		model := &UserLikedTrackModel{}
		err := rows.Scan(&model.UserID, &model.TrackID, &model.AddedAt)
		if err != nil {
			return nil, err
		}
		likedTracks = append(likedTracks, model)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return likedTracks, nil
}

func (r *TrackRepo) AddToLiked(ctx context.Context, currentUserID, trackID int64) error {
	query := `
		INSERT INTO user_liked_tracks (user_id, track_id, added_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, track_id) DO UPDATE
		SET added_at = $3
	`

	_, err := r.postgres.ExecContext(ctx, query, currentUserID, trackID, time.Now())
	return err
}

func (r *TrackRepo) RemoveFromLiked(ctx context.Context, currentUserID, trackID int64) error {
	query := `DELETE FROM user_liked_tracks WHERE user_id = $1 AND track_id = $2`

	_, err := r.postgres.ExecContext(ctx, query, currentUserID, trackID)
	return err
}
