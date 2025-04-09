package suite

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

const (
	AdminUsername = "admin"
	AdminUserID   = 1

	TestTrackID1 = 1
	TestTrackID2 = 2
	TestTrackID3 = 3

	TestPlaylistID = 1
)

func init() {
	err := gofakeit.Seed(0)
	if err != nil {
		panic(err)
	}
}

func FakePlaylistTitle() string {
	return gofakeit.Word() + " " + gofakeit.Word()
}

func FakeChangeableID() string {
	return gofakeit.Username()
}

func FakeAuthToken() string {
	return "test-token-" + uuid.NewString()
}

func FakeTrackPosition() int {
	return gofakeit.IntRange(1, 100)
}
