package playlisttracks

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/ocenb/music-go/content-service/internal/modules/playlist"
	"github.com/ocenb/music-go/content-service/internal/modules/track"
	"github.com/ocenb/music-go/content-service/internal/storage"
)

type PlaylistTracksServiceInterface interface {
	GetMany(ctx context.Context, currentUserID, playlistID int64, take int) ([]*TrackInPlaylistModel, error)
	Add(ctx context.Context, userID, playlistID, trackID int64, position int) (*PlaylistTrackModel, error)
	UpdatePosition(ctx context.Context, userID, playlistID, trackID int64, position int) error
	Remove(ctx context.Context, userID, playlistID, trackID int64) error
}

type PlaylistTracksService struct {
	log                *slog.Logger
	playlistTracksRepo PlaylistTracksRepoInterface
	playlistRepo       playlist.PlaylistRepoInterface
	trackRepo          track.TrackRepoInterface
}

func NewPlaylistTracksService(
	log *slog.Logger,
	playlistTracksRepo PlaylistTracksRepoInterface,
	playlistRepo playlist.PlaylistRepoInterface,
	trackRepo track.TrackRepoInterface,
) PlaylistTracksServiceInterface {
	return &PlaylistTracksService{
		log:                log,
		playlistTracksRepo: playlistTracksRepo,
		playlistRepo:       playlistRepo,
		trackRepo:          trackRepo,
	}
}

func (s *PlaylistTracksService) GetMany(ctx context.Context, currentUserID, playlistID int64, take int) ([]*TrackInPlaylistModel, error) {
	_, err := s.playlistRepo.GetByID(ctx, playlistID, currentUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}

	tracks, err := s.playlistTracksRepo.GetMany(ctx, playlistID, currentUserID, take)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func (s *PlaylistTracksService) Add(ctx context.Context, userID, playlistID, trackID int64, position int) (*PlaylistTrackModel, error) {
	hasPermission, err := s.playlistRepo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrPermissionDenied
	}

	_, err = s.trackRepo.GetByID(ctx, trackID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTrackNotFound
		}
		return nil, err
	}

	trackInPlaylist, err := s.playlistTracksRepo.GetOne(ctx, playlistID, trackID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if trackInPlaylist != nil {
		return nil, ErrTrackAlreadyInPlaylist
	}

	lastPosition, err := s.playlistTracksRepo.GetLastPosition(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	newPosition := lastPosition + 1
	if position > 0 && position <= lastPosition+1 {
		newPosition = position
	}

	var playlistTrack *PlaylistTrackModel
	err = storage.WithTransaction(ctx, s.playlistTracksRepo, func(txCtx context.Context) error {
		if position > 0 && position <= lastPosition {
			err = s.playlistTracksRepo.IncrementPositions(txCtx, playlistID, position)
			if err != nil {
				return err
			}
		}

		playlistTrack, err = s.playlistTracksRepo.Add(txCtx, playlistID, trackID, newPosition)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return playlistTrack, nil
}

func (s *PlaylistTracksService) UpdatePosition(ctx context.Context, userID, playlistID, trackID int64, position int) error {
	hasPermission, err := s.playlistRepo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrPermissionDenied
	}

	trackInPlaylist, err := s.playlistTracksRepo.GetOne(ctx, playlistID, trackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTrackNotInPlaylist
		}
		return err
	}

	if trackInPlaylist.Position == position {
		return ErrPositionConflict
	}

	err = storage.WithTransaction(ctx, s.playlistTracksRepo, func(txCtx context.Context) error {
		err = s.playlistTracksRepo.MovePositions(txCtx, playlistID, trackInPlaylist.Position, position)
		if err != nil {
			return err
		}

		err = s.playlistTracksRepo.UpdatePosition(txCtx, playlistID, trackID, position)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *PlaylistTracksService) Remove(ctx context.Context, userID, playlistID, trackID int64) error {
	hasPermission, err := s.playlistRepo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrPermissionDenied
	}

	_, err = s.playlistTracksRepo.GetOne(ctx, playlistID, trackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTrackNotInPlaylist
		}
		return err
	}

	err = s.playlistTracksRepo.Remove(ctx, playlistID, trackID)
	if err != nil {
		return err
	}

	return nil
}
