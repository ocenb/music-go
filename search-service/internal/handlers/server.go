package handlers

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/ocenb/music-protos/gen/searchservice"

	"github.com/ocenb/music-go/search-service/internal/errs"
	"github.com/ocenb/music-go/search-service/internal/logger"
	"github.com/ocenb/music-go/search-service/internal/logger/logattr"
)

type SearchService interface {
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

type SearchServer struct {
	searchservice.UnimplementedSearchServiceServer
	searchService SearchService
}

func NewSearchServer(gRPCServer *grpc.Server, searchService SearchService) {
	searchservice.RegisterSearchServiceServer(gRPCServer, &SearchServer{searchService: searchService})
}

func handleError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	log := logger.FromContext(ctx)
	if logErr, ok := errors.AsType[*LogWrappedError](err); ok {
		log = logErr.Logger
		err = logErr.Err
	}

	domainErr, ok := errs.As(err)
	if ok && domainErr.Kind() == errs.KindInternal {
		log.ErrorContext(ctx, "internal server error", logattr.Err(err))
	}

	return errs.ToGRPC(err)
}

type LogWrappedError struct {
	Err    error
	Logger *slog.Logger
}

func (e *LogWrappedError) Error() string { return e.Err.Error() }
func (e *LogWrappedError) Unwrap() error { return e.Err }
