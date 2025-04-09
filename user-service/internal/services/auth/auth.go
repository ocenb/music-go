package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/ocenb/music-go/user-service/internal/clients/notificationclient"
	"github.com/ocenb/music-go/user-service/internal/config"
	"github.com/ocenb/music-go/user-service/internal/models"
	authrepo "github.com/ocenb/music-go/user-service/internal/repos/auth"
	"github.com/ocenb/music-go/user-service/internal/services/token"
	"github.com/ocenb/music-go/user-service/internal/services/user"
	"github.com/ocenb/music-go/user-service/internal/storage"
	"github.com/ocenb/music-go/user-service/internal/utils"
	"github.com/ocenb/music-protos/gen/userservice"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, username, email, password string) (*userservice.UserPrivateModel, error)
	Login(ctx context.Context, email, password string) (*userservice.UserPrivateModel, string, string, error)
	Logout(ctx context.Context, tokenId string) error
	LogoutAll(ctx context.Context, userID int64) error
	Refresh(ctx context.Context, oldRefreshToken string) (*userservice.UserPrivateModel, string, string, error)
	Verify(ctx context.Context, verifyToken string) (*userservice.UserPrivateModel, string, string, error)
	NewVerification(ctx context.Context, email, password string) (*userservice.UserPrivateModel, error)
	ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, string, string, error)
	ChangePassword(ctx context.Context, userID int64, truePassword, oldPassword, newPassword string) (*userservice.UserPrivateModel, string, string, error)
	ValidateAccessToken(ctx context.Context, accessToken string) (*models.UserFullModel, string, error)
	validateRefreshToken(ctx context.Context, refreshToken string) (*models.UserFullModel, string, error)
}

type AuthService struct {
	cfg                *config.Config
	userService        user.UserServiceInterface
	tokenService       token.TokenServiceInterface
	authRepo           authrepo.AuthRepoInterface
	notificationClient notificationclient.NotificationClientInterface
	log                *slog.Logger
}

