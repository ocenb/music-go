package handlers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ocenb/music-protos/gen/userservice"
	"google.golang.org/grpc"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/logger"
	"github.com/ocenb/music-go/user-service/internal/logger/logattr"
	"github.com/ocenb/music-go/user-service/internal/models"
)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*userservice.UserPrivateModel, error)
	Login(ctx context.Context, email, password string) (*userservice.UserPrivateModel, string, string, error)
	Logout(ctx context.Context, tokenID string) error
	LogoutAll(ctx context.Context, userID int64) error
	Refresh(ctx context.Context, oldRefreshToken string) (*userservice.UserPrivateModel, string, string, error)
	Verify(ctx context.Context, verifyToken string) (*userservice.UserPrivateModel, string, string, error)
	NewVerification(ctx context.Context, email, password string) (*userservice.UserPrivateModel, error)
	ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, string, string, error)
	ChangePassword(ctx context.Context, userID int64, truePassword, oldPassword, newPassword string) (*userservice.UserPrivateModel, string, string, error)
	ValidateAccessToken(ctx context.Context, accessToken string) (*models.UserFullModel, string, error)
}

type UserService interface {
	GetByUsername(ctx context.Context, username string) (*userservice.UserPublicModel, error)
	ChangeUsername(ctx context.Context, userID int64, username string) (*userservice.UserPublicModel, error)
	Delete(ctx context.Context, userID int64) error
	CheckFollow(ctx context.Context, userID, targetUserID int64) (bool, error)
	Follow(ctx context.Context, userID, targetUserID int64) error
	Unfollow(ctx context.Context, userID, targetUserID int64) error
}

type UserServer struct {
	userservice.UnimplementedUserServiceServer `exhaustruct:"optional"`
	authService                                AuthService
	userService                                UserService
}

func NewUserServer(gRPCServer *grpc.Server, authService AuthService, userService UserService) {
	userservice.RegisterUserServiceServer(gRPCServer, &UserServer{authService: authService, userService: userService})
}

func handleError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	log := logger.FromContext(ctx)
	if logErr, ok := errors.AsType[*LogWrappedError](err); ok {
		log = logErr.Logger
		err = logErr.Err
	}

	domainErr, ok := errs.As(err)
	if ok && domainErr.Kind() == errs.KindInternal {
		log.ErrorContext(ctx, "internal server error", logattr.Err(err))
	}

	return errs.ToGRPC(err)
}

type LogWrappedError struct {
	Err    error
	Logger *slog.Logger
}

func (e *LogWrappedError) Error() string { return e.Err.Error() }
func (e *LogWrappedError) Unwrap() error { return e.Err }
