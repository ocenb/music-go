package playlisttracks

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/ocenb/music-go/content-service/internal/modules/track"
)

type PlaylistTracksRepoInterface interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	GetMany(ctx context.Context, playlistID, currentUserID int64, take int) ([]*TrackInPlaylistModel, error)
	Add(ctx context.Context, playlistID, trackID int64, position int) (*PlaylistTrackModel, error)
	UpdatePosition(ctx context.Context, playlistID, trackID int64, position int) error
	Remove(ctx context.Context, playlistID, trackID int64) error
	GetOne(ctx context.Context, playlistID, trackID int64) (*PlaylistTrackModel, error)
	GetLastPosition(ctx context.Context, playlistID int64) (int, error)
	IncrementPositions(ctx context.Context, playlistID int64, fromPosition int) error
	DecrementPositions(ctx context.Context, playlistID int64, fromPosition int) error
	MovePositions(ctx context.Context, playlistID int64, fromPosition, toPosition int) error
}

type PlaylistTracksRepo struct {
	postgres *sql.DB
	log      *slog.Logger
}

func NewPlaylistTracksRepo(postgres *sql.DB, log *slog.Logger) PlaylistTracksRepoInterface {
	return &PlaylistTracksRepo{postgres: postgres, log: log}
}

func (r *PlaylistTracksRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.postgres.BeginTx(ctx, opts)
}

func (r *PlaylistTracksRepo) GetMany(ctx context.Context, playlistID, currentUserID int64, take int) ([]*TrackInPlaylistModel, error) {
	query := `
		SELECT pt.track_id, pt.position, pt.added_at,
			t.id, t.user_id, t.username, t.title, t.changeable_id, t.audio, t.image, t.duration, t.plays, t.created_at, t.updated_at,
			CASE WHEN ult.user_id IS NOT NULL THEN true ELSE false END as is_liked
		FROM playlist_tracks pt
		JOIN tracks t ON pt.track_id = t.id
		LEFT JOIN user_liked_tracks ult ON ult.track_id = t.id AND ult.user_id = $1
		WHERE pt.playlist_id = $2
		ORDER BY pt.position ASC
	`

	if take > 0 {
		query += " LIMIT $3"
	}

	var rows *sql.Rows
	var err error

	if take > 0 {
		rows, err = r.postgres.QueryContext(ctx, query, currentUserID, playlistID, take)
	} else {
		rows, err = r.postgres.QueryContext(ctx, query, currentUserID, playlistID)
	}

	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			r.log.Error("Failed to close rows", "error", err)
		}
	}()

	var tracks []*TrackInPlaylistModel

	for rows.Next() {
		var trackInPlaylist TrackInPlaylistModel
		var trackModel track.TrackWithLikedModel
		var createdAt, updatedAt, addedAt time.Time

		err := rows.Scan(
			&trackInPlaylist.TrackID,
			&trackInPlaylist.Position,
			&addedAt,
			&trackModel.ID,
			&trackModel.UserID,
			&trackModel.Username,
			&trackModel.Title,
			&trackModel.ChangeableID,
			&trackModel.Audio,
			&trackModel.Image,
			&trackModel.Duration,
			&trackModel.Plays,
			&createdAt,
			&updatedAt,
			&trackModel.IsLiked,
		)

		if err != nil {
			return nil, err
		}

		trackModel.CreatedAt = createdAt
		trackModel.UpdatedAt = updatedAt

		trackInPlaylist.PlaylistID = playlistID
		trackInPlaylist.Title = trackModel.Title
		trackInPlaylist.Artist = trackModel.Username
		trackInPlaylist.Duration = int(trackModel.Duration)
		trackInPlaylist.CoverImagePath = trackModel.Image
		trackInPlaylist.CreatedAt = addedAt

		tracks = append(tracks, &trackInPlaylist)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tracks, nil
}

