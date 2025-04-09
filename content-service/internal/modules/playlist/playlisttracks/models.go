package playlisttracks

import (
	"time"
)

type PlaylistTrackModel struct {
	PlaylistID int64     `json:"playlistId"`
	TrackID    int64     `json:"trackId"`
	Position   int       `json:"position"`
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
