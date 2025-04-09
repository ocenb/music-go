package history

type GetRequest struct {
	Take int64 `uri:"take" binding:"omitempty,min=1"`
}

type AddRequest struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}
