package user

import (
	"github.com/ocenb/music-go/user-service/internal/utils"
)

var (
	ErrUserNotFound             = utils.NotFoundError("user not found")
	ErrUserEmailExists          = utils.AlreadyExistsError("user with the same email already exists")
	ErrUserUsernameExists       = utils.AlreadyExistsError("user with the same username already exists")
	ErrInvalidVerificationToken = utils.InvalidArgumentError("invalid verification token")
)
