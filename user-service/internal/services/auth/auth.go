package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ocenb/music-protos/gen/userservice"
	"golang.org/x/crypto/bcrypt"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/models"
	"github.com/ocenb/music-go/user-service/internal/storage/transactor"
)

type UserService interface {
	GetByUsername(ctx context.Context, username string) (*userservice.UserPublicModel, error)
	GetByID(ctx context.Context, id int64) (*models.UserFullModel, error)
	GetByEmail(ctx context.Context, email string) (*models.UserFullModel, error)
	GetByVerificationToken(ctx context.Context, verificationToken string) (*models.UserFullModel, error)
	UpdateVerificationToken(ctx context.Context, userID int64, newVerificationToken string) (*userservice.UserPrivateModel, error)
	SetVerified(ctx context.Context, userID int64) (*userservice.UserPrivateModel, error)
	Create(ctx context.Context, username, email, password, verificationToken string) (*userservice.UserPrivateModel, error)
	ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, error)
	ChangePassword(ctx context.Context, userID int64, password string) (*userservice.UserPrivateModel, error)
}

type TokenService interface {
	GetTokenByID(ctx context.Context, tokenID string) (*models.TokenModel, error)
	CreateTokens(ctx context.Context, userID int64) (string, string, error)
	ValidateToken(tokenString string) (jwt.MapClaims, error)
	RevokeToken(ctx context.Context, tokenID string) error
	RevokeAllTokens(ctx context.Context, userID int64) error
	GenerateVerificationToken() string
}

type NotificationClient interface {
	SendEmailNotification(ctx context.Context, email, msg string) error
}

type Service struct {
	userService        UserService
	tokenService       TokenService
	notificationClient NotificationClient
	tm                 *transactor.Manager
	bcryptCost         int
}

func New(
	userService UserService,
	tokenService TokenService,
	notificationClient NotificationClient,
	tm *transactor.Manager,
	bcryptCost int,
) *Service {
	return &Service{
		userService:        userService,
		tokenService:       tokenService,
		notificationClient: notificationClient,
		tm:                 tm,
		bcryptCost:         bcryptCost,
	}
}

