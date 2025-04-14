package track

import "mime/multipart"

type GetByTrackIDUri struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type GetOneForm struct {
	Username     string `form:"username" binding:"required"`
	ChangeableID string `form:"changeableId" binding:"required"`
}

type GetManyForm struct {
	UserID int64 `form:"userId" binding:"required"`
	Take   int   `form:"take" binding:"omitempty,min=1"`
	LastID int64 `form:"lastId" binding:"omitempty,min=1"`
}

type UploadTrackForm struct {
	Title        string                `form:"title" binding:"required,min=1,max=20"`
	ChangeableID string                `form:"changeableId" binding:"required,min=1,max=20"`
	AudioFile    *multipart.FileHeader `form:"audioFile" binding:"required"`
	ImageFile    *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type AddPlayUri struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type ChangeTitleUri struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type ChangeTitleForm struct {
	Title string `form:"title" binding:"required,min=1,max=20"`
}

type ChangeChangeableIdUri struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type ChangeChangeableIdForm struct {
	ChangeableID string `form:"changeableId" binding:"required,min=1,max=20"`
}

type ChangeImageUri struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type ChangeImageForm struct {
	ImageFile *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type DeleteUri struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}
