package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/middlewares"
	"github.com/ocenb/music-go/content-service/internal/models"
)

func (h *ContentHandler) registerHistoryRoutes(router *gin.RouterGroup) {
	historyRouter := router.Group("/history")
	historyRouter.GET("", h.getHistory)
	historyRouter.POST("/:trackId", h.addHistory)
	historyRouter.DELETE("", h.clearHistory)
}

func (h *ContentHandler) getHistory(c *gin.Context) {
	var params models.HistoryGetForm
	if err := c.ShouldBindQuery(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	history, err := h.historySvc.Get(c.Request.Context(), user.Id, params.Take)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *ContentHandler) addHistory(c *gin.Context) {
	var params models.HistoryAddURI
	if err := c.ShouldBindUri(&params); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return
	}

	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	err = h.historySvc.Add(c.Request.Context(), user.Id, params.TrackID)
	if err != nil {
		if errors.Is(err, errs.ErrTrackNotFound) {
			handleError(c, errs.NotFound(err.Error()))
			return
		}
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ContentHandler) clearHistory(c *gin.Context) {
	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return
	}

	err = h.historySvc.Clear(c.Request.Context(), user.Id)
	if err != nil {
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
