package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/middlewares"
	"github.com/ocenb/music-go/content-service/internal/models"
)

func (h *ContentHandler) registerPlaylistTracksRoutes(router *gin.RouterGroup) {
	playlistTracksRouter := router.Group("/playlist-tracks")
	playlistTracksRouter.GET("/:playlistId", h.getPlaylistTracks)
	playlistTracksRouter.POST("/:playlistId/tracks/:trackId", h.addPlaylistTrack)
	playlistTracksRouter.PUT("/:playlistId/tracks/:trackId/position", h.updatePlaylistTrackPosition)
	playlistTracksRouter.DELETE("/:playlistId/tracks/:trackId", h.removePlaylistTrack)
}

func (h *ContentHandler) getPlaylistTracks(c *gin.Context) {
	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Unauthenticated(err.Error()))
		return
	}

	var playlistReq models.GetByPlaylistIDURI
	if err := c.ShouldBindUri(&playlistReq); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	var getManyReq models.PlaylistTracksGetManyForm
	if err := c.ShouldBindQuery(&getManyReq); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	tracks, err := h.playlistTracksSvc.GetMany(c, user.Id, playlistReq.PlaylistID, getManyReq.Take)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPlaylistNotFound):
			handleError(c, errs.NotFound(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.JSON(http.StatusOK, tracks)
}

func (h *ContentHandler) addPlaylistTrack(c *gin.Context) {
	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Unauthenticated(err.Error()))
		return
	}

	var playlistTrackIDsReq models.PlaylistTrackIDsURI
	if err := c.ShouldBindUri(&playlistTrackIDsReq); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	var addTrackReq models.AddTrackJSON
	if err := c.ShouldBindJSON(&addTrackReq); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	playlistTrack, err := h.playlistTracksSvc.Add(
		c,
		user.Id,
		playlistTrackIDsReq.PlaylistID,
		playlistTrackIDsReq.TrackID,
		addTrackReq.Position,
	)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPlaylistNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrTrackNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		case errors.Is(err, errs.ErrTrackAlreadyInPlaylist):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.JSON(http.StatusCreated, playlistTrack)
}

func (h *ContentHandler) updatePlaylistTrackPosition(c *gin.Context) {
	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Unauthenticated(err.Error()))
		return
	}

	var playlistTrackIDsReq models.PlaylistTrackIDsURI
	if err := c.ShouldBindUri(&playlistTrackIDsReq); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	var updatePositionReq models.UpdatePositionJSON
	if err := c.ShouldBindJSON(&updatePositionReq); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	err = h.playlistTracksSvc.UpdatePosition(
		c,
		user.Id,
		playlistTrackIDsReq.PlaylistID,
		playlistTrackIDsReq.TrackID,
		updatePositionReq.Position,
	)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPlaylistNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrTrackNotInPlaylist):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		case errors.Is(err, errs.ErrPositionConflict):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusOK)
}

func (h *ContentHandler) removePlaylistTrack(c *gin.Context) {
	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Unauthenticated(err.Error()))
		return
	}

	var playlistTrackIDsReq models.PlaylistTrackIDsURI
	if err := c.ShouldBindUri(&playlistTrackIDsReq); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	err = h.playlistTracksSvc.Remove(c, user.Id, playlistTrackIDsReq.PlaylistID, playlistTrackIDsReq.TrackID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPlaylistNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrTrackNotInPlaylist):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusOK)
}
