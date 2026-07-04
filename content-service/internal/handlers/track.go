package handlers

import (
	"errors"
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
	var params models.TrackGetOneForm
	if err := c.ShouldBindQuery(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	trackModel, err := h.trackSvc.GetOne(c.Request.Context(), user.Id, params.Username, params.ChangeableID)
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
	var params models.TrackChangeTitleURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	var request models.TrackChangeTitleForm
	if err := c.ShouldBind(&request); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	err = h.trackSvc.ChangeTitle(c.Request.Context(), user.Id, params.TrackID, request.Title)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrTrackNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		case errors.Is(err, errs.ErrTrackAlreadyExists):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) changeTrackChangeableID(c *gin.Context) {
	var params models.TrackChangeChangeableIDUri
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	var request models.TrackChangeChangeableIDForm
	if err := c.ShouldBind(&request); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	err = h.trackSvc.ChangeChangeableID(c.Request.Context(), user.Id, params.TrackID, request.ChangeableID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrTrackNotFound):
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

func (h *ContentHandler) changeTrackImage(c *gin.Context) {
	var params models.TrackChangeImageURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	var request models.TrackChangeImageForm
	if err := c.ShouldBind(&request); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	err = h.trackSvc.ChangeImage(c.Request.Context(), user.Id, params.TrackID, request.ImageFile)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrTrackNotFound):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		case errors.Is(err, errs.ErrInvalidImageFormat):
			handleError(c, errs.InvalidArgument(err.Error()))
		case errors.Is(err, errs.ErrImageFileTooLarge):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) deleteTrack(c *gin.Context) {
	var params models.TrackDeleteURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	if err := h.trackSvc.Delete(c.Request.Context(), user.Id, params.TrackID); err != nil {
		switch {
		case errors.Is(err, errs.ErrTrackNotFound):
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
