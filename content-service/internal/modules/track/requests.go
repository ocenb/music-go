package track

import "mime/multipart"

type GetByTrackIDRequest struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type GetOneRequest struct {
	Username     string `form:"username" binding:"required"`
	ChangeableID string `form:"changeableId" binding:"required"`
}

type GetManyRequest struct {
	UserID int64 `form:"userId" binding:"required"`
	Take   int   `form:"take" binding:"omitempty,min=1"`
	LastID int64 `form:"lastId" binding:"omitempty,min=1"`
}

type UploadTrackRequest struct {
	Title        string                `form:"title" binding:"required,min=1,max=20"`
	ChangeableID string                `form:"changeableId" binding:"required,min=1,max=20"`
	AudioFile    *multipart.FileHeader `form:"audioFile" binding:"required"`
	ImageFile    *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type AddPlayRequest struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type ChangeTitleRequest struct {
	TrackID int64  `uri:"trackId" binding:"required"`
	Title   string `form:"title" binding:"required,min=1,max=20"`
}

type ChangeChangeableIdRequest struct {
	TrackID      int64  `uri:"trackId" binding:"required"`
	ChangeableID string `form:"changeableId" binding:"required,min=1,max=20"`
}

type ChangeImageRequest struct {
	TrackID   int64                 `uri:"trackId" binding:"required"`
	ImageFile *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type DeleteRequest struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}
