package handlers

import (
	"context"
	"log/slog"

	"github.com/ocenb/music-protos/gen/userservice"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/ocenb/music-go/user-service/internal/logger"
	"github.com/ocenb/music-go/user-service/internal/middlewares"
)

func (h *UserServer) Register(ctx context.Context, req *userservice.RegisterRequest) (*userservice.RegisterResponse, error) {
	log := logger.FromContext(ctx).With(slog.String("username", req.Username))
	ctx = logger.IntoContext(ctx, log)

	user, err := h.authService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.RegisterResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

func (h *UserServer) Login(ctx context.Context, req *userservice.LoginRequest) (*userservice.LoginResponse, error) {
	user, accessToken, refreshToken, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	return &userservice.LoginResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserServer) Logout(ctx context.Context, _ *emptypb.Empty) (*userservice.LogoutResponse, error) {
	user, tokenID, err := middlewares.AuthFromContext(ctx)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	log := logger.FromContext(ctx).With(slog.Int64("user_id", user.ID))
	ctx = logger.IntoContext(ctx, log)

	if err := h.authService.Logout(ctx, tokenID); err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.LogoutResponse{Success: true}, nil
}

func (h *UserServer) LogoutAll(ctx context.Context, _ *emptypb.Empty) (*userservice.LogoutAllResponse, error) {
	user, _, err := middlewares.AuthFromContext(ctx)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	log := logger.FromContext(ctx).With(slog.Int64("user_id", user.ID))
	ctx = logger.IntoContext(ctx, log)

	if err := h.authService.LogoutAll(ctx, user.ID); err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.LogoutAllResponse{Success: true}, nil
}

func (h *UserServer) Refresh(ctx context.Context, req *userservice.RefreshRequest) (*userservice.RefreshResponse, error) {
	user, accessToken, refreshToken, err := h.authService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	return &userservice.RefreshResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserServer) Verify(ctx context.Context, req *userservice.VerifyRequest) (*userservice.VerifyResponse, error) {
	user, accessToken, refreshToken, err := h.authService.Verify(ctx, req.VerifyToken)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	return &userservice.VerifyResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserServer) NewVerification(ctx context.Context, req *userservice.NewVerificationRequest) (*userservice.NewVerificationResponse, error) {
	user, err := h.authService.NewVerification(ctx, req.Email, req.Password)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	return &userservice.NewVerificationResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

func (h *UserServer) ChangeEmail(ctx context.Context, req *userservice.ChangeEmailRequest) (*userservice.ChangeEmailResponse, error) {
	user, _, err := middlewares.AuthFromContext(ctx)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	log := logger.FromContext(ctx).With(slog.Int64("user_id", user.ID))
	ctx = logger.IntoContext(ctx, log)

	updatedUser, accessToken, refreshToken, err := h.authService.ChangeEmail(ctx, user.ID, req.Email)
	if err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.ChangeEmailResponse{
		User: &userservice.UserPrivateModel{
			Id:        updatedUser.Id,
			Username:  updatedUser.Username,
			Email:     updatedUser.Email,
			CreatedAt: updatedUser.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserServer) ChangePassword(ctx context.Context, req *userservice.ChangePasswordRequest) (*userservice.ChangePasswordResponse, error) {
	user, _, err := middlewares.AuthFromContext(ctx)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	log := logger.FromContext(ctx).With(slog.Int64("user_id", user.ID))
	ctx = logger.IntoContext(ctx, log)

	updatedUser, accessToken, refreshToken, err := h.authService.ChangePassword(ctx, user.ID, user.Password, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.ChangePasswordResponse{
		User: &userservice.UserPrivateModel{
			Id:        updatedUser.Id,
			Username:  updatedUser.Username,
			Email:     updatedUser.Email,
			CreatedAt: updatedUser.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserServer) CheckAuth(ctx context.Context, _ *emptypb.Empty) (*userservice.CheckAuthResponse, error) {
	user, tokenID, err := middlewares.AuthFromContext(ctx)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	return &userservice.CheckAuthResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		TokenId: tokenID,
	}, nil
}
