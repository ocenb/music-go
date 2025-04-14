package playlist

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ocenb/music-go/content-service/internal/modules/file"
	"github.com/ocenb/music-go/content-service/internal/utils"
)

type PlaylistHandlerInterface interface {
	getOne(c *gin.Context)
	getMany(c *gin.Context)
	getManyWithSaved(c *gin.Context)
	create(c *gin.Context)
	changeTitle(c *gin.Context)
	changeChangeableId(c *gin.Context)
	changeImage(c *gin.Context)
	delete(c *gin.Context)
	savePlaylist(c *gin.Context)
	removeFromSaved(c *gin.Context)
	RegisterHandlers(router *gin.RouterGroup)
}

type PlaylistHandler struct {
	playlistService PlaylistServiceInterface
}

func NewPlaylistHandler(playlistService PlaylistServiceInterface) PlaylistHandlerInterface {
	return &PlaylistHandler{
		playlistService: playlistService,
	}
}

func (h *PlaylistHandler) getOne(c *gin.Context) {
	var params GetOneForm
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	playlist, err := h.playlistService.GetOne(c.Request.Context(), user.Id, params.Username, params.ChangeableID)
	if err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			utils.NotFoundError(c, err)
			return
		}
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, playlist)
}

func (h *PlaylistHandler) getMany(c *gin.Context) {
	var params GetManyForm
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	playlists, err := h.playlistService.GetMany(c.Request.Context(), params.UserID, user.Id, params.Take, params.LastID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, playlists)
}

func (h *PlaylistHandler) getManyWithSaved(c *gin.Context) {
	var params GetManyWithSavedForm
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	playlists, err := h.playlistService.GetManyWithSaved(c.Request.Context(), user.Id, params.Take, params.LastID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, playlists)
}

func (h *PlaylistHandler) create(c *gin.Context) {
	var request CreatePlaylistForm
	if err := c.ShouldBind(&request); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	playlist, err := h.playlistService.Create(
		c.Request.Context(),
		user.Id,
		user.Username,
		request.Title,
		request.ChangeableID,
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

	c.JSON(http.StatusCreated, playlist)
}

func (h *PlaylistHandler) changeTitle(c *gin.Context) {
	var params ChangeTitleUri
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	var request ChangeTitleForm
	if err := c.ShouldBind(&request); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	err = h.playlistService.ChangeTitle(
		c.Request.Context(),
		user.Id,
		params.PlaylistID,
		request.Title,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrPlaylistNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPermissionDenied):
			utils.PermissionDeniedError(c, err)
		case errors.Is(err, ErrPlaylistAlreadyExists):
			utils.BadRequestError(c, err)
		default:
			utils.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PlaylistHandler) changeChangeableId(c *gin.Context) {
	var params ChangeChangeableIdUri
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	var request ChangeChangeableIdForm
	if err := c.ShouldBind(&request); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	err = h.playlistService.ChangeChangeableId(
		c.Request.Context(),
		user.Id,
		params.PlaylistID,
		request.ChangeableID,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrPlaylistNotFound):
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

func (h *PlaylistHandler) changeImage(c *gin.Context) {
	var params ChangeImageUri
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	var request ChangeImageForm
	if err := c.ShouldBind(&request); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	err = h.playlistService.ChangeImage(
		c.Request.Context(),
		user.Id,
		params.PlaylistID,
		request.ImageFile,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrPlaylistNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPermissionDenied):
			utils.PermissionDeniedError(c, err)
		case errors.Is(err, file.ErrInvalidImageFormat):
			utils.BadRequestError(c, err)
		default:
			utils.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PlaylistHandler) delete(c *gin.Context) {
	var params DeleteUri
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	if err := h.playlistService.Delete(c.Request.Context(), user.Id, params.PlaylistID); err != nil {
		switch {
		case errors.Is(err, ErrPlaylistNotFound):
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

func (h *PlaylistHandler) savePlaylist(c *gin.Context) {
	var params GetByPlaylistIDUri
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	if err := h.playlistService.SavePlaylist(c.Request.Context(), user.Id, params.PlaylistID); err != nil {
		switch {
		case errors.Is(err, ErrPlaylistNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPlaylistIsYours), errors.Is(err, ErrPlaylistAlreadySaved):
			utils.BadRequestError(c, err)
		default:
			utils.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PlaylistHandler) removeFromSaved(c *gin.Context) {
	var params GetByPlaylistIDUri
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	if err := h.playlistService.RemoveFromSaved(c.Request.Context(), user.Id, params.PlaylistID); err != nil {
		switch {
		case errors.Is(err, ErrPlaylistNotFound):
			utils.NotFoundError(c, err)
		case errors.Is(err, ErrPlaylistIsNotSaved):
			utils.BadRequestError(c, err)
		default:
			utils.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PlaylistHandler) RegisterHandlers(router *gin.RouterGroup) {
	playlistRouter := router.Group("/playlist")
	playlistRouter.GET("/one", h.getOne)
	playlistRouter.GET("", h.getMany)
	playlistRouter.GET("/with-saved", h.getManyWithSaved)
	playlistRouter.POST("", h.create)
	playlistRouter.PATCH("/:playlistId/title", h.changeTitle)
	playlistRouter.PATCH("/:playlistId/changeable-id", h.changeChangeableId)
	playlistRouter.PATCH("/:playlistId/image", h.changeImage)
	playlistRouter.DELETE("/:playlistId", h.delete)
	playlistRouter.POST("/:playlistId/save", h.savePlaylist)
	playlistRouter.DELETE("/:playlistId/save", h.removeFromSaved)
}
