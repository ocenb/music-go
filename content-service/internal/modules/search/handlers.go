package search

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ocenb/music-go/content-service/internal/clients/searchclient"
	"github.com/ocenb/music-go/content-service/internal/utils"
	"github.com/ocenb/music-protos/gen/searchservice"
)

type SearchHandlerInterface interface {
	searchUsers(c *gin.Context)
	searchTracks(c *gin.Context)
	RegisterHandlers(router *gin.RouterGroup)
}

type SearchHandler struct {
	searchClient *searchclient.SearchServiceClient
}

func NewSearchHandler(searchClient *searchclient.SearchServiceClient) SearchHandlerInterface {
	return &SearchHandler{
		searchClient: searchClient,
	}
}

func (h *SearchHandler) searchUsers(c *gin.Context) {
	var params SearchForm
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	response, err := h.searchClient.Client.SearchUsers(c.Request.Context(), &searchservice.SearchRequest{
		Query: params.Query,
	})
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Ids)
}

func (h *SearchHandler) searchTracks(c *gin.Context) {
	var params SearchForm
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestError(c, err)
		return
	}

	response, err := h.searchClient.Client.SearchTracks(c.Request.Context(), &searchservice.SearchRequest{
		Query: params.Query,
	})
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Ids)
}

func (h *SearchHandler) RegisterHandlers(router *gin.RouterGroup) {
	searchRouter := router.Group("/search")
	searchRouter.GET("/users", h.searchUsers)
	searchRouter.GET("/tracks", h.searchTracks)
}
