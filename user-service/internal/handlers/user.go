package handlers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/ocenb/music-go/user-service/internal/utils"
	"github.com/ocenb/music-protos/gen/userservice"
)

func (s *UserServer) GetUserByUsername(ctx context.Context, req *userservice.GetUserByUsernameRequest) (*userservice.GetUserByUsernameResponse, error) {
	user, err := s.userService.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	res := &userservice.GetUserByUsernameResponse{
		User: &userservice.UserPublicModel{
			Id:             user.Id,
			Username:       user.Username,
			FollowersCount: user.FollowersCount,
		},
	}
	return res, nil
}

func (s *UserServer) ChangeUsername(ctx context.Context, req *userservice.ChangeUsernameRequest) (*userservice.ChangeUsernameResponse, error) {
	user, _, err := utils.GetInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	updatedUser, err := s.userService.ChangeUsername(ctx, user.ID, req.Username)
	if err != nil {
		return nil, err
	}

	res := &userservice.ChangeUsernameResponse{
		User: &userservice.UserPublicModel{
			Id:       updatedUser.Id,
			Username: updatedUser.Username,
		},
	}
	return res, nil
}

func (s *UserServer) DeleteUser(ctx context.Context, req *emptypb.Empty) (*userservice.DeleteUserResponse, error) {
	user, _, err := utils.GetInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.userService.Delete(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	res := &userservice.DeleteUserResponse{Success: true}
	return res, nil
}
