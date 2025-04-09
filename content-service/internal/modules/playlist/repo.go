package playlist

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

type PlaylistRepoInterface interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	GetByID(ctx context.Context, playlistID int64, currentUserID int64) (*PlaylistWithSavedModel, error)
	GetByChangeableID(ctx context.Context, username, changeableID string, currentUserID int64) (*PlaylistWithSavedModel, error)
	GetMany(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*PlaylistWithSavedModel, error)
	Create(ctx context.Context, userID int64, username, title, changeableID, image string) (*PlaylistModel, error)
	CheckPermission(ctx context.Context, userID, playlistID int64) (bool, error)
	Delete(ctx context.Context, playlistID int64) error
	ChangeTitle(ctx context.Context, playlistID int64, title string) error
	ChangeChangeableID(ctx context.Context, playlistID int64, changeableID string) error
	ChangeImage(ctx context.Context, playlistID int64, image string) error
	CheckTitle(ctx context.Context, userID int64, title string) (bool, error)
	CheckChangeableID(ctx context.Context, userID int64, changeableID string) (bool, error)
	SavePlaylist(ctx context.Context, userID, playlistID int64) error
	RemoveFromSaved(ctx context.Context, userID, playlistID int64) error
	GetManySaved(ctx context.Context, userID int64, take int, lastID int64) ([]*UserSavedPlaylistModel, error)
	GetManyWithSaved(ctx context.Context, userID int64, take int, lastID int64) ([]*PlaylistWithSavedModel, error)
}

type PlaylistRepo struct {
	postgres *sql.DB
	log      *slog.Logger
}

func NewPlaylistRepo(postgres *sql.DB, log *slog.Logger) PlaylistRepoInterface {
	return &PlaylistRepo{postgres: postgres, log: log}
}

func (r *PlaylistRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.postgres.BeginTx(ctx, opts)
}

func (r *PlaylistRepo) GetByID(ctx context.Context, playlistID int64, currentUserID int64) (*PlaylistWithSavedModel, error) {
	query := `
		SELECT p.id, p.user_id, p.title, p.changeable_id, p.image, p.created_at, p.updated_at,
			CASE WHEN usp.user_id IS NOT NULL THEN true ELSE false END as is_saved,
			usp.added_at as saved_at
		FROM playlists p
		LEFT JOIN user_saved_playlists usp ON usp.playlist_id = p.id AND usp.user_id = $1
		WHERE p.id = $2
	`

	var playlist PlaylistWithSavedModel
	var createdAt, updatedAt time.Time
	var savedAt sql.NullTime

	err := r.postgres.QueryRowContext(ctx, query, currentUserID, playlistID).Scan(
		&playlist.ID,
		&playlist.UserID,
		&playlist.Title,
		&playlist.ChangeableID,
		&playlist.Image,
		&createdAt,
		&updatedAt,
		&playlist.IsSaved,
		&savedAt,
	)

	if err != nil {
		return nil, err
	}

	playlist.CreatedAt = createdAt
	playlist.UpdatedAt = updatedAt
	if savedAt.Valid {
		playlist.SavedAt = &savedAt.Time
	}

	return &playlist, nil
}

func (r *PlaylistRepo) GetByChangeableID(ctx context.Context, username, changeableID string, currentUserID int64) (*PlaylistWithSavedModel, error) {
	query := `
		SELECT p.id, p.user_id, p.title, p.changeable_id, p.image, p.created_at, p.updated_at,
			CASE WHEN usp.user_id IS NOT NULL THEN true ELSE false END as is_saved,
			usp.added_at as saved_at
		FROM playlists p
		LEFT JOIN user_saved_playlists usp ON usp.playlist_id = p.id AND usp.user_id = $1
		WHERE p.changeable_id = $2 AND p.username = $3
	`

	var playlist PlaylistWithSavedModel
	var createdAt, updatedAt time.Time
	var savedAt sql.NullTime

	err := r.postgres.QueryRowContext(ctx, query, currentUserID, changeableID, username).Scan(
		&playlist.ID,
		&playlist.UserID,
		&playlist.Title,
		&playlist.ChangeableID,
		&playlist.Image,
		&createdAt,
		&updatedAt,
		&playlist.IsSaved,
		&savedAt,
	)

	if err != nil {
		return nil, err
	}

	playlist.CreatedAt = createdAt
	playlist.UpdatedAt = updatedAt
	if savedAt.Valid {
		playlist.SavedAt = &savedAt.Time
	}

	return &playlist, nil
}

