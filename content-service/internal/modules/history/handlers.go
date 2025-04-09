package history

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ocenb/music-go/content-service/internal/utils"
)

type HistoryHandlersInterface interface {
	get(c *gin.Context)
	add(c *gin.Context)
	clear(c *gin.Context)
	RegisterHandlers(router *gin.RouterGroup)
}

type HistoryHandlers struct {
	historyService HistoryServiceInterface
}

func NewHistoryHandler(historyService HistoryServiceInterface) HistoryHandlersInterface {
	return &HistoryHandlers{
		historyService: historyService,
	}
}

func (h *HistoryHandlers) get(c *gin.Context) {
	var params GetRequest
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	history, err := h.historyService.Get(c.Request.Context(), user.Id, params.Take)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *HistoryHandlers) add(c *gin.Context) {
	var params AddRequest
	if err := c.ShouldBindUri(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	err = h.historyService.Add(c.Request.Context(), user.Id, params.TrackID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *HistoryHandlers) clear(c *gin.Context) {
	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	err = h.historyService.Clear(c.Request.Context(), user.Id)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *HistoryHandlers) RegisterHandlers(router *gin.RouterGroup) {
	historyRouter := router.Group("/history")
	historyRouter.GET("", h.get)
	historyRouter.POST("", h.add)
	historyRouter.DELETE("", h.clear)
}
