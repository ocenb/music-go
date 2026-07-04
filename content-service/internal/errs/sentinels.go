package errs

var (
	ErrTrackNotFound          = NotFound("track not found")
	ErrTrackAlreadyExists     = AlreadyExists("track with this title already exists")
	ErrPlaylistNotFound       = NotFound("playlist not found")
	ErrPlaylistAlreadyExists  = AlreadyExists("playlist with this title already exists")
	ErrTrackNotInPlaylist     = NotFound("track is not in this playlist")
	ErrTrackAlreadyLiked      = AlreadyExists("track already liked")
	ErrTrackNotLiked          = NotFound("track not liked")
	ErrPlaylistAlreadySaved   = AlreadyExists("playlist is already saved")
	ErrPlaylistNotSaved       = NotFound("playlist not saved")
	ErrPlaylistIsYours        = InvalidArgument("playlist is yours")
	ErrPlaylistIsNotSaved     = InvalidArgument("playlist is not saved")
	ErrTrackAlreadyInPlaylist = AlreadyExists("track already in playlist")
	ErrPositionConflict       = InvalidArgument("track already in this position")
	ErrPermissionDenied       = PermissionDenied("permission denied")
	ErrTitleExists            = AlreadyExists("title already exists")
	ErrChangeableIDExists     = AlreadyExists("changeable id already exists")
	ErrAudioFileTooLarge      = InvalidArgument("audio file too large")
	ErrImageFileTooLarge      = InvalidArgument("image file too large")
	ErrInvalidImageFormat     = InvalidArgument("invalid image format")
	ErrInvalidAudioFormat     = InvalidArgument("invalid audio format")
	ErrUnauthorized           = Unauthenticated("unauthorized")
)

var TrackUploadBadRequestErrors = []error{
	ErrTrackNotFound,
	ErrTrackAlreadyExists,
	ErrChangeableIDExists,
	ErrPermissionDenied,
	ErrInvalidImageFormat,
	ErrInvalidAudioFormat,
	ErrAudioFileTooLarge,
	ErrImageFileTooLarge,
}

var PlaylistCreateBadRequestErrors = []error{
	ErrPlaylistAlreadyExists,
	ErrChangeableIDExists,
	ErrInvalidImageFormat,
	ErrPlaylistIsYours,
	ErrPlaylistAlreadySaved,
}

var PlaylistTracksBadRequestErrors = []error{
	ErrTrackAlreadyInPlaylist,
	ErrPositionConflict,
}
