package playlist

import "mime/multipart"

type GetByPlaylistIDRequest struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
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

type CreatePlaylistRequest struct {
	Title        string                `form:"title" binding:"required,min=1,max=20"`
	ChangeableID string                `form:"changeableId" binding:"required,min=1,max=20"`
	ImageFile    *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type ChangeTitleRequest struct {
	PlaylistID int64  `uri:"playlistId" binding:"required"`
	Title      string `form:"title" binding:"required,min=1,max=20"`
}

type ChangeChangeableIdRequest struct {
	PlaylistID   int64  `uri:"playlistId" binding:"required"`
	ChangeableID string `form:"changeableId" binding:"required,min=1,max=20"`
}

type ChangeImageRequest struct {
	PlaylistID int64                 `uri:"playlistId" binding:"required"`
	ImageFile  *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type DeleteRequest struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}
