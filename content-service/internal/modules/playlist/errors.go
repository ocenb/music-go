package playlist

import (
	"errors"

	"github.com/ocenb/music-go/content-service/internal/modules/file"
)

var (
	ErrPlaylistNotFound      = errors.New("playlist not found")
	ErrPlaylistAlreadyExists = errors.New("playlist with this title already exists")
	ErrChangeableIDExists    = errors.New("playlist with this changeableID already exists")
	ErrPermissionDenied      = errors.New("permission denied")
	ErrPlaylistIsYours       = errors.New("playlist is yours")
	ErrPlaylistAlreadySaved  = errors.New("playlist is already saved")
	ErrPlaylistIsNotSaved    = errors.New("playlist is not saved")
)

var BadRequestErrors = []error{
	ErrPlaylistAlreadyExists,
	ErrChangeableIDExists,
	file.ErrInvalidImageFormat,
	ErrPlaylistIsYours,
	ErrPlaylistAlreadySaved,
}
