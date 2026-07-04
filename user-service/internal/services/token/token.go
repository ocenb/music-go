package token

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/models"
)

const (
	ClaimUserID  = "userId"
	ClaimTokenID = "tokenId"
	ClaimExp     = "exp"
	ClaimIAT     = "iat"

	tokenCleanupTimeout = 30 * time.Second
)

type Repo interface {
	GetTokenByID(ctx context.Context, tokenID string) (*models.TokenModel, error)
	CreateToken(ctx context.Context, tokenID string, userID int64, refreshToken string, expiresAt time.Time) error
	DeleteTokenByID(ctx context.Context, tokenID string) error
	DeleteAllUserTokens(ctx context.Context, userID int64) error
	DeleteExpiredTokens(ctx context.Context) error
}

type Service struct {
	repo                 Repo
	jwtSecret            string
	accessTokenLiveTime  time.Duration
	refreshTokenLiveTime time.Duration
}

func New(
	repo Repo,
	jwtSecret string,
	accessTokenLiveTime time.Duration,
	refreshTokenLiveTime time.Duration,
) *Service {
	return &Service{
		repo:                 repo,
		jwtSecret:            jwtSecret,
		accessTokenLiveTime:  accessTokenLiveTime,
		refreshTokenLiveTime: refreshTokenLiveTime,
	}
}

func (s *Service) GetTokenByID(ctx context.Context, tokenID string) (*models.TokenModel, error) {
	token, err := s.repo.GetTokenByID(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("TokenService.GetTokenByID: %w", err)
	}
	return token, nil
}

func (s *Service) CreateTokens(ctx context.Context, userID int64) (string, string, error) {
	accessToken, refreshToken, tokenID, expiresAt, err := s.generateTokens(userID)
	if err != nil {
		return "", "", fmt.Errorf("TokenService.CreateTokens: %w", err)
	}

	if err := s.repo.CreateToken(ctx, tokenID, userID, refreshToken, expiresAt); err != nil {
		return "", "", fmt.Errorf("TokenService.CreateTokens: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.ErrInvalidSigningMethod
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("TokenService.ValidateToken: %w: %w", errs.ErrInvalidToken, err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("TokenService.ValidateToken: %w", errs.ErrInvalidToken)
}

func (s *Service) RevokeToken(ctx context.Context, tokenID string) error {
	if err := s.repo.DeleteTokenByID(ctx, tokenID); err != nil {
		return fmt.Errorf("TokenService.RevokeToken: %w", err)
	}
	return nil
}

func (s *Service) RevokeAllTokens(ctx context.Context, userID int64) error {
	if err := s.repo.DeleteAllUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("TokenService.RevokeAllTokens: %w", err)
	}
	return nil
}

func (s *Service) CleanupExpiredTokens() error {
	ctx, cancel := context.WithTimeout(context.Background(), tokenCleanupTimeout)
	defer cancel()

	if err := s.repo.DeleteExpiredTokens(ctx); err != nil {
		return fmt.Errorf("TokenService.CleanupExpiredTokens: %w", err)
	}
	return nil
}

func (s *Service) GenerateVerificationToken() string {
	return uuid.New().String()
}

func (s *Service) generateTokens(userID int64) (string, string, string, time.Time, error) {
	tokenID := uuid.New().String()
	refreshExpiresAt := time.Now().Add(s.refreshTokenLiveTime)
	accessExpiresAt := time.Now().Add(s.accessTokenLiveTime)

	accessPayload := jwt.MapClaims{
		ClaimUserID:  userID,
		ClaimTokenID: tokenID,
		ClaimExp:     accessExpiresAt.Unix(),
		ClaimIAT:     time.Now().Unix(),
	}
	refreshPayload := jwt.MapClaims{
		ClaimUserID:  userID,
		ClaimTokenID: tokenID,
		ClaimExp:     refreshExpiresAt.Unix(),
		ClaimIAT:     time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessPayload)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", "", time.Time{}, fmt.Errorf("TokenService.generateTokens: failed to sign access token: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshPayload)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", "", time.Time{}, fmt.Errorf("TokenService.generateTokens: failed to sign refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, tokenID, refreshExpiresAt, nil
}
