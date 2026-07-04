package history

import (
	"context"
	"fmt"

	"github.com/ocenb/music-go/content-service/internal/models"
)

type Repo interface {
	Get(ctx context.Context, currentUserID int64, take int64) ([]*models.ListeningHistoryModel, error)
	Add(ctx context.Context, currentUserID, trackID int64) error
	Clear(ctx context.Context, currentUserID int64) error
}

type TrackService interface {
	GetOneByID(ctx context.Context, currentUserID, trackID int64) (*models.TrackWithLikedModel, error)
}

type Service struct {
	repo         Repo
	trackService TrackService
}

func New(repo Repo, trackService TrackService) *Service {
	return &Service{
		repo:         repo,
		trackService: trackService,
	}
}

func (s *Service) Get(ctx context.Context, currentUserID int64, take int64) ([]*models.ListeningHistoryModel, error) {
	history, err := s.repo.Get(ctx, currentUserID, take)
	if err != nil {
		return nil, fmt.Errorf("HistoryService.Get: %w", err)
	}
	return history, nil
}

func (s *Service) Add(ctx context.Context, currentUserID, trackID int64) error {
	if _, err := s.trackService.GetOneByID(ctx, currentUserID, trackID); err != nil {
		return fmt.Errorf("HistoryService.Add: %w", err)
	}

	if err := s.repo.Add(ctx, currentUserID, trackID); err != nil {
		return fmt.Errorf("HistoryService.Add: %w", err)
	}
	return nil
}

func (s *Service) Clear(ctx context.Context, currentUserID int64) error {
	if err := s.repo.Clear(ctx, currentUserID); err != nil {
		return fmt.Errorf("HistoryService.Clear: %w", err)
	}
	return nil
}
