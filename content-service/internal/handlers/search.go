package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/models"
)

func (h *ContentHandler) registerSearchRoutes(router *gin.RouterGroup) {
	searchRouter := router.Group("/search")
	searchRouter.GET("/users", h.searchUsers)
	searchRouter.GET("/tracks", h.searchTracks)
}

func (h *ContentHandler) searchUsers(c *gin.Context) {
	var params models.SearchForm
	if err := c.ShouldBindQuery(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	ids, err := h.searchSvc.SearchUsers(c.Request.Context(), params.Query)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, ids)
}

func (h *ContentHandler) searchTracks(c *gin.Context) {
	var params models.SearchForm
	if err := c.ShouldBindQuery(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	ids, err := h.searchSvc.SearchTracks(c.Request.Context(), params.Query)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, ids)
}
