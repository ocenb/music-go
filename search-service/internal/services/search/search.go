package search

import (
	"context"
	"errors"
	"fmt"

	"github.com/ocenb/music-go/search-service/internal/errs"
)

type Repo interface {
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

type Service struct {
	repo Repo
}

func New(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) SearchUsers(ctx context.Context, query string) ([]int64, error) {
	ids, err := s.repo.SearchUsers(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SearchService.SearchUsers: %w", err)
	}
	return ids, nil
}

func (s *Service) SearchAlbums(ctx context.Context, query string) ([]int64, error) {
	ids, err := s.repo.SearchAlbums(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SearchService.SearchAlbums: %w", err)
	}
	return ids, nil
}

func (s *Service) SearchTracks(ctx context.Context, query string) ([]int64, error) {
	ids, err := s.repo.SearchTracks(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SearchService.SearchTracks: %w", err)
	}
	return ids, nil
}

func (s *Service) AddUser(ctx context.Context, id int64, username string) error {
	err := s.repo.AddUser(ctx, id, username)
	if err != nil {
		if errors.Is(err, errs.ErrUserAlreadyExists) {
			return err
		}
		return fmt.Errorf("SearchService.AddUser: %w", err)
	}
	return nil
}

func (s *Service) AddAlbum(ctx context.Context, id int64, title string) error {
	err := s.repo.AddAlbum(ctx, id, title)
	if err != nil {
		if errors.Is(err, errs.ErrAlbumAlreadyExists) {
			return err
		}
		return fmt.Errorf("SearchService.AddAlbum: %w", err)
	}
	return nil
}

func (s *Service) AddTrack(ctx context.Context, id int64, title string) error {
	err := s.repo.AddTrack(ctx, id, title)
	if err != nil {
		if errors.Is(err, errs.ErrTrackAlreadyExists) {
			return err
		}
		return fmt.Errorf("SearchService.AddTrack: %w", err)
	}
	return nil
}

func (s *Service) UpdateUser(ctx context.Context, id int64, username string) error {
	if err := s.repo.UpdateUser(ctx, id, username); err != nil {
		return fmt.Errorf("SearchService.UpdateUser: %w", err)
	}
	return nil
}

func (s *Service) UpdateAlbum(ctx context.Context, id int64, title string) error {
	if err := s.repo.UpdateAlbum(ctx, id, title); err != nil {
		return fmt.Errorf("SearchService.UpdateAlbum: %w", err)
	}
	return nil
}

func (s *Service) UpdateTrack(ctx context.Context, id int64, title string) error {
	if err := s.repo.UpdateTrack(ctx, id, title); err != nil {
		return fmt.Errorf("SearchService.UpdateTrack: %w", err)
	}
	return nil
}

func (s *Service) DeleteUser(ctx context.Context, id int64) error {
	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("SearchService.DeleteUser: %w", err)
	}
	return nil
}

func (s *Service) DeleteAlbum(ctx context.Context, id int64) error {
	err := s.repo.DeleteAlbum(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrAlbumNotFound) {
			return err
		}
		return fmt.Errorf("SearchService.DeleteAlbum: %w", err)
	}
	return nil
}

func (s *Service) DeleteTrack(ctx context.Context, id int64) error {
	err := s.repo.DeleteTrack(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrTrackNotFound) {
			return err
		}
		return fmt.Errorf("SearchService.DeleteTrack: %w", err)
	}
	return nil
}
