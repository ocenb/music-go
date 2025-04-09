package track

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ocenb/music-go/content-service/internal/utils"
)

type TrackHandlerInterface interface {
	getOneByID(c *gin.Context)
	getOne(c *gin.Context)
	getMany(c *gin.Context)
	getManyPopular(c *gin.Context)
	upload(c *gin.Context)
	addPlay(c *gin.Context)
	changeTitle(c *gin.Context)
	changeChangeableId(c *gin.Context)
	changeImage(c *gin.Context)
	delete(c *gin.Context)
	getManyLiked(c *gin.Context)
	addToLiked(c *gin.Context)
	removeFromLiked(c *gin.Context)
	RegisterHandlers(router *gin.RouterGroup)
}

type TrackHandler struct {
	trackService TrackServiceInterface
}

func NewTrackHandler(trackService TrackServiceInterface) TrackHandlerInterface {
	return &TrackHandler{
		trackService: trackService,
	}
}

func (h *TrackHandler) getOneByID(c *gin.Context) {
	var params GetByTrackIDRequest
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	track, err := h.trackService.GetOneById(c.Request.Context(), user.Id, params.TrackID)
	if err != nil {
		if errors.Is(err, ErrTrackNotFound) {
			utils.NotFoundError(c, err)
			return
		}
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, track)
}

func (h *TrackHandler) getOne(c *gin.Context) {
	var params GetOneRequest
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	track, err := h.trackService.GetOne(c.Request.Context(), user.Id, params.Username, params.ChangeableID)
	if err != nil {
		if errors.Is(err, ErrTrackNotFound) {
			utils.NotFoundError(c, err)
			return
		}
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, track)
}

func (h *TrackHandler) getMany(c *gin.Context) {
	var params GetManyRequest
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	tracks, err := h.trackService.GetMany(c.Request.Context(), user.Id, params.UserID, params.Take, params.LastID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, tracks)
}

func (h *TrackHandler) getManyPopular(c *gin.Context) {
	var params GetManyRequest
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	tracks, err := h.trackService.GetManyPopular(c.Request.Context(), user.Id, params.UserID, params.Take, params.LastID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, tracks)
}

func (h *TrackHandler) upload(c *gin.Context) {
	var request UploadTrackRequest
	if err := c.ShouldBind(&request); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	track, err := h.trackService.Upload(
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
		for _, badRequestError := range BadRequestErrors {
			if errors.Is(err, badRequestError) {
				utils.BadRequestError(c, err)
				return
			}
		}
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, track)
}

func (h *TrackHandler) addPlay(c *gin.Context) {
	var params AddPlayRequest
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	if err := h.trackService.AddPlay(c.Request.Context(), params.TrackID); err != nil {
		if errors.Is(err, ErrTrackNotFound) {
			utils.NotFoundError(c, err)
			return
		}
		utils.InternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TrackHandler) changeTitle(c *gin.Context) {
	var params ChangeTitleRequest
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	var request ChangeTitleRequest
	if err := c.ShouldBind(&request); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	err = h.trackService.ChangeTitle(
		c.Request.Context(),
		user.Id,
		params.TrackID,
		request.Title,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrTrackNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPermissionDenied):
			utils.PermissionDeniedError(c, err)
		case errors.Is(err, ErrTrackAlreadyExists):
			utils.BadRequestError(c, err)
		default:
			utils.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TrackHandler) changeChangeableId(c *gin.Context) {
	var params ChangeChangeableIdRequest
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	var request ChangeChangeableIdRequest
	if err := c.ShouldBind(&request); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	err = h.trackService.ChangeChangeableId(
		c.Request.Context(),
		user.Id,
		params.TrackID,
		request.ChangeableID,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrTrackNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPermissionDenied):
			utils.PermissionDeniedError(c, err)
		case errors.Is(err, ErrChangeableIDExists):
			utils.BadRequestError(c, err)
		default:
			utils.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TrackHandler) changeImage(c *gin.Context) {
	var params ChangeImageRequest
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	var request ChangeImageRequest
	if err := c.ShouldBind(&request); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	err = h.trackService.ChangeImage(
		c.Request.Context(),
		user.Id,
		params.TrackID,
		request.ImageFile,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrTrackNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPermissionDenied):
			utils.PermissionDeniedError(c, err)
		case errors.Is(err, ErrInvalidImageFormat):
			utils.BadRequestError(c, err)
		default:
			utils.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TrackHandler) delete(c *gin.Context) {
	var params DeleteRequest
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	if err := h.trackService.Delete(c.Request.Context(), user.Id, params.TrackID); err != nil {
		switch {
		case errors.Is(err, ErrTrackNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPermissionDenied):
			utils.PermissionDeniedError(c, err)
		default:
			utils.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TrackHandler) getManyLiked(c *gin.Context) {
	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	likedTracks, err := h.trackService.GetManyLiked(c.Request.Context(), user.Id)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, likedTracks)
}

func (h *TrackHandler) addToLiked(c *gin.Context) {
	var params GetByTrackIDRequest
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	if err := h.trackService.AddToLiked(c.Request.Context(), user.Id, params.TrackID); err != nil {
		if errors.Is(err, ErrTrackNotFound) {
			utils.NotFoundError(c, err)
			return
		}
		utils.InternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TrackHandler) removeFromLiked(c *gin.Context) {
	var params GetByTrackIDRequest
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	if err := h.trackService.RemoveFromLiked(c.Request.Context(), user.Id, params.TrackID); err != nil {
		utils.InternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TrackHandler) RegisterHandlers(router *gin.RouterGroup) {
	trackRouter := router.Group("/track")
	trackRouter.GET("/oneById/:trackId", h.getOneByID)
	trackRouter.GET("/one", h.getOne)
	trackRouter.GET("", h.getMany)
	trackRouter.GET("/popular", h.getManyPopular)
	trackRouter.POST("", h.upload)
	trackRouter.PATCH("/:trackId/add-play", h.addPlay)
	trackRouter.PATCH("/:trackId/title", h.changeTitle)
	trackRouter.PATCH("/:trackId/changeable-id", h.changeChangeableId)
	trackRouter.PATCH("/:trackId/image", h.changeImage)
	trackRouter.DELETE("/:trackId", h.delete)
	trackRouter.GET("/liked", h.getManyLiked)
	trackRouter.POST("/:trackId/like", h.addToLiked)
	trackRouter.DELETE("/:trackId/like", h.removeFromLiked)
}
