package handlers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/ocenb/music-go/user-service/internal/utils"
	"github.com/ocenb/music-protos/gen/userservice"
)

func (s *UserServer) Register(ctx context.Context, req *userservice.RegisterRequest) (*userservice.RegisterResponse, error) {
	user, err := s.authService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	res := &userservice.RegisterResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}
	return res, nil
}

func (s *UserServer) Login(ctx context.Context, req *userservice.LoginRequest) (*userservice.LoginResponse, error) {
	user, accessToken, refreshToken, err := s.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	res := &userservice.LoginResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return res, nil
}

func (s *UserServer) Logout(ctx context.Context, req *emptypb.Empty) (*userservice.LogoutResponse, error) {
	_, tokenId, err := utils.GetInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.authService.Logout(ctx, tokenId)
	if err != nil {
		return nil, err
	}

	res := &userservice.LogoutResponse{Success: true}
	return res, nil
}

func (s *UserServer) LogoutAll(ctx context.Context, req *emptypb.Empty) (*userservice.LogoutAllResponse, error) {
	user, _, err := utils.GetInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.authService.LogoutAll(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	res := &userservice.LogoutAllResponse{Success: true}
	return res, nil
}

func (s *UserServer) Refresh(ctx context.Context, req *userservice.RefreshRequest) (*userservice.RefreshResponse, error) {
	user, accessToken, refreshToken, err := s.authService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	res := &userservice.RefreshResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return res, nil
}

func (s *UserServer) Verify(ctx context.Context, req *userservice.VerifyRequest) (*userservice.VerifyResponse, error) {
	user, accessToken, refreshToken, err := s.authService.Verify(ctx, req.VerifyToken)
	if err != nil {
		return nil, err
	}

	res := &userservice.VerifyResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return res, nil
}

func (s *UserServer) NewVerification(ctx context.Context, req *userservice.NewVerificationRequest) (*userservice.NewVerificationResponse, error) {
	user, err := s.authService.NewVerification(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	res := &userservice.NewVerificationResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}
	return res, nil
}

func (s *UserServer) ChangeEmail(ctx context.Context, req *userservice.ChangeEmailRequest) (*userservice.ChangeEmailResponse, error) {
	user, _, err := utils.GetInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	updatedUser, accessToken, refreshToken, err := s.authService.ChangeEmail(ctx, user.ID, req.Email)
	if err != nil {
		return nil, err
	}

	res := &userservice.ChangeEmailResponse{
		User: &userservice.UserPrivateModel{
			Id:        updatedUser.Id,
			Username:  updatedUser.Username,
			Email:     updatedUser.Email,
			CreatedAt: updatedUser.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return res, nil
}

func (s *UserServer) ChangePassword(ctx context.Context, req *userservice.ChangePasswordRequest) (*userservice.ChangePasswordResponse, error) {
	user, _, err := utils.GetInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	updatedUser, accessToken, refreshToken, err := s.authService.ChangePassword(ctx, user.ID, user.Password, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}

	res := &userservice.ChangePasswordResponse{
		User: &userservice.UserPrivateModel{
			Id:        updatedUser.Id,
			Username:  updatedUser.Username,
			Email:     updatedUser.Email,
			CreatedAt: updatedUser.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return res, nil
}

func (s *UserServer) CheckAuth(ctx context.Context, req *emptypb.Empty) (*userservice.CheckAuthResponse, error) {
	user, tokenId, err := utils.GetInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	res := &userservice.CheckAuthResponse{
		User: &userservice.UserPrivateModel{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		TokenId: tokenId,
	}
	return res, nil
}
