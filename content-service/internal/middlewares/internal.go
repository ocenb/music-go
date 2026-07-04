package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InternalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Internal-Secret") != secret {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}
