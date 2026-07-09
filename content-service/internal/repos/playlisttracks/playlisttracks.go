package playlisttracks

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

func (r *Repo) GetMany(ctx context.Context, playlistID, currentUserID int64, take int) ([]*models.TrackInPlaylistModel, error) {
	q := r.tm.GetQueryEngine(ctx)

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

	var rows pgx.Rows
	var err error

	if take > 0 {
		rows, err = q.Query(ctx, query, currentUserID, playlistID, take)
	} else {
		rows, err = q.Query(ctx, query, currentUserID, playlistID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get playlist tracks: %w", err)
	}
	defer rows.Close()

	var tracks []*models.TrackInPlaylistModel

	for rows.Next() {
		var trackInPlaylist models.TrackInPlaylistModel
		var trackModel models.TrackWithLikedModel
		var createdAt, updatedAt, addedAt time.Time

		if err := rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("failed to scan playlist track: %w", err)
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
		return nil, fmt.Errorf("failed to iterate playlist tracks: %w", err)
	}

	return tracks, nil
}

func (r *Repo) Add(ctx context.Context, playlistID, trackID int64, position int) (*models.PlaylistTrackModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		INSERT INTO playlist_tracks (playlist_id, track_id, position, added_at)
		VALUES ($1, $2, $3, $4)
		RETURNING playlist_id, track_id, position, added_at
	`

	var model models.PlaylistTrackModel
	var addedAt time.Time

	err := q.QueryRow(ctx, query, playlistID, trackID, position, time.Now()).Scan(
		&model.PlaylistID,
		&model.TrackID,
		&model.Position,
		&addedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add track to playlist: %w", err)
	}

	model.AddedAt = addedAt

	return &model, nil
}

func (r *Repo) UpdatePosition(ctx context.Context, playlistID, trackID int64, position int) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE playlist_tracks
		SET position = $1
		WHERE playlist_id = $2 AND track_id = $3
	`

	if _, err := q.Exec(ctx, query, position, playlistID, trackID); err != nil {
		return fmt.Errorf("failed to update playlist track position: %w", err)
	}

	return nil
}

func (r *Repo) Remove(ctx context.Context, playlistID, trackID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		DELETE FROM playlist_tracks
		WHERE playlist_id = $1 AND track_id = $2
		RETURNING position
	`

	var position int
	if err := q.QueryRow(ctx, query, playlistID, trackID).Scan(&position); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrTrackNotInPlaylist
		}
		return fmt.Errorf("failed to remove track from playlist: %w", err)
	}

	return r.DecrementPositions(ctx, playlistID, position)
}

func (r *Repo) GetOne(ctx context.Context, playlistID, trackID int64) (*models.PlaylistTrackModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT playlist_id, track_id, position, added_at
		FROM playlist_tracks
		WHERE playlist_id = $1 AND track_id = $2
	`

	var model models.PlaylistTrackModel
	var addedAt time.Time

	err := q.QueryRow(ctx, query, playlistID, trackID).Scan(
		&model.PlaylistID,
		&model.TrackID,
		&model.Position,
		&addedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrTrackNotInPlaylist
		}
		return nil, fmt.Errorf("failed to get playlist track: %w", err)
	}

	model.AddedAt = addedAt

	return &model, nil
}

func (r *Repo) GetLastPosition(ctx context.Context, playlistID int64) (int, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT COALESCE(MAX(position), 0)
		FROM playlist_tracks
		WHERE playlist_id = $1
	`

	var lastPosition int
	if err := q.QueryRow(ctx, query, playlistID).Scan(&lastPosition); err != nil {
		return 0, fmt.Errorf("failed to get last playlist track position: %w", err)
	}

	return lastPosition, nil
}

func (r *Repo) IncrementPositions(ctx context.Context, playlistID int64, fromPosition int) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE playlist_tracks
		SET position = position + 1
		WHERE playlist_id = $1 AND position >= $2
	`

	if _, err := q.Exec(ctx, query, playlistID, fromPosition); err != nil {
		return fmt.Errorf("failed to increment playlist track positions: %w", err)
	}

	return nil
}

func (r *Repo) DecrementPositions(ctx context.Context, playlistID int64, fromPosition int) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE playlist_tracks
		SET position = position - 1
		WHERE playlist_id = $1 AND position > $2
	`

	if _, err := q.Exec(ctx, query, playlistID, fromPosition); err != nil {
		return fmt.Errorf("failed to decrement playlist track positions: %w", err)
	}

	return nil
}

func (r *Repo) MovePositions(ctx context.Context, playlistID int64, fromPosition, toPosition int) error {
	q := r.tm.GetQueryEngine(ctx)

	if fromPosition < toPosition {
		query := `
			UPDATE playlist_tracks
			SET position = position - 1
			WHERE playlist_id = $1 AND position > $2 AND position <= $3
		`
		if _, err := q.Exec(ctx, query, playlistID, fromPosition, toPosition); err != nil {
			return fmt.Errorf("failed to move playlist track positions: %w", err)
		}
	} else if fromPosition > toPosition {
		query := `
			UPDATE playlist_tracks
			SET position = position + 1
			WHERE playlist_id = $1 AND position >= $2 AND position < $3
		`
		if _, err := q.Exec(ctx, query, playlistID, toPosition, fromPosition); err != nil {
			return fmt.Errorf("failed to move playlist track positions: %w", err)
		}
	}

	return nil
}
