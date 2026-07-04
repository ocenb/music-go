package token_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/models"
	tokenservice "github.com/ocenb/music-go/user-service/internal/services/token"
)

type TokenServiceSuite struct {
	suite.Suite
	service   *tokenservice.Service
	jwtSecret string
}

func (s *TokenServiceSuite) SetupTest() {
	s.jwtSecret = "test-secret"
	s.service = tokenservice.New(mockTokenRepo{}, s.jwtSecret, time.Hour, time.Hour*24*30)
}

func TestTokenServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(TokenServiceSuite))
}

func (s *TokenServiceSuite) TestValidateToken_Valid() {
	claims := jwt.MapClaims{
		tokenservice.ClaimUserID:  int64(123),
		tokenservice.ClaimTokenID: "test-token-id",
		tokenservice.ClaimExp:     time.Now().Add(time.Hour).Unix(),
		tokenservice.ClaimIAT:     time.Now().Unix(),
	}
	signedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := signedToken.SignedString([]byte(s.jwtSecret))
	s.Require().NoError(err)

	resultClaims, err := s.service.ValidateToken(tokenString)
	s.Require().NoError(err)
	s.NotNil(resultClaims)
	s.InDelta(float64(123), resultClaims[tokenservice.ClaimUserID], 0)
	s.Equal("test-token-id", resultClaims[tokenservice.ClaimTokenID])
}

func (s *TokenServiceSuite) TestValidateToken_InvalidSignature() {
	claims := jwt.MapClaims{
		tokenservice.ClaimUserID:  int64(123),
		tokenservice.ClaimTokenID: "test-token-id",
		tokenservice.ClaimExp:     time.Now().Add(time.Hour).Unix(),
		tokenservice.ClaimIAT:     time.Now().Unix(),
	}
	signedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := signedToken.SignedString([]byte("wrong-secret"))
	s.Require().NoError(err)

	resultClaims, err := s.service.ValidateToken(tokenString)
	s.Require().Error(err)
	s.Nil(resultClaims)
	s.Require().ErrorIs(err, errs.ErrInvalidToken)
}

func (s *TokenServiceSuite) TestValidateToken_ExpiredToken() {
	claims := jwt.MapClaims{
		tokenservice.ClaimUserID:  int64(123),
		tokenservice.ClaimTokenID: "test-token-id",
		tokenservice.ClaimExp:     time.Now().Add(-time.Hour).Unix(),
		tokenservice.ClaimIAT:     time.Now().Add(-time.Hour * 2).Unix(),
	}
	signedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := signedToken.SignedString([]byte(s.jwtSecret))
	s.Require().NoError(err)

	resultClaims, err := s.service.ValidateToken(tokenString)
	s.Require().Error(err)
	s.Nil(resultClaims)
	s.Require().ErrorIs(err, errs.ErrInvalidToken)
}

func (s *TokenServiceSuite) TestValidateToken_InvalidMethod() {
	claims := jwt.MapClaims{
		tokenservice.ClaimUserID:  int64(123),
		tokenservice.ClaimTokenID: "test-token-id",
		tokenservice.ClaimExp:     time.Now().Add(time.Hour).Unix(),
		tokenservice.ClaimIAT:     time.Now().Unix(),
	}
	signedToken := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := signedToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	s.Require().NoError(err)

	resultClaims, err := s.service.ValidateToken(tokenString)
	s.Require().Error(err)
	s.Nil(resultClaims)
	s.Require().ErrorIs(err, errs.ErrInvalidToken)
}

func (s *TokenServiceSuite) TestGenerateTokens() {
	userID := int64(123)

	accessToken, refreshToken, tokenID, expiresAt, err := tokenservice.GenerateTokens(s.service, userID)
	s.Require().NoError(err)
	s.NotEmpty(accessToken)
	s.NotEmpty(refreshToken)
	s.NotEmpty(tokenID)
	s.True(expiresAt.After(time.Now().Add(time.Hour * 24 * 29)))

	accessClaims, err := s.service.ValidateToken(accessToken)
	s.Require().NoError(err)
	s.InDelta(float64(userID), accessClaims[tokenservice.ClaimUserID], 0)
	s.Equal(tokenID, accessClaims[tokenservice.ClaimTokenID])

	refreshClaims, err := s.service.ValidateToken(refreshToken)
	s.Require().NoError(err)
	s.InDelta(float64(userID), refreshClaims[tokenservice.ClaimUserID], 0)
	s.Equal(tokenID, refreshClaims[tokenservice.ClaimTokenID])
}

type mockTokenRepo struct{}

func (mockTokenRepo) GetTokenByID(context.Context, string) (*models.TokenModel, error) {
	return nil, errs.ErrTokenNotFound
}

func (mockTokenRepo) CreateToken(context.Context, string, int64, string, time.Time) error {
	return nil
}

func (mockTokenRepo) DeleteTokenByID(context.Context, string) error {
	return nil
}

func (mockTokenRepo) DeleteAllUserTokens(context.Context, int64) error {
	return nil
}

func (mockTokenRepo) DeleteExpiredTokens(context.Context) error {
	return nil
}