func (r *PlaylistRepo) GetMany(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*PlaylistWithSavedModel, error) {
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

	var playlists []*PlaylistWithSavedModel

	for rows.Next() {
		var playlist PlaylistWithSavedModel
		var createdAt, updatedAt time.Time
		var savedAt sql.NullTime

		err := rows.Scan(
			&playlist.ID,
			&playlist.UserID,
			&playlist.Title,
			&playlist.ChangeableID,
			&playlist.Image,
			&createdAt,
			&updatedAt,
			&playlist.IsSaved,
			&savedAt,
		)

		if err != nil {
			return nil, err
		}

		playlist.CreatedAt = createdAt
		playlist.UpdatedAt = updatedAt
		if savedAt.Valid {
			playlist.SavedAt = &savedAt.Time
		}

		playlists = append(playlists, &playlist)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return playlists, nil
}

func (r *PlaylistRepo) Create(ctx context.Context, userID int64, username, title, changeableID, image string) (*PlaylistModel, error) {
	query := `
		INSERT INTO playlists (user_id, username, title, changeable_id, image)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, username, title, changeable_id, image, created_at, updated_at
	`

	var playlist PlaylistModel
	var createdAt, updatedAt time.Time

	err := r.postgres.QueryRowContext(
		ctx, query, userID, username, title, changeableID, image,
	).Scan(
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
		return nil, err
	}

	playlist.CreatedAt = createdAt
	playlist.UpdatedAt = updatedAt

	return &playlist, nil
}

func (r *PlaylistRepo) CheckPermission(ctx context.Context, userID, playlistID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM playlists
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.postgres.QueryRowContext(ctx, query, playlistID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *PlaylistRepo) Delete(ctx context.Context, playlistID int64) error {
	query := `
		DELETE FROM playlists
		WHERE id = $1
	`

	_, err := r.postgres.ExecContext(ctx, query, playlistID)
	return err
}

func (r *PlaylistRepo) ChangeTitle(ctx context.Context, playlistID int64, title string) error {
	query := `
		UPDATE playlists
		SET title = $1
		WHERE id = $2
	`

	_, err := r.postgres.ExecContext(ctx, query, title, playlistID)
	return err
}

func (r *PlaylistRepo) ChangeChangeableID(ctx context.Context, playlistID int64, changeableID string) error {
	query := `
		UPDATE playlists
		SET changeable_id = $1
		WHERE id = $2
	`

	_, err := r.postgres.ExecContext(ctx, query, changeableID, playlistID)
	return err
}

func (r *PlaylistRepo) ChangeImage(ctx context.Context, playlistID int64, image string) error {
	query := `
		UPDATE playlists
		SET image = $1
		WHERE id = $2
	`

	_, err := r.postgres.ExecContext(ctx, query, image, playlistID)
	return err
}

func (r *PlaylistRepo) CheckTitle(ctx context.Context, userID int64, title string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM playlists
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

func (r *PlaylistRepo) CheckChangeableID(ctx context.Context, userID int64, changeableID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM playlists
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

func (r *PlaylistRepo) SavePlaylist(ctx context.Context, userID, playlistID int64) error {
	query := `
		INSERT INTO user_saved_playlists (user_id, playlist_id, added_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.postgres.ExecContext(ctx, query, userID, playlistID, time.Now())
	return err
}

func (r *PlaylistRepo) RemoveFromSaved(ctx context.Context, userID, playlistID int64) error {
	query := `
		DELETE FROM user_saved_playlists
		WHERE user_id = $1 AND playlist_id = $2
	`

	_, err := r.postgres.ExecContext(ctx, query, userID, playlistID)
	return err
}

func (r *PlaylistRepo) GetManyWithSaved(ctx context.Context, userID int64, take int, lastID int64) ([]*PlaylistWithSavedModel, error) {
	query := `
		WITH my_playlists AS (
			SELECT p.id, p.user_id, p.title, p.changeable_id, p.image, p.created_at, p.updated_at,
				false as is_saved, NULL as saved_at
			FROM playlists p
			WHERE p.user_id = $1 AND ($2 = 0 OR p.id < $2)
		),
		saved_playlists AS (
			SELECT p.id, p.user_id, p.title, p.changeable_id, p.image, p.created_at, p.updated_at,
				true as is_saved, usp.added_at as saved_at
			FROM playlists p
			JOIN user_saved_playlists usp ON p.id = usp.playlist_id
			WHERE usp.user_id = $1 AND ($2 = 0 OR p.id < $2)
		)
		SELECT * FROM my_playlists
		UNION ALL
		SELECT * FROM saved_playlists
		ORDER BY COALESCE(saved_at, created_at) DESC
		LIMIT $3
	`

	rows, err := r.postgres.QueryContext(ctx, query, userID, lastID, take)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			r.log.Error("Failed to close rows", "error", err)
		}
	}()

	var playlists []*PlaylistWithSavedModel

	for rows.Next() {
		var playlist PlaylistWithSavedModel
		var createdAt, updatedAt time.Time
		var savedAt sql.NullTime

		err := rows.Scan(
			&playlist.ID,
			&playlist.UserID,
			&playlist.Title,
			&playlist.ChangeableID,
			&playlist.Image,
			&createdAt,
			&updatedAt,
			&playlist.IsSaved,
			&savedAt,
		)

		if err != nil {
			return nil, err
		}

		playlist.CreatedAt = createdAt
		playlist.UpdatedAt = updatedAt
		if savedAt.Valid {
			playlist.SavedAt = &savedAt.Time
		}

		playlists = append(playlists, &playlist)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return playlists, nil
}

func (r *PlaylistRepo) GetManySaved(ctx context.Context, userID int64, take int, lastID int64) ([]*UserSavedPlaylistModel, error) {
	query := `
		SELECT usp.user_id, usp.playlist_id, usp.added_at
		FROM user_saved_playlists usp
		WHERE usp.user_id = $1 AND ($2 = 0 OR usp.playlist_id < $2)
		ORDER BY usp.added_at DESC
		LIMIT $3
	`

	rows, err := r.postgres.QueryContext(ctx, query, userID, lastID, take)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			r.log.Error("Failed to close rows", "error", err)
		}
	}()

	var savedPlaylists []*UserSavedPlaylistModel
	for rows.Next() {
		model := &UserSavedPlaylistModel{}
		err := rows.Scan(&model.UserID, &model.PlaylistID, &model.AddedAt)
		if err != nil {
			return nil, err
		}
		savedPlaylists = append(savedPlaylists, model)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return savedPlaylists, nil
}
