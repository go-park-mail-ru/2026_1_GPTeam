package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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

func NewJwtPostgres(db *pgxpool.Pool) *JwtPostgres {
	return &JwtPostgres{db: db}
}

func (obj *JwtPostgres) Create(ctx context.Context, token models.RefreshTokenModel) error {
	query := `insert into jwt (uuid, user_id, expired_at) values ($1, $2, $3);`
	_, err := obj.db.Exec(ctx, query, token.Uuid, token.UserId, token.ExpiredAt)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return DuplicatedDataError
		case pgerrcode.CheckViolation:
			return ConstraintError
		default:
			return pgErr
		}
	}
	return err
}

func (obj *JwtPostgres) DeleteByUuid(ctx context.Context, uuid string) error {
	query := `delete from jwt where uuid = $1;`
	_, err := obj.db.Exec(ctx, query, uuid)
	if errors.Is(err, pgx.ErrNoRows) {
		return NothingInTableError
	}
	return err
}

func (obj *JwtPostgres) DeleteByUserId(ctx context.Context, userID int) error {
	query := `delete from jwt where user_id = $1;`
	_, err := obj.db.Exec(ctx, query, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return NothingInTableError
	}
	return err
}

func (obj *JwtPostgres) Get(ctx context.Context, uuid string) (models.RefreshTokenModel, error) {
	query := `select user_id, expired_at from jwt where uuid = $1;`
	token := models.RefreshTokenModel{Uuid: uuid}
	err := obj.db.QueryRow(ctx, query, uuid).Scan(&token.UserId, &token.ExpiredAt)
	if err != nil {
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
