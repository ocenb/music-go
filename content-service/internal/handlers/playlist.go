package handlers

import (
	"errors"
	"mime/multipart"
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
	respondGetOneByChangeableID(c, errs.ErrPlaylistNotFound, h.playlistSvc.GetOne)
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
	handleChangeTitle(
		c,
		func(p models.PlaylistChangeTitleURI) int64 { return p.PlaylistID },
		func(f models.PlaylistChangeTitleForm) string { return f.Title },
		errs.ErrPlaylistNotFound,
		errs.ErrPlaylistAlreadyExists,
		h.playlistSvc.ChangeTitle,
	)
}

func (h *ContentHandler) changePlaylistChangeableID(c *gin.Context) {
	handleChangeChangeableID(
		c,
		func(p models.PlaylistChangeChangeableIDUri) int64 { return p.PlaylistID },
		func(f models.PlaylistChangeChangeableIDForm) string { return f.ChangeableID },
		errs.ErrPlaylistNotFound,
		h.playlistSvc.ChangeChangeableID,
	)
}

func (h *ContentHandler) changePlaylistImage(c *gin.Context) {
	handleChangeImage(
		c,
		func(p models.PlaylistChangeImageURI) int64 { return p.PlaylistID },
		func(f models.PlaylistChangeImageForm) *multipart.FileHeader { return f.ImageFile },
		errs.ErrPlaylistNotFound,
		nil,
		h.playlistSvc.ChangeImage,
	)
}

func (h *ContentHandler) deletePlaylist(c *gin.Context) {
	handleDeleteByID(
		c,
		func(p models.PlaylistDeleteURI) int64 { return p.PlaylistID },
		errs.ErrPlaylistNotFound,
		h.playlistSvc.Delete,
	)
}

func (h *ContentHandler) savePlaylist(c *gin.Context) {
	handlePlaylistIDAction(
		c,
		errs.ErrPlaylistNotFound,
		[]error{errs.ErrPlaylistIsYours, errs.ErrPlaylistAlreadySaved},
		h.playlistSvc.SavePlaylist,
	)
}

func (h *ContentHandler) removePlaylistFromSaved(c *gin.Context) {
	handlePlaylistIDAction(
		c,
		errs.ErrPlaylistNotFound,
		[]error{errs.ErrPlaylistIsNotSaved},
		h.playlistSvc.RemoveFromSaved,
	)
}