func (s *Service) Register(ctx context.Context, username, email, password string) (*userservice.UserPrivateModel, error) {
	userByEmail, err := s.userService.GetByEmail(ctx, email)
	if err == nil && userByEmail != nil {
		return nil, errs.ErrUserEmailExists
	}
	if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
		return nil, fmt.Errorf("AuthService.Register: %w", err)
	}

	userByName, err := s.userService.GetByUsername(ctx, username)
	if err == nil && userByName != nil {
		return nil, errs.ErrUserUsernameExists
	}
	if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
		return nil, fmt.Errorf("AuthService.Register: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Register: failed to hash password: %w", err)
	}

	verificationToken := s.tokenService.GenerateVerificationToken()

	user, err := s.userService.Create(ctx, username, email, string(hashedPassword), verificationToken)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Register: %w", err)
	}

	if err := s.notificationClient.SendEmailNotification(ctx, user.Email, verificationToken); err != nil {
		return nil, fmt.Errorf("AuthService.Register: %w", err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*userservice.UserPrivateModel, string, string, error) {
	user, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", errs.ErrUserEmailNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", "", errs.ErrInvalidPassword
	}

	if !user.IsVerified {
		return nil, "", "", errs.ErrUserNotVerified
	}

	accessToken, refreshToken, err := s.tokenService.CreateTokens(ctx, user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("AuthService.Login: %w", err)
	}

	return &userservice.UserPrivateModel{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, accessToken, refreshToken, nil
}

func (s *Service) Logout(ctx context.Context, tokenID string) error {
	if err := s.tokenService.RevokeToken(ctx, tokenID); err != nil {
		return fmt.Errorf("AuthService.Logout: %w", err)
	}
	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userID int64) error {
	if err := s.tokenService.RevokeAllTokens(ctx, userID); err != nil {
		return fmt.Errorf("AuthService.LogoutAll: %w", err)
	}
	return nil
}

func (s *Service) Refresh(ctx context.Context, oldRefreshToken string) (*userservice.UserPrivateModel, string, string, error) {
	user, tokenID, err := s.validateRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		return nil, "", "", err
	}

	var accessToken, refreshToken string

	err = s.tm.Run(ctx, func(txCtx context.Context) error {
		if err := s.tokenService.RevokeToken(txCtx, tokenID); err != nil {
			return fmt.Errorf("failed to revoke old refresh token: %w", err)
		}

		var tokenErr error
		accessToken, refreshToken, tokenErr = s.tokenService.CreateTokens(txCtx, user.ID)
		if tokenErr != nil {
			return fmt.Errorf("failed to create new tokens: %w", tokenErr)
		}

		return nil
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("AuthService.Refresh: %w", err)
	}

	return &userservice.UserPrivateModel{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, accessToken, refreshToken, nil
}

func (s *Service) Verify(ctx context.Context, verificationToken string) (*userservice.UserPrivateModel, string, string, error) {
	user, err := s.userService.GetByVerificationToken(ctx, verificationToken)
	if err != nil {
		return nil, "", "", errs.ErrTokenNotFound
	}

	expDate, err := time.Parse("2006-01-02", *user.VerificationTokenExpiresAt)
	if err != nil {
		return nil, "", "", fmt.Errorf("AuthService.Verify: failed to parse expiration date: %w", err)
	}
	if expDate.Before(time.Now()) {
		return nil, "", "", errs.ErrTokenExpired
	}

	var verifiedUser *userservice.UserPrivateModel
	var accessToken, refreshToken string

	err = s.tm.Run(ctx, func(txCtx context.Context) error {
		var err error

		verifiedUser, err = s.userService.SetVerified(txCtx, user.ID)
		if err != nil {
			return fmt.Errorf("failed to set user as verified: %w", err)
		}

		accessToken, refreshToken, err = s.tokenService.CreateTokens(txCtx, verifiedUser.Id)
		if err != nil {
			return fmt.Errorf("failed to create tokens: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("AuthService.Verify: %w", err)
	}

	return verifiedUser, accessToken, refreshToken, nil
}

func (s *Service) NewVerification(ctx context.Context, email, password string) (*userservice.UserPrivateModel, error) {
	user, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		return nil, errs.ErrUserEmailNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errs.ErrInvalidPassword
	}

	if user.IsVerified {
		return nil, errs.ErrUserAlreadyVerified
	}

	newVerificationToken := s.tokenService.GenerateVerificationToken()
	updatedUser, err := s.userService.UpdateVerificationToken(ctx, user.ID, newVerificationToken)
	if err != nil {
		return nil, fmt.Errorf("AuthService.NewVerification: %w", err)
	}

	if err := s.notificationClient.SendEmailNotification(ctx, user.Email, newVerificationToken); err != nil {
		return nil, fmt.Errorf("AuthService.NewVerification: %w", err)
	}

	return updatedUser, nil
}

func (s *Service) ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, string, string, error) {
	var user *userservice.UserPrivateModel
	var newAccessToken, newRefreshToken string

	err := s.tm.Run(ctx, func(txCtx context.Context) error {
		var err error

		user, err = s.userService.ChangeEmail(txCtx, userID, email)
		if err != nil {
			return fmt.Errorf("failed to change email: %w", err)
		}

		if err := s.tokenService.RevokeAllTokens(txCtx, userID); err != nil {
			return fmt.Errorf("failed to revoke all tokens: %w", err)
		}

		newAccessToken, newRefreshToken, err = s.tokenService.CreateTokens(txCtx, userID)
		if err != nil {
			return fmt.Errorf("failed to create new tokens: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("AuthService.ChangeEmail: %w", err)
	}

	return user, newAccessToken, newRefreshToken, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, truePassword, oldPassword, newPassword string) (*userservice.UserPrivateModel, string, string, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(truePassword), []byte(oldPassword)); err != nil {
		return nil, "", "", errs.ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.bcryptCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("AuthService.ChangePassword: failed to hash password: %w", err)
	}

	var user *userservice.UserPrivateModel
	var newAccessToken, newRefreshToken string

	err = s.tm.Run(ctx, func(txCtx context.Context) error {
		var err error

		user, err = s.userService.ChangePassword(txCtx, userID, string(hashedPassword))
		if err != nil {
			return fmt.Errorf("failed to change password: %w", err)
		}

		if err := s.tokenService.RevokeAllTokens(txCtx, userID); err != nil {
			return fmt.Errorf("failed to revoke all tokens: %w", err)
		}

		newAccessToken, newRefreshToken, err = s.tokenService.CreateTokens(txCtx, userID)
		if err != nil {
			return fmt.Errorf("failed to create new tokens: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("AuthService.ChangePassword: %w", err)
	}

	return user, newAccessToken, newRefreshToken, nil
}

func (s *Service) ValidateAccessToken(ctx context.Context, accessToken string) (*models.UserFullModel, string, error) {
	claims, err := s.tokenService.ValidateToken(accessToken)
	if err != nil {
		return nil, "", err
	}

	userID, ok := claims["userId"].(float64)
	if !ok {
		return nil, "", errs.ErrInvalidToken
	}

	tokenID, ok := claims["tokenId"].(string)
	if !ok {
		return nil, "", errs.ErrInvalidToken
	}

	if _, err = s.tokenService.GetTokenByID(ctx, tokenID); err != nil {
		return nil, "", err
	}

	user, err := s.userService.GetByID(ctx, int64(userID))
	if err != nil {
		return nil, "", err
	}

	return user, tokenID, nil
}

func (s *Service) validateRefreshToken(ctx context.Context, refreshToken string) (*models.UserFullModel, string, error) {
	claims, err := s.tokenService.ValidateToken(refreshToken)
	if err != nil {
		return nil, "", errs.ErrInvalidRefreshToken
	}

	tokenID, ok := claims["tokenId"].(string)
	if !ok {
		return nil, "", errs.ErrInvalidTokenID
	}

	userID, ok := claims["userId"].(float64)
	if !ok {
		return nil, "", errs.ErrInvalidUserID
	}

	tokenByID, err := s.tokenService.GetTokenByID(ctx, tokenID)
	if err != nil && !errors.Is(err, errs.ErrTokenNotFound) {
		return nil, "", fmt.Errorf("AuthService.validateRefreshToken: %w", err)
	}

	if err != nil || tokenByID == nil || tokenByID.RefreshToken != refreshToken {
		return nil, "", errs.ErrInvalidRefreshToken
	}

	user, err := s.userService.GetByID(ctx, int64(userID))
	if err != nil {
		return nil, "", errs.ErrUserNotFound
	}

	return user, tokenID, nil
}
