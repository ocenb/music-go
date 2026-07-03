package token

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/models"
)

type TokenServiceSuite struct {
	suite.Suite
	service   *Service
	jwtSecret string
}

func (s *TokenServiceSuite) SetupTest() {
	s.jwtSecret = "test-secret"
	s.service = New(mockTokenRepo{}, s.jwtSecret, time.Hour, time.Hour*24*30)
}

func TestTokenServiceSuite(t *testing.T) {
	suite.Run(t, new(TokenServiceSuite))
}

func (s *TokenServiceSuite) TestValidateToken_Valid() {
	claims := jwt.MapClaims{
		"userId":  int64(123),
		"tokenId": "test-token-id",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	s.Require().NoError(err)

	resultClaims, err := s.service.ValidateToken(tokenString)
	s.Require().NoError(err)
	s.NotNil(resultClaims)
	s.Equal(float64(123), resultClaims["userId"])
	s.Equal("test-token-id", resultClaims["tokenId"])
}

func (s *TokenServiceSuite) TestValidateToken_InvalidSignature() {
	claims := jwt.MapClaims{
		"userId":  int64(123),
		"tokenId": "test-token-id",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("wrong-secret"))
	s.Require().NoError(err)

	resultClaims, err := s.service.ValidateToken(tokenString)
	s.Require().Error(err)
	s.Nil(resultClaims)
	s.ErrorIs(err, errs.ErrInvalidToken)
}

func (s *TokenServiceSuite) TestValidateToken_ExpiredToken() {
	claims := jwt.MapClaims{
		"userId":  int64(123),
		"tokenId": "test-token-id",
		"exp":     time.Now().Add(-time.Hour).Unix(),
		"iat":     time.Now().Add(-time.Hour * 2).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	s.Require().NoError(err)

	resultClaims, err := s.service.ValidateToken(tokenString)
	s.Require().Error(err)
	s.Nil(resultClaims)
	s.ErrorIs(err, errs.ErrInvalidToken)
}

func (s *TokenServiceSuite) TestValidateToken_InvalidMethod() {
	claims := jwt.MapClaims{
		"userId":  int64(123),
		"tokenId": "test-token-id",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	s.Require().NoError(err)

	resultClaims, err := s.service.ValidateToken(tokenString)
	s.Require().Error(err)
	s.Nil(resultClaims)
	s.ErrorIs(err, errs.ErrInvalidToken)
}

func (s *TokenServiceSuite) TestGenerateTokens() {
	userID := int64(123)

	accessToken, refreshToken, tokenID, expiresAt, err := s.service.generateTokens(userID)
	s.Require().NoError(err)
	s.NotEmpty(accessToken)
	s.NotEmpty(refreshToken)
	s.NotEmpty(tokenID)
	s.True(expiresAt.After(time.Now().Add(time.Hour * 24 * 29)))

	accessClaims, err := s.service.ValidateToken(accessToken)
	s.Require().NoError(err)
	s.Equal(float64(userID), accessClaims["userId"])
	s.Equal(tokenID, accessClaims["tokenId"])

	refreshClaims, err := s.service.ValidateToken(refreshToken)
	s.Require().NoError(err)
	s.Equal(float64(userID), refreshClaims["userId"])
	s.Equal(tokenID, refreshClaims["tokenId"])
}

type mockTokenRepo struct{}

func (mockTokenRepo) GetTokenByID(context.Context, string) (*models.TokenModel, error) {
	return nil, nil
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
