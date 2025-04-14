package playlist

import "mime/multipart"

type GetByPlaylistIDUri struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
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

type GetManyWithSavedForm struct {
	Take   int   `form:"take" binding:"omitempty,min=1"`
	LastID int64 `form:"lastId" binding:"omitempty,min=1"`
}

type CreatePlaylistForm struct {
	Title        string                `form:"title" binding:"required,min=1,max=20"`
	ChangeableID string                `form:"changeableId" binding:"required,min=1,max=20"`
	ImageFile    *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type ChangeTitleUri struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}

type ChangeTitleForm struct {
	Title string `form:"title" binding:"required,min=1,max=20"`
}

type ChangeChangeableIdUri struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}

type ChangeChangeableIdForm struct {
	ChangeableID string `form:"changeableId" binding:"required,min=1,max=20"`
}

type ChangeImageUri struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}

type ChangeImageForm struct {
	ImageFile *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type DeleteUri struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}
