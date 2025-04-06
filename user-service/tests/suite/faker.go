package suite

import (
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ocenb/music-go/user-service/internal/models"
)

var (
	AdminUsername    = "admin"
	AdminEmail       = "admin@example.com"
	AdminPassword    = "Password123!"
	ToDeleteUsername = "todelete"
	ToDeleteEmail    = "todelete@example.com"
	ToDeletePassword = "Password123!"
	ToChangeUsername = "tochange"
	ToChangeEmail    = "tochange@example.com"
	ToChangePassword = "Password123!"
	ToFollowUsername = "tofollow"
	ToFollowEmail    = "tofollow@example.com"
	ToFollowPassword = "Password123!"
)

func init() {
	err := gofakeit.Seed(0)
	if err != nil {
		panic(err)
	}
}

func ValidUsername() string {
	starters := "abcdefghijklmnopqrstuvwxyz0123456789"
	chars := "abcdefghijklmnopqrstuvwxyz0123456789_-"

	length := rand.Intn(13) + 3

	username := string(starters[rand.Intn(len(starters))])

	for i := 1; i < length; i++ {
		username += string(chars[rand.Intn(len(chars))])
	}

	return username
}

func validPassword() string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*?-"

	length := rand.Intn(9) + 8

	password := ""
	for range length {
		password += string(chars[rand.Intn(len(chars))])
	}

	return password
}

func FakeUser() *models.UserFullModel {
	return &models.UserFullModel{
		ID:       gofakeit.Int64(),
		Username: ValidUsername(),
		Email:    gofakeit.Email(),
		Password: validPassword(),
	}
}

func FakeRegisterRequest() (string, string, string) {
	return ValidUsername(),
		gofakeit.Email(),
		validPassword()
}

func FakeTokenID() string {
	return uuid.NewString()
}

func FakeUserWithFollowers() *models.UserFullModel {
	user := FakeUser()
	user.FollowersCount = int64(gofakeit.IntRange(0, 1000))
	return user
}

func FakePasswordUpdate() (string, string) {
	return validPassword(), validPassword()
}

func FakeAccessToken() string {
	userId := gofakeit.Int64()
	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"jti": uuid.NewString(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, _ := token.SignedString([]byte("test-secret-key"))
	return tokenString
}

func FakeRefreshToken() string {
	return uuid.NewString() + uuid.NewString()
}
