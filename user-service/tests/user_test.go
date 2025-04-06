package tests

import (
	"testing"

	"github.com/ocenb/music-go/user-service/tests/suite"
	"github.com/ocenb/music-protos/gen/userservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGetUserByUsername(t *testing.T) {
	ctx, s := suite.New(t)

	loginReq := &userservice.LoginRequest{
		Email:    suite.AdminEmail,
		Password: suite.AdminPassword,
	}

	loginResp, err := s.UserClient.Login(ctx, loginReq)
	require.NoError(t, err)

	md := metadata.New(map[string]string{"authorization": "Bearer " + loginResp.AccessToken})
	authCtx := metadata.NewOutgoingContext(ctx, md)

	req := &userservice.GetUserByUsernameRequest{
		Username: suite.AdminUsername,
	}

	resp, err := s.UserClient.GetUserByUsername(authCtx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.User)
	assert.Equal(t, suite.AdminUsername, resp.User.Username)

	nonexistentUsername := "nonexistent-" + suite.ValidUsername()
	if len(nonexistentUsername) > 20 {
		nonexistentUsername = nonexistentUsername[:20]
	}

	nonexistentReq := &userservice.GetUserByUsernameRequest{
		Username: nonexistentUsername,
	}

	_, err = s.UserClient.GetUserByUsername(authCtx, nonexistentReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestChangeUsername(t *testing.T) {
	ctx, s := suite.New(t)

	loginReq := &userservice.LoginRequest{
		Email:    suite.ToChangeEmail,
		Password: suite.ToChangePassword,
	}

	loginResp, err := s.UserClient.Login(ctx, loginReq)
	require.NoError(t, err)

	md := metadata.New(map[string]string{"authorization": "Bearer " + loginResp.AccessToken})
	authCtx := metadata.NewOutgoingContext(ctx, md)

	newUsername := suite.ValidUsername()

	changeReq := &userservice.ChangeUsernameRequest{
		Username: newUsername,
	}

	changeResp, err := s.UserClient.ChangeUsername(authCtx, changeReq)
	require.NoError(t, err)
	require.NotNil(t, changeResp)
	require.NotNil(t, changeResp.User)
	assert.Equal(t, newUsername, changeResp.User.Username)

	anotherUsername, anotherEmail, anotherPassword := suite.FakeRegisterRequest()
	anotherRegisterReq := &userservice.RegisterRequest{
		Username: anotherUsername,
		Email:    anotherEmail,
		Password: anotherPassword,
	}

	_, err = s.UserClient.Register(ctx, anotherRegisterReq)
	require.NoError(t, err)

	conflictReq := &userservice.ChangeUsernameRequest{
		Username: anotherUsername,
	}

	_, err = s.UserClient.ChangeUsername(authCtx, conflictReq)
	require.Error(t, err)
}

func TestDeleteUser(t *testing.T) {
	ctx, s := suite.New(t)

	loginReq := &userservice.LoginRequest{
		Email:    suite.ToDeleteEmail,
		Password: suite.ToDeletePassword,
	}

	loginResp, err := s.UserClient.Login(ctx, loginReq)
	require.NoError(t, err)

	md := metadata.New(map[string]string{"authorization": "Bearer " + loginResp.AccessToken})
	authCtx := metadata.NewOutgoingContext(ctx, md)

	resp, err := s.UserClient.DeleteUser(authCtx, &emptypb.Empty{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestFollowUnfollow(t *testing.T) {
	ctx, s := suite.New(t)

	login1Req := &userservice.LoginRequest{
		Email:    suite.AdminEmail,
		Password: suite.AdminPassword,
	}
	login1Resp, err := s.UserClient.Login(ctx, login1Req)
	require.NoError(t, err)

	login2Req := &userservice.LoginRequest{
		Email:    suite.ToFollowEmail,
		Password: suite.ToFollowPassword,
	}
	login2Resp, err := s.UserClient.Login(ctx, login2Req)
	require.NoError(t, err)

	md1 := metadata.New(map[string]string{"authorization": "Bearer " + login1Resp.AccessToken})
	authCtx1 := metadata.NewOutgoingContext(ctx, md1)

	checkReq := &userservice.CheckFollowRequest{
		UserId: login2Resp.User.Id,
	}
	checkResp, err := s.UserClient.CheckFollow(authCtx1, checkReq)
	require.NoError(t, err)
	require.NotNil(t, checkResp)
	assert.False(t, checkResp.IsFollowed)

	followReq := &userservice.FollowRequest{
		UserId: login2Resp.User.Id,
	}
	followResp, err := s.UserClient.Follow(authCtx1, followReq)
	require.NoError(t, err)
	require.NotNil(t, followResp)
	assert.True(t, followResp.Success)

	checkRespAfterFollow, err := s.UserClient.CheckFollow(authCtx1, checkReq)
	require.NoError(t, err)
	require.NotNil(t, checkRespAfterFollow)
	assert.True(t, checkRespAfterFollow.IsFollowed)

	unfollowReq := &userservice.UnfollowRequest{
		UserId: login2Resp.User.Id,
	}
	unfollowResp, err := s.UserClient.Unfollow(authCtx1, unfollowReq)
	require.NoError(t, err)
	require.NotNil(t, unfollowResp)
	assert.True(t, unfollowResp.Success)

	checkRespAfterUnfollow, err := s.UserClient.CheckFollow(authCtx1, checkReq)
	require.NoError(t, err)
	require.NotNil(t, checkRespAfterUnfollow)
	assert.False(t, checkRespAfterUnfollow.IsFollowed)
}
