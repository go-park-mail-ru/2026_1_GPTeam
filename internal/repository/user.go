package repository

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/metrics"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=user.go -destination=mocks/mock_user.go -package=mocks
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
	log := logger.GetLoggerWithRequestId(ctx)
	query := `insert into "user" (username, password, email, last_login, is_staff) VALUES ($1, $2, $3, $4, $5) returning id;`
	lastLogin := pgtype.Timestamp{Valid: false}
	args := []any{userInfo.Username, userInfo.Password, userInfo.Email, lastLogin, userInfo.IsStaff}
	var id int
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(&id)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	appMetrics := metrics.GetMetrics()
	appMetrics.DbQueryDuration.WithLabelValues(query, "user").Observe(float64(duration.Milliseconds()))
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
	log.Info("Query executed")
	return id, nil
}

func (obj *UserPostgres) GetByID(ctx context.Context, id int) (*models.UserModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select id, username, password, email, created_at, last_login, avatar_url, updated_at, active, is_staff from "user" where id = $1;`
	args := []any{id}
	var lastLogin pgtype.Timestamp
	user := models.UserModel{Id: id}
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(
		&user.Id, &user.Username, &user.Password, &user.Email,
		&user.CreatedAt, &lastLogin, &user.AvatarUrl,
		&user.UpdatedAt, &user.Active, &user.IsStaff,
	)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	appMetrics := metrics.GetMetrics()
	appMetrics.DbQueryDuration.WithLabelValues(query, "user").Observe(float64(duration.Milliseconds()))
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
	log.Info("Query executed")
	return &user, nil
}

func (obj *UserPostgres) GetByUsername(ctx context.Context, username string) (*models.UserModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select id, username, password, email, created_at, last_login, avatar_url, updated_at, active, is_staff from "user" where username = $1;`
	args := []any{username}
	var lastLogin pgtype.Timestamp
	user := models.UserModel{}
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(
		&user.Id, &user.Username, &user.Password, &user.Email,
		&user.CreatedAt, &lastLogin, &user.AvatarUrl,
		&user.UpdatedAt, &user.Active, &user.IsStaff,
	)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	appMetrics := metrics.GetMetrics()
	appMetrics.DbQueryDuration.WithLabelValues(query, "user").Observe(float64(duration.Milliseconds()))
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
	log.Info("Query executed")
	return &user, nil
}

func (obj *UserPostgres) GetByEmail(ctx context.Context, email string) (*models.UserModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select id, username, password, email, created_at, last_login, avatar_url, updated_at, active, is_staff from "user" where email = $1;`
	args := []any{email}
	var lastLogin pgtype.Timestamp
	user := models.UserModel{}
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(
		&user.Id, &user.Username, &user.Password, &user.Email,
		&user.CreatedAt, &lastLogin, &user.AvatarUrl,
		&user.UpdatedAt, &user.Active, &user.IsStaff,
	)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	appMetrics := metrics.GetMetrics()
	appMetrics.DbQueryDuration.WithLabelValues(query, "user").Observe(float64(duration.Milliseconds()))
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
	log.Info("Query executed")
	return &user, nil
}

func (obj *UserPostgres) UpdateLastLogin(ctx context.Context, userId int, lastLogin time.Time) error {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `UPDATE "user" SET last_login = $1 WHERE id = $2;`
	args := []any{lastLogin, userId}
	startTime := time.Now()
	_, err := obj.db.Exec(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	appMetrics := metrics.GetMetrics()
	appMetrics.DbQueryDuration.WithLabelValues(query, "user").Observe(float64(duration.Milliseconds()))
	if err != nil {
		log.Error("failed to update last login for user (not db error)",
			zap.Error(err))
		return err
	}
	log.Info("Query executed")
	return nil
}

func (obj *UserPostgres) Update(ctx context.Context, profile models.UpdateUserProfile) (*models.UserModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `UPDATE "user" SET username   = COALESCE($1, username), password   = COALESCE($2, password), email      = COALESCE($3, email), avatar_url = COALESCE($4, avatar_url), updated_at = $5 WHERE id = $6 RETURNING id, username, password, email, created_at, last_login, avatar_url, updated_at, active`
	args := []any{
		profile.Username,
		profile.Password,
		profile.Email,
		profile.AvatarUrl,
		profile.UpdatedAt,
		profile.Id,
	}
	var lastLogin pgtype.Timestamp
	var user models.UserModel
	startTime := time.Now()
	err := obj.db.QueryRow(ctx, query, args...).Scan(
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
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, []any{
		profile.Username,
		profile.Email,
		profile.AvatarUrl,
		profile.UpdatedAt,
		profile.Id,
	}, duration)
	appMetrics := metrics.GetMetrics()
	appMetrics.DbQueryDuration.WithLabelValues(query, "user").Observe(float64(duration.Milliseconds()))
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
	log.Info("Query executed")
	return &user, nil
}

func (obj *UserPostgres) UpdateAvatar(ctx context.Context, id int, avatarUrl string) error {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `update "user" set avatar_url = $1 where id = $2;`
	args := []any{
		avatarUrl,
		id,
	}
	startTime := time.Now()
	result, err := obj.db.Exec(ctx, query, args...)
	duration := time.Since(startTime)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	appMetrics := metrics.GetMetrics()
	appMetrics.DbQueryDuration.WithLabelValues(query, "user").Observe(float64(duration.Milliseconds()))
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
	log.Info("Query executed")
	return nil
}
