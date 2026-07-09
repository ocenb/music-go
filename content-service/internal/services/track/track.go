package track

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/ocenb/music-protos/gen/searchservice"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/models"
	fileservice "github.com/ocenb/music-go/content-service/internal/services/file"
	"github.com/ocenb/music-go/content-service/internal/storage/transactor"
)

type Repo interface {
	GetByID(ctx context.Context, trackID int64, currentUserID int64) (*models.TrackWithLikedModel, error)
	GetByChangeableID(ctx context.Context, username, changeableID string, currentUserID int64) (*models.TrackWithLikedModel, error)
	GetMany(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*models.TrackWithLikedModel, error)
	GetManyPopular(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*models.TrackWithLikedModel, error)
	Create(ctx context.Context, userID int64, username, title, changeableID, audio, image string, duration int64) (*models.TrackModel, error)
	AddPlay(ctx context.Context, trackID int64) error
	CheckPermission(ctx context.Context, userID, trackID int64) (bool, error)
	Delete(ctx context.Context, trackID int64) error
	ChangeTitle(ctx context.Context, trackID int64, title string) error
	ChangeChangeableID(ctx context.Context, trackID int64, changeableID string) error
	ChangeImage(ctx context.Context, trackID int64, image string) error
	CheckTitle(ctx context.Context, userID int64, title string) (bool, error)
	CheckChangeableID(ctx context.Context, userID int64, changeableID string) (bool, error)
	GetManyLiked(ctx context.Context, currentUserID int64) ([]*models.UserLikedTrackModel, error)
	AddToLiked(ctx context.Context, currentUserID, trackID int64) error
	RemoveFromLiked(ctx context.Context, currentUserID, trackID int64) error
}

type FileService interface {
	SaveAudio(ctx context.Context, file *multipart.FileHeader) (*fileservice.AudioResult, error)
	SaveImage(ctx context.Context, file *multipart.FileHeader) (string, error)
	DeleteFile(ctx context.Context, fileName string, category fileservice.Category) error
}

type SearchClient interface {
	AddTrack(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error)
	UpdateTrack(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error)
	DeleteTrack(ctx context.Context, req *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error)
}

type NotificationClient interface {
	SendEmailNotification(ctx context.Context, email, msg string) error
}

type Service struct {
	repo               Repo
	fileService        FileService
	searchClient       SearchClient
	notificationClient NotificationClient
	tm                 transactor.Runner
}

func New(
	repo Repo,
	fileService FileService,
	searchClient SearchClient,
	notificationClient NotificationClient,
	tm transactor.Runner,
) *Service {
	return &Service{
		repo:               repo,
		fileService:        fileService,
		searchClient:       searchClient,
		notificationClient: notificationClient,
		tm:                 tm,
	}
}

func (s *Service) GetOneByID(ctx context.Context, currentUserID, trackID int64) (*models.TrackWithLikedModel, error) {
	track, err := s.repo.GetByID(ctx, trackID, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("TrackService.GetOneByID: %w", err)
	}
	return track, nil
}

func (s *Service) GetOne(ctx context.Context, currentUserID int64, username, changeableID string) (*models.TrackWithLikedModel, error) {
	track, err := s.repo.GetByChangeableID(ctx, username, changeableID, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("TrackService.GetOne: %w", err)
	}
	return track, nil
}

func (s *Service) GetMany(ctx context.Context, currentUserID, userID int64, take int, lastID int64) ([]*models.TrackWithLikedModel, error) {
	tracks, err := s.repo.GetMany(ctx, userID, currentUserID, take, lastID)
	if err != nil {
		return nil, fmt.Errorf("TrackService.GetMany: %w", err)
	}
	return tracks, nil
}

func (s *Service) GetManyPopular(ctx context.Context, currentUserID, userID int64, take int, lastID int64) ([]*models.TrackWithLikedModel, error) {
	tracks, err := s.repo.GetManyPopular(ctx, userID, currentUserID, take, lastID)
	if err != nil {
		return nil, fmt.Errorf("TrackService.GetManyPopular: %w", err)
	}
	return tracks, nil
}

