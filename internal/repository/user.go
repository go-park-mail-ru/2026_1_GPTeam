package repository

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type UserRepository interface {
	Create(ctx context.Context, userInfo models.UserModel) (int, error)
	GetByID(ctx context.Context, id int) (*models.UserModel, error)
	GetByUsername(ctx context.Context, username string) (*models.UserModel, error)
	GetByEmail(ctx context.Context, email string) (*models.UserModel, error)
	UpdateLastLogin(ctx context.Context, userId int, lastLogin time.Time) error
	Update(ctx context.Context, profile models.UpdateUserProfile) (*models.UserModel, error)
	UpdateAvatar(ctx context.Context, id int, avatarUrl string) error
}

type UserPostgres struct {
	db DB
}

func NewUserPostgres(db DB) *UserPostgres {
	return &UserPostgres{
		db: db,
	}
}

func (obj *UserPostgres) Create(ctx context.Context, userInfo models.UserModel) (int, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `insert into "user" (username, password, email, last_login) VALUES ($1, $2, $3, $4) returning id;`
	var id int
	lastLogin := pgtype.Timestamp{Valid: false}
	err := obj.db.QueryRow(ctx, query, userInfo.Username, userInfo.Password, userInfo.Email, lastLogin).Scan(&id)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to create user (db error)",
			zap.Error(pgErr))
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return -1, DuplicatedDataError
		case pgerrcode.CheckViolation:
			return -1, ConstraintError
		default:
			return -1, pgErr
		}
	}
	if err != nil {
		log.Error("failed to create user (not db error)",
			zap.Error(err))
		return -1, err
	}
	return id, nil
}

func (obj *UserPostgres) GetByID(ctx context.Context, id int) (*models.UserModel, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `select id, username, password, email, created_at, last_login, avatar_url, updated_at, active from "user" where id = $1;`
	var lastLogin pgtype.Timestamp
	user := models.UserModel{Id: id}
	err := obj.db.QueryRow(ctx, query, id).Scan(
		&user.Id, &user.Username, &user.Password, &user.Email,
		&user.CreatedAt, &lastLogin, &user.AvatarUrl,
		&user.UpdatedAt, &user.Active,
	)
	if err != nil {
		log.Error("failed to get user (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NothingInTableError
		}
		return nil, err
	}
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}
	return &user, nil
}

func (obj *UserPostgres) GetByUsername(ctx context.Context, username string) (*models.UserModel, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `select id, username, password, email, created_at, last_login, avatar_url, updated_at, active from "user" where username = $1;`
	var lastLogin pgtype.Timestamp
	user := models.UserModel{}
	err := obj.db.QueryRow(ctx, query, username).Scan(
		&user.Id, &user.Username, &user.Password, &user.Email,
		&user.CreatedAt, &lastLogin, &user.AvatarUrl,
		&user.UpdatedAt, &user.Active,
	)
	if err != nil {
		log.Error("failed to get user by username (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NothingInTableError
		}
		return nil, err
	}
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}
	return &user, nil
}

func (obj *UserPostgres) GetByEmail(ctx context.Context, email string) (*models.UserModel, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `select id, username, password, email, created_at, last_login, avatar_url, updated_at, active from "user" where email = $1;`
	var lastLogin pgtype.Timestamp
	user := models.UserModel{}
	err := obj.db.QueryRow(ctx, query, email).Scan(
		&user.Id, &user.Username, &user.Password, &user.Email,
		&user.CreatedAt, &lastLogin, &user.AvatarUrl,
		&user.UpdatedAt, &user.Active,
	)
	if err != nil {
		log.Error("failed to get user by email (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NothingInTableError
		}
		return nil, err
	}
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}
	return &user, nil
}

func (obj *UserPostgres) UpdateLastLogin(ctx context.Context, userId int, lastLogin time.Time) error {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `UPDATE "user" SET last_login = $1 WHERE id = $2;`
	_, err := obj.db.Exec(ctx, query, lastLogin, userId)
	if err != nil {
		log.Error("failed to update last login for user (not db error)",
			zap.Error(err))
		return err
	}
	return nil
}

func (obj *UserPostgres) Update(ctx context.Context, profile models.UpdateUserProfile) (*models.UserModel, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `UPDATE "user" SET username   = COALESCE($1, username), password   = COALESCE($2, password), email      = COALESCE($3, email), avatar_url = COALESCE($4, avatar_url), updated_at = $5 WHERE id = $6 RETURNING id, username, password, email, created_at, last_login, avatar_url, updated_at, active`

	var lastLogin pgtype.Timestamp
	var user models.UserModel
	err := obj.db.QueryRow(
		ctx, query,
		profile.Username,
		profile.Password,
		profile.Email,
		profile.AvatarUrl,
		profile.UpdatedAt,
		profile.Id,
	).Scan(
		&user.Id,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
		&lastLogin,
		&user.AvatarUrl,
		&user.UpdatedAt,
		&user.Active,
	)
	if err != nil {
		log.Error("failed to update user (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NothingInTableError
		}
		return nil, err
	}
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}
	return &user, nil
}

func (obj *UserPostgres) UpdateAvatar(ctx context.Context, id int, avatarUrl string) error {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `update "user" set avatar_url = $1 where id = $2;`
	result, err := obj.db.Exec(ctx, query, avatarUrl, id)
	if err != nil {
		log.Warn("failed to update avatar (not db error)",
			zap.Int("user_id", id),
			zap.Error(err))
		return err
	}

	if result.RowsAffected() == 0 {
		log.Warn("no rows affected",
			zap.Int("user_id", id))
		return NothingInTableError
	}

	return nil
}
