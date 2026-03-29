package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type JwtRepository interface {
	Create(ctx context.Context, token models.RefreshTokenModel) error
	DeleteByUuid(ctx context.Context, uuid string) error
	DeleteByUserId(ctx context.Context, userId int) error
	Get(ctx context.Context, uuid string) (models.RefreshTokenModel, error)
}

type JwtPostgres struct {
	db DB
}

func NewJwtPostgres(db DB) *JwtPostgres {
	return &JwtPostgres{
		db: db,
	}
}

func (obj *JwtPostgres) Create(ctx context.Context, token models.RefreshTokenModel) error {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `insert into jwt (uuid, user_id, expired_at) values ($1, $2, $3);`
	_, err := obj.db.Exec(ctx, query, token.Uuid, token.UserId, token.ExpiredAt)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		log.Error("failed to create refresh token (db error)",
			zap.Error(pgErr))
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return DuplicatedDataError
		case pgerrcode.CheckViolation:
			return ConstraintError
		default:
			return pgErr
		}
	}
	if err != nil {
		log.Error("failed to create refresh token (not db error)",
			zap.Error(err))
		return err
	}
	return nil
}

func (obj *JwtPostgres) DeleteByUuid(ctx context.Context, uuid string) error {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `delete from jwt where uuid = $1;`
	_, err := obj.db.Exec(ctx, query, uuid)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Error("failed to delete refresh token (no such uuid)",
			zap.Error(pgx.ErrNoRows))
		return NothingInTableError
	}
	if err != nil {
		log.Error("failed to delete refresh token (not db error)",
			zap.Error(err))
		return err
	}
	return nil
}

func (obj *JwtPostgres) DeleteByUserId(ctx context.Context, userID int) error {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `delete from jwt where user_id = $1;`
	_, err := obj.db.Exec(ctx, query, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Error("failed to delete refresh token (no such user)",
			zap.Error(pgx.ErrNoRows))
		return NothingInTableError
	}
	if err != nil {
		log.Error("failed to delete refresh token by user (not db error)",
			zap.Error(err))
		return err
	}
	return nil
}

func (obj *JwtPostgres) Get(ctx context.Context, uuid string) (models.RefreshTokenModel, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	query := `select user_id, expired_at from jwt where uuid = $1;`
	token := models.RefreshTokenModel{Uuid: uuid}
	err := obj.db.QueryRow(ctx, query, uuid).Scan(&token.UserId, &token.ExpiredAt)
	if err != nil {
		log.Error("failed to get refresh token (not db error)",
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RefreshTokenModel{}, NothingInTableError
		}
		if errors.Is(err, pgx.ErrTooManyRows) {
			return models.RefreshTokenModel{}, TooManyRowsError
		}
		return models.RefreshTokenModel{}, err
	}
	return token, nil
}