func (r *PlaylistTracksRepo) Add(ctx context.Context, playlistID, trackID int64, position int) (*PlaylistTrackModel, error) {
	query := `
		INSERT INTO playlist_tracks (playlist_id, track_id, position, added_at)
		VALUES ($1, $2, $3, $4)
		RETURNING playlist_id, track_id, position, added_at
	`

	var model PlaylistTrackModel
	var addedAt time.Time

	err := r.postgres.QueryRowContext(
		ctx, query, playlistID, trackID, position, time.Now(),
	).Scan(
		&model.PlaylistID,
		&model.TrackID,
		&model.Position,
		&addedAt,
	)

	if err != nil {
		return nil, err
	}

	model.AddedAt = addedAt

	return &model, nil
}

func (r *PlaylistTracksRepo) UpdatePosition(ctx context.Context, playlistID, trackID int64, position int) error {
	query := `
		UPDATE playlist_tracks
		SET position = $1
		WHERE playlist_id = $2 AND track_id = $3
	`

	_, err := r.postgres.ExecContext(ctx, query, position, playlistID, trackID)
	return err
}

func (r *PlaylistTracksRepo) Remove(ctx context.Context, playlistID, trackID int64) error {
	query := `
		DELETE FROM playlist_tracks
		WHERE playlist_id = $1 AND track_id = $2
		RETURNING position
	`

	var position int
	err := r.postgres.QueryRowContext(ctx, query, playlistID, trackID).Scan(&position)
	if err != nil {
		return err
	}

	err = r.DecrementPositions(ctx, playlistID, position)
	return err
}

func (r *PlaylistTracksRepo) GetOne(ctx context.Context, playlistID, trackID int64) (*PlaylistTrackModel, error) {
	query := `
		SELECT playlist_id, track_id, position, added_at
		FROM playlist_tracks
		WHERE playlist_id = $1 AND track_id = $2
	`

	var model PlaylistTrackModel
	var addedAt time.Time

	err := r.postgres.QueryRowContext(ctx, query, playlistID, trackID).Scan(
		&model.PlaylistID,
		&model.TrackID,
		&model.Position,
		&addedAt,
	)

	if err != nil {
		return nil, err
	}

	model.AddedAt = addedAt

	return &model, nil
}

func (r *PlaylistTracksRepo) GetLastPosition(ctx context.Context, playlistID int64) (int, error) {
	query := `
		SELECT COALESCE(MAX(position), 0)
		FROM playlist_tracks
		WHERE playlist_id = $1
	`

	var lastPosition int
	err := r.postgres.QueryRowContext(ctx, query, playlistID).Scan(&lastPosition)
	if err != nil {
		return 0, err
	}

	return lastPosition, nil
}

func (r *PlaylistTracksRepo) IncrementPositions(ctx context.Context, playlistID int64, fromPosition int) error {
	query := `
		UPDATE playlist_tracks
		SET position = position + 1
		WHERE playlist_id = $1 AND position >= $2
	`

	_, err := r.postgres.ExecContext(ctx, query, playlistID, fromPosition)
	return err
}

func (r *PlaylistTracksRepo) DecrementPositions(ctx context.Context, playlistID int64, fromPosition int) error {
	query := `
		UPDATE playlist_tracks
		SET position = position - 1
		WHERE playlist_id = $1 AND position > $2
	`

	_, err := r.postgres.ExecContext(ctx, query, playlistID, fromPosition)
	return err
}

func (r *PlaylistTracksRepo) MovePositions(ctx context.Context, playlistID int64, fromPosition, toPosition int) error {
	if fromPosition < toPosition {
		query := `
			UPDATE playlist_tracks
			SET position = position - 1
			WHERE playlist_id = $1 AND position > $2 AND position <= $3
		`
		_, err := r.postgres.ExecContext(ctx, query, playlistID, fromPosition, toPosition)
		return err
	} else if fromPosition > toPosition {
		query := `
			UPDATE playlist_tracks
			SET position = position + 1
			WHERE playlist_id = $1 AND position >= $2 AND position < $3
		`
		_, err := r.postgres.ExecContext(ctx, query, playlistID, toPosition, fromPosition)
		return err
	}

	return nil
}
