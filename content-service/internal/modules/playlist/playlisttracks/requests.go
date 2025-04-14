package playlisttracks

type PlaylistTrackIDsUri struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
	TrackID    int64 `uri:"trackId" binding:"required"`
}

type PlaylistUri struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}

type GetManyForm struct {
	Take int `form:"take" binding:"omitempty,min=1"`
}

type AddTrackJSON struct {
	Position int `json:"position" binding:"omitempty,min=1"`
}

type UpdatePositionJSON struct {
	Position int `json:"position" binding:"required,min=1"`
}
