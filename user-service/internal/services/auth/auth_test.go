package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ocenb/music-go/user-service/internal/config"
	"github.com/ocenb/music-go/user-service/internal/logger"
	"github.com/ocenb/music-go/user-service/internal/mocks/authmocks"
	"github.com/ocenb/music-go/user-service/internal/mocks/notificationmocks"
	"github.com/ocenb/music-go/user-service/internal/mocks/tokenmocks"
	"github.com/ocenb/music-go/user-service/internal/mocks/usermocks"
	"github.com/ocenb/music-go/user-service/internal/models"
	"github.com/ocenb/music-go/user-service/internal/services/token"
	"github.com/ocenb/music-go/user-service/internal/services/user"
	"github.com/stretchr/testify/assert"
)

func suite() (context.Context, *tokenmocks.MockTokenService, *usermocks.MockUserService, AuthServiceInterface) {
	ctx := context.Background()
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	log := logger.NewForTest()
	mockTokenService := new(tokenmocks.MockTokenService)
	mockUserService := new(usermocks.MockUserService)
	mockAuthRepo := new(authmocks.MockAuthRepo)
	mockNotificationClient := new(notificationmocks.MockNotificationClient)
	authService := NewAuthService(cfg, log, mockUserService, mockTokenService, mockAuthRepo, mockNotificationClient)

	return ctx, mockTokenService, mockUserService, authService
}

func TestValidateAccessToken_Valid(t *testing.T) {
	ctx, mockTokenService, mockUserService, authService := suite()

	tokenId := "test-token-id"
	userId := int64(123)
	claims := jwt.MapClaims{
		"userId":  float64(userId),
		"tokenId": tokenId,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	mockTokenService.On("ValidateToken", "valid-token").Return(claims, nil)
	mockTokenService.On("GetTokenByID", ctx, tokenId).Return(&models.TokenModel{ID: tokenId}, nil)

	expectedUser := &models.UserFullModel{
		ID:         userId,
		Username:   "testuser",
		Email:      "test@example.com",
		IsVerified: true,
	}
	mockUserService.On("GetById", ctx, userId).Return(expectedUser, nil)

	user, resultTokenId, err := authService.ValidateAccessToken(ctx, "valid-token")
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	assert.Equal(t, tokenId, resultTokenId)

	mockTokenService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	ctx, mockTokenService, _, authService := suite()

	mockTokenService.On("ValidateToken", "invalid-token").Return(nil, token.ErrInvalidToken)

	user, tokenId, err := authService.ValidateAccessToken(ctx, "invalid-token")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, tokenId)

	mockTokenService.AssertExpectations(t)
}

func TestValidateAccessToken_InvalidUserID(t *testing.T) {
	ctx, mockTokenService, _, authService := suite()

	claims := jwt.MapClaims{
		"tokenId": "test-token-id",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	mockTokenService.On("ValidateToken", "token-without-userid").Return(claims, nil)

	user, tokenId, err := authService.ValidateAccessToken(ctx, "token-without-userid")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Nil(t, user)
	assert.Empty(t, tokenId)

	mockTokenService.AssertExpectations(t)
}

func TestValidateAccessToken_InvalidTokenID(t *testing.T) {
	ctx, mockTokenService, _, authService := suite()

	claims := jwt.MapClaims{
		"userId": float64(123),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	mockTokenService.On("ValidateToken", "token-without-tokenid").Return(claims, nil)

	user, tokenId, err := authService.ValidateAccessToken(ctx, "token-without-tokenid")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Nil(t, user)
	assert.Empty(t, tokenId)

	mockTokenService.AssertExpectations(t)
}

func TestValidateAccessToken_UserNotFound(t *testing.T) {
	ctx, mockTokenService, mockUserService, authService := suite()

	tokenId := "test-token-id"
	userId := int64(123)
	claims := jwt.MapClaims{
		"userId":  float64(userId),
		"tokenId": tokenId,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	mockTokenService.On("ValidateToken", "valid-token-nonexistent-user").Return(claims, nil)
	mockTokenService.On("GetTokenByID", ctx, tokenId).Return(&models.TokenModel{ID: tokenId}, nil)
	mockUserService.On("GetById", ctx, userId).Return(nil, user.ErrUserNotFound)

	user, resultTokenId, err := authService.ValidateAccessToken(ctx, "valid-token-nonexistent-user")
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)
	assert.Empty(t, resultTokenId)

	mockTokenService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestValidateRefreshToken_Valid(t *testing.T) {
	ctx, mockTokenService, mockUserService, authService := suite()

	tokenId := "test-token-id"
	userId := int64(123)
	refreshToken := "valid-refresh-token"

	claims := jwt.MapClaims{
		"userId":  float64(userId),
		"tokenId": tokenId,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	mockTokenService.On("ValidateToken", refreshToken).Return(claims, nil)

	mockTokenService.On("GetTokenByID", ctx, tokenId).Return(&models.TokenModel{
		ID:           tokenId,
		RefreshToken: refreshToken,
		UserId:       userId,
	}, nil)

	expectedUser := &models.UserFullModel{
		ID:         userId,
		Username:   "testuser",
		Email:      "test@example.com",
		IsVerified: true,
	}
	mockUserService.On("GetById", ctx, userId).Return(expectedUser, nil)

	user, resultTokenId, err := authService.validateRefreshToken(ctx, refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	assert.Equal(t, tokenId, resultTokenId)

	mockTokenService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestValidateRefreshToken_InvalidToken(t *testing.T) {
	ctx, mockTokenService, _, authService := suite()

	mockTokenService.On("ValidateToken", "invalid-token").Return(nil, token.ErrInvalidToken)

	user, tokenId, err := authService.validateRefreshToken(ctx, "invalid-token")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRefreshToken, err)
	assert.Nil(t, user)
	assert.Empty(t, tokenId)

	mockTokenService.AssertExpectations(t)
}

func TestValidateRefreshToken_RefreshTokenMismatch(t *testing.T) {
	ctx, mockTokenService, _, authService := suite()

	tokenId := "test-token-id"
	userId := int64(123)
	refreshToken := "refresh-token"

	claims := jwt.MapClaims{
		"userId":  float64(userId),
		"tokenId": tokenId,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	mockTokenService.On("ValidateToken", refreshToken).Return(claims, nil)
	mockTokenService.On("GetTokenByID", ctx, tokenId).Return(&models.TokenModel{
		ID:           tokenId,
		RefreshToken: "different-refresh-token",
		UserId:       userId,
	}, nil)

	user, resultTokenId, err := authService.validateRefreshToken(ctx, refreshToken)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRefreshToken, err)
	assert.Nil(t, user)
	assert.Empty(t, resultTokenId)

	mockTokenService.AssertExpectations(t)
}
