package history

type GetUri struct {
	Take int64 `form:"take" binding:"omitempty,min=1"`
}

type AddUri struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}
