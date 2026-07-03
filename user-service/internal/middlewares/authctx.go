package middlewares

import (
	"context"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/models"
)

type authCtxKey struct {
	name string
}

var (
	userCtxKey    = &authCtxKey{name: "user"}
	tokenIDCtxKey = &authCtxKey{name: "token_id"}
)

func ContextWithAuth(ctx context.Context, user *models.UserFullModel, tokenID string) context.Context {
	ctx = context.WithValue(ctx, userCtxKey, user)
	return context.WithValue(ctx, tokenIDCtxKey, tokenID)
}

func AuthFromContext(ctx context.Context) (*models.UserFullModel, string, error) {
	user, ok := ctx.Value(userCtxKey).(*models.UserFullModel)
	if !ok || user == nil {
		return nil, "", errs.Unauthenticated("user is not in context")
	}

	tokenID, ok := ctx.Value(tokenIDCtxKey).(string)
	if !ok || tokenID == "" {
		return nil, "", errs.Unauthenticated("token id is not in context")
	}

	return user, tokenID, nil
}
