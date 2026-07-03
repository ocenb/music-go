package user

import (
	"context"
	"fmt"
	"time"

	"github.com/ocenb/music-protos/gen/searchservice"
	"github.com/ocenb/music-protos/gen/userservice"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/models"
	"github.com/ocenb/music-go/user-service/internal/storage/transactor"
)

type Repo interface {
	GetByUsername(ctx context.Context, username string) (*userservice.UserPublicModel, error)
	GetByID(ctx context.Context, id int64) (*models.UserFullModel, error)
	GetByEmail(ctx context.Context, email string) (*models.UserFullModel, error)
	GetByVerificationToken(ctx context.Context, verificationToken string) (*models.UserFullModel, error)
	UpdateVerificationToken(ctx context.Context, userID int64, newVerificationToken string, expiresAt time.Time) (*userservice.UserPrivateModel, error)
	SetVerified(ctx context.Context, userID int64) (*userservice.UserPrivateModel, error)
	Create(ctx context.Context, username, email, password, verificationToken string, verificationTokenExpiresAt time.Time) (*userservice.UserPrivateModel, error)
	ChangeUsername(ctx context.Context, userID int64, username string) (*userservice.UserPublicModel, error)
	ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, error)
	ChangePassword(ctx context.Context, userID int64, password string) (*userservice.UserPrivateModel, error)
	Delete(ctx context.Context, userID int64) error
	CheckFollow(ctx context.Context, userID int64, targetUserID int64) (bool, error)
	Follow(ctx context.Context, userID int64, targetUserID int64) error
	Unfollow(ctx context.Context, userID int64, targetUserID int64) error
}

type SearchClient interface {
	AddUser(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error)
	UpdateUser(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error)
	DeleteUser(ctx context.Context, req *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error)
}

type ContentClient interface {
	DeleteUserContent(ctx context.Context, userID int64) error
}

type Service struct {
	repo          Repo
	tm            *transactor.Manager
	searchClient  SearchClient
	contentClient ContentClient
}

func New(
	repo Repo,
	tm *transactor.Manager,
	searchClient SearchClient,
	contentClient ContentClient,
) *Service {
	return &Service{
		repo:          repo,
		tm:            tm,
		searchClient:  searchClient,
		contentClient: contentClient,
	}
}

func (s *Service) GetByUsername(ctx context.Context, username string) (*userservice.UserPublicModel, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("UserService.GetByUsername: %w", err)
	}
	return user, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*models.UserFullModel, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("UserService.GetByID: %w", err)
	}
	return user, nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*models.UserFullModel, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("UserService.GetByEmail: %w", err)
	}
	return user, nil
}

func (s *Service) GetByVerificationToken(ctx context.Context, verificationToken string) (*models.UserFullModel, error) {
	user, err := s.repo.GetByVerificationToken(ctx, verificationToken)
	if err != nil {
		return nil, fmt.Errorf("UserService.GetByVerificationToken: %w", err)
	}
	return user, nil
}

func (s *Service) UpdateVerificationToken(ctx context.Context, userID int64, newVerificationToken string) (*userservice.UserPrivateModel, error) {
	expiresAt := time.Now().Add(24 * time.Hour)
	user, err := s.repo.UpdateVerificationToken(ctx, userID, newVerificationToken, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("UserService.UpdateVerificationToken: %w", err)
	}
	return user, nil
}

func (s *Service) SetVerified(ctx context.Context, userID int64) (*userservice.UserPrivateModel, error) {
	user, err := s.repo.SetVerified(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("UserService.SetVerified: %w", err)
	}
	return user, nil
}

func (s *Service) Create(ctx context.Context, username, email, password, verificationToken string) (*userservice.UserPrivateModel, error) {
	verificationTokenExpiresAt := time.Now().Add(24 * time.Hour)

	user, err := s.repo.Create(ctx, username, email, password, verificationToken, verificationTokenExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("UserService.Create: %w", err)
	}

	addResp, err := s.searchClient.AddUser(ctx, &searchservice.AddOrUpdateRequest{
		Id:   user.Id,
		Name: username,
	})
	if err != nil || !addResp.Success {
		return nil, fmt.Errorf("UserService.Create: %w", err)
	}

	return user, nil
}

func (s *Service) ChangeUsername(ctx context.Context, userID int64, username string) (*userservice.UserPublicModel, error) {
	existingUser, err := s.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, errs.ErrUserUsernameExists
	}

	var updatedUser *userservice.UserPublicModel
	err = s.tm.Run(ctx, func(txCtx context.Context) error {
		var err error
		updatedUser, err = s.repo.ChangeUsername(txCtx, userID, username)
		if err != nil {
			return fmt.Errorf("failed to change username: %w", err)
		}

		updateResp, err := s.searchClient.UpdateUser(txCtx, &searchservice.AddOrUpdateRequest{
			Id:   userID,
			Name: username,
		})
		if err != nil || !updateResp.Success {
			return fmt.Errorf("failed to update user in search service: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("UserService.ChangeUsername: %w", err)
	}

	return updatedUser, nil
}

func (s *Service) ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, error) {
	existingUser, err := s.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errs.ErrUserEmailExists
	}

	updatedUser, err := s.repo.ChangeEmail(ctx, userID, email)
	if err != nil {
		return nil, fmt.Errorf("UserService.ChangeEmail: %w", err)
	}

	return updatedUser, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, password string) (*userservice.UserPrivateModel, error) {
	updatedUser, err := s.repo.ChangePassword(ctx, userID, password)
	if err != nil {
		return nil, fmt.Errorf("UserService.ChangePassword: %w", err)
	}

	return updatedUser, nil
}

func (s *Service) Delete(ctx context.Context, userID int64) error {
	if _, err := s.GetByID(ctx, userID); err != nil {
		return err
	}

	if err := s.contentClient.DeleteUserContent(ctx, userID); err != nil {
		return fmt.Errorf("UserService.Delete: %w", err)
	}

	err := s.tm.Run(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, userID); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		deleteResp, err := s.searchClient.DeleteUser(txCtx, &searchservice.DeleteRequest{
			Id: userID,
		})
		if err != nil || !deleteResp.Success {
			return fmt.Errorf("failed to delete user in search service: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("UserService.Delete: %w", err)
	}

	return nil
}

func (s *Service) CheckFollow(ctx context.Context, userID int64, targetUserID int64) (bool, error) {
	exists, err := s.repo.CheckFollow(ctx, userID, targetUserID)
	if err != nil {
		return false, fmt.Errorf("UserService.CheckFollow: %w", err)
	}
	return exists, nil
}

func (s *Service) Follow(ctx context.Context, userID int64, targetUserID int64) error {
	exists, err := s.repo.CheckFollow(ctx, userID, targetUserID)
	if err != nil {
		return fmt.Errorf("UserService.Follow: %w", err)
	}
	if exists {
		return errs.ErrUserAlreadyFollowed
	}
	if err := s.repo.Follow(ctx, userID, targetUserID); err != nil {
		return fmt.Errorf("UserService.Follow: %w", err)
	}
	return nil
}

func (s *Service) Unfollow(ctx context.Context, userID int64, targetUserID int64) error {
	exists, err := s.repo.CheckFollow(ctx, userID, targetUserID)
	if err != nil {
		return fmt.Errorf("UserService.Unfollow: %w", err)
	}
	if !exists {
		return errs.ErrUserNotFollowed
	}
	if err := s.repo.Unfollow(ctx, userID, targetUserID); err != nil {
		return fmt.Errorf("UserService.Unfollow: %w", err)
	}
	return nil
}
