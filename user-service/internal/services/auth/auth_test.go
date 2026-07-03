package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/models"
	"github.com/ocenb/music-go/user-service/internal/storage/transactor"
)

type AuthServiceSuite struct {
	suite.Suite
	mockUserService         *MockUserService
	mockTokenService        *MockTokenService
	mockNotificationClient  *MockNotificationClient
	tm                      *transactor.Manager
	service                 *Service
}

func (s *AuthServiceSuite) SetupTest() {
	s.mockUserService = NewMockUserService(s.T())
	s.mockTokenService = NewMockTokenService(s.T())
	s.mockNotificationClient = NewMockNotificationClient(s.T())
	s.tm = transactor.NewMock()
	s.service = New(s.mockUserService, s.mockTokenService, s.mockNotificationClient, s.tm, 12)
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceSuite))
}

func (s *AuthServiceSuite) TestValidateAccessToken_Valid() {
	ctx := context.Background()
	tokenID := "test-token-id"
	userID := int64(123)
	claims := jwt.MapClaims{
		"userId":  float64(userID),
		"tokenId": tokenID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	s.mockTokenService.On("ValidateToken", "valid-token").Return(claims, nil)
	s.mockTokenService.On("GetTokenByID", ctx, tokenID).Return(&models.TokenModel{ID: tokenID}, nil)

	expectedUser := &models.UserFullModel{
		ID:         userID,
		Username:   "testuser",
		Email:      "test@example.com",
		IsVerified: true,
	}
	s.mockUserService.On("GetById", ctx, userID).Return(expectedUser, nil)

	user, resultTokenID, err := s.service.ValidateAccessToken(ctx, "valid-token")
	s.Require().NoError(err)
	s.Equal(expectedUser, user)
	s.Equal(tokenID, resultTokenID)
}

func (s *AuthServiceSuite) TestValidateAccessToken_InvalidToken() {
	ctx := context.Background()

	s.mockTokenService.On("ValidateToken", "invalid-token").Return(nil, errs.ErrInvalidToken)

	user, tokenID, err := s.service.ValidateAccessToken(ctx, "invalid-token")
	s.Require().Error(err)
	s.Nil(user)
	s.Empty(tokenID)
}

func (s *AuthServiceSuite) TestValidateAccessToken_InvalidUserID() {
	ctx := context.Background()
	claims := jwt.MapClaims{
		"tokenId": "test-token-id",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	s.mockTokenService.On("ValidateToken", "token-without-userid").Return(claims, nil)

	user, tokenID, err := s.service.ValidateAccessToken(ctx, "token-without-userid")
	s.Require().Error(err)
	s.ErrorIs(err, errs.ErrInvalidToken)
	s.Nil(user)
	s.Empty(tokenID)
}

func (s *AuthServiceSuite) TestValidateAccessToken_InvalidTokenID() {
	ctx := context.Background()
	claims := jwt.MapClaims{
		"userId": float64(123),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	s.mockTokenService.On("ValidateToken", "token-without-tokenid").Return(claims, nil)

	user, tokenID, err := s.service.ValidateAccessToken(ctx, "token-without-tokenid")
	s.Require().Error(err)
	s.ErrorIs(err, errs.ErrInvalidToken)
	s.Nil(user)
	s.Empty(tokenID)
}

func (s *AuthServiceSuite) TestValidateAccessToken_UserNotFound() {
	ctx := context.Background()
	tokenID := "test-token-id"
	userID := int64(123)
	claims := jwt.MapClaims{
		"userId":  float64(userID),
		"tokenId": tokenID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	s.mockTokenService.On("ValidateToken", "valid-token-nonexistent-user").Return(claims, nil)
	s.mockTokenService.On("GetTokenByID", ctx, tokenID).Return(&models.TokenModel{ID: tokenID}, nil)
	s.mockUserService.On("GetById", ctx, userID).Return(nil, errs.ErrUserNotFound)

	user, resultTokenID, err := s.service.ValidateAccessToken(ctx, "valid-token-nonexistent-user")
	s.Require().Error(err)
	s.ErrorIs(err, errs.ErrUserNotFound)
	s.Nil(user)
	s.Empty(resultTokenID)
}

func (s *AuthServiceSuite) TestValidateRefreshToken_Valid() {
	ctx := context.Background()
	tokenID := "test-token-id"
	userID := int64(123)
	refreshToken := "valid-refresh-token"
	claims := jwt.MapClaims{
		"userId":  float64(userID),
		"tokenId": tokenID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	s.mockTokenService.On("ValidateToken", refreshToken).Return(claims, nil)
	s.mockTokenService.On("GetTokenByID", ctx, tokenID).Return(&models.TokenModel{
		ID:           tokenID,
		RefreshToken: refreshToken,
		UserId:       userID,
	}, nil)

	expectedUser := &models.UserFullModel{
		ID:         userID,
		Username:   "testuser",
		Email:      "test@example.com",
		IsVerified: true,
	}
	s.mockUserService.On("GetById", ctx, userID).Return(expectedUser, nil)

	user, resultTokenID, err := s.service.validateRefreshToken(ctx, refreshToken)
	s.Require().NoError(err)
	s.Equal(expectedUser, user)
	s.Equal(tokenID, resultTokenID)
}

func (s *AuthServiceSuite) TestValidateRefreshToken_InvalidToken() {
	ctx := context.Background()

	s.mockTokenService.On("ValidateToken", "invalid-token").Return(nil, errs.ErrInvalidToken)

	user, tokenID, err := s.service.validateRefreshToken(ctx, "invalid-token")
	s.Require().Error(err)
	s.ErrorIs(err, errs.ErrInvalidRefreshToken)
	s.Nil(user)
	s.Empty(tokenID)
}

func (s *AuthServiceSuite) TestValidateRefreshToken_RefreshTokenMismatch() {
	ctx := context.Background()
	tokenID := "test-token-id"
	userID := int64(123)
	refreshToken := "refresh-token"
	claims := jwt.MapClaims{
		"userId":  float64(userID),
		"tokenId": tokenID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	s.mockTokenService.On("ValidateToken", refreshToken).Return(claims, nil)
	s.mockTokenService.On("GetTokenByID", ctx, tokenID).Return(&models.TokenModel{
		ID:           tokenID,
		RefreshToken: "different-refresh-token",
		UserId:       userID,
	}, nil)

	user, resultTokenID, err := s.service.validateRefreshToken(ctx, refreshToken)
	s.Require().Error(err)
	s.ErrorIs(err, errs.ErrInvalidRefreshToken)
	s.Nil(user)
	s.Empty(resultTokenID)
}
