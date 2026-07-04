package all

import (
	"context"
	"fmt"

	fileservice "github.com/ocenb/music-go/content-service/internal/services/file"
)

type Repo interface {
	DeleteAll(ctx context.Context, userID int64) ([]string, []string, []string, error)
}

type FileService interface {
	DeleteFile(ctx context.Context, fileName string, category fileservice.Category) error
}

type Service struct {
	repo        Repo
	fileService FileService
}

func New(repo Repo, fileService FileService) *Service {
	return &Service{
		repo:        repo,
		fileService: fileService,
	}
}

func (s *Service) DeleteAll(ctx context.Context, userID int64) error {
	tracksAudios, tracksImages, playlistsImages, err := s.repo.DeleteAll(ctx, userID)
	if err != nil {
		return fmt.Errorf("AllService.DeleteAll: %w", err)
	}

	for _, audio := range tracksAudios {
		if err := s.fileService.DeleteFile(ctx, audio, fileservice.AudioCategory); err != nil {
			return fmt.Errorf("AllService.DeleteAll: %w", err)
		}
	}

	for _, image := range tracksImages {
		if err := s.fileService.DeleteFile(ctx, image, fileservice.ImagesCategory); err != nil {
			return fmt.Errorf("AllService.DeleteAll: %w", err)
		}
	}

	for _, image := range playlistsImages {
		if err := s.fileService.DeleteFile(ctx, image, fileservice.ImagesCategory); err != nil {
			return fmt.Errorf("AllService.DeleteAll: %w", err)
		}
	}

	return nil
}
