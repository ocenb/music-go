package playlisttracks

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ocenb/music-go/content-service/internal/utils"
)

type PlaylistTracksHandlersInterface interface {
	GetMany(c *gin.Context)
	Add(c *gin.Context)
	UpdatePosition(c *gin.Context)
	Remove(c *gin.Context)
	RegisterHandlers(router *gin.RouterGroup)
}

type PlaylistTracksHandlers struct {
	playlistTracksService PlaylistTracksServiceInterface
}

func NewHandlers(playlistTracksService PlaylistTracksServiceInterface) PlaylistTracksHandlersInterface {
	return &PlaylistTracksHandlers{
		playlistTracksService: playlistTracksService,
	}
}

func (h *PlaylistTracksHandlers) GetMany(c *gin.Context) {
	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.UnauthenticatedError(c, err)
		return
	}

	var playlistReq PlaylistUri
	if err := c.ShouldBindUri(&playlistReq); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	var getManyReq GetManyForm
	if err := c.ShouldBindQuery(&getManyReq); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	tracks, err := h.playlistTracksService.GetMany(c, user.Id, playlistReq.PlaylistID, getManyReq.Take)
	if err != nil {
		switch {
		case errors.Is(err, ErrPlaylistNotFound):
			utils.NotFoundError(c, err)
		default:
			utils.InternalError(c, err)
		}

		return
	}

	c.JSON(http.StatusOK, tracks)
}

func (h *PlaylistTracksHandlers) Add(c *gin.Context) {
	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.UnauthenticatedError(c, err)
		return
	}

	var playlistTrackIDsReq PlaylistTrackIDsUri
	if err := c.ShouldBindUri(&playlistTrackIDsReq); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	var addTrackReq AddTrackJSON
	if err := c.ShouldBindJSON(&addTrackReq); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	playlistTrack, err := h.playlistTracksService.Add(c, user.Id, playlistTrackIDsReq.PlaylistID, playlistTrackIDsReq.TrackID, addTrackReq.Position)
	if err != nil {
		switch {
		case errors.Is(err, ErrPlaylistNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrTrackNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPermissionDenied):
			utils.PermissionDeniedError(c, err)
		case errors.Is(err, ErrTrackAlreadyInPlaylist):
			utils.BadRequestError(c, err)
		default:
			utils.InternalError(c, err)
		}

		return
	}

	c.JSON(http.StatusCreated, playlistTrack)
}

func (h *PlaylistTracksHandlers) UpdatePosition(c *gin.Context) {
	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.UnauthenticatedError(c, err)
		return
	}

	var playlistTrackIDsReq PlaylistTrackIDsUri
	if err := c.ShouldBindUri(&playlistTrackIDsReq); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	var updatePositionReq UpdatePositionJSON
	if err := c.ShouldBindJSON(&updatePositionReq); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	err = h.playlistTracksService.UpdatePosition(c, user.Id, playlistTrackIDsReq.PlaylistID, playlistTrackIDsReq.TrackID, updatePositionReq.Position)
	if err != nil {
		switch {
		case errors.Is(err, ErrPlaylistNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrTrackNotInPlaylist):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPermissionDenied):
			utils.PermissionDeniedError(c, err)
		case errors.Is(err, ErrPositionConflict):
			utils.BadRequestError(c, err)
		default:
			utils.InternalError(c, err)
		}

		return
	}

	c.Status(http.StatusOK)
}

func (h *PlaylistTracksHandlers) Remove(c *gin.Context) {
	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.UnauthenticatedError(c, err)
		return
	}

	var playlistTrackIDsReq PlaylistTrackIDsUri
	if err := c.ShouldBindUri(&playlistTrackIDsReq); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	err = h.playlistTracksService.Remove(c, user.Id, playlistTrackIDsReq.PlaylistID, playlistTrackIDsReq.TrackID)
	if err != nil {
		switch {
		case errors.Is(err, ErrPlaylistNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrTrackNotInPlaylist):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPermissionDenied):
			utils.PermissionDeniedError(c, err)
		default:
			utils.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusOK)
}

func (h *PlaylistTracksHandlers) RegisterHandlers(router *gin.RouterGroup) {
	playlistTracksRouter := router.Group("/playlist-tracks")
	playlistTracksRouter.GET("/:playlistId", h.GetMany)
	playlistTracksRouter.POST("/:playlistId/tracks/:trackId", h.Add)
	playlistTracksRouter.PUT("/:playlistId/tracks/:trackId/position", h.UpdatePosition)
	playlistTracksRouter.DELETE("/:playlistId/tracks/:trackId", h.Remove)
}
