package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	InternalError = func(c *gin.Context, err error) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	NotFoundError = func(c *gin.Context, err error) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}

	BadRequestError = func(c *gin.Context, err error) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	AlreadyExistsError = func(c *gin.Context, err error) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	}

	UnauthenticatedError = func(c *gin.Context, err error) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	}

	PermissionDeniedError = func(c *gin.Context, err error) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	}
)
