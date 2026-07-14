package playlist

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/models"
	fileservice "github.com/ocenb/music-go/content-service/internal/services/file"
	"github.com/ocenb/music-go/content-service/internal/storage/transactor"
)

type Repo interface {
	GetByID(ctx context.Context, playlistID int64, currentUserID int64) (*models.PlaylistWithSavedModel, error)
	GetByChangeableID(ctx context.Context, username, changeableID string, currentUserID int64) (*models.PlaylistWithSavedModel, error)
	GetMany(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*models.PlaylistWithSavedModel, error)
	GetManyWithSaved(ctx context.Context, userID int64, take int, lastID int64) ([]*models.PlaylistWithSavedModel, error)
	Create(ctx context.Context, userID int64, username, title, changeableID, image string) (*models.PlaylistModel, error)
	CheckPermission(ctx context.Context, userID, playlistID int64) (bool, error)
	Delete(ctx context.Context, playlistID int64) error
	ChangeTitle(ctx context.Context, playlistID int64, title string) error
	ChangeChangeableID(ctx context.Context, playlistID int64, changeableID string) error
	ChangeImage(ctx context.Context, playlistID int64, image string) error
	CheckTitle(ctx context.Context, userID int64, title string) (bool, error)
	CheckChangeableID(ctx context.Context, userID int64, changeableID string) (bool, error)
	SavePlaylist(ctx context.Context, userID, playlistID int64) error
	RemoveFromSaved(ctx context.Context, userID, playlistID int64) error
}

type FileService interface {
	SaveImage(ctx context.Context, file *multipart.FileHeader) (string, error)
	DeleteFile(ctx context.Context, fileName string, category fileservice.Category) error
}

type Service struct {
	repo        Repo
	fileService FileService
	tm          transactor.Runner
}

func New(repo Repo, fileService FileService, tm transactor.Runner) *Service {
	return &Service{
		repo:        repo,
		fileService: fileService,
		tm:          tm,
	}
}

func (s *Service) GetOne(ctx context.Context, currentUserID int64, username, changeableID string) (*models.PlaylistWithSavedModel, error) {
	playlist, err := s.repo.GetByChangeableID(ctx, username, changeableID, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("PlaylistService.GetOne: %w", err)
	}
	return playlist, nil
}

func (s *Service) GetMany(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*models.PlaylistWithSavedModel, error) {
	playlists, err := s.repo.GetMany(ctx, userID, currentUserID, take, lastID)
	if err != nil {
		return nil, fmt.Errorf("PlaylistService.GetMany: %w", err)
	}
	return playlists, nil
}

func (s *Service) GetManyWithSaved(ctx context.Context, currentUserID int64, take int, lastID int64) ([]*models.PlaylistWithSavedModel, error) {
	playlists, err := s.repo.GetManyWithSaved(ctx, currentUserID, take, lastID)
	if err != nil {
		return nil, fmt.Errorf("PlaylistService.GetManyWithSaved: %w", err)
	}
	return playlists, nil
}

func (s *Service) Create(ctx context.Context, userID int64, username, title, changeableID string, imageFile *multipart.FileHeader) (*models.PlaylistModel, error) {
	if err := s.validatePlaylistTitle(ctx, userID, title); err != nil {
		return nil, err
	}

	if err := s.validateChangeableID(ctx, userID, changeableID); err != nil {
		return nil, err
	}

	imageName, err := s.fileService.SaveImage(ctx, imageFile)
	if err != nil {
		return nil, fmt.Errorf("PlaylistService.Create: %w", err)
	}

	playlist, err := s.repo.Create(ctx, userID, username, title, changeableID, imageName)
	if err != nil {
		return nil, fmt.Errorf("PlaylistService.Create: %w", err)
	}

	return playlist, nil
}

func (s *Service) Delete(ctx context.Context, userID, playlistID int64) error {
	playlist, err := s.repo.GetByID(ctx, playlistID, userID)
	if err != nil {
		return fmt.Errorf("PlaylistService.Delete: %w", err)
	}

	hasPermission, err := s.repo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return fmt.Errorf("PlaylistService.Delete: %w", err)
	}
	if !hasPermission {
		return errs.ErrPermissionDenied
	}

	err = s.tm.Run(ctx, func(txCtx context.Context) error {
		if delErr := s.repo.Delete(txCtx, playlistID); delErr != nil {
			return delErr
		}

		return s.fileService.DeleteFile(txCtx, playlist.Image, fileservice.ImagesCategory)
	})
	if err != nil {
		return fmt.Errorf("PlaylistService.Delete: %w", err)
	}
	return nil
}

