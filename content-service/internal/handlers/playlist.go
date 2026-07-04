package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/middlewares"
	"github.com/ocenb/music-go/content-service/internal/models"
)

func (h *ContentHandler) registerPlaylistRoutes(router *gin.RouterGroup) {
	playlistRouter := router.Group("/playlist")
	playlistRouter.GET("/one", h.getPlaylistOne)
	playlistRouter.GET("", h.getManyPlaylists)
	playlistRouter.GET("/with-saved", h.getManyPlaylistsWithSaved)
	playlistRouter.POST("", h.createPlaylist)
	playlistRouter.PATCH("/:playlistId/title", h.changePlaylistTitle)
	playlistRouter.PATCH("/:playlistId/changeable-id", h.changePlaylistChangeableID)
	playlistRouter.PATCH("/:playlistId/image", h.changePlaylistImage)
	playlistRouter.DELETE("/:playlistId", h.deletePlaylist)
	playlistRouter.POST("/:playlistId/save", h.savePlaylist)
	playlistRouter.DELETE("/:playlistId/save", h.removePlaylistFromSaved)
}

func (h *ContentHandler) getPlaylistOne(c *gin.Context) {
	var params models.PlaylistGetOneForm
	if err := c.ShouldBindQuery(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	playlistModel, err := h.playlistSvc.GetOne(c.Request.Context(), user.Id, params.Username, params.ChangeableID)
	if err != nil {
		if errors.Is(err, errs.ErrPlaylistNotFound) {
			handleError(c, errs.NotFound(err.Error()))
			return
		}
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, playlistModel)
}

func (h *ContentHandler) getManyPlaylists(c *gin.Context) {
	var params models.PlaylistGetManyForm
	if err := c.ShouldBindQuery(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	playlists, err := h.playlistSvc.GetMany(c.Request.Context(), params.UserID, user.Id, params.Take, params.LastID)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, playlists)
}

func (h *ContentHandler) getManyPlaylistsWithSaved(c *gin.Context) {
	var params models.PlaylistGetManyWithSavedForm
	if err := c.ShouldBindQuery(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	playlists, err := h.playlistSvc.GetManyWithSaved(c.Request.Context(), user.Id, params.Take, params.LastID)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, playlists)
}

func (h *ContentHandler) createPlaylist(c *gin.Context) {
	var request models.CreatePlaylistForm
	if err := c.ShouldBind(&request); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	playlistModel, err := h.playlistSvc.Create(
		c.Request.Context(),
		user.Id,
		user.Username,
		request.Title,
		request.ChangeableID,
		request.ImageFile,
	)
	if err != nil {
		for _, badRequestError := range errs.PlaylistCreateBadRequestErrors {
			if errors.Is(err, badRequestError) {
				handleError(c, errs.InvalidArgument(err.Error()))
				return
			}
		}
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusCreated, playlistModel)
}

func (h *ContentHandler) changePlaylistTitle(c *gin.Context) {
	var params models.PlaylistChangeTitleURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	var request models.PlaylistChangeTitleForm
	if err := c.ShouldBind(&request); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	err = h.playlistSvc.ChangeTitle(c.Request.Context(), user.Id, params.PlaylistID, request.Title)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPlaylistNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		case errors.Is(err, errs.ErrPlaylistAlreadyExists):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) changePlaylistChangeableID(c *gin.Context) {
	var params models.PlaylistChangeChangeableIDUri
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	var request models.PlaylistChangeChangeableIDForm
	if err := c.ShouldBind(&request); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	err = h.playlistSvc.ChangeChangeableID(c.Request.Context(), user.Id, params.PlaylistID, request.ChangeableID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPlaylistNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		case errors.Is(err, errs.ErrChangeableIDExists):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) changePlaylistImage(c *gin.Context) {
	var params models.PlaylistChangeImageURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	var request models.PlaylistChangeImageForm
	if err := c.ShouldBind(&request); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	err = h.playlistSvc.ChangeImage(c.Request.Context(), user.Id, params.PlaylistID, request.ImageFile)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPlaylistNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		case errors.Is(err, errs.ErrInvalidImageFormat):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) deletePlaylist(c *gin.Context) {
	var params models.PlaylistDeleteURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	if err := h.playlistSvc.Delete(c.Request.Context(), user.Id, params.PlaylistID); err != nil {
		switch {
		case errors.Is(err, errs.ErrPlaylistNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) savePlaylist(c *gin.Context) {
	var params models.GetByPlaylistIDURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	if err := h.playlistSvc.SavePlaylist(c.Request.Context(), user.Id, params.PlaylistID); err != nil {
		switch {
		case errors.Is(err, errs.ErrPlaylistNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPlaylistIsYours), errors.Is(err, errs.ErrPlaylistAlreadySaved):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) removePlaylistFromSaved(c *gin.Context) {
	var params models.GetByPlaylistIDURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	if err := h.playlistSvc.RemoveFromSaved(c.Request.Context(), user.Id, params.PlaylistID); err != nil {
		switch {
		case errors.Is(err, errs.ErrPlaylistNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPlaylistIsNotSaved):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}
