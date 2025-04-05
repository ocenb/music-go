package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/ocenb/music-go/user-service/internal/models"
	"github.com/ocenb/music-go/user-service/internal/utils"
	"github.com/ocenb/music-protos/gen/userservice"

	_ "github.com/lib/pq"
)

type UserRepoInterface interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	GetByUsername(ctx context.Context, username string) (*userservice.UserPublicModel, error)
	GetById(ctx context.Context, id int64) (*models.UserFullModel, error)
	GetByEmail(ctx context.Context, email string) (*models.UserFullModel, error)
	GetByVerificationToken(ctx context.Context, verificationToken string) (*models.UserFullModel, error)
	UpdateVerificationToken(ctx context.Context, userID int64, newVerificationToken string, expiresAt time.Time) (*userservice.UserPrivateModel, error)
	SetVerified(ctx context.Context, userID int64) (*userservice.UserPrivateModel, error)
	Create(ctx context.Context, username, email, password, verificationToken string, verificationTokenExpiresAt time.Time) (*userservice.UserPrivateModel, error)
	ChangeUsername(ctx context.Context, userID int64, username string) (*userservice.UserPublicModel, error)
	ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, error)
	ChangePassword(ctx context.Context, userID int64, password string) (*userservice.UserPrivateModel, error)
	Delete(ctx context.Context, userID int64) error
}

type UserRepo struct {
	postgres *sql.DB
}

func NewUserRepo(postgres *sql.DB) UserRepoInterface {
	return &UserRepo{postgres: postgres}
}

func (r *UserRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.postgres.BeginTx(ctx, opts)
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*userservice.UserPublicModel, error) {
	query := `
		SELECT u.id, u.username, u.followers_count
		FROM users u
		WHERE u.username = $1 AND u.is_verified = TRUE
	`
	row := r.postgres.QueryRowContext(ctx, query, username)

	var user userservice.UserPublicModel
	err := row.Scan(&user.Id, &user.Username, &user.FollowersCount)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetById(ctx context.Context, id int64) (*models.UserFullModel, error) {
	query := `SELECT * FROM users WHERE id = $1`
	row := r.postgres.QueryRowContext(ctx, query, id)

	var user models.UserFullModel
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsVerified, &user.VerificationToken, &user.VerificationTokenExpiresAt, &user.CreatedAt, &user.FollowersCount)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.UserFullModel, error) {
	query := `SELECT * FROM users WHERE email = $1`
	row := r.postgres.QueryRowContext(ctx, query, email)

	var user models.UserFullModel
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsVerified, &user.VerificationToken, &user.VerificationTokenExpiresAt, &user.CreatedAt, &user.FollowersCount)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetByVerificationToken(ctx context.Context, verificationToken string) (*models.UserFullModel, error) {
	query := `SELECT * FROM users WHERE verification_token = $1`
	row := r.postgres.QueryRowContext(ctx, query, verificationToken)

	var user models.UserFullModel
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsVerified, &user.VerificationToken, &user.VerificationTokenExpiresAt, &user.CreatedAt, &user.FollowersCount)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) UpdateVerificationToken(ctx context.Context, userID int64, newVerificationToken string, expiresAt time.Time) (*userservice.UserPrivateModel, error) {
	query := `
		UPDATE users
		SET verification_token = $2, verification_token_expires_at = $3
		WHERE id = $1
		RETURNING id, username, email, created_at
	`
	var user userservice.UserPrivateModel

	tx, hasTx := utils.GetTxFromContext(ctx)
	if hasTx {
		_, err := tx.ExecContext(ctx, query, userID, newVerificationToken, expiresAt)
		if err != nil {
			return nil, err
		}
	} else {
		err := r.postgres.QueryRowContext(ctx, query, userID, newVerificationToken, expiresAt).Scan(&user.Id, &user.Username, &user.Email, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (r *UserRepo) SetVerified(ctx context.Context, userID int64) (*userservice.UserPrivateModel, error) {
	query := `
		UPDATE users
		SET is_verified = TRUE, verification_token = NULL, verification_token_expires_at = NULL
		WHERE id = $1
		RETURNING id, username, email, created_at
	`
	var user userservice.UserPrivateModel

	tx, hasTx := utils.GetTxFromContext(ctx)
	if hasTx {
		_, err := tx.ExecContext(ctx, query, userID)
		if err != nil {
			return nil, err
		}
	} else {
		err := r.postgres.QueryRowContext(ctx, query, userID).Scan(&user.Id, &user.Username, &user.Email, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (r *UserRepo) Create(ctx context.Context, username, email, password, verificationToken string, verificationTokenExpiresAt time.Time) (*userservice.UserPrivateModel, error) {
	query := `
		INSERT INTO users (username, email, password, verification_token, verification_token_expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, created_at
	`
	tx, hasTx := utils.GetTxFromContext(ctx)

	var row *sql.Row
	if hasTx {
		row = tx.QueryRowContext(ctx, query, username, email, password, verificationToken, verificationTokenExpiresAt)
	} else {
		row = r.postgres.QueryRowContext(ctx, query, username, email, password, verificationToken, verificationTokenExpiresAt)
	}

	var user userservice.UserPrivateModel
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) ChangeUsername(ctx context.Context, userID int64, username string) (*userservice.UserPublicModel, error) {
	query := `
		UPDATE users
		SET username = $2
		WHERE id = $1
		RETURNING id, username, followers_count
	`
	tx, hasTx := utils.GetTxFromContext(ctx)

	var row *sql.Row
	if hasTx {
		row = tx.QueryRowContext(ctx, query, userID, username)
	} else {
		row = r.postgres.QueryRowContext(ctx, query, userID, username)
	}

	var user userservice.UserPublicModel
	err := row.Scan(&user.Id, &user.Username, &user.FollowersCount)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) ChangeEmail(ctx context.Context, userID int64, email string) (*userservice.UserPrivateModel, error) {
	query := `
		UPDATE users
		SET email = $2
		WHERE id = $1
		RETURNING id, username, email, created_at
	`
	tx, hasTx := utils.GetTxFromContext(ctx)

	var row *sql.Row
	if hasTx {
		row = tx.QueryRowContext(ctx, query, userID, email)
	} else {
		row = r.postgres.QueryRowContext(ctx, query, userID, email)
	}

	var user userservice.UserPrivateModel
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) ChangePassword(ctx context.Context, userID int64, password string) (*userservice.UserPrivateModel, error) {
	query := `
		UPDATE users
		SET password = $2
		WHERE id = $1
		RETURNING id, username, email, created_at
	`
	tx, hasTx := utils.GetTxFromContext(ctx)

	var row *sql.Row
	if hasTx {
		row = tx.QueryRowContext(ctx, query, userID, password)
	} else {
		row = r.postgres.QueryRowContext(ctx, query, userID, password)
	}

	var user userservice.UserPrivateModel
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) Delete(ctx context.Context, userID int64) error {
	query := `DELETE FROM users WHERE id = $1`

	tx, hasTx := utils.GetTxFromContext(ctx)
	if hasTx {
		_, err := tx.ExecContext(ctx, query, userID)
		return err
	}
	_, err := r.postgres.ExecContext(ctx, query, userID)
	return err
}
