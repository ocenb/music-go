package models

import "time"

type TrackModel struct {
	ID           int64     `json:"id"`
	ChangeableID string    `json:"changeableId"`
	Title        string    `json:"title"`
	Duration     int64     `json:"duration"`
	Plays        int64     `json:"plays"`
	Audio        string    `json:"audio"`
	Image        string    `json:"image"`
	UserID       int64     `json:"userId"`
	Username     string    `json:"username"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type TrackWithLikedModel struct {
	TrackModel
	IsLiked bool       `json:"isLiked"`
	LikedAt *time.Time `json:"likedAt,omitempty"`
}

type UserLikedTrackModel struct {
	UserID  int64     `json:"userId"`
	TrackID int64     `json:"trackId"`
	AddedAt time.Time `json:"addedAt"`
}

type PlaylistModel struct {
	ID           int64     `json:"id"`
	ChangeableID string    `json:"changeableId"`
	Title        string    `json:"title"`
	Image        string    `json:"image"`
	UserID       int64     `json:"userId"`
	Username     string    `json:"username"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type PlaylistWithSavedModel struct {
	PlaylistModel
	IsSaved bool       `json:"isSaved"`
	SavedAt *time.Time `json:"savedAt,omitempty"`
}

type PlaylistTrackModel struct {
	PlaylistID int64     `json:"playlistId"`
	TrackID    int64     `json:"trackId"`
	Position   int       `json:"position"`
	AddedAt    time.Time `json:"addedAt"`
}

type UserSavedPlaylistModel struct {
	UserID     int64     `json:"userId"`
	PlaylistID int64     `json:"playlistId"`
	AddedAt    time.Time `json:"addedAt"`
}

type TrackInPlaylistModel struct {
	PlaylistID     int64     `json:"playlistId"`
	TrackID        int64     `json:"trackId"`
	Position       int       `json:"position"`
	Title          string    `json:"title"`
	Artist         string    `json:"artist"`
	Duration       int       `json:"duration"`
	CoverImagePath string    `json:"coverImagePath"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ListeningHistoryModel struct {
	UserID   int64     `json:"userId"`
	TrackID  int64     `json:"trackId"`
	PlayedAt time.Time `json:"playedAt"`
}
