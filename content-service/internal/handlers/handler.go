package handlers

import (
	"context"
	"mime/multipart"

	"github.com/gin-gonic/gin"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/logger"
	"github.com/ocenb/music-go/content-service/internal/logger/logattr"
	"github.com/ocenb/music-go/content-service/internal/middlewares"
	"github.com/ocenb/music-go/content-service/internal/models"
)

type TrackService interface {
	GetOneByID(ctx context.Context, currentUserID, trackID int64) (*models.TrackWithLikedModel, error)
	GetOne(ctx context.Context, currentUserID int64, username, changeableID string) (*models.TrackWithLikedModel, error)
	GetMany(ctx context.Context, currentUserID, userID int64, take int, lastID int64) ([]*models.TrackWithLikedModel, error)
	GetManyPopular(ctx context.Context, currentUserID, userID int64, take int, lastID int64) ([]*models.TrackWithLikedModel, error)
	Upload(ctx context.Context, userID int64, username, email, title, changeableID string, audioFile, imageFile *multipart.FileHeader) (*models.TrackModel, error)
	AddPlay(ctx context.Context, trackID int64) error
	Delete(ctx context.Context, userID, trackID int64) error
	ChangeTitle(ctx context.Context, userID, trackID int64, title string) error
	ChangeChangeableID(ctx context.Context, userID, trackID int64, changeableID string) error
	ChangeImage(ctx context.Context, userID, trackID int64, imageFile *multipart.FileHeader) error
	GetManyLiked(ctx context.Context, currentUserID int64) ([]*models.UserLikedTrackModel, error)
	AddToLiked(ctx context.Context, currentUserID, trackID int64) error
	RemoveFromLiked(ctx context.Context, currentUserID, trackID int64) error
}

type PlaylistService interface {
	GetOne(ctx context.Context, currentUserID int64, username, changeableID string) (*models.PlaylistWithSavedModel, error)
	GetMany(ctx context.Context, userID, currentUserID int64, take int, lastID int64) ([]*models.PlaylistWithSavedModel, error)
	GetManyWithSaved(ctx context.Context, currentUserID int64, take int, lastID int64) ([]*models.PlaylistWithSavedModel, error)
	Create(ctx context.Context, userID int64, username, title, changeableID string, imageFile *multipart.FileHeader) (*models.PlaylistModel, error)
	Delete(ctx context.Context, userID, playlistID int64) error
	ChangeTitle(ctx context.Context, userID, playlistID int64, title string) error
	ChangeChangeableID(ctx context.Context, userID, playlistID int64, changeableID string) error
	ChangeImage(ctx context.Context, userID, playlistID int64, imageFile *multipart.FileHeader) error
	SavePlaylist(ctx context.Context, userID, playlistID int64) error
	RemoveFromSaved(ctx context.Context, userID, playlistID int64) error
}

type PlaylistTracksService interface {
	GetMany(ctx context.Context, currentUserID, playlistID int64, take int) ([]*models.TrackInPlaylistModel, error)
	Add(ctx context.Context, userID, playlistID, trackID int64, position int) (*models.PlaylistTrackModel, error)
	UpdatePosition(ctx context.Context, userID, playlistID, trackID int64, position int) error
	Remove(ctx context.Context, userID, playlistID, trackID int64) error
}

type HistoryService interface {
	Get(ctx context.Context, currentUserID, take int64) ([]*models.ListeningHistoryModel, error)
	Add(ctx context.Context, currentUserID, trackID int64) error
	Clear(ctx context.Context, currentUserID int64) error
}

type SearchService interface {
	SearchUsers(ctx context.Context, query string) ([]int64, error)
	SearchTracks(ctx context.Context, query string) ([]int64, error)
}

type AllService interface {
	DeleteAll(ctx context.Context, userID int64) error
}

type ContentHandler struct {
	trackSvc          TrackService
	playlistSvc       PlaylistService
	playlistTracksSvc PlaylistTracksService
	historySvc        HistoryService
	searchSvc         SearchService
	allSvc            AllService
	userClient        middlewares.UserClient
	internalSecret    string
}

func NewContentHandler(
	trackSvc TrackService,
	playlistSvc PlaylistService,
	playlistTracksSvc PlaylistTracksService,
	historySvc HistoryService,
	searchSvc SearchService,
	allSvc AllService,
	userClient middlewares.UserClient,
	internalSecret string,
) *ContentHandler {
	return &ContentHandler{
		trackSvc:          trackSvc,
		playlistSvc:       playlistSvc,
		playlistTracksSvc: playlistTracksSvc,
		historySvc:        historySvc,
		searchSvc:         searchSvc,
		allSvc:            allSvc,
		userClient:        userClient,
		internalSecret:    internalSecret,
	}
}

func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	log := logger.FromContext(c.Request.Context())
	if domainErr, ok := errs.As(err); ok && domainErr.Kind() == errs.KindInternal {
		log.ErrorContext(c.Request.Context(), "internal server error", logattr.Err(err))
	}

	c.JSON(errs.HTTPStatus(err), gin.H{"error": errs.HTTPMessage(err)})
}

func (h *ContentHandler) RegisterRoutes(engine *gin.Engine) {
	api := engine.Group("/api/content")
	api.Use(middlewares.Auth(h.userClient))
	h.registerTrackRoutes(api)
	h.registerPlaylistRoutes(api)
	h.registerPlaylistTracksRoutes(api)
	h.registerHistoryRoutes(api)
	h.registerSearchRoutes(api)
	h.registerAllRoutes(api)

	internalAPI := engine.Group("/api/content/internal")
	internalAPI.Use(middlewares.InternalAuth(h.internalSecret))
	h.registerInternalRoutes(internalAPI)
}
