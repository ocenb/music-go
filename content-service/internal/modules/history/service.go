package history

import (
	"context"
	"log/slog"

	"github.com/ocenb/music-go/content-service/internal/modules/track"
)

type HistoryServiceInterface interface {
	Get(ctx context.Context, currentUserID int64, take int64) ([]*ListeningHistoryModel, error)
	Add(ctx context.Context, currentUserID, trackID int64) error
	Clear(ctx context.Context, currentUserID int64) error
}

type HistoryService struct {
	log          *slog.Logger
	historyRepo  HistoryRepoInterface
	trackService track.TrackServiceInterface
}

func NewHistoryService(log *slog.Logger, historyRepo HistoryRepoInterface, trackService track.TrackServiceInterface) HistoryServiceInterface {
	return &HistoryService{
		log:          log,
		historyRepo:  historyRepo,
		trackService: trackService,
	}
}

func (s *HistoryService) Get(ctx context.Context, currentUserID int64, take int64) ([]*ListeningHistoryModel, error) {
	history, err := s.historyRepo.Get(ctx, currentUserID, take)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func (s *HistoryService) Add(ctx context.Context, currentUserID, trackID int64) error {
	_, err := s.trackService.GetOneById(ctx, currentUserID, trackID)
	if err != nil {
		return err
	}
	err = s.historyRepo.Add(ctx, currentUserID, trackID)
	if err != nil {
		return err
	}
	return nil
}

func (s *HistoryService) Clear(ctx context.Context, currentUserID int64) error {
	err := s.historyRepo.Clear(ctx, currentUserID)
	if err != nil {
		return err
	}
	return nil
}
