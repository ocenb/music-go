package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/middlewares"
)

func (h *ContentHandler) registerAllRoutes(router *gin.RouterGroup) {
	router.DELETE("/all", h.deleteAll)
}

func (h *ContentHandler) registerInternalRoutes(router *gin.RouterGroup) {
	router.DELETE("/users/:userID", h.deleteUserContent)
}

func (h *ContentHandler) deleteAll(c *gin.Context) {
	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	err = h.allSvc.DeleteAll(c.Request.Context(), user.Id)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) deleteUserContent(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		handleError(c, errs.InvalidArgument("invalid user id"))
		return
	}

	err = h.allSvc.DeleteAll(c.Request.Context(), userID)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.Status(http.StatusNoContent)
}
