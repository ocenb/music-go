package history

import "time"

type ListeningHistoryModel struct {
	UserID   int64     `json:"userId"`
	TrackID  int64     `json:"trackId"`
	PlayedAt time.Time `json:"playedAt"`
}
