package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/ocenb/music-go/user-service/internal/clients/contentservice"
	searchclient "github.com/ocenb/music-go/user-service/internal/clients/searchservice"
	"github.com/ocenb/music-go/user-service/internal/config"
	"github.com/ocenb/music-go/user-service/internal/models"
	userrepo "github.com/ocenb/music-go/user-service/internal/repos/user"
	"github.com/ocenb/music-go/user-service/internal/storage"
	"github.com/ocenb/music-go/user-service/internal/utils"
	"github.com/ocenb/music-protos/gen/searchservice"
	"github.com/ocenb/music-protos/gen/userservice"
)

type UserServiceInterface interface {
	GetByUsername(ctx context.Context, username string) (*userservice.UserPublicModel, error)
	GetById(ctx context.Context, id int64) (*models.UserFullModel, error)
	GetByEmail(ctx context.Context, email string) (*models.UserFullModel, error)
	GetByVerificationToken(ctx context.Context, verificationToken string) (*models.UserFullModel, error)
	UpdateVerificationToken(ctx context.Context, userID int64, newVerificationToken string) (*userservice.UserPrivateModel, error)
	SetVerified(ctx context.Context, userID int64) (*userservice.UserPrivateModel, error)
	Create(ctx context.Context, username, email, password, verificationToken string) (*userservice.UserPrivateModel, error)
	ChangeUsername(ctx context.Context, userID int64, email string) (*userservice.UserPublicModel, error)
	ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, error)
	ChangePassword(ctx context.Context, userID int64, password string) (*userservice.UserPrivateModel, error)
	Delete(ctx context.Context, userID int64) error
	CheckFollow(ctx context.Context, userID int64, targetUserID int64) (bool, error)
	Follow(ctx context.Context, userID int64, targetUserID int64) error
	Unfollow(ctx context.Context, userID int64, targetUserID int64) error
}

type UserService struct {
	cfg                  *config.Config
	userRepo             userrepo.UserRepoInterface
	log                  *slog.Logger
	searchServiceClient  *searchclient.SearchServiceClient
	contentServiceClient *contentservice.ContentServiceClient
}

func NewUserService(cfg *config.Config, log *slog.Logger, userRepo userrepo.UserRepoInterface, searchServiceClient *searchclient.SearchServiceClient, contentServiceClient *contentservice.ContentServiceClient) UserServiceInterface {
	return &UserService{
		cfg:                  cfg,
		userRepo:             userRepo,
		log:                  log,
		searchServiceClient:  searchServiceClient,
		contentServiceClient: contentServiceClient,
	}
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*userservice.UserPublicModel, error) {
	s.log.Debug("Getting user by username", slog.String("username", username))
	user, err := s.userRepo.GetByUsername(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		s.log.Debug("User not found by username", slog.String("username", username))
		return nil, ErrUserNotFound
	}
	if err != nil {
		s.log.Error("Failed to get user by username", slog.String("username", username), utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to get user by username")
	}
	s.log.Debug("User found by username", slog.String("username", username), slog.Int64("user_id", user.Id))
	return user, nil
}

func (s *UserService) GetById(ctx context.Context, id int64) (*models.UserFullModel, error) {
	s.log.Debug("Getting user by id", slog.Int64("user_id", id))
	user, err := s.userRepo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Debug("User not found by id", slog.Int64("user_id", id))
			return nil, ErrUserNotFound
		}
		s.log.Error("Failed to get user by id", slog.Int64("user_id", id), utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to get user by id")
	}
	s.log.Debug("User found by id", slog.Int64("user_id", id))
	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.UserFullModel, error) {
	s.log.Debug("Getting user by email", slog.String("email", email))
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Debug("User not found by email", slog.String("email", email))
			return nil, ErrUserNotFound
		}
		s.log.Error("Failed to get user by email", slog.String("email", email), utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to get user by email")
	}
	s.log.Debug("User found by email", slog.String("email", email), slog.Int64("user_id", user.ID))
	return user, nil
}

func (s *UserService) GetByVerificationToken(ctx context.Context, verificationToken string) (*models.UserFullModel, error) {
	user, err := s.userRepo.GetByVerificationToken(ctx, verificationToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, utils.InternalError(err, "failed to get user by verification token")
	}
	return user, nil
}

func (s *UserService) UpdateVerificationToken(ctx context.Context, userID int64, newVerificationToken string) (*userservice.UserPrivateModel, error) {
	expiresAt := time.Now().Add(24 * time.Hour)
	user, err := s.userRepo.UpdateVerificationToken(ctx, userID, newVerificationToken, expiresAt)
	if err != nil {
		return nil, utils.InternalError(err, "failed to update verification token")
	}
	return user, nil
}

func (s *UserService) SetVerified(ctx context.Context, userID int64) (*userservice.UserPrivateModel, error) {
	user, err := s.userRepo.SetVerified(ctx, userID)
	if err != nil {
		return nil, utils.InternalError(err, "failed to set user as verified")
	}
	return user, nil
}

