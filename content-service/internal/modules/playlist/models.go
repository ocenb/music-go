package playlist

import "time"

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
