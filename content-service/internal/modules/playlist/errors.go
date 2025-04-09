package playlist

import "errors"

var (
	ErrPlaylistNotFound      = errors.New("playlist not found")
	ErrPlaylistAlreadyExists = errors.New("playlist with this title already exists")
	ErrChangeableIDExists    = errors.New("playlist with this changeableID already exists")
	ErrPermissionDenied      = errors.New("permission denied")
	ErrInvalidImageFormat    = errors.New("invalid image format")
	ErrPlaylistIsYours       = errors.New("playlist is yours")
	ErrPlaylistAlreadySaved  = errors.New("playlist is already saved")
	ErrPlaylistIsNotSaved    = errors.New("playlist is not saved")
)

var BadRequestErrors = []error{
	ErrPlaylistAlreadyExists,
	ErrChangeableIDExists,
	ErrInvalidImageFormat,
	ErrPlaylistIsYours,
	ErrPlaylistAlreadySaved,
}
