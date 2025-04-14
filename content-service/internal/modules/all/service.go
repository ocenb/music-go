package all

import (
	"context"
	"log/slog"

	"github.com/ocenb/music-go/content-service/internal/modules/file"
)

type AllServiceInterface interface {
	DeleteAll(ctx context.Context, userID int64) error
}

type AllService struct {
	log         *slog.Logger
	allRepo     AllRepoInterface
	fileService file.FileServiceInterface
}

func NewAllService(log *slog.Logger, allRepo AllRepoInterface, fileService file.FileServiceInterface) AllServiceInterface {
	return &AllService{
		log:         log,
		allRepo:     allRepo,
		fileService: fileService,
	}
}

func (s *AllService) DeleteAll(ctx context.Context, userID int64) error {
	s.log.Info("Starting deletion of all user content", "user_id", userID)

	tracksAudios, tracksImages, playlistsImages, err := s.allRepo.DeleteAll(ctx, userID)
	if err != nil {
		s.log.Error("Failed to delete user content from database", "error", err, "user_id", userID)
		return err
	}

	for _, audio := range tracksAudios {
		err = s.fileService.DeleteFile(ctx, audio, file.AudioCategory)
		if err != nil {
			s.log.Error("Failed to delete audio file", "error", err, "audio", audio)
			return err
		}
	}

	for _, image := range tracksImages {
		err = s.fileService.DeleteFile(ctx, image, file.ImagesCategory)
		if err != nil {
			s.log.Error("Failed to delete image file", "error", err, "image", image)
			return err
		}
	}

	for _, image := range playlistsImages {
		err = s.fileService.DeleteFile(ctx, image, file.ImagesCategory)
		if err != nil {
			s.log.Error("Failed to delete image file", "error", err, "image", image)
			return err
		}
	}

	s.log.Info("Successfully deleted all user content", "user_id", userID)
	return nil
}
