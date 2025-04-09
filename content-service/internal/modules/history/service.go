package history

import (
	"context"
	"log/slog"
)

type HistoryServiceInterface interface {
	Get(ctx context.Context, currentUserID int64, take int64) ([]*ListeningHistoryModel, error)
	Add(ctx context.Context, currentUserID, trackID int64) error
	Clear(ctx context.Context, currentUserID int64) error
}

type HistoryService struct {
	log         *slog.Logger
	historyRepo HistoryRepoInterface
}

func NewHistoryService(log *slog.Logger, historyRepo HistoryRepoInterface) HistoryServiceInterface {
	return &HistoryService{
		log:         log,
		historyRepo: historyRepo,
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
	err := s.historyRepo.Add(ctx, currentUserID, trackID)
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
