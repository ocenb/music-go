package auth

import (
	"github.com/ocenb/music-go/user-service/internal/utils"
)

var (
	ErrInvalidAccessToken  = utils.UnauthenticatedError("invalid access token")
	ErrInvalidRefreshToken = utils.UnauthenticatedError("invalid refresh token")
	ErrInvalidToken        = utils.UnauthenticatedError("invalid token")
	ErrTokenNotFound       = utils.NotFoundError("token not found")
	ErrTokenExpired        = utils.InvalidArgumentError("token expired")

	ErrUserEmailExists     = utils.AlreadyExistsError("user with the same email already exists")
	ErrUserUsernameExists  = utils.AlreadyExistsError("user with the same username already exists")
	ErrUserNotFound        = utils.NotFoundError("user not found")
	ErrUserEmailNotFound   = utils.NotFoundError("user with this email does not exist")
	ErrUserNotVerified     = utils.PermissionDeniedError("user is not verified")
	ErrUserAlreadyVerified = utils.AlreadyExistsError("user is already verified")

	ErrInvalidPassword = utils.InvalidArgumentError("wrong password")

	ErrInvalidTokenID = utils.InvalidArgumentError("invalid token id in token")
	ErrInvalidUserID  = utils.InvalidArgumentError("invalid user id in token")
)