func (s *Service) ChangeTitle(ctx context.Context, userID, playlistID int64, title string) error {
	hasPermission, err := s.repo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return fmt.Errorf("PlaylistService.ChangeTitle: %w", err)
	}
	if !hasPermission {
		return errs.ErrPermissionDenied
	}

	if err := s.validatePlaylistTitle(ctx, userID, title); err != nil {
		return err
	}

	if err := s.repo.ChangeTitle(ctx, playlistID, title); err != nil {
		return fmt.Errorf("PlaylistService.ChangeTitle: %w", err)
	}

	return nil
}

func (s *Service) ChangeChangeableID(ctx context.Context, userID, playlistID int64, changeableID string) error {
	playlist, err := s.repo.GetByID(ctx, playlistID, userID)
	if err != nil {
		return fmt.Errorf("PlaylistService.ChangeChangeableID: %w", err)
	}

	hasPermission, err := s.repo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return fmt.Errorf("PlaylistService.ChangeChangeableID: %w", err)
	}
	if !hasPermission {
		return errs.ErrPermissionDenied
	}

	if err := s.validateChangeableID(ctx, playlist.UserID, changeableID); err != nil {
		return err
	}

	if err := s.repo.ChangeChangeableID(ctx, playlistID, changeableID); err != nil {
		return fmt.Errorf("PlaylistService.ChangeChangeableID: %w", err)
	}

	return nil
}

func (s *Service) ChangeImage(ctx context.Context, userID, playlistID int64, imageFile *multipart.FileHeader) error {
	hasPermission, err := s.repo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return fmt.Errorf("PlaylistService.ChangeImage: %w", err)
	}
	if !hasPermission {
		return errs.ErrPermissionDenied
	}

	imageName, err := s.fileService.SaveImage(ctx, imageFile)
	if err != nil {
		return fmt.Errorf("PlaylistService.ChangeImage: %w", err)
	}

	if err := s.repo.ChangeImage(ctx, playlistID, imageName); err != nil {
		return fmt.Errorf("PlaylistService.ChangeImage: %w", err)
	}

	if err := s.fileService.DeleteFile(ctx, imageName, fileservice.ImagesCategory); err != nil {
		return fmt.Errorf("PlaylistService.ChangeImage: %w", err)
	}

	return nil
}

func (s *Service) SavePlaylist(ctx context.Context, userID, playlistID int64) error {
	playlist, err := s.repo.GetByID(ctx, playlistID, userID)
	if err != nil {
		return fmt.Errorf("PlaylistService.SavePlaylist: %w", err)
	}
	if playlist.UserID == userID {
		return errs.ErrPlaylistIsYours
	}
	if playlist.IsSaved {
		return errs.ErrPlaylistAlreadySaved
	}

	if err := s.repo.SavePlaylist(ctx, userID, playlistID); err != nil {
		return fmt.Errorf("PlaylistService.SavePlaylist: %w", err)
	}
	return nil
}

func (s *Service) RemoveFromSaved(ctx context.Context, userID, playlistID int64) error {
	playlist, err := s.repo.GetByID(ctx, playlistID, userID)
	if err != nil {
		return fmt.Errorf("PlaylistService.RemoveFromSaved: %w", err)
	}
	if !playlist.IsSaved {
		return errs.ErrPlaylistNotSaved
	}

	if err := s.repo.RemoveFromSaved(ctx, userID, playlistID); err != nil {
		return fmt.Errorf("PlaylistService.RemoveFromSaved: %w", err)
	}
	return nil
}

func (s *Service) validatePlaylistTitle(ctx context.Context, userID int64, title string) error {
	exists, err := s.repo.CheckTitle(ctx, userID, title)
	if err != nil {
		return fmt.Errorf("PlaylistService.validatePlaylistTitle: %w", err)
	}
	if exists {
		return errs.ErrTitleExists
	}
	return nil
}

func (s *Service) validateChangeableID(ctx context.Context, userID int64, changeableID string) error {
	exists, err := s.repo.CheckChangeableID(ctx, userID, changeableID)
	if err != nil {
		return fmt.Errorf("PlaylistService.validateChangeableID: %w", err)
	}
	if exists {
		return errs.ErrChangeableIDExists
	}
	return nil
}
