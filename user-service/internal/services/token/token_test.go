package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ocenb/music-go/user-service/internal/config"
	"github.com/ocenb/music-go/user-service/internal/logger"
	"github.com/ocenb/music-go/user-service/internal/mocks/tokenmocks"
	"github.com/stretchr/testify/assert"
)

func suite() (TokenServiceInterface, *config.Config) {
	cfg := &config.Config{
		JWTSecret:            "test-secret",
		RefreshTokenLiveTime: time.Hour * 24 * 30,
	}
	log := logger.NewForTest()
	mockRepo := new(tokenmocks.MockTokenRepo)
	service := NewTokenService(cfg, log, mockRepo)

	return service, cfg
}

func TestValidateToken_Valid(t *testing.T) {
	service, cfg := suite()

	claims := jwt.MapClaims{
		"userId":  int64(123),
		"tokenId": "test-token-id",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	assert.NoError(t, err)

	resultClaims, err := service.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, resultClaims)
	assert.Equal(t, float64(123), resultClaims["userId"])
	assert.Equal(t, "test-token-id", resultClaims["tokenId"])
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	service, _ := suite()

	claims := jwt.MapClaims{
		"userId":  int64(123),
		"tokenId": "test-token-id",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("wrong-secret"))
	assert.NoError(t, err)

	resultClaims, err := service.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, resultClaims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	service, cfg := suite()

	claims := jwt.MapClaims{
		"userId":  int64(123),
		"tokenId": "test-token-id",
		"exp":     time.Now().Add(-time.Hour).Unix(),
		"iat":     time.Now().Add(-time.Hour * 2).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	assert.NoError(t, err)

	resultClaims, err := service.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, resultClaims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidateToken_InvalidMethod(t *testing.T) {
	service, _ := suite()

	claims := jwt.MapClaims{
		"userId":  int64(123),
		"tokenId": "test-token-id",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	resultClaims, err := service.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, resultClaims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestGenerateTokens(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:            "test-secret",
		RefreshTokenLiveTime: time.Hour * 24 * 30,
		AccessTokenLiveTime:  time.Hour * 24 * 30,
	}
	log := logger.NewForTest()
	mockRepo := new(tokenmocks.MockTokenRepo)
	service := &TokenService{
		tokenRepo: mockRepo,
		cfg:       cfg,
		log:       log,
	}

	userID := int64(123)

	accessToken, refreshToken, tokenID, expiresAt, err := service.generateTokens(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEmpty(t, tokenID)
	assert.True(t, expiresAt.After(time.Now().Add(time.Hour*24*29)))

	accessClaims, err := service.ValidateToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, float64(userID), accessClaims["userId"])
	assert.Equal(t, tokenID, accessClaims["tokenId"])

	refreshClaims, err := service.ValidateToken(refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, float64(userID), refreshClaims["userId"])
	assert.Equal(t, tokenID, refreshClaims["tokenId"])
}
