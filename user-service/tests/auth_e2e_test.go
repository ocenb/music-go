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

func TestAuthE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := setupTestEnv(ctx, t)
	client := env.Client

	t.Run("Register", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			username, email, password := fakeRegisterRequest()

			resp, err := client.Register(ctx, &userservice.RegisterRequest{
				Username: username,
				Email:    email,
				Password: password,
			})
			require.NoError(t, err)
			require.NotNil(t, resp.User)
			require.Equal(t, username, resp.User.Username)
			require.Equal(t, email, resp.User.Email)
		})

		t.Run("Duplicate", func(t *testing.T) {
			username, email, password := fakeRegisterRequest()

			_, err := client.Register(ctx, &userservice.RegisterRequest{
				Username: username,
				Email:    email,
				Password: password,
			})
			require.NoError(t, err)

			_, err = client.Register(ctx, &userservice.RegisterRequest{
				Username: username,
				Email:    email,
				Password: password,
			})
			require.Error(t, err)
			require.Equal(t, codes.AlreadyExists, status.Code(err))
		})
	})

	t.Run("Login", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			resp := login(ctx, t, client, adminEmail, adminPassword)
			require.Equal(t, adminUsername, resp.User.Username)
		})

		t.Run("WrongPassword", func(t *testing.T) {
			_, err := client.Login(ctx, &userservice.LoginRequest{
				Email:    adminEmail,
				Password: adminPassword + "!",
			})
			require.Error(t, err)
			require.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	})

	t.Run("Logout", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			resp := login(ctx, t, client, adminEmail, adminPassword)
			logout(ctx, t, client, resp.AccessToken)
		})
	})

	t.Run("Refresh", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			loginResp := login(ctx, t, client, adminEmail, adminPassword)

			refreshResp, err := client.Refresh(ctx, &userservice.RefreshRequest{
				RefreshToken: loginResp.RefreshToken,
			})
			require.NoError(t, err)
			require.Equal(t, adminUsername, refreshResp.User.Username)
			require.NotEmpty(t, refreshResp.AccessToken)
			require.NotEmpty(t, refreshResp.RefreshToken)
			require.NotEqual(t, loginResp.AccessToken, refreshResp.AccessToken)
			require.NotEqual(t, loginResp.RefreshToken, refreshResp.RefreshToken)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			_, err := client.Refresh(ctx, &userservice.RefreshRequest{
				RefreshToken: fakeRefreshToken(),
			})
			require.Error(t, err)
			require.Equal(t, codes.Unauthenticated, status.Code(err))
		})
	})

	t.Run("CheckAuth", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			loginResp := login(ctx, t, client, adminEmail, adminPassword)

			resp, err := client.CheckAuth(authContext(ctx, loginResp.AccessToken), &emptypb.Empty{})
			require.NoError(t, err)
			require.Equal(t, adminUsername, resp.User.Username)
			require.NotEmpty(t, resp.TokenId)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			_, err := client.CheckAuth(authContext(ctx, fakeAccessToken()), &emptypb.Empty{})
			require.Error(t, err)
			require.Equal(t, codes.Unauthenticated, status.Code(err))
		})
	})
}
