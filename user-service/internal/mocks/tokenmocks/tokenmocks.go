package tokenmocks

import (
	"context"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ocenb/music-go/user-service/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GetTokenByID(ctx context.Context, tokenID string) (*models.TokenModel, error) {
	args := m.Called(ctx, tokenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TokenModel), args.Error(1)
}

func (m *MockTokenService) CreateTokens(ctx context.Context, userID int64) (string, string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockTokenService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jwt.MapClaims), args.Error(1)
}

func (m *MockTokenService) RevokeToken(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockTokenService) RevokeAllTokens(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTokenService) CleanupExpiredTokens(log *slog.Logger) {
	m.Called(log)
}

func (m *MockTokenService) GenerateVerificationToken() string {
	args := m.Called()
	return args.String(0)
}

type MockTokenRepo struct {
	mock.Mock
}

func (m *MockTokenRepo) GetTokenByID(ctx context.Context, tokenID string) (*models.TokenModel, error) {
	args := m.Called(ctx, tokenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TokenModel), args.Error(1)
}

func (m *MockTokenRepo) CreateToken(ctx context.Context, tokenID string, userID int64, refreshToken string, expiresAt time.Time) error {
	args := m.Called(ctx, tokenID, userID, refreshToken, expiresAt)
	return args.Error(0)
}

func (m *MockTokenRepo) DeleteTokenByID(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockTokenRepo) DeleteAllUserTokens(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTokenRepo) DeleteExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
