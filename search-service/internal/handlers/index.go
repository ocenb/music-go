package handlers

import (
	"context"

	"github.com/ocenb/music-protos/gen/searchservice"
)

func (s *SearchServer) AddUser(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	if err := s.searchService.AddUser(ctx, req.Id, req.Name); err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SuccessResponse{Success: true}, nil
}

func (s *SearchServer) AddAlbum(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	if err := s.searchService.AddAlbum(ctx, req.Id, req.Name); err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SuccessResponse{Success: true}, nil
}

func (s *SearchServer) AddTrack(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	if err := s.searchService.AddTrack(ctx, req.Id, req.Name); err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SuccessResponse{Success: true}, nil
}

func (s *SearchServer) UpdateUser(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	if err := s.searchService.UpdateUser(ctx, req.Id, req.Name); err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SuccessResponse{Success: true}, nil
}

func (s *SearchServer) UpdateAlbum(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	if err := s.searchService.UpdateAlbum(ctx, req.Id, req.Name); err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SuccessResponse{Success: true}, nil
}

func (s *SearchServer) UpdateTrack(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	if err := s.searchService.UpdateTrack(ctx, req.Id, req.Name); err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SuccessResponse{Success: true}, nil
}

func (s *SearchServer) DeleteUser(ctx context.Context, req *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error) {
	if err := s.searchService.DeleteUser(ctx, req.Id); err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SuccessResponse{Success: true}, nil
}

func (s *SearchServer) DeleteAlbum(ctx context.Context, req *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error) {
	if err := s.searchService.DeleteAlbum(ctx, req.Id); err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SuccessResponse{Success: true}, nil
}

func (s *SearchServer) DeleteTrack(ctx context.Context, req *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error) {
	if err := s.searchService.DeleteTrack(ctx, req.Id); err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SuccessResponse{Success: true}, nil
}
