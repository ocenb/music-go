package track

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"mime/multipart"

	"github.com/ocenb/music-go/content-service/internal/clients/notificationclient"
	"github.com/ocenb/music-go/content-service/internal/clients/searchclient"
	"github.com/ocenb/music-go/content-service/internal/modules/file"
	"github.com/ocenb/music-go/content-service/internal/storage"
	"github.com/ocenb/music-protos/gen/searchservice"
)

type TrackServiceInterface interface {
	GetOneById(ctx context.Context, currentUserID, trackID int64) (*TrackWithLikedModel, error)
	GetOne(ctx context.Context, currentUserID int64, username, changeableID string) (*TrackWithLikedModel, error)
	GetMany(ctx context.Context, currentUserID, userID int64, take int, lastID int64) ([]*TrackWithLikedModel, error)
	GetManyPopular(ctx context.Context, currentUserID, userID int64, take int, lastID int64) ([]*TrackWithLikedModel, error)
	Upload(ctx context.Context, userID int64, username, email, title, changeableID string, audioFile *multipart.FileHeader, imageFile *multipart.FileHeader) (*TrackModel, error)
	AddPlay(ctx context.Context, trackID int64) error
	Delete(ctx context.Context, userID, trackID int64) error
	ChangeTitle(ctx context.Context, userID, trackID int64, title string) error
	ChangeChangeableId(ctx context.Context, userID, trackID int64, changeableID string) error
	ChangeImage(ctx context.Context, userID, trackID int64, imageFile *multipart.FileHeader) error
	GetManyLiked(ctx context.Context, currentUserID int64) ([]*UserLikedTrackModel, error)
	AddToLiked(ctx context.Context, currentUserID, trackID int64) error
	RemoveFromLiked(ctx context.Context, currentUserID, trackID int64) error
}

type TrackService struct {
	log                *slog.Logger
	trackRepo          TrackRepoInterface
	fileService        file.FileServiceInterface
	searchClient       *searchclient.SearchServiceClient
	notificationClient notificationclient.NotificationClientInterface
}

func NewTrackService(log *slog.Logger, trackRepo TrackRepoInterface, fileService file.FileServiceInterface, searchClient *searchclient.SearchServiceClient, notificationClient notificationclient.NotificationClientInterface) TrackServiceInterface {
	return &TrackService{
		log:                log,
		trackRepo:          trackRepo,
		fileService:        fileService,
		searchClient:       searchClient,
		notificationClient: notificationClient,
	}
}

func (s *TrackService) GetOneById(ctx context.Context, currentUserID, trackID int64) (*TrackWithLikedModel, error) {
	track, err := s.trackRepo.GetByID(ctx, trackID, currentUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTrackNotFound
		}
		return nil, err
	}

	return track, nil
}

func (s *TrackService) GetOne(ctx context.Context, currentUserID int64, username, changeableID string) (*TrackWithLikedModel, error) {
	track, err := s.trackRepo.GetByChangeableID(ctx, username, changeableID, currentUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTrackNotFound
		}
		return nil, err
	}

	return track, nil
}

