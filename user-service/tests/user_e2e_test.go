package tests_test

import (
	"context"
	"testing"
	"time"

	"github.com/ocenb/music-protos/gen/userservice"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestUserE2E(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := setupTestEnv(ctx, t)
	client := env.Client

	t.Run("GetUserByUsername", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			loginResp := login(ctx, t, client, adminEmail)

			resp, err := client.GetUserByUsername(authContext(ctx, loginResp.AccessToken), &userservice.GetUserByUsernameRequest{
				Username: adminUsername,
			})
			require.NoError(t, err)
			require.Equal(t, adminUsername, resp.User.Username)
		})

		t.Run("NotFound", func(t *testing.T) {
			loginResp := login(ctx, t, client, adminEmail)

			username := validUsername()
			_, err := client.GetUserByUsername(authContext(ctx, loginResp.AccessToken), &userservice.GetUserByUsernameRequest{
				Username: username,
			})
			require.Error(t, err)
			require.Equal(t, codes.NotFound, status.Code(err))
		})
	})

	t.Run("ChangeUsername", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			loginResp := login(ctx, t, client, changeNameEmail)
			newUsername := validUsername()

			changeResp, err := client.ChangeUsername(authContext(ctx, loginResp.AccessToken), &userservice.ChangeUsernameRequest{
				Username: newUsername,
			})
			require.NoError(t, err)
			require.Equal(t, newUsername, changeResp.User.Username)
		})

		t.Run("Conflict", func(t *testing.T) {
			loginResp := login(ctx, t, client, toChangeEmail)

			_, err := client.ChangeUsername(authContext(ctx, loginResp.AccessToken), &userservice.ChangeUsernameRequest{
				Username: adminUsername,
			})
			require.Error(t, err)
			require.Equal(t, codes.AlreadyExists, status.Code(err))
		})
	})

	t.Run("DeleteUser", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			loginResp := login(ctx, t, client, toDeleteEmail)

			resp, err := client.DeleteUser(authContext(ctx, loginResp.AccessToken), &emptypb.Empty{})
			require.NoError(t, err)
			require.True(t, resp.Success)
		})
	})

	t.Run("Follow", func(t *testing.T) {
		t.Run("FollowUnfollow", func(t *testing.T) {
			login1 := login(ctx, t, client, adminEmail)
			login2 := login(ctx, t, client, toFollowEmail)
			authCtx1 := authContext(ctx, login1.AccessToken)

			checkResp, err := client.CheckFollow(authCtx1, &userservice.CheckFollowRequest{
				UserId: login2.User.Id,
			})
			require.NoError(t, err)
			require.False(t, checkResp.IsFollowed)

			followResp, err := client.Follow(authCtx1, &userservice.FollowRequest{
				UserId: login2.User.Id,
			})
			require.NoError(t, err)
			require.True(t, followResp.Success)

			checkResp, err = client.CheckFollow(authCtx1, &userservice.CheckFollowRequest{
				UserId: login2.User.Id,
			})
			require.NoError(t, err)
			require.True(t, checkResp.IsFollowed)

			unfollowResp, err := client.Unfollow(authCtx1, &userservice.UnfollowRequest{
				UserId: login2.User.Id,
			})
			require.NoError(t, err)
			require.True(t, unfollowResp.Success)

			checkResp, err = client.CheckFollow(authCtx1, &userservice.CheckFollowRequest{
				UserId: login2.User.Id,
			})
			require.NoError(t, err)
			require.False(t, checkResp.IsFollowed)
		})
	})
}
