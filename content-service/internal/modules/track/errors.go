package track

import (
	"errors"

	"github.com/ocenb/music-go/content-service/internal/modules/file"
)

var (
	ErrTrackNotFound      = errors.New("track not found")
	ErrTrackAlreadyExists = errors.New("track with this title already exists")
	ErrChangeableIDExists = errors.New("track with this changeableId already exists")
	ErrPermissionDenied   = errors.New("you don't have permission for this action")
)

var BadRequestErrors = []error{
	ErrTrackNotFound,
	ErrTrackAlreadyExists,
	ErrChangeableIDExists,
	ErrPermissionDenied,
	file.ErrInvalidImageFormat,
	file.ErrInvalidAudioFormat,
	file.ErrAudioFileTooLarge,
	file.ErrImageFileTooLarge,
}
