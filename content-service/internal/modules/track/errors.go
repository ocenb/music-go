package track

import "errors"

var (
	ErrTrackNotFound      = errors.New("track not found")
	ErrTrackAlreadyExists = errors.New("track with this title already exists")
	ErrChangeableIDExists = errors.New("track with this changeableId already exists")
	ErrPermissionDenied   = errors.New("you don't have permission for this action")
	ErrInvalidImageFormat = errors.New("invalid image format")
	ErrInvalidAudioFormat = errors.New("invalid audio format")
)

var BadRequestErrors = []error{
	ErrTrackNotFound,
	ErrTrackAlreadyExists,
	ErrChangeableIDExists,
	ErrInvalidImageFormat,
	ErrInvalidAudioFormat,
	ErrPermissionDenied,
}
