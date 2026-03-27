package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type JwtRepository interface {
	Create(ctx context.Context, token models.RefreshTokenModel) error
	DeleteByUuid(ctx context.Context, uuid string) error
	DeleteByUserId(ctx context.Context, userId int) error
	Get(ctx context.Context, uuid string) (models.RefreshTokenModel, error)
}

type JwtPostgres struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewJwtPostgres(db *pgxpool.Pool) *JwtPostgres {
	return &JwtPostgres{
		db:  db,
		log: logger.GetLogger(),
	}
}

func (obj *JwtPostgres) Create(ctx context.Context, token models.RefreshTokenModel) error {
	query := `insert into jwt (uuid, user_id, expired_at) values ($1, $2, $3);`
	_, err := obj.db.Exec(ctx, query, token.Uuid, token.UserId, token.ExpiredAt)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		obj.log.Error("failed to create refresh token (db error)",
			zap.String("request_id", ctx.Value("request_id").(string)),
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
		obj.log.Error("failed to create refresh token (not db error)",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return err
	}
	obj.log.Info("creating refresh token query executed",
		zap.String("request_id", ctx.Value("request_id").(string)))
	return nil
}

func (obj *JwtPostgres) DeleteByUuid(ctx context.Context, uuid string) error {
	query := `delete from jwt where uuid = $1;`
	_, err := obj.db.Exec(ctx, query, uuid)
	if errors.Is(err, pgx.ErrNoRows) {
		obj.log.Error("failed to delete refresh token (no such uuid)",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(pgx.ErrNoRows))
		return NothingInTableError
	}
	if err != nil {
		obj.log.Error("failed to delete refresh token (not db error)",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return err
	}
	obj.log.Info("deleting refresh token query executed",
		zap.String("request_id", ctx.Value("request_id").(string)))
	return nil
}

func (obj *JwtPostgres) DeleteByUserId(ctx context.Context, userID int) error {
	query := `delete from jwt where user_id = $1;`
	_, err := obj.db.Exec(ctx, query, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		obj.log.Error("failed to delete refresh token (no such user)",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(pgx.ErrNoRows))
		return NothingInTableError
	}
	if err != nil {
		obj.log.Error("failed to delete refresh token by user (not db error)",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return err
	}
	obj.log.Info("deleting refresh token by user query executed",
		zap.String("request_id", ctx.Value("request_id").(string)))
	return nil
}

func (obj *JwtPostgres) Get(ctx context.Context, uuid string) (models.RefreshTokenModel, error) {
	query := `select user_id, expired_at from jwt where uuid = $1;`
	token := models.RefreshTokenModel{Uuid: uuid}
	err := obj.db.QueryRow(ctx, query, uuid).Scan(&token.UserId, &token.ExpiredAt)
	if err != nil {
		obj.log.Error("failed to get refresh token (not db error)",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RefreshTokenModel{}, NothingInTableError
		}
		if errors.Is(err, pgx.ErrTooManyRows) {
			return models.RefreshTokenModel{}, TooManyRowsError
		}
		return models.RefreshTokenModel{}, err
	}
	obj.log.Info("getting refresh token query executed",
		zap.String("request_id", ctx.Value("request_id").(string)))
	return token, nil
}