func NewAuthService(cfg *config.Config, log *slog.Logger, userService user.UserServiceInterface, tokenService token.TokenServiceInterface, authRepo authrepo.AuthRepoInterface, notificationClient notificationclient.NotificationClientInterface) AuthServiceInterface {
	return &AuthService{
		cfg:                cfg,
		userService:        userService,
		tokenService:       tokenService,
		authRepo:           authRepo,
		notificationClient: notificationClient,
		log:                log,
	}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (*userservice.UserPrivateModel, error) {
	s.log.Info("Registering new user", slog.String("username", username), slog.String("email", email))

	userByEmail, err := s.userService.GetByEmail(ctx, email)
	if err == nil && userByEmail != nil {
		s.log.Info("Registration failed: email already exists", slog.String("email", email))
		return nil, ErrUserEmailExists
	}
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		s.log.Error("Registration failed: error checking email existence", slog.String("email", email), utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to get user by email")
	}

	userByName, err := s.userService.GetByUsername(ctx, username)
	if err == nil && userByName != nil {
		s.log.Info("Registration failed: username already exists", slog.String("username", username))
		return nil, ErrUserUsernameExists
	}
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		s.log.Error("Registration failed: error checking username existence", slog.String("username", username), utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to get user by username")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BCryptCost)
	if err != nil {
		s.log.Error("Registration failed: error hashing password", utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to hash password")
	}

	verificationToken := s.tokenService.GenerateVerificationToken()

	user, err := s.userService.Create(ctx, username, email, string(hashedPassword), verificationToken)
	if err != nil {
		s.log.Error("Registration failed: error creating user", slog.String("username", username), slog.String("email", email), utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to create user")
	}

	s.log.Info("User registered successfully", slog.String("username", username), slog.String("email", email), slog.Int64("user_id", user.Id))

	err = s.notificationClient.SendEmailNotification(user.Email, verificationToken)
	if err != nil {
		s.log.Error("Registration failed: error sending verification email", slog.String("email", email), utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to send verification email")
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*userservice.UserPrivateModel, string, string, error) {
	s.log.Info("User login attempt", slog.String("email", email))

	user, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		s.log.Info("Login failed: email not found", slog.String("email", email))
		return nil, "", "", ErrUserEmailNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		s.log.Info("Login failed: invalid password", slog.String("email", email), slog.Int64("user_id", user.ID))
		return nil, "", "", ErrInvalidPassword
	}

	if !user.IsVerified {
		s.log.Info("Login failed: user not verified", slog.String("email", email), slog.Int64("user_id", user.ID))
		return nil, "", "", ErrUserNotVerified
	}

	accessToken, refreshToken, err := s.tokenService.CreateTokens(ctx, user.ID)
	if err != nil {
		s.log.Error("Login failed: error creating tokens", slog.String("email", email), slog.Int64("user_id", user.ID), utils.ErrLog(err))
		return nil, "", "", utils.InternalError(err, "failed to create tokens")
	}

	s.log.Info("User logged in successfully", slog.String("email", email), slog.Int64("user_id", user.ID))
	return &userservice.UserPrivateModel{Id: user.ID, Username: user.Username, Email: user.Email, CreatedAt: user.CreatedAt}, accessToken, refreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, tokenID string) error {
	err := s.tokenService.RevokeToken(ctx, tokenID)
	if err != nil {
		return utils.InternalError(err, "failed to revoke token")
	}
	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID int64) error {
	err := s.tokenService.RevokeAllTokens(ctx, userID)
	if err != nil {
		return utils.InternalError(err, "failed to revoke all tokens")
	}
	return nil
}

func (s *AuthService) Refresh(ctx context.Context, oldRefreshToken string) (*userservice.UserPrivateModel, string, string, error) {
	user, tokenId, err := s.validateRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		return nil, "", "", err
	}

	var accessToken, refreshToken string

	err = storage.WithTransaction(ctx, s.authRepo, func(txCtx context.Context) error {
		err := s.tokenService.RevokeToken(txCtx, tokenId)
		if err != nil {
			return utils.InternalError(err, "failed to revoke old refresh token")
		}

		var tokenErr error
		accessToken, refreshToken, tokenErr = s.tokenService.CreateTokens(txCtx, user.ID)
		if tokenErr != nil {
			return utils.InternalError(tokenErr, "failed to create new tokens")
		}

		return nil
	})

	if err != nil {
		return nil, "", "", err
	}

	return &userservice.UserPrivateModel{Id: user.ID, Username: user.Username, Email: user.Email, CreatedAt: user.CreatedAt}, accessToken, refreshToken, nil
}

func (s *AuthService) Verify(ctx context.Context, verificationToken string) (*userservice.UserPrivateModel, string, string, error) {
	user, err := s.userService.GetByVerificationToken(ctx, verificationToken)
	if err != nil {
		return nil, "", "", ErrTokenNotFound
	}

	expDate, err := time.Parse("2006-01-02", *user.VerificationTokenExpiresAt)
	if err != nil {
		return nil, "", "", utils.InternalError(err, "failed to parse expiration date")
	}
	if expDate.Before(time.Now()) {
		return nil, "", "", ErrTokenExpired
	}

	var verifiedUser *userservice.UserPrivateModel
	var accessToken, refreshToken string

	err = storage.WithTransaction(ctx, s.authRepo, func(txCtx context.Context) error {
		var err error

		verifiedUser, err = s.userService.SetVerified(txCtx, user.ID)
		if err != nil {
			return utils.InternalError(err, "failed to set user as verified")
		}

		accessToken, refreshToken, err = s.tokenService.CreateTokens(txCtx, verifiedUser.Id)
		if err != nil {
			return utils.InternalError(err, "failed to create tokens")
		}

		return nil
	})

	if err != nil {
		return nil, "", "", err
	}

	return verifiedUser, accessToken, refreshToken, nil
}

func (s *AuthService) NewVerification(ctx context.Context, email, password string) (*userservice.UserPrivateModel, error) {
	user, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserEmailNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidPassword
	}

	if user.IsVerified {
		return nil, ErrUserAlreadyVerified
	}

	newVerificationToken := s.tokenService.GenerateVerificationToken()
	updatedUser, err := s.userService.UpdateVerificationToken(ctx, user.ID, newVerificationToken)
	if err != nil {
		return nil, utils.InternalError(err, "failed to update verification token")
	}

	err = s.notificationClient.SendEmailNotification(user.Email, newVerificationToken)
	if err != nil {
		s.log.Error("New verification failed: error sending verification email", slog.String("email", user.Email), utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to send verification email")
	}

	return updatedUser, nil
}

func (s *AuthService) ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, string, string, error) {
	var user *userservice.UserPrivateModel
	var newAccessToken, newRefreshToken string

	err := storage.WithTransaction(ctx, s.authRepo, func(txCtx context.Context) error {
		var err error

		user, err = s.userService.ChangeEmail(txCtx, userID, email)
		if err != nil {
			return utils.InternalError(err, "failed to change email")
		}

		err = s.tokenService.RevokeAllTokens(txCtx, userID)
		if err != nil {
			return utils.InternalError(err, "failed to revoke all tokens")
		}

		newAccessToken, newRefreshToken, err = s.tokenService.CreateTokens(txCtx, userID)
		if err != nil {
			return utils.InternalError(err, "failed to create new tokens")
		}

		return nil
	})

	if err != nil {
		return nil, "", "", err
	}

	return user, newAccessToken, newRefreshToken, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, truePassword, oldPassword, newPassword string) (*userservice.UserPrivateModel, string, string, error) {
	err := bcrypt.CompareHashAndPassword([]byte(truePassword), []byte(oldPassword))
	if err != nil {
		return nil, "", "", ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.BCryptCost)
	if err != nil {
		return nil, "", "", utils.InternalError(err, "failed to hash password")
	}

	var user *userservice.UserPrivateModel
	var newAccessToken, newRefreshToken string

	err = storage.WithTransaction(ctx, s.authRepo, func(txCtx context.Context) error {
		var err error

		user, err = s.userService.ChangePassword(txCtx, userID, string(hashedPassword))
		if err != nil {
			return utils.InternalError(err, "failed to change password")
		}

		err = s.tokenService.RevokeAllTokens(txCtx, userID)
		if err != nil {
			return utils.InternalError(err, "failed to revoke all tokens")
		}

		newAccessToken, newRefreshToken, err = s.tokenService.CreateTokens(txCtx, userID)
		if err != nil {
			return utils.InternalError(err, "failed to create new tokens")
		}

		return nil
	})

	if err != nil {
		return nil, "", "", err
	}

	return user, newAccessToken, newRefreshToken, nil
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, accessToken string) (*models.UserFullModel, string, error) {
	s.log.Debug("Validating access token")

	claims, err := s.tokenService.ValidateToken(accessToken)
	if err != nil {
		s.log.Info("Token validation failed: invalid token", utils.ErrLog(err))
		return nil, "", err
	}

	userID, ok := claims["userId"].(float64)
	if !ok {
		s.log.Error("Token validation failed: userId not found in token")
		return nil, "", ErrInvalidToken
	}

	tokenID, ok := claims["tokenId"].(string)
	if !ok {
		s.log.Error("Token validation failed: tokenId not found in token")
		return nil, "", ErrInvalidToken
	}

	_, err = s.tokenService.GetTokenByID(ctx, tokenID)
	if err != nil {
		s.log.Info("Token validation failed: token not found or expired", slog.String("token_id", tokenID))
		return nil, "", err
	}

	user, err := s.userService.GetById(ctx, int64(userID))
	if err != nil {
		s.log.Error("Token validation failed: user not found", slog.Int64("user_id", int64(userID)))
		return nil, "", err
	}

	s.log.Debug("Access token validated successfully", slog.Int64("user_id", int64(userID)), slog.String("token_id", tokenID))
	return user, tokenID, nil
}

func (s *AuthService) validateRefreshToken(ctx context.Context, refreshToken string) (*models.UserFullModel, string, error) {
	claims, err := s.tokenService.ValidateToken(refreshToken)
	if err != nil {
		return nil, "", ErrInvalidRefreshToken
	}

	tokenId, ok := claims["tokenId"].(string)
	if !ok {
		return nil, "", ErrInvalidTokenID
	}

	userID, ok := claims["userId"].(float64)
	if !ok {
		return nil, "", ErrInvalidUserID
	}

	tokenById, err := s.tokenService.GetTokenByID(ctx, tokenId)
	if err != nil && !errors.Is(err, token.ErrTokenNotFound) {
		return nil, "", utils.InternalError(err, "failed to get token by id")
	}

	if tokenById.RefreshToken != refreshToken {
		return nil, "", ErrInvalidRefreshToken
	}

	user, err := s.userService.GetById(ctx, int64(userID))
	if err != nil {
		return nil, "", ErrUserNotFound
	}

	return user, tokenId, nil
}
