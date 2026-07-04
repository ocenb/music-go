package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ocenb/music-protos/gen/userservice"
)

type UserClient interface {
	CheckAuth(ctx context.Context, authorizationHeader string) (*userservice.CheckAuthResponse, error)
}

func Auth(userClient UserClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		res, err := userClient.CheckAuth(c.Request.Context(), authorization)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		ctx := ContextWithUser(c.Request.Context(), res.User)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