func (s *Service) Upload(ctx context.Context, userID int64, username, email, title, changeableID string, audioFile *multipart.FileHeader, imageFile *multipart.FileHeader) (*models.TrackModel, error) {
	if err := s.validateTrackTitle(ctx, userID, title); err != nil {
		return nil, err
	}

	if err := s.validateChangeableID(ctx, userID, changeableID); err != nil {
		return nil, err
	}

	audioResult, err := s.fileService.SaveAudio(ctx, audioFile)
	if err != nil {
		return nil, fmt.Errorf("TrackService.Upload: %w", err)
	}

	imageName, err := s.fileService.SaveImage(ctx, imageFile)
	if err != nil {
		return nil, fmt.Errorf("TrackService.Upload: %w", err)
	}

	var newTrack *models.TrackModel
	err = s.tm.Run(ctx, func(txCtx context.Context) error {
		var createErr error
		newTrack, createErr = s.repo.Create(txCtx, userID, username, title, changeableID, audioResult.FileName, imageName, int64(audioResult.Duration))
		if createErr != nil {
			return createErr
		}

		_, createErr = s.searchClient.AddTrack(txCtx, &searchservice.AddOrUpdateRequest{
			Id:   newTrack.ID,
			Name: newTrack.Title,
		})
		if createErr != nil {
			return createErr
		}

		if notifyErr := s.notificationClient.SendEmailNotification(txCtx, email, newTrack.Title); notifyErr != nil {
			return fmt.Errorf("failed to send email notification: %w", notifyErr)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("TrackService.Upload: %w", err)
	}

	return newTrack, nil
}

func (s *Service) AddPlay(ctx context.Context, trackID int64) error {
	if err := s.repo.AddPlay(ctx, trackID); err != nil {
		return fmt.Errorf("TrackService.AddPlay: %w", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, userID, trackID int64) error {
	track, err := s.repo.GetByID(ctx, trackID, userID)
	if err != nil {
		return fmt.Errorf("TrackService.Delete: %w", err)
	}

	hasPermission, err := s.repo.CheckPermission(ctx, userID, trackID)
	if err != nil {
		return fmt.Errorf("TrackService.Delete: %w", err)
	}
	if !hasPermission {
		return errs.ErrPermissionDenied
	}

	err = s.tm.Run(ctx, func(txCtx context.Context) error {
		if delErr := s.repo.Delete(txCtx, trackID); delErr != nil {
			return delErr
		}

		deleteResp, deleteErr := s.searchClient.DeleteTrack(txCtx, &searchservice.DeleteRequest{
			Id: trackID,
		})
		if deleteErr != nil || !deleteResp.Success {
			return fmt.Errorf("failed to delete track in search service: %w", deleteErr)
		}

		if audioErr := s.fileService.DeleteFile(txCtx, track.Audio, fileservice.AudioCategory); audioErr != nil {
			return audioErr
		}

		return s.fileService.DeleteFile(txCtx, track.Image, fileservice.ImagesCategory)
	})
	if err != nil {
		return fmt.Errorf("TrackService.Delete: %w", err)
	}
	return nil
}

func (s *Service) ChangeTitle(ctx context.Context, userID, trackID int64, title string) error {
	hasPermission, err := s.repo.CheckPermission(ctx, userID, trackID)
	if err != nil {
		return fmt.Errorf("TrackService.ChangeTitle: %w", err)
	}
	if !hasPermission {
		return errs.ErrPermissionDenied
	}

	if titleErr := s.validateTrackTitle(ctx, userID, title); titleErr != nil {
		return titleErr
	}

	if changeErr := s.repo.ChangeTitle(ctx, trackID, title); changeErr != nil {
		return fmt.Errorf("TrackService.ChangeTitle: %w", changeErr)
	}

	updateResp, err := s.searchClient.UpdateTrack(ctx, &searchservice.AddOrUpdateRequest{
		Id:   trackID,
		Name: title,
	})
	if err != nil || !updateResp.Success {
		return fmt.Errorf("failed to update track in search service: %w", err)
	}

	return nil
}

func (s *Service) ChangeChangeableID(ctx context.Context, userID, trackID int64, changeableID string) error {
	track, err := s.repo.GetByID(ctx, trackID, userID)
	if err != nil {
		return fmt.Errorf("TrackService.ChangeChangeableID: %w", err)
	}

	hasPermission, err := s.repo.CheckPermission(ctx, userID, trackID)
	if err != nil {
		return fmt.Errorf("TrackService.ChangeChangeableID: %w", err)
	}
	if !hasPermission {
		return errs.ErrPermissionDenied
	}

	if err := s.validateChangeableID(ctx, track.UserID, changeableID); err != nil {
		return err
	}

	if err := s.repo.ChangeChangeableID(ctx, trackID, changeableID); err != nil {
		return fmt.Errorf("TrackService.ChangeChangeableID: %w", err)
	}

	return nil
}

func (s *Service) ChangeImage(ctx context.Context, userID, trackID int64, imageFile *multipart.FileHeader) error {
	track, err := s.repo.GetByID(ctx, trackID, userID)
	if err != nil {
		return fmt.Errorf("TrackService.ChangeImage: %w", err)
	}

	hasPermission, err := s.repo.CheckPermission(ctx, userID, trackID)
	if err != nil {
		return fmt.Errorf("TrackService.ChangeImage: %w", err)
	}
	if !hasPermission {
		return errs.ErrPermissionDenied
	}

	imageName, err := s.fileService.SaveImage(ctx, imageFile)
	if err != nil {
		return fmt.Errorf("TrackService.ChangeImage: %w", err)
	}

	if err := s.repo.ChangeImage(ctx, trackID, imageName); err != nil {
		return fmt.Errorf("TrackService.ChangeImage: %w", err)
	}

	if err := s.fileService.DeleteFile(ctx, track.Image, fileservice.ImagesCategory); err != nil {
		return fmt.Errorf("TrackService.ChangeImage: %w", err)
	}

	return nil
}

func (s *Service) validateTrackTitle(ctx context.Context, userID int64, title string) error {
	exists, err := s.repo.CheckTitle(ctx, userID, title)
	if err != nil {
		return fmt.Errorf("TrackService.validateTrackTitle: %w", err)
	}
	if exists {
		return errs.ErrTitleExists
	}
	return nil
}

func (s *Service) validateChangeableID(ctx context.Context, userID int64, changeableID string) error {
	exists, err := s.repo.CheckChangeableID(ctx, userID, changeableID)
	if err != nil {
		return fmt.Errorf("TrackService.validateChangeableID: %w", err)
	}
	if exists {
		return errs.ErrChangeableIDExists
	}
	return nil
}

func (s *Service) GetManyLiked(ctx context.Context, currentUserID int64) ([]*models.UserLikedTrackModel, error) {
	likedTracks, err := s.repo.GetManyLiked(ctx, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("TrackService.GetManyLiked: %w", err)
	}
	return likedTracks, nil
}

func (s *Service) AddToLiked(ctx context.Context, currentUserID, trackID int64) error {
	_, err := s.repo.GetByID(ctx, trackID, currentUserID)
	if err != nil {
		return fmt.Errorf("TrackService.AddToLiked: %w", err)
	}

	if err := s.repo.AddToLiked(ctx, currentUserID, trackID); err != nil {
		return fmt.Errorf("TrackService.AddToLiked: %w", err)
	}
	return nil
}

func (s *Service) RemoveFromLiked(ctx context.Context, currentUserID, trackID int64) error {
	if err := s.repo.RemoveFromLiked(ctx, currentUserID, trackID); err != nil {
		return fmt.Errorf("TrackService.RemoveFromLiked: %w", err)
	}
	return nil
}
