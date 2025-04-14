package all

import (
	"context"
	"database/sql"
	"log/slog"
)

type AllRepoInterface interface {
	DeleteAll(ctx context.Context, userID int64) ([]string, []string, []string, error)
}

type AllRepo struct {
	postgres *sql.DB
	log      *slog.Logger
}

func NewAllRepo(postgres *sql.DB, log *slog.Logger) AllRepoInterface {
	return &AllRepo{postgres: postgres, log: log}
}

func (r *AllRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.postgres.BeginTx(ctx, opts)
}

func (r *AllRepo) DeleteAll(ctx context.Context, userID int64) ([]string, []string, []string, error) {
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		r.log.Error("Failed to begin transaction", "error", err)
		return nil, nil, nil, err
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			r.log.Error("Failed to rollback transaction", "error", err)
		}
	}()

	var trackAudios []string
	rows, err := tx.QueryContext(ctx, "SELECT audio FROM tracks WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to get track audios", "error", err, "user_id", userID)
		return nil, nil, nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.log.Error("Failed to close rows", "error", err)
		}
	}()

	for rows.Next() {
		var audio string
		if err := rows.Scan(&audio); err != nil {
			r.log.Error("Failed to scan track audio", "error", err)
			return nil, nil, nil, err
		}
		trackAudios = append(trackAudios, audio)
	}
	if err := rows.Err(); err != nil {
		r.log.Error("Error iterating track audios", "error", err)
		return nil, nil, nil, err
	}

	var trackImages []string
	rows, err = tx.QueryContext(ctx, "SELECT image FROM tracks WHERE user_id = $1 AND image != 'default'", userID)
	if err != nil {
		r.log.Error("Failed to get track images", "error", err, "user_id", userID)
		return nil, nil, nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.log.Error("Failed to close rows", "error", err)
		}
	}()

	for rows.Next() {
		var image string
		if err := rows.Scan(&image); err != nil {
			r.log.Error("Failed to scan track image", "error", err)
			return nil, nil, nil, err
		}
		trackImages = append(trackImages, image)
	}
	if err := rows.Err(); err != nil {
		r.log.Error("Error iterating track images", "error", err)
		return nil, nil, nil, err
	}

	var playlistImages []string
	rows, err = tx.QueryContext(ctx, "SELECT image FROM playlists WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to get playlist images", "error", err, "user_id", userID)
		return nil, nil, nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.log.Error("Failed to close rows", "error", err)
		}
	}()

	for rows.Next() {
		var image string
		if err := rows.Scan(&image); err != nil {
			r.log.Error("Failed to scan playlist image", "error", err)
			return nil, nil, nil, err
		}
		playlistImages = append(playlistImages, image)
	}
	if err := rows.Err(); err != nil {
		r.log.Error("Error iterating playlist images", "error", err)
		return nil, nil, nil, err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM listening_history WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to delete listening history", "error", err, "user_id", userID)
		return nil, nil, nil, err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM user_liked_tracks WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to delete user liked tracks", "error", err, "user_id", userID)
		return nil, nil, nil, err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM user_saved_playlists WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to delete user saved playlists", "error", err, "user_id", userID)
		return nil, nil, nil, err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM playlists WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to delete playlists", "error", err, "user_id", userID)
		return nil, nil, nil, err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM tracks WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to delete tracks", "error", err, "user_id", userID)
		return nil, nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		r.log.Error("Failed to commit transaction", "error", err)
		return nil, nil, nil, err
	}

	return trackAudios, trackImages, playlistImages, nil
}
