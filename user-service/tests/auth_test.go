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

func TestRegister(t *testing.T) {
	ctx, s := suite.New(t)

	username, email, password := suite.FakeRegisterRequest()
	req := &userservice.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	resp, err := s.UserClient.Register(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.User)
	assert.Equal(t, username, resp.User.Username)
	assert.Equal(t, email, resp.User.Email)

	duplicateReq := &userservice.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	_, err = s.UserClient.Register(ctx, duplicateReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestLogin(t *testing.T) {
	ctx, s := suite.New(t)

	loginReq := &userservice.LoginRequest{
		Email:    suite.AdminEmail,
		Password: suite.AdminPassword,
	}

	loginResp, err := s.UserClient.Login(ctx, loginReq)
	require.NoError(t, err)
	require.NotNil(t, loginResp)
	require.NotNil(t, loginResp.User)
	assert.Equal(t, suite.AdminUsername, loginResp.User.Username)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.RefreshToken)

	wrongPassword := suite.AdminPassword + "!"
	if len(wrongPassword) > 50 {
		wrongPassword = suite.AdminPassword[:len(suite.AdminPassword)-1] + "!"
	}

	wrongLoginReq := &userservice.LoginRequest{
		Email:    suite.AdminEmail,
		Password: wrongPassword,
	}

	_, err = s.UserClient.Login(ctx, wrongLoginReq)
	require.Error(t, err)
}

func TestLogout(t *testing.T) {
	ctx, s := suite.New(t)

	loginReq := &userservice.LoginRequest{
		Email:    suite.AdminEmail,
		Password: suite.AdminPassword,
	}

	loginResp, err := s.UserClient.Login(ctx, loginReq)
	require.NoError(t, err)

	md := metadata.New(map[string]string{"authorization": "Bearer " + loginResp.AccessToken})
	authCtx := metadata.NewOutgoingContext(ctx, md)

	logoutResp, err := s.UserClient.Logout(authCtx, &emptypb.Empty{})
	require.NoError(t, err)
	require.NotNil(t, logoutResp)
	assert.True(t, logoutResp.Success)
}

func TestRefresh(t *testing.T) {
	ctx, s := suite.New(t)

	loginReq := &userservice.LoginRequest{
		Email:    suite.AdminEmail,
		Password: suite.AdminPassword,
	}

	loginResp, err := s.UserClient.Login(ctx, loginReq)
	require.NoError(t, err)

	refreshReq := &userservice.RefreshRequest{
		RefreshToken: loginResp.RefreshToken,
	}

	refreshResp, err := s.UserClient.Refresh(ctx, refreshReq)
	require.NoError(t, err)
	require.NotNil(t, refreshResp)
	require.NotNil(t, refreshResp.User)
	assert.Equal(t, suite.AdminUsername, refreshResp.User.Username)
	assert.NotEmpty(t, refreshResp.AccessToken)
	assert.NotEmpty(t, refreshResp.RefreshToken)
	assert.NotEqual(t, loginResp.AccessToken, refreshResp.AccessToken)
	assert.NotEqual(t, loginResp.RefreshToken, refreshResp.RefreshToken)

	refreshReq = &userservice.RefreshRequest{
		RefreshToken: suite.FakeRefreshToken(),
	}

	_, err = s.UserClient.Refresh(ctx, refreshReq)
	require.Error(t, err)
}
