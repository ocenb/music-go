package playlisttracks

type PlaylistTrackIDsRequest struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
	TrackID    int64 `uri:"trackId" binding:"required"`
}

type PlaylistRequest struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}

type GetManyRequest struct {
	Take int `form:"take" binding:"omitempty,min=1"`
}

type AddTrackRequest struct {
	Position int `json:"position" binding:"omitempty,min=1"`
}

type UpdatePositionRequest struct {
	Position int `json:"position" binding:"required,min=1"`
}
