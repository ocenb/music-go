package file

import "errors"

var (
	ErrAudioFileTooLarge  = errors.New("audio file too large")
	ErrImageFileTooLarge  = errors.New("image file too large")
	ErrInvalidImageFormat = errors.New("invalid image format")
	ErrInvalidAudioFormat = errors.New("invalid audio format")
)