func (s *TrackService) GetMany(ctx context.Context, currentUserID, userID int64, take int, lastID int64) ([]*TrackWithLikedModel, error) {
	tracks, err := s.trackRepo.GetMany(ctx, userID, currentUserID, take, lastID)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func (s *TrackService) GetManyPopular(ctx context.Context, currentUserID, userID int64, take int, lastID int64) ([]*TrackWithLikedModel, error) {
	tracks, err := s.trackRepo.GetManyPopular(ctx, userID, currentUserID, take, lastID)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func (s *TrackService) Upload(ctx context.Context, userID int64, username, email, title, changeableID string, audioFile *multipart.FileHeader, imageFile *multipart.FileHeader) (*TrackModel, error) {
	if err := s.validateTrackTitle(ctx, userID, title); err != nil {
		return nil, err
	}

	if err := s.validateChangeableId(ctx, userID, changeableID); err != nil {
		return nil, err
	}

	audioResult, err := s.fileService.SaveAudio(ctx, audioFile)
	if err != nil {
		return nil, err
	}

	imageName, err := s.fileService.SaveImage(ctx, imageFile)
	if err != nil {
		return nil, err
	}

	var newTrack *TrackModel
	err = storage.WithTransaction(ctx, s.trackRepo, func(txCtx context.Context) error {
		newTrack, err = s.trackRepo.Create(txCtx, userID, username, title, changeableID, audioResult.FileName, imageName, int64(audioResult.Duration))
		if err != nil {
			return err
		}
		_, err = s.searchClient.Client.AddTrack(txCtx, &searchservice.AddOrUpdateRequest{
			Id:   newTrack.ID,
			Name: newTrack.Title,
		})
		if err != nil {
			return err
		}

		err = s.notificationClient.SendEmailNotification(email, newTrack.Title)
		if err != nil {
			s.log.Error("Failed to send email notification", "error", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newTrack, nil
}

func (s *TrackService) AddPlay(ctx context.Context, trackID int64) error {
	err := s.trackRepo.AddPlay(ctx, trackID)
	if err != nil {
		return err
	}
	return nil
}

func (s *TrackService) Delete(ctx context.Context, userID, trackID int64) error {
	track, err := s.trackRepo.GetByID(ctx, trackID, userID)
	if err != nil {
		return err
	}

	hasPermission, err := s.trackRepo.CheckPermission(ctx, userID, trackID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrPermissionDenied
	}

	err = storage.WithTransaction(ctx, s.trackRepo, func(txCtx context.Context) error {
		if err := s.trackRepo.Delete(txCtx, trackID); err != nil {
			return err
		}

		_, err = s.searchClient.Client.DeleteTrack(txCtx, &searchservice.DeleteRequest{
			Id: trackID,
		})
		if err != nil {
			return err
		}

		err = s.fileService.DeleteFile(txCtx, track.Audio, file.AudioCategory)
		if err != nil {
			return err
		}

		err = s.fileService.DeleteFile(txCtx, track.Image, file.ImagesCategory)
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

func (s *TrackService) ChangeTitle(ctx context.Context, userID, trackID int64, title string) error {
	hasPermission, err := s.trackRepo.CheckPermission(ctx, userID, trackID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrPermissionDenied
	}

	if err := s.validateTrackTitle(ctx, userID, title); err != nil {
		return err
	}

	if err := s.trackRepo.ChangeTitle(ctx, trackID, title); err != nil {
		return err
	}

	return nil
}

func (s *TrackService) ChangeChangeableId(ctx context.Context, userID, trackID int64, changeableID string) error {
	track, err := s.trackRepo.GetByID(ctx, trackID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTrackNotFound
		}
		return err
	}

	hasPermission, err := s.trackRepo.CheckPermission(ctx, userID, trackID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrPermissionDenied
	}

	if err := s.validateChangeableId(ctx, track.UserID, changeableID); err != nil {
		return err
	}

	if err := s.trackRepo.ChangeChangeableID(ctx, trackID, changeableID); err != nil {
		return err
	}

	return nil
}

func (s *TrackService) ChangeImage(ctx context.Context, userID, trackID int64, imageFile *multipart.FileHeader) error {
	hasPermission, err := s.trackRepo.CheckPermission(ctx, userID, trackID)
	if err != nil {
		return err
	}

	if !hasPermission {
		return ErrPermissionDenied
	}

	imageName, err := s.fileService.SaveImage(ctx, imageFile)
	if err != nil {
		return err
	}

	if err := s.trackRepo.ChangeImage(ctx, trackID, imageName); err != nil {
		return err
	}

	return nil
}

func (s *TrackService) validateTrackTitle(ctx context.Context, userID int64, title string) error {
	exists, err := s.trackRepo.CheckTitle(ctx, userID, title)
	if err != nil {
		return err
	}
	if exists {
		return ErrTrackAlreadyExists
	}

	return nil
}

func (s *TrackService) validateChangeableId(ctx context.Context, userID int64, changeableID string) error {
	exists, err := s.trackRepo.CheckChangeableID(ctx, userID, changeableID)
	if err != nil {
		return err
	}
	if exists {
		return ErrChangeableIDExists
	}

	return nil
}

func (s *TrackService) GetManyLiked(ctx context.Context, currentUserID int64) ([]*UserLikedTrackModel, error) {
	likedTracks, err := s.trackRepo.GetManyLiked(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	return likedTracks, nil
}

func (s *TrackService) AddToLiked(ctx context.Context, currentUserID, trackID int64) error {
	_, err := s.trackRepo.GetByID(ctx, trackID, currentUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTrackNotFound
		}
		return err
	}

	err = s.trackRepo.AddToLiked(ctx, currentUserID, trackID)
	if err != nil {
		return err
	}
	return nil
}

func (s *TrackService) RemoveFromLiked(ctx context.Context, currentUserID, trackID int64) error {
	err := s.trackRepo.RemoveFromLiked(ctx, currentUserID, trackID)
	if err != nil {
		return err
	}
	return nil
}
