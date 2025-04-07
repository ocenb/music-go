package handlers

import (
	"context"
	"log/slog"

	"github.com/ocenb/music-go/search-service/internal/config"
	"github.com/ocenb/music-go/search-service/internal/service"
	"github.com/ocenb/music-go/search-service/internal/utils"
	"github.com/ocenb/music-protos/gen/searchservice"
	"google.golang.org/grpc"
)

type SearchServer struct {
	searchservice.UnimplementedSearchServiceServer
	searchService service.SearchServiceInterface
	cfg           *config.Config
	log           *slog.Logger
}

func NewSearchServer(gRPCServer *grpc.Server, cfg *config.Config, searchService service.SearchServiceInterface, log *slog.Logger) {
	searchservice.RegisterSearchServiceServer(gRPCServer, &SearchServer{
		searchService: searchService,
		cfg:           cfg,
		log:           log,
	})
}

func (s *SearchServer) SearchUsers(ctx context.Context, req *searchservice.SearchRequest) (*searchservice.SearchResponse, error) {
	s.log.Info("Received search users request", slog.String("query", req.Query))

	ids, err := s.searchService.SearchUsers(ctx, req.Query)
	if err != nil {
		s.log.Error("Failed to search users", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SearchResponse{
		Ids: ids,
	}, nil
}

func (s *SearchServer) SearchAlbums(ctx context.Context, req *searchservice.SearchRequest) (*searchservice.SearchResponse, error) {
	s.log.Info("Received search albums request", slog.String("query", req.Query))

	ids, err := s.searchService.SearchAlbums(ctx, req.Query)
	if err != nil {
		s.log.Error("Failed to search albums", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SearchResponse{
		Ids: ids,
	}, nil
}

func (s *SearchServer) SearchTracks(ctx context.Context, req *searchservice.SearchRequest) (*searchservice.SearchResponse, error) {
	s.log.Info("Received search tracks request", slog.String("query", req.Query))

	ids, err := s.searchService.SearchTracks(ctx, req.Query)
	if err != nil {
		s.log.Error("Failed to search tracks", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SearchResponse{
		Ids: ids,
	}, nil
}

func (s *SearchServer) AddUser(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	s.log.Info("Received add user request", slog.Int64("user_id", req.Id), slog.String("username", req.Name))

	err := s.searchService.AddUser(ctx, req.Id, req.Name)
	if err != nil {
		s.log.Error("Failed to add user", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SuccessResponse{
		Success: true,
	}, nil
}

func (s *SearchServer) AddAlbum(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	s.log.Info("Received add album request", slog.Int64("album_id", req.Id), slog.String("title", req.Name))

	err := s.searchService.AddAlbum(ctx, req.Id, req.Name)
	if err != nil {
		s.log.Error("Failed to add album", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SuccessResponse{
		Success: true,
	}, nil
}

func (s *SearchServer) AddTrack(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	s.log.Info("Received add track request", slog.Int64("track_id", req.Id), slog.String("title", req.Name))

	err := s.searchService.AddTrack(ctx, req.Id, req.Name)
	if err != nil {
		s.log.Error("Failed to add track", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SuccessResponse{
		Success: true,
	}, nil
}

func (s *SearchServer) UpdateUser(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	s.log.Info("Received update user request", slog.Int64("user_id", req.Id), slog.String("username", req.Name))

	err := s.searchService.UpdateUser(ctx, req.Id, req.Name)
	if err != nil {
		s.log.Error("Failed to update user", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SuccessResponse{
		Success: true,
	}, nil
}

func (s *SearchServer) UpdateAlbum(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	s.log.Info("Received update album request", slog.Int64("album_id", req.Id), slog.String("title", req.Name))

	err := s.searchService.UpdateAlbum(ctx, req.Id, req.Name)
	if err != nil {
		s.log.Error("Failed to update album", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SuccessResponse{
		Success: true,
	}, nil
}

func (s *SearchServer) UpdateTrack(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	s.log.Info("Received update track request", slog.Int64("track_id", req.Id), slog.String("title", req.Name))

	err := s.searchService.UpdateTrack(ctx, req.Id, req.Name)
	if err != nil {
		s.log.Error("Failed to update track", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SuccessResponse{
		Success: true,
	}, nil
}

func (s *SearchServer) DeleteUser(ctx context.Context, req *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error) {
	s.log.Info("Received delete user request", slog.Int64("user_id", req.Id))

	err := s.searchService.DeleteUser(ctx, req.Id)
	if err != nil {
		s.log.Error("Failed to delete user", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SuccessResponse{
		Success: true,
	}, nil
}

func (s *SearchServer) DeleteAlbum(ctx context.Context, req *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error) {
	s.log.Info("Received delete album request", slog.Int64("album_id", req.Id))

	err := s.searchService.DeleteAlbum(ctx, req.Id)
	if err != nil {
		s.log.Error("Failed to delete album", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SuccessResponse{
		Success: true,
	}, nil
}

func (s *SearchServer) DeleteTrack(ctx context.Context, req *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error) {
	s.log.Info("Received delete track request", slog.Int64("track_id", req.Id))

	err := s.searchService.DeleteTrack(ctx, req.Id)
	if err != nil {
		s.log.Error("Failed to delete track", utils.ErrLog(err))
		return nil, err
	}

	return &searchservice.SuccessResponse{
		Success: true,
	}, nil
}
