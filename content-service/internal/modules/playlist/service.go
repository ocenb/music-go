package playlist

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"mime/multipart"

	"github.com/ocenb/music-go/content-service/internal/modules/file"
	"github.com/ocenb/music-go/content-service/internal/storage"
)

type PlaylistServiceInterface interface {
	GetOne(ctx context.Context, currentUserID int64, username, changeableID string) (*PlaylistWithSavedModel, error)
	GetMany(ctx context.Context, userID int64, currentUserID int64, take int, lastID int64) ([]*PlaylistWithSavedModel, error)
	GetManyWithSaved(ctx context.Context, currentUserID int64, take int, lastID int64) ([]*PlaylistWithSavedModel, error)
	Create(ctx context.Context, userID int64, username, title, changeableID string, imageFile *multipart.FileHeader) (*PlaylistModel, error)
	Delete(ctx context.Context, userID, playlistID int64) error
	ChangeTitle(ctx context.Context, userID, playlistID int64, title string) error
	ChangeChangeableId(ctx context.Context, userID, playlistID int64, changeableID string) error
	ChangeImage(ctx context.Context, userID, playlistID int64, imageFile *multipart.FileHeader) error
	SavePlaylist(ctx context.Context, userID, playlistID int64) error
	RemoveFromSaved(ctx context.Context, userID, playlistID int64) error
}

type PlaylistService struct {
	log          *slog.Logger
	playlistRepo PlaylistRepoInterface
	fileService  file.FileServiceInterface
}

func NewPlaylistService(log *slog.Logger, playlistRepo PlaylistRepoInterface, fileService file.FileServiceInterface) PlaylistServiceInterface {
	return &PlaylistService{
		log:          log,
		playlistRepo: playlistRepo,
		fileService:  fileService,
	}
}

func (s *PlaylistService) GetOne(ctx context.Context, currentUserID int64, username, changeableID string) (*PlaylistWithSavedModel, error) {
	playlist, err := s.playlistRepo.GetByChangeableID(ctx, username, changeableID, currentUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}

	return playlist, nil
}

func (s *PlaylistService) GetMany(ctx context.Context, userID int64, currentUserID int64, take int, lastID int64) ([]*PlaylistWithSavedModel, error) {
	playlists, err := s.playlistRepo.GetMany(ctx, userID, currentUserID, take, lastID)
	if err != nil {
		return nil, err
	}

	return playlists, nil
}

func (s *PlaylistService) GetManyWithSaved(ctx context.Context, currentUserID int64, take int, lastID int64) ([]*PlaylistWithSavedModel, error) {
	playlists, err := s.playlistRepo.GetManyWithSaved(ctx, currentUserID, take, lastID)
	if err != nil {
		return nil, err
	}

	return playlists, nil
}

func (s *PlaylistService) Create(ctx context.Context, userID int64, username, title, changeableID string, imageFile *multipart.FileHeader) (*PlaylistModel, error) {
	if err := s.validatePlaylistTitle(ctx, userID, title); err != nil {
		return nil, err
	}

	if err := s.validateChangeableId(ctx, userID, changeableID); err != nil {
		return nil, err
	}

	imageName, err := s.fileService.SaveImage(ctx, imageFile)
	if err != nil {
		return nil, err
	}

	playlist, err := s.playlistRepo.Create(ctx, userID, username, title, changeableID, imageName)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func (s *PlaylistService) Delete(ctx context.Context, userID, playlistID int64) error {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPlaylistNotFound
		}
		return err
	}

	hasPermission, err := s.playlistRepo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrPermissionDenied
	}

	err = storage.WithTransaction(ctx, s.playlistRepo, func(txCtx context.Context) error {
		if err := s.playlistRepo.Delete(txCtx, playlistID); err != nil {
			return err
		}

		err = s.fileService.DeleteFile(txCtx, playlist.Image, file.ImagesCategory)
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

func (s *PlaylistService) ChangeTitle(ctx context.Context, userID, playlistID int64, title string) error {
	hasPermission, err := s.playlistRepo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrPermissionDenied
	}

	if err := s.validatePlaylistTitle(ctx, userID, title); err != nil {
		return err
	}

	if err := s.playlistRepo.ChangeTitle(ctx, playlistID, title); err != nil {
		return err
	}

	return nil
}

func (s *PlaylistService) ChangeChangeableId(ctx context.Context, userID, playlistID int64, changeableID string) error {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPlaylistNotFound
		}
		return err
	}

	hasPermission, err := s.playlistRepo.CheckPermission(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrPermissionDenied
	}

	if err := s.validateChangeableId(ctx, playlist.UserID, changeableID); err != nil {
		return err
	}

	if err := s.playlistRepo.ChangeChangeableID(ctx, playlistID, changeableID); err != nil {
		return err
	}

	return nil
}

func (s *PlaylistService) ChangeImage(ctx context.Context, userID, playlistID int64, imageFile *multipart.FileHeader) error {
	hasPermission, err := s.playlistRepo.CheckPermission(ctx, userID, playlistID)
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

	if err := s.playlistRepo.ChangeImage(ctx, playlistID, imageName); err != nil {
		return err
	}

	if err := s.fileService.DeleteFile(ctx, imageName, file.ImagesCategory); err != nil {
		return err
	}

	return nil
}

func (s *PlaylistService) SavePlaylist(ctx context.Context, userID, playlistID int64) error {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPlaylistNotFound
		}
		return err
	}
	if playlist.UserID == userID {
		return ErrPlaylistIsYours
	}
	if playlist.IsSaved {
		return ErrPlaylistAlreadySaved
	}

	return s.playlistRepo.SavePlaylist(ctx, userID, playlistID)
}

func (s *PlaylistService) RemoveFromSaved(ctx context.Context, userID, playlistID int64) error {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPlaylistNotFound
		}
		return err
	}
	if !playlist.IsSaved {
		return ErrPlaylistIsNotSaved
	}

	return s.playlistRepo.RemoveFromSaved(ctx, userID, playlistID)
}

func (s *PlaylistService) validatePlaylistTitle(ctx context.Context, userID int64, title string) error {
	exists, err := s.playlistRepo.CheckTitle(ctx, userID, title)
	if err != nil {
		return err
	}
	if exists {
		return ErrPlaylistAlreadyExists
	}

	return nil
}

func (s *PlaylistService) validateChangeableId(ctx context.Context, userID int64, changeableID string) error {
	exists, err := s.playlistRepo.CheckChangeableID(ctx, userID, changeableID)
	if err != nil {
		return err
	}
	if exists {
		return ErrChangeableIDExists
	}

	return nil
}
