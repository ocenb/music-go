package playlisttracks

import "errors"

var (
	ErrPlaylistNotFound       = errors.New("playlist not found")
	ErrTrackNotFound          = errors.New("track not found")
	ErrPermissionDenied       = errors.New("permission denied")
	ErrTrackAlreadyInPlaylist = errors.New("track already in playlist")
	ErrTrackNotInPlaylist     = errors.New("track is not in this playlist")
	ErrPositionConflict       = errors.New("track already in this position")
)

var BadRequestErrors = []error{
	ErrTrackAlreadyInPlaylist,
	ErrPositionConflict,
}
