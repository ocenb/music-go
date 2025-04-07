package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ocenb/music-go/search-service/internal/clients/elastic"
	"github.com/ocenb/music-go/search-service/internal/config"
	"github.com/ocenb/music-go/search-service/internal/utils"
)

type SearchServiceInterface interface {
	SearchUsers(ctx context.Context, query string) ([]int64, error)
	SearchAlbums(ctx context.Context, query string) ([]int64, error)
	SearchTracks(ctx context.Context, query string) ([]int64, error)
	AddUser(ctx context.Context, id int64, username string) error
	AddAlbum(ctx context.Context, id int64, title string) error
	AddTrack(ctx context.Context, id int64, title string) error
	UpdateUser(ctx context.Context, id int64, username string) error
	UpdateAlbum(ctx context.Context, id int64, title string) error
	UpdateTrack(ctx context.Context, id int64, title string) error
	DeleteUser(ctx context.Context, id int64) error
	DeleteAlbum(ctx context.Context, id int64) error
	DeleteTrack(ctx context.Context, id int64) error
}

type SearchService struct {
	cfg     *config.Config
	log     *slog.Logger
	elastic *elastic.ElasticClient
}

func NewSearchService(cfg *config.Config, log *slog.Logger, elastic *elastic.ElasticClient) *SearchService {
	return &SearchService{
		cfg:     cfg,
		log:     log,
		elastic: elastic,
	}
}

func (s *SearchService) SearchUsers(ctx context.Context, query string) ([]int64, error) {
	s.log.Info("Searching users", slog.String("query", query))
	ids, err := s.elastic.SearchUsers(ctx, query)
	if err != nil {
		return nil, utils.InternalError(err, "failed to search users")
	}
	return ids, nil
}

func (s *SearchService) SearchAlbums(ctx context.Context, query string) ([]int64, error) {
	s.log.Info("Searching albums", slog.String("query", query))
	ids, err := s.elastic.SearchAlbums(ctx, query)
	if err != nil {
		return nil, utils.InternalError(err, "failed to search albums")
	}
	return ids, nil
}

func (s *SearchService) SearchTracks(ctx context.Context, query string) ([]int64, error) {
	s.log.Info("Searching tracks", slog.String("query", query))
	ids, err := s.elastic.SearchTracks(ctx, query)
	if err != nil {
		return nil, utils.InternalError(err, "failed to search tracks")
	}
	return ids, nil
}

func (s *SearchService) AddUser(ctx context.Context, id int64, username string) error {
	s.log.Info("Adding user", slog.Int64("id", id), slog.String("username", username))
	err := s.elastic.AddUser(ctx, id, username)
	if err != nil {
		if errors.Is(err, elastic.ErrUserAlreadyExists) {
			return utils.AlreadyExistsError(err)
		}
		return utils.InternalError(err, "failed to add user")
	}
	return nil
}

func (s *SearchService) AddAlbum(ctx context.Context, id int64, title string) error {
	s.log.Info("Adding album", slog.Int64("id", id), slog.String("title", title))
	err := s.elastic.AddAlbum(ctx, id, title)
	if err != nil {
		if errors.Is(err, elastic.ErrAlbumAlreadyExists) {
			return utils.AlreadyExistsError(err)
		}
		return utils.InternalError(err, "failed to add album")
	}
	return nil
}

func (s *SearchService) AddTrack(ctx context.Context, id int64, title string) error {
	s.log.Info("Adding track", slog.Int64("id", id), slog.String("title", title))
	err := s.elastic.AddTrack(ctx, id, title)
	if err != nil {
		if errors.Is(err, elastic.ErrTrackAlreadyExists) {
			return utils.AlreadyExistsError(err)
		}
		return utils.InternalError(err, "failed to add track")
	}
	return nil
}

func (s *SearchService) UpdateUser(ctx context.Context, id int64, username string) error {
	s.log.Info("Updating user", slog.Int64("id", id), slog.String("username", username))
	err := s.elastic.UpdateUser(ctx, id, username)
	if err != nil {
		return utils.InternalError(err, "failed to update user")
	}
	return nil
}

func (s *SearchService) UpdateAlbum(ctx context.Context, id int64, title string) error {
	s.log.Info("Updating album", slog.Int64("id", id), slog.String("title", title))
	err := s.elastic.UpdateAlbum(ctx, id, title)
	if err != nil {
		return utils.InternalError(err, "failed to update album")
	}
	return nil
}

func (s *SearchService) UpdateTrack(ctx context.Context, id int64, title string) error {
	s.log.Info("Updating track", slog.Int64("id", id), slog.String("title", title))
	err := s.elastic.UpdateTrack(ctx, id, title)
	if err != nil {
		return utils.InternalError(err, "failed to update track")
	}
	return nil
}

func (s *SearchService) DeleteUser(ctx context.Context, id int64) error {
	s.log.Info("Deleting user", slog.Int64("id", id))
	err := s.elastic.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, elastic.ErrUserNotFound) {
			return utils.NotFoundError(err)
		}
		return utils.InternalError(err, "failed to delete user")
	}
	return nil
}

func (s *SearchService) DeleteAlbum(ctx context.Context, id int64) error {
	s.log.Info("Deleting album", slog.Int64("id", id))
	err := s.elastic.DeleteAlbum(ctx, id)
	if err != nil {
		if errors.Is(err, elastic.ErrAlbumNotFound) {
			return utils.NotFoundError(err)
		}
		return utils.InternalError(err, "failed to delete album")
	}
	return nil
}

func (s *SearchService) DeleteTrack(ctx context.Context, id int64) error {
	s.log.Info("Deleting track", slog.Int64("id", id))
	err := s.elastic.DeleteTrack(ctx, id)
	if err != nil {
		if errors.Is(err, elastic.ErrTrackNotFound) {
			return utils.NotFoundError(err)
		}
		return utils.InternalError(err, "failed to delete track")
	}
	return nil
}
