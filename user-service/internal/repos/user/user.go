package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ocenb/music-protos/gen/userservice"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/models"
	"github.com/ocenb/music-go/user-service/internal/storage/transactor"
)

type Repo struct {
	tm transactor.Querier
}

func New(tm transactor.Querier) *Repo {
	return &Repo{tm: tm}
}

func (r *Repo) GetByUsername(ctx context.Context, username string) (*userservice.UserPublicModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT u.id, u.username, u.followers_count
		FROM users u
		WHERE u.username = $1 AND u.is_verified = TRUE
	`

	var user userservice.UserPublicModel
	err := q.QueryRow(ctx, query, username).Scan(&user.Id, &user.Username, &user.FollowersCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (r *Repo) GetByID(ctx context.Context, id int64) (*models.UserFullModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT id, username, email, password, is_verified, verification_token,
		       verification_token_expires_at::text, created_at::text, followers_count
		FROM users
		WHERE id = $1
	`

	return scanUserFullModel(q.QueryRow(ctx, query, id))
}

func (r *Repo) GetByEmail(ctx context.Context, email string) (*models.UserFullModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT id, username, email, password, is_verified, verification_token,
		       verification_token_expires_at::text, created_at::text, followers_count
		FROM users
		WHERE email = $1
	`

	return scanUserFullModel(q.QueryRow(ctx, query, email))
}

func (r *Repo) GetByVerificationToken(ctx context.Context, verificationToken string) (*models.UserFullModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		SELECT id, username, email, password, is_verified, verification_token,
		       verification_token_expires_at::text, created_at::text, followers_count
		FROM users
		WHERE verification_token = $1
	`

	return scanUserFullModel(q.QueryRow(ctx, query, verificationToken))
}

func (r *Repo) UpdateVerificationToken(
	ctx context.Context,
	userID int64,
	newVerificationToken string,
	expiresAt time.Time,
) (*userservice.UserPrivateModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE users
		SET verification_token = $2, verification_token_expires_at = $3
		WHERE id = $1
		RETURNING id, username, email, created_at::text
	`

	var user userservice.UserPrivateModel
	err := q.QueryRow(ctx, query, userID, newVerificationToken, expiresAt).Scan(&user.Id, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update verification token: %w", err)
	}

	return &user, nil
}

func (r *Repo) SetVerified(ctx context.Context, userID int64) (*userservice.UserPrivateModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE users
		SET is_verified = TRUE, verification_token = NULL, verification_token_expires_at = NULL
		WHERE id = $1
		RETURNING id, username, email, created_at::text
	`

	var user userservice.UserPrivateModel
	err := q.QueryRow(ctx, query, userID).Scan(&user.Id, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to set user as verified: %w", err)
	}

	return &user, nil
}

func (r *Repo) Create(
	ctx context.Context,
	username, email, password, verificationToken string,
	verificationTokenExpiresAt time.Time,
) (*userservice.UserPrivateModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		INSERT INTO users (username, email, password, verification_token, verification_token_expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, created_at::text
	`

	var user userservice.UserPrivateModel
	err := q.QueryRow(ctx, query, username, email, password, verificationToken, verificationTokenExpiresAt).
		Scan(&user.Id, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return nil, errs.ErrUserUsernameExists
			case "users_email_key":
				return nil, errs.ErrUserEmailExists
			case "users_verification_token_key":
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (r *Repo) ChangeUsername(ctx context.Context, userID int64, username string) (*userservice.UserPublicModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE users
		SET username = $2
		WHERE id = $1
		RETURNING id, username, followers_count
	`

	var user userservice.UserPublicModel
	err := q.QueryRow(ctx, query, userID, username).Scan(&user.Id, &user.Username, &user.FollowersCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return nil, errs.ErrUserUsernameExists
		}
		return nil, fmt.Errorf("failed to change username: %w", err)
	}

	return &user, nil
}

func (r *Repo) ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE users
		SET email = $2
		WHERE id = $1
		RETURNING id, username, email, created_at::text
	`

	var user userservice.UserPrivateModel
	err := q.QueryRow(ctx, query, userID, email).Scan(&user.Id, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return nil, errs.ErrUserEmailExists
		}
		return nil, fmt.Errorf("failed to change email: %w", err)
	}

	return &user, nil
}

func (r *Repo) ChangePassword(ctx context.Context, userID int64, password string) (*userservice.UserPrivateModel, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `
		UPDATE users
		SET password = $2
		WHERE id = $1
		RETURNING id, username, email, created_at::text
	`

	var user userservice.UserPrivateModel
	err := q.QueryRow(ctx, query, userID, password).Scan(&user.Id, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to change password: %w", err)
	}

	return &user, nil
}

func (r *Repo) Delete(ctx context.Context, userID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `DELETE FROM users WHERE id = $1`
	cmdTag, err := q.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errs.ErrUserNotFound
	}

	return nil
}

func (r *Repo) CheckFollow(ctx context.Context, userID int64, targetUserID int64) (bool, error) {
	q := r.tm.GetQueryEngine(ctx)

	query := `SELECT EXISTS(SELECT 1 FROM user_followers WHERE user_id = $1 AND follower_id = $2)`

	var exists bool
	if err := q.QueryRow(ctx, query, targetUserID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check follow: %w", err)
	}

	return exists, nil
}

func (r *Repo) Follow(ctx context.Context, userID int64, targetUserID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `INSERT INTO user_followers (user_id, follower_id) VALUES ($1, $2)`
	if _, err := q.Exec(ctx, query, targetUserID, userID); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return errs.ErrUserAlreadyFollowed
		}
		return fmt.Errorf("failed to follow user: %w", err)
	}

	return nil
}

func (r *Repo) Unfollow(ctx context.Context, userID int64, targetUserID int64) error {
	q := r.tm.GetQueryEngine(ctx)

	query := `DELETE FROM user_followers WHERE user_id = $1 AND follower_id = $2`
	cmdTag, err := q.Exec(ctx, query, targetUserID, userID)
	if err != nil {
		return fmt.Errorf("failed to unfollow user: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errs.ErrUserNotFollowed
	}

	return nil
}

func scanUserFullModel(row pgx.Row) (*models.UserFullModel, error) {
	var user models.UserFullModel
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.IsVerified,
		&user.VerificationToken,
		&user.VerificationTokenExpiresAt,
		&user.CreatedAt,
		&user.FollowersCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
