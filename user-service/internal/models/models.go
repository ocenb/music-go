package models

type UserFullModel struct {
	ID                         int64
	Username                   string
	Email                      string
	Password                   string
	IsVerified                 bool
	VerificationToken          *string
	VerificationTokenExpiresAt *string
	FollowersCount             int64
	CreatedAt                  string
}

type TokenModel struct {
	ID           string
	CreatedAt    string
	RefreshToken string
	UserId       int64
	ExpiresAt    string
}
