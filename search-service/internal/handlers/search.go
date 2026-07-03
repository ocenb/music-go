package handlers

import (
	"context"

	"github.com/ocenb/music-protos/gen/searchservice"
)

func (s *SearchServer) SearchUsers(ctx context.Context, req *searchservice.SearchRequest) (*searchservice.SearchResponse, error) {
	ids, err := s.searchService.SearchUsers(ctx, req.Query)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SearchResponse{Ids: ids}, nil
}

func (s *SearchServer) SearchAlbums(ctx context.Context, req *searchservice.SearchRequest) (*searchservice.SearchResponse, error) {
	ids, err := s.searchService.SearchAlbums(ctx, req.Query)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SearchResponse{Ids: ids}, nil
}

func (s *SearchServer) SearchTracks(ctx context.Context, req *searchservice.SearchRequest) (*searchservice.SearchResponse, error) {
	ids, err := s.searchService.SearchTracks(ctx, req.Query)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	return &searchservice.SearchResponse{Ids: ids}, nil
}
