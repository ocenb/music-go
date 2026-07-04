package search

import (
	"context"
	"fmt"

	"github.com/ocenb/music-protos/gen/searchservice"
)

type Client interface {
	SearchUsers(ctx context.Context, req *searchservice.SearchRequest) (*searchservice.SearchResponse, error)
	SearchTracks(ctx context.Context, req *searchservice.SearchRequest) (*searchservice.SearchResponse, error)
}

type Service struct {
	client Client
}

func New(client Client) *Service {
	return &Service{client: client}
}

func (s *Service) SearchUsers(ctx context.Context, query string) ([]int64, error) {
	response, err := s.client.SearchUsers(ctx, &searchservice.SearchRequest{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("SearchService.SearchUsers: %w", err)
	}
	return response.Ids, nil
}

func (s *Service) SearchTracks(ctx context.Context, query string) ([]int64, error) {
	response, err := s.client.SearchTracks(ctx, &searchservice.SearchRequest{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("SearchService.SearchTracks: %w", err)
	}
	return response.Ids, nil
}
