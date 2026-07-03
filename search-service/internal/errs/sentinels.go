package errs

var (
	ErrUserAlreadyExists  = AlreadyExists("user already exists")
	ErrAlbumAlreadyExists = AlreadyExists("album already exists")
	ErrTrackAlreadyExists = AlreadyExists("track already exists")
	ErrUserNotFound       = NotFound("user not found")
	ErrAlbumNotFound      = NotFound("album not found")
	ErrTrackNotFound      = NotFound("track not found")
)
