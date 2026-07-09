package playlisttracks

import (
	"context"
	"errors"
	"fmt"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/models"
	"github.com/ocenb/music-go/content-service/internal/storage/transactor"
)

type Repo interface {
	GetMany(ctx context.Context, playlistID, currentUserID int64, take int) ([]*models.TrackInPlaylistModel, error)
	Add(ctx context.Context, playlistID, trackID int64, position int) (*models.PlaylistTrackModel, error)
	UpdatePosition(ctx context.Context, playlistID, trackID int64, position int) error
	Remove(ctx context.Context, playlistID, trackID int64) error
	GetOne(ctx context.Context, playlistID, trackID int64) (*models.PlaylistTrackModel, error)
	GetLastPosition(ctx context.Context, playlistID int64) (int, error)
	IncrementPositions(ctx context.Context, playlistID int64, fromPosition int) error
	MovePositions(ctx context.Context, playlistID int64, fromPosition, toPosition int) error
}

type PlaylistRepo interface {
	GetByID(ctx context.Context, playlistID int64, currentUserID int64) (*models.PlaylistWithSavedModel, error)
	CheckPermission(ctx context.Context, userID, playlistID int64) (bool, error)
}

type TrackRepo interface {
	GetByID(ctx context.Context, trackID int64, currentUserID int64) (*models.TrackWithLikedModel, error)
}

type Service struct {
	repo         Repo
	playlistRepo PlaylistRepo
	trackRepo    TrackRepo
	tm           transactor.Runner
}

func New(
	repo Repo,
	playlistRepo PlaylistRepo,
	trackRepo TrackRepo,
	tm transactor.Runner,
) *Service {
	return &Service{
		repo:         repo,
		playlistRepo: playlistRepo,
		trackRepo:    trackRepo,
		tm:           tm,
	}
}

func (s *Service) GetMany(ctx context.Context, currentUserID, playlistID int64, take int) ([]*models.TrackInPlaylistModel, error) {
	if _, err := s.playlistRepo.GetByID(ctx, playlistID, currentUserID); err != nil {
		return nil, fmt.Errorf("PlaylistTracksService.GetMany: %w", err)
	}

	tracks, err := s.repo.GetMany(ctx, playlistID, currentUserID, take)
	if err != nil {
		return nil, fmt.Errorf("PlaylistTracksService.GetMany: %w", err)
	}

	return tracks, nil
}

func (s *Service) Add(ctx context.Context, userID, playlistID, trackID int64, position int) (*models.PlaylistTrackModel, error) {
	hasPermission, err := s.playlistRepo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return nil, fmt.Errorf("PlaylistTracksService.Add: %w", err)
	}
	if !hasPermission {
		return nil, errs.ErrPermissionDenied
	}

	if _, getErr := s.trackRepo.GetByID(ctx, trackID, userID); getErr != nil {
		return nil, fmt.Errorf("PlaylistTracksService.Add: %w", getErr)
	}

	trackInPlaylist, err := s.repo.GetOne(ctx, playlistID, trackID)
	if err != nil && !errors.Is(err, errs.ErrTrackNotInPlaylist) {
		return nil, fmt.Errorf("PlaylistTracksService.Add: %w", err)
	}
	if trackInPlaylist != nil {
		return nil, errs.ErrTrackAlreadyInPlaylist
	}

	lastPosition, err := s.repo.GetLastPosition(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("PlaylistTracksService.Add: %w", err)
	}

	newPosition := lastPosition + 1
	if position > 0 && position <= lastPosition+1 {
		newPosition = position
	}

	var playlistTrack *models.PlaylistTrackModel
	err = s.tm.Run(ctx, func(txCtx context.Context) error {
		if position > 0 && position <= lastPosition {
			if incErr := s.repo.IncrementPositions(txCtx, playlistID, position); incErr != nil {
				return incErr
			}
		}

		var addErr error
		playlistTrack, addErr = s.repo.Add(txCtx, playlistID, trackID, newPosition)
		return addErr
	})
	if err != nil {
		return nil, fmt.Errorf("PlaylistTracksService.Add: %w", err)
	}

	return playlistTrack, nil
}

func (s *Service) UpdatePosition(ctx context.Context, userID, playlistID, trackID int64, position int) error {
	hasPermission, err := s.playlistRepo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return fmt.Errorf("PlaylistTracksService.UpdatePosition: %w", err)
	}
	if !hasPermission {
		return errs.ErrPermissionDenied
	}

	trackInPlaylist, err := s.repo.GetOne(ctx, playlistID, trackID)
	if err != nil {
		return fmt.Errorf("PlaylistTracksService.UpdatePosition: %w", err)
	}

	if trackInPlaylist.Position == position {
		return errs.ErrPositionConflict
	}

	err = s.tm.Run(ctx, func(txCtx context.Context) error {
		if moveErr := s.repo.MovePositions(txCtx, playlistID, trackInPlaylist.Position, position); moveErr != nil {
			return moveErr
		}

		return s.repo.UpdatePosition(txCtx, playlistID, trackID, position)
	})
	if err != nil {
		return fmt.Errorf("PlaylistTracksService.UpdatePosition: %w", err)
	}

	return nil
}

func (s *Service) Remove(ctx context.Context, userID, playlistID, trackID int64) error {
	hasPermission, err := s.playlistRepo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return fmt.Errorf("PlaylistTracksService.Remove: %w", err)
	}
	if !hasPermission {
		return errs.ErrPermissionDenied
	}

	if _, err := s.repo.GetOne(ctx, playlistID, trackID); err != nil {
		return fmt.Errorf("PlaylistTracksService.Remove: %w", err)
	}

	if err := s.repo.Remove(ctx, playlistID, trackID); err != nil {
		return fmt.Errorf("PlaylistTracksService.Remove: %w", err)
	}

	return nil
}
