package handlers

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ocenb/music-protos/gen/userservice"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/middlewares"
	"github.com/ocenb/music-go/content-service/internal/models"
)

type getOneByChangeableIDForm struct {
	Username     string `form:"username" binding:"required"`
	ChangeableID string `form:"changeableId" binding:"required"`
}

func bindQuery(c *gin.Context, dest any) bool {
	if err := c.ShouldBindQuery(dest); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return false
	}
	return true
}

func bindURI(c *gin.Context, dest any) bool {
	if err := c.ShouldBindUri(dest); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return false
	}
	return true
}

func bindForm(c *gin.Context, dest any) bool {
	if err := c.ShouldBind(dest); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return false
	}
	return true
}

func bindJSON(c *gin.Context, dest any) bool {
	if err := c.ShouldBindJSON(dest); err != nil {
		handleError(c, errs.InvalidArgument(err.Error()))
		return false
	}
	return true
}

func authenticatedUser(c *gin.Context) (*userservice.UserPrivateModel, bool) {
	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Internal(err, "failed to get user from context"))
		return nil, false
	}
	return user, true
}

func requiredUser(c *gin.Context) (*userservice.UserPrivateModel, bool) {
	user, err := middlewares.UserFromContext(c.Request.Context())
	if err != nil {
		handleError(c, errs.Unauthenticated(err.Error()))
		return nil, false
	}
	return user, true
}

func respondGetOneByChangeableID[T any](
	c *gin.Context,
	notFoundErr error,
	fetch func(ctx context.Context, userID int64, username, changeableID string) (T, error),
) {
	var params getOneByChangeableIDForm
	if !bindQuery(c, &params) {
		return
	}

	user, ok := authenticatedUser(c)
	if !ok {
		return
	}

	result, err := fetch(c.Request.Context(), user.Id, params.Username, params.ChangeableID)
	if err != nil {
		if errors.Is(err, notFoundErr) {
			handleError(c, errs.NotFound(err.Error()))
			return
		}
		handleError(c, errs.Internal(err, ""))
		return
	}

	c.JSON(http.StatusOK, result)
}

func handleChangeTitle[TURI any, TForm any](
	c *gin.Context,
	getEntityID func(TURI) int64,
	getTitle func(TForm) string,
	notFoundErr, existsErr error,
	mutate func(ctx context.Context, userID, entityID int64, title string) error,
) {
	var uriParams TURI
	var formParams TForm
	if !bindURI(c, &uriParams) || !bindForm(c, &formParams) {
		return
	}

	user, ok := authenticatedUser(c)
	if !ok {
		return
	}

	err := mutate(c.Request.Context(), user.Id, getEntityID(uriParams), getTitle(formParams))
	if err != nil {
		switch {
		case errors.Is(err, notFoundErr):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		case errors.Is(err, existsErr):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func handleChangeChangeableID[TURI any, TForm any](
	c *gin.Context,
	getEntityID func(TURI) int64,
	getChangeableID func(TForm) string,
	notFoundErr error,
	mutate func(ctx context.Context, userID, entityID int64, changeableID string) error,
) {
	var uriParams TURI
	var formParams TForm
	if !bindURI(c, &uriParams) || !bindForm(c, &formParams) {
		return
	}

	user, ok := authenticatedUser(c)
	if !ok {
		return
	}

	err := mutate(c.Request.Context(), user.Id, getEntityID(uriParams), getChangeableID(formParams))
	if err != nil {
		switch {
		case errors.Is(err, notFoundErr):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		case errors.Is(err, errs.ErrChangeableIDExists):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func handleChangeImage[TURI any, TForm any](
	c *gin.Context,
	getEntityID func(TURI) int64,
	getImage func(TForm) *multipart.FileHeader,
	notFoundErr error,
	extraInvalidArgErrs []error,
	mutate func(ctx context.Context, userID, entityID int64, image *multipart.FileHeader) error,
) {
	var uriParams TURI
	var formParams TForm
	if !bindURI(c, &uriParams) || !bindForm(c, &formParams) {
		return
	}

	user, ok := authenticatedUser(c)
	if !ok {
		return
	}

	err := mutate(c.Request.Context(), user.Id, getEntityID(uriParams), getImage(formParams))
	if err != nil {
		switch {
		case errors.Is(err, notFoundErr):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		case errors.Is(err, errs.ErrInvalidImageFormat):
			handleError(c, errs.InvalidArgument(err.Error()))
		default:
			for _, invalidErr := range extraInvalidArgErrs {
				if errors.Is(err, invalidErr) {
					handleError(c, errs.InvalidArgument(err.Error()))
					return
				}
			}
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func handleDeleteByID[TURI any](
	c *gin.Context,
	getEntityID func(TURI) int64,
	notFoundErr error,
	deleteFn func(ctx context.Context, userID, entityID int64) error,
) {
	var uriParams TURI
	if !bindURI(c, &uriParams) {
		return
	}

	user, ok := authenticatedUser(c)
	if !ok {
		return
	}

	if err := deleteFn(c.Request.Context(), user.Id, getEntityID(uriParams)); err != nil {
		switch {
		case errors.Is(err, notFoundErr):
			handleError(c, errs.NotFound(err.Error()))
		case errors.Is(err, errs.ErrPermissionDenied):
			handleError(c, errs.PermissionDenied(err.Error()))
		default:
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func handlePlaylistIDAction(
	c *gin.Context,
	notFoundErr error,
	invalidArgErrs []error,
	action func(ctx context.Context, userID, playlistID int64) error,
) {
	var params models.GetByPlaylistIDURI
	if !bindURI(c, &params) {
		return
	}

	user, ok := authenticatedUser(c)
	if !ok {
		return
	}

	if err := action(c.Request.Context(), user.Id, params.PlaylistID); err != nil {
		switch {
		case errors.Is(err, notFoundErr):
			handleError(c, errs.NotFound(err.Error()))
		default:
			for _, invalidErr := range invalidArgErrs {
				if errors.Is(err, invalidErr) {
					handleError(c, errs.InvalidArgument(err.Error()))
					return
				}
			}
			handleError(c, errs.Internal(err, ""))
		}
		return
	}

	c.Status(http.StatusNoContent)
}
