package usermocks

import (
	"context"

	"github.com/ocenb/music-go/user-service/internal/models"
	"github.com/ocenb/music-protos/gen/userservice"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetById(ctx context.Context, id int64) (*models.UserFullModel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserFullModel), args.Error(1)
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (*userservice.UserPublicModel, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userservice.UserPublicModel), args.Error(1)
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*models.UserFullModel, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserFullModel), args.Error(1)
}

func (m *MockUserService) GetByVerificationToken(ctx context.Context, token string) (*models.UserFullModel, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserFullModel), args.Error(1)
}

func (m *MockUserService) UpdateVerificationToken(ctx context.Context, userID int64, token string) (*userservice.UserPrivateModel, error) {
	args := m.Called(ctx, userID, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userservice.UserPrivateModel), args.Error(1)
}

func (m *MockUserService) SetVerified(ctx context.Context, userID int64) (*userservice.UserPrivateModel, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userservice.UserPrivateModel), args.Error(1)
}

func (m *MockUserService) Create(ctx context.Context, username, email, password, token string) (*userservice.UserPrivateModel, error) {
	args := m.Called(ctx, username, email, password, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userservice.UserPrivateModel), args.Error(1)
}

func (m *MockUserService) ChangeUsername(ctx context.Context, userID int64, username string) (*userservice.UserPublicModel, error) {
	args := m.Called(ctx, userID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userservice.UserPublicModel), args.Error(1)
}

func (m *MockUserService) ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, error) {
	args := m.Called(ctx, userID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userservice.UserPrivateModel), args.Error(1)
}

func (m *MockUserService) ChangePassword(ctx context.Context, userID int64, password string) (*userservice.UserPrivateModel, error) {
	args := m.Called(ctx, userID, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userservice.UserPrivateModel), args.Error(1)
}

func (m *MockUserService) Delete(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
