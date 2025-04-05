package token

import (
	"github.com/ocenb/music-go/user-service/internal/utils"
)

var (
	ErrInvalidToken         = utils.UnauthenticatedError("invalid token")
	ErrInvalidSigningMethod = utils.UnauthenticatedError("invalid signing method")
	ErrTokenExpired         = utils.UnauthenticatedError("token has expired")
	ErrTokenNotFound        = utils.NotFoundError("token not found")
)
