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

func (h *ContentHandler) registerTrackRoutes(router *gin.RouterGroup) {
	trackRouter := router.Group("/track")
	trackRouter.GET("/oneById/:trackId", h.getOneByID)
	trackRouter.GET("/one", h.getOne)
	trackRouter.GET("", h.getMany)
	trackRouter.GET("/popular", h.getManyPopular)
	trackRouter.POST("", h.upload)
	trackRouter.PATCH("/:trackId/add-play", h.addPlay)
	trackRouter.PATCH("/:trackId/title", h.changeTrackTitle)
	trackRouter.PATCH("/:trackId/changeable-id", h.changeTrackChangeableID)
	trackRouter.PATCH("/:trackId/image", h.changeTrackImage)
	trackRouter.DELETE("/:trackId", h.deleteTrack)
	trackRouter.GET("/liked", h.getManyLiked)
	trackRouter.POST("/:trackId/like", h.addToLiked)
	trackRouter.DELETE("/:trackId/like", h.removeFromLiked)
}

func (h *ContentHandler) getOneByID(c *gin.Context) {
	var params models.GetByTrackIDURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	trackModel, err := h.trackSvc.GetOneByID(c.Request.Context(), user.Id, params.TrackID)
	if err != nil {
		if errors.Is(err, errs.ErrTrackNotFound) {
			handleError(c, errs.NotFound(err.Error()))
			return
		}
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, trackModel)
}

func (h *ContentHandler) getOne(c *gin.Context) {
	respondGetOneByChangeableID(c, errs.ErrTrackNotFound, h.trackSvc.GetOne)
}

func (h *ContentHandler) getMany(c *gin.Context) {
	var params models.TrackGetManyForm
	if err := c.ShouldBindQuery(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	tracks, err := h.trackSvc.GetMany(c.Request.Context(), user.Id, params.UserID, params.Take, params.LastID)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, tracks)
}

func (h *ContentHandler) getManyPopular(c *gin.Context) {
	var params models.TrackGetManyForm
	if err := c.ShouldBindQuery(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	tracks, err := h.trackSvc.GetManyPopular(c.Request.Context(), user.Id, params.UserID, params.Take, params.LastID)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, tracks)
}

func (h *ContentHandler) upload(c *gin.Context) {
	var request models.UploadTrackForm
	if err := c.ShouldBind(&request); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	trackModel, err := h.trackSvc.Upload(
		c.Request.Context(),
		user.Id,
		user.Username,
		user.Email,
		request.Title,
		request.ChangeableID,
		request.AudioFile,
		request.ImageFile,
	)
	if err != nil {
		for _, badRequestError := range errs.TrackUploadBadRequestErrors {
			if errors.Is(err, badRequestError) {
				handleError(c, errs.InvalidArgument(err.Error()))
				return
			}
		}
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusCreated, trackModel)
}

func (h *ContentHandler) addPlay(c *gin.Context) {
	var params models.TrackAddPlayURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	if err := h.trackSvc.AddPlay(c.Request.Context(), params.TrackID); err != nil {
		if errors.Is(err, errs.ErrTrackNotFound) {
			handleError(c, errs.NotFound(err.Error()))
			return
		}
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) changeTrackTitle(c *gin.Context) {
	handleChangeTitle(
		c,
		func(p models.TrackChangeTitleURI) int64 { return p.TrackID },
		func(f models.TrackChangeTitleForm) string { return f.Title },
		errs.ErrTrackNotFound,
		errs.ErrTrackAlreadyExists,
		h.trackSvc.ChangeTitle,
	)
}

func (h *ContentHandler) changeTrackChangeableID(c *gin.Context) {
	handleChangeChangeableID(
		c,
		func(p models.TrackChangeChangeableIDUri) int64 { return p.TrackID },
		func(f models.TrackChangeChangeableIDForm) string { return f.ChangeableID },
		errs.ErrTrackNotFound,
		h.trackSvc.ChangeChangeableID,
	)
}

func (h *ContentHandler) changeTrackImage(c *gin.Context) {
	handleChangeImage(
		c,
		func(p models.TrackChangeImageURI) int64 { return p.TrackID },
		func(f models.TrackChangeImageForm) *multipart.FileHeader { return f.ImageFile },
		errs.ErrTrackNotFound,
		[]error{errs.ErrImageFileTooLarge},
		h.trackSvc.ChangeImage,
	)
}

func (h *ContentHandler) deleteTrack(c *gin.Context) {
	handleDeleteByID(
		c,
		func(p models.TrackDeleteURI) int64 { return p.TrackID },
		errs.ErrTrackNotFound,
		h.trackSvc.Delete,
	)
}

func (h *ContentHandler) getManyLiked(c *gin.Context) {
	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	likedTracks, err := h.trackSvc.GetManyLiked(c.Request.Context(), user.Id)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, likedTracks)
}

func (h *ContentHandler) addToLiked(c *gin.Context) {
	var params models.GetByTrackIDURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	if err := h.trackSvc.AddToLiked(c.Request.Context(), user.Id, params.TrackID); err != nil {
		if errors.Is(err, errs.ErrTrackNotFound) {
			handleError(c, errs.NotFound(err.Error()))
			return
		}
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) removeFromLiked(c *gin.Context) {
	var params models.GetByTrackIDURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	if err := h.trackSvc.RemoveFromLiked(c.Request.Context(), user.Id, params.TrackID); err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.Status(http.StatusNoContent)
}
