package all

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ocenb/music-go/content-service/internal/utils"
)

type AllHandlerInterface interface {
	deleteAll(c *gin.Context)
	deleteUserContent(c *gin.Context)
	RegisterHandlers(router *gin.RouterGroup)
	RegisterInternalHandlers(router *gin.RouterGroup)
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

func (h *AllHandler) deleteUserContent(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		utils.BadRequestError(c, errors.New("invalid user id"))
		return
	}

	err = h.allService.DeleteAll(c.Request.Context(), userID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AllHandler) RegisterHandlers(router *gin.RouterGroup) {
	router.DELETE("/all", h.deleteAll)
}

func (h *AllHandler) RegisterInternalHandlers(router *gin.RouterGroup) {
	router.DELETE("/users/:userID", h.deleteUserContent)
}
