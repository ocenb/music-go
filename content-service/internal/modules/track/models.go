package track

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
