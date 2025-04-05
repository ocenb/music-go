package token

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ocenb/music-go/user-service/internal/config"
	"github.com/ocenb/music-go/user-service/internal/models"
	tokenrepo "github.com/ocenb/music-go/user-service/internal/repos/token"
	"github.com/ocenb/music-go/user-service/internal/utils"
)

type TokenServiceInterface interface {
	GetTokenByID(ctx context.Context, tokenID string) (*models.TokenModel, error)
	CreateTokens(ctx context.Context, userID int64) (string, string, error)
	ValidateToken(tokenString string) (jwt.MapClaims, error)
	RevokeToken(ctx context.Context, tokenID string) error
	RevokeAllTokens(ctx context.Context, userID int64) error
	CleanupExpiredTokens(log *slog.Logger)
	GenerateVerificationToken() string
}

type TokenService struct {
	tokenRepo tokenrepo.TokenRepoInterface
	cfg       *config.Config
	log       *slog.Logger
}

func NewTokenService(cfg *config.Config, log *slog.Logger, tokenRepo tokenrepo.TokenRepoInterface) TokenServiceInterface {
	return &TokenService{
		tokenRepo: tokenRepo,
		cfg:       cfg,
		log:       log,
	}
}

func (s *TokenService) GetTokenByID(ctx context.Context, tokenID string) (*models.TokenModel, error) {
	s.log.Debug("Getting token by ID", slog.String("token_id", tokenID))
	token, err := s.tokenRepo.GetTokenByID(ctx, tokenID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Debug("Token not found", slog.String("token_id", tokenID))
			return nil, ErrTokenNotFound
		}
		s.log.Error("Failed to get token by ID", slog.String("token_id", tokenID), utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to get token by ID")
	}
	s.log.Debug("Token found", slog.String("token_id", tokenID), slog.Int64("user_id", token.UserId))
	return token, nil
}

func (s *TokenService) CreateTokens(ctx context.Context, userID int64) (string, string, error) {
	s.log.Debug("Creating tokens for user", slog.Int64("user_id", userID))
	accessToken, refreshToken, tokenId, expiresAt, err := s.generateTokens(userID)
	if err != nil {
		s.log.Error("Failed to generate tokens", slog.Int64("user_id", userID), utils.ErrLog(err))
		return "", "", err
	}

	err = s.tokenRepo.CreateToken(ctx, tokenId, userID, refreshToken, expiresAt)
	if err != nil {
		s.log.Error("Failed to create token in db", slog.Int64("user_id", userID), slog.String("token_id", tokenId), utils.ErrLog(err))
		return "", "", utils.InternalError(err, "failed to create token in db")
	}

	s.log.Info("Tokens created successfully", slog.Int64("user_id", userID), slog.String("token_id", tokenId), slog.Time("expires_at", expiresAt))
	return accessToken, refreshToken, nil
}

func (s *TokenService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (s *TokenService) RevokeToken(ctx context.Context, tokenID string) error {
	s.log.Debug("Revoking token", slog.String("token_id", tokenID))
	err := s.tokenRepo.DeleteTokenByID(ctx, tokenID)
	if err != nil {
		s.log.Error("Failed to revoke token", slog.String("token_id", tokenID), utils.ErrLog(err))
		return utils.InternalError(err, "failed to revoke token")
	}
	s.log.Info("Token revoked successfully", slog.String("token_id", tokenID))
	return nil
}

func (s *TokenService) RevokeAllTokens(ctx context.Context, userID int64) error {
	s.log.Debug("Revoking all tokens for user", slog.Int64("user_id", userID))
	err := s.tokenRepo.DeleteAllUserTokens(ctx, userID)
	if err != nil {
		s.log.Error("Failed to revoke all tokens", slog.Int64("user_id", userID), utils.ErrLog(err))
		return utils.InternalError(err, "failed to revoke all tokens")
	}
	s.log.Info("All tokens revoked successfully for user", slog.Int64("user_id", userID))
	return nil
}

func (s *TokenService) CleanupExpiredTokens(log *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.tokenRepo.DeleteExpiredTokens(ctx); err != nil {
		log.Error("Failed to cleanup expired tokens", utils.ErrLog(err))
	} else {
		log.Info("Successfully cleaned up expired tokens")
	}
}

func (s *TokenService) GenerateVerificationToken() string {
	token := uuid.New().String()
	s.log.Debug("Generated verification token")
	return token
}

func (s *TokenService) generateTokens(userID int64) (string, string, string, time.Time, error) {
	tokenId := uuid.New().String()
	expiresAt := time.Now().Add(s.cfg.RefreshTokenLiveTime)

	payload := jwt.MapClaims{
		"userId":  userID,
		"tokenId": tokenId,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", "", time.Time{}, utils.InternalError(err, "failed to sign access token")
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", "", time.Time{}, utils.InternalError(err, "failed to sign refresh token")
	}

	return accessTokenString, refreshTokenString, tokenId, expiresAt, nil
}
