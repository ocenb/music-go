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