func (s *UserService) Create(ctx context.Context, username, email, password, verificationToken string) (*userservice.UserPrivateModel, error) {
	s.log.Info("Creating new user", slog.String("username", username), slog.String("email", email))
	verificationTokenExpiresAt := time.Now().Add(24 * time.Hour)

	user, err := s.userRepo.Create(ctx, username, email, password, verificationToken, verificationTokenExpiresAt)
	if err != nil {
		s.log.Error("Failed to create user", slog.String("username", username), slog.String("email", email), utils.ErrLog(err))
		return nil, utils.InternalError(err, "failed to create user")
	}

	s.log.Info("User created successfully", slog.String("username", username), slog.String("email", email), slog.Int64("user_id", user.Id))

	addResp, err := s.searchServiceClient.Client.AddUser(ctx, &searchservice.AddOrUpdateRequest{
		Id:   user.Id,
		Name: username,
	})
	if err != nil || !addResp.Success {
		return nil, utils.InternalError(err, "failed to add user to search service")
	}

	return user, nil
}

func (s *UserService) ChangeUsername(ctx context.Context, userID int64, username string) (*userservice.UserPublicModel, error) {
	existingUser, err := s.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, ErrUserUsernameExists
	}

	var updatedUser *userservice.UserPublicModel
	err = storage.WithTransaction(ctx, s.userRepo, func(txCtx context.Context) error {
		var err error
		updatedUser, err = s.userRepo.ChangeUsername(txCtx, userID, username)
		if err != nil {
			return utils.InternalError(err, "failed to change username")
		}

		updateResp, err := s.searchServiceClient.Client.UpdateUser(ctx, &searchservice.AddOrUpdateRequest{
			Id:   userID,
			Name: username,
		})
		if err != nil || !updateResp.Success {
			return utils.InternalError(err, "failed to update user in search service")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *UserService) ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, error) {
	existingUser, err := s.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrUserEmailExists
	}

	updatedUser, err := s.userRepo.ChangeEmail(ctx, userID, email)
	if err != nil {
		return nil, utils.InternalError(err, "failed to change email")
	}

	return updatedUser, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID int64, password string) (*userservice.UserPrivateModel, error) {
	updatedUser, err := s.userRepo.ChangePassword(ctx, userID, password)
	if err != nil {
		return nil, utils.InternalError(err, "failed to change password")
	}

	return updatedUser, nil
}

func (s *UserService) Delete(ctx context.Context, userID int64) error {
	s.log.Info("Deleting user", slog.Int64("user_id", userID))
	if _, err := s.GetById(ctx, userID); err != nil {
		s.log.Error("User not found for deletion", slog.Int64("user_id", userID))
		return ErrUserNotFound
	}

	if err := s.contentServiceClient.DeleteUserContent(ctx, userID); err != nil {
		s.log.Error("Failed to delete user content", slog.Int64("user_id", userID), utils.ErrLog(err))
		return utils.InternalError(err, "failed to delete user content")
	}

	err := storage.WithTransaction(ctx, s.userRepo, func(txCtx context.Context) error {
		if err := s.userRepo.Delete(txCtx, userID); err != nil {
			s.log.Error("Failed to delete user in transaction", slog.Int64("user_id", userID), utils.ErrLog(err))
			return utils.InternalError(err, "failed to delete user")
		}

		deleteResp, err := s.searchServiceClient.Client.DeleteUser(ctx, &searchservice.DeleteRequest{
			Id: userID,
		})
		if err != nil || !deleteResp.Success {
			return utils.InternalError(err, "failed to delete user in search service")
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.log.Info("User deleted successfully", slog.Int64("user_id", userID))
	return nil
}

func (s *UserService) CheckFollow(ctx context.Context, userID int64, targetUserID int64) (bool, error) {
	exists, err := s.userRepo.CheckFollow(ctx, userID, targetUserID)
	if err != nil {
		return false, utils.InternalError(err, "failed to check if user is followed")
	}
	return exists, nil
}

func (s *UserService) Follow(ctx context.Context, userID int64, targetUserID int64) error {
	exists, err := s.userRepo.CheckFollow(ctx, userID, targetUserID)
	if err != nil {
		return utils.InternalError(err, "failed to check if user is followed")
	}
	if exists {
		return utils.AlreadyExistsError("user already followed")
	}
	err = s.userRepo.Follow(ctx, userID, targetUserID)
	if err != nil {
		return utils.InternalError(err, "failed to follow user")
	}
	return nil
}

func (s *UserService) Unfollow(ctx context.Context, userID int64, targetUserID int64) error {
	exists, err := s.userRepo.CheckFollow(ctx, userID, targetUserID)
	if err != nil {
		return utils.InternalError(err, "failed to check if user is followed")
	}
	if !exists {
		return utils.NotFoundError("user not followed")
	}
	err = s.userRepo.Unfollow(ctx, userID, targetUserID)
	if err != nil {
		return utils.InternalError(err, "failed to unfollow user")
	}
	return nil
}
