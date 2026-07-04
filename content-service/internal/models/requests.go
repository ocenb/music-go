package models

import "mime/multipart"

type GetByTrackIDURI struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type TrackGetOneForm struct {
	Username     string `form:"username" binding:"required"`
	ChangeableID string `form:"changeableId" binding:"required"`
}

type TrackGetManyForm struct {
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

type TrackAddPlayURI struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type TrackChangeTitleURI struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type TrackChangeTitleForm struct {
	Title string `form:"title" binding:"required,min=1,max=20"`
}

type TrackChangeChangeableIDUri struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type TrackChangeChangeableIDForm struct {
	ChangeableID string `form:"changeableId" binding:"required,min=1,max=20"`
}

type TrackChangeImageURI struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type TrackChangeImageForm struct {
	ImageFile *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type TrackDeleteURI struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type GetByPlaylistIDURI struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}

type PlaylistGetOneForm struct {
	Username     string `form:"username" binding:"required"`
	ChangeableID string `form:"changeableId" binding:"required"`
}

type PlaylistGetManyForm struct {
	UserID int64 `form:"userId" binding:"required"`
	Take   int   `form:"take" binding:"omitempty,min=1"`
	LastID int64 `form:"lastId" binding:"omitempty,min=1"`
}

type PlaylistGetManyWithSavedForm struct {
	Take   int   `form:"take" binding:"omitempty,min=1"`
	LastID int64 `form:"lastId" binding:"omitempty,min=1"`
}

type CreatePlaylistForm struct {
	Title        string                `form:"title" binding:"required,min=1,max=20"`
	ChangeableID string                `form:"changeableId" binding:"required,min=1,max=20"`
	ImageFile    *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type PlaylistChangeTitleURI struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}

type PlaylistChangeTitleForm struct {
	Title string `form:"title" binding:"required,min=1,max=20"`
}

type PlaylistChangeChangeableIDUri struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}

type PlaylistChangeChangeableIDForm struct {
	ChangeableID string `form:"changeableId" binding:"required,min=1,max=20"`
}

type PlaylistChangeImageURI struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}

type PlaylistChangeImageForm struct {
	ImageFile *multipart.FileHeader `form:"imageFile" binding:"required"`
}

type PlaylistDeleteURI struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
}

type PlaylistTrackIDsURI struct {
	PlaylistID int64 `uri:"playlistId" binding:"required"`
	TrackID    int64 `uri:"trackId" binding:"required"`
}

type PlaylistTracksGetManyForm struct {
	Take int `form:"take" binding:"omitempty,min=1"`
}

type AddTrackJSON struct {
	Position int `json:"position" binding:"omitempty,min=1"`
}

type UpdatePositionJSON struct {
	Position int `json:"position" binding:"required,min=1"`
}

type HistoryGetForm struct {
	Take int64 `form:"take" binding:"omitempty,min=1"`
}

type HistoryAddURI struct {
	TrackID int64 `uri:"trackId" binding:"required"`
}

type SearchForm struct {
	Query string `form:"query" binding:"required"`
}
