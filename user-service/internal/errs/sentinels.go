package errs

var (
	ErrInvalidAccessToken  = Unauthenticated("invalid access token")
	ErrInvalidRefreshToken = Unauthenticated("invalid refresh token")
	ErrInvalidToken        = Unauthenticated("invalid token")
	ErrTokenNotFound       = NotFound("token not found")
	ErrTokenExpired        = InvalidArgument("token expired")

	ErrUserEmailExists     = AlreadyExists("user with the same email already exists")
	ErrUserUsernameExists  = AlreadyExists("user with the same username already exists")
	ErrUserNotFound        = NotFound("user not found")
	ErrUserEmailNotFound   = NotFound("user with this email does not exist")
	ErrUserNotVerified     = PermissionDenied("user is not verified")
	ErrUserAlreadyVerified = AlreadyExists("user is already verified")

	ErrInvalidPassword = InvalidArgument("wrong password")

	ErrInvalidTokenID = InvalidArgument("invalid token id in token")
	ErrInvalidUserID  = InvalidArgument("invalid user id in token")

	ErrInvalidVerificationToken = InvalidArgument("invalid verification token")
	ErrUserAlreadyFollowed      = AlreadyExists("user already followed")
	ErrUserNotFollowed          = NotFound("user not followed")

	ErrInvalidSigningMethod = Unauthenticated("invalid signing method")
)
