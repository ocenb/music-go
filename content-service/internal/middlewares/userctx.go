package middlewares

import (
	"context"
	"errors"

	"github.com/ocenb/music-protos/gen/userservice"
)

type userCtxKey struct{}

func ContextWithUser(ctx context.Context, user *userservice.UserPrivateModel) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func UserFromContext(ctx context.Context) (*userservice.UserPrivateModel, error) {
	user, ok := ctx.Value(userCtxKey{}).(*userservice.UserPrivateModel)
	if !ok || user == nil {
		return nil, errors.New("failed to get user from context")
	}
	return user, nil
}
