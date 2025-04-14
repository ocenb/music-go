package all

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ocenb/music-go/content-service/internal/utils"
)

type AllHandlerInterface interface {
	deleteAll(c *gin.Context)
	RegisterHandlers(router *gin.RouterGroup)
}

type AllHandler struct {
	allService AllServiceInterface
}

func NewAllHandler(allService AllServiceInterface) AllHandlerInterface {
	return &AllHandler{
		allService: allService,
	}
}

func (h *AllHandler) deleteAll(c *gin.Context) {
	user, err := utils.GetInfoFromContext(c)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	err = h.allService.DeleteAll(c.Request.Context(), user.Id)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AllHandler) RegisterHandlers(router *gin.RouterGroup) {
	trackRouter := router.Group("/all")
	trackRouter.DELETE("", h.deleteAll)
}
