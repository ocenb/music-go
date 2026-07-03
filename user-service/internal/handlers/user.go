package handlers

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/ocenb/music-protos/gen/userservice"

	"github.com/ocenb/music-go/user-service/internal/logger"
	"github.com/ocenb/music-go/user-service/internal/middlewares"
)

func (h *UserServer) GetUserByUsername(ctx context.Context, req *userservice.GetUserByUsernameRequest) (*userservice.GetUserByUsernameResponse, error) {
	log := logger.FromContext(ctx).With(slog.String("username", req.Username))
	ctx = logger.IntoContext(ctx, log)

	user, err := h.userService.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.GetUserByUsernameResponse{
		User: &userservice.UserPublicModel{
			Id:             user.Id,
			Username:       user.Username,
			FollowersCount: user.FollowersCount,
		},
	}, nil
}

func (h *UserServer) ChangeUsername(ctx context.Context, req *userservice.ChangeUsernameRequest) (*userservice.ChangeUsernameResponse, error) {
	user, _, err := middlewares.AuthFromContext(ctx)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	log := logger.FromContext(ctx).With(
		slog.Int64("user_id", user.ID),
		slog.String("username", req.Username),
	)
	ctx = logger.IntoContext(ctx, log)

	updatedUser, err := h.userService.ChangeUsername(ctx, user.ID, req.Username)
	if err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.ChangeUsernameResponse{
		User: &userservice.UserPublicModel{
			Id:       updatedUser.Id,
			Username: updatedUser.Username,
		},
	}, nil
}

func (h *UserServer) DeleteUser(ctx context.Context, _ *emptypb.Empty) (*userservice.DeleteUserResponse, error) {
	user, _, err := middlewares.AuthFromContext(ctx)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	log := logger.FromContext(ctx).With(slog.Int64("user_id", user.ID))
	ctx = logger.IntoContext(ctx, log)

	if err := h.userService.Delete(ctx, user.ID); err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.DeleteUserResponse{Success: true}, nil
}

func (h *UserServer) CheckFollow(ctx context.Context, req *userservice.CheckFollowRequest) (*userservice.CheckFollowResponse, error) {
	user, _, err := middlewares.AuthFromContext(ctx)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	log := logger.FromContext(ctx).With(
		slog.Int64("user_id", user.ID),
		slog.Int64("target_user_id", req.UserId),
	)
	ctx = logger.IntoContext(ctx, log)

	exists, err := h.userService.CheckFollow(ctx, user.ID, req.UserId)
	if err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.CheckFollowResponse{IsFollowed: exists}, nil
}

func (h *UserServer) Follow(ctx context.Context, req *userservice.FollowRequest) (*userservice.FollowResponse, error) {
	user, _, err := middlewares.AuthFromContext(ctx)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	log := logger.FromContext(ctx).With(
		slog.Int64("user_id", user.ID),
		slog.Int64("target_user_id", req.UserId),
	)
	ctx = logger.IntoContext(ctx, log)

	if err := h.userService.Follow(ctx, user.ID, req.UserId); err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.FollowResponse{Success: true}, nil
}

func (h *UserServer) Unfollow(ctx context.Context, req *userservice.UnfollowRequest) (*userservice.UnfollowResponse, error) {
	user, _, err := middlewares.AuthFromContext(ctx)
	if err != nil {
		return nil, handleError(ctx, err)
	}

	log := logger.FromContext(ctx).With(
		slog.Int64("user_id", user.ID),
		slog.Int64("target_user_id", req.UserId),
	)
	ctx = logger.IntoContext(ctx, log)

	if err := h.userService.Unfollow(ctx, user.ID, req.UserId); err != nil {
		return nil, handleError(ctx, &LogWrappedError{Err: err, Logger: log})
	}

	return &userservice.UnfollowResponse{Success: true}, nil
}
