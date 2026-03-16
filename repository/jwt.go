package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type PostgresJWT struct {
	db *pgx.Conn
}

func NewPostgresJWT(db *pgx.Conn) *PostgresJWT {
	return &PostgresJWT{
		db: db,
	}
}

func (obj *PostgresJWT) Create(ctx context.Context, uuid string, token models.RefreshTokenInfo) error { // ToDo: use token.Uuid
	query := `insert into jwt (uuid, user_id, expired_at) values ($1, $2, $3);`
	pk := pgtype.Text{
		String: uuid,
		Valid:  true,
	}
	userID := pgtype.Int4{
		Int32: int32(token.UserID),
		Valid: true,
	}
	expiredAt := pgtype.Timestamp{
		Time:  token.ExpiredAt,
		Valid: true,
	}
	_, err := obj.db.Exec(ctx, query, pk, userID, expiredAt)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok {
		switch pgErr.Code {
		case "23505":
			return DuplicatedDataError(pgErr.Message)
		case "23514":
			return ConstraintError(pgErr.Message)
		default:
			return pgErr
		}
	}
	if err != nil {
		fmt.Printf("Unable to create token: %v\n", err)
		return err
	}
	return nil
}

func (obj *PostgresJWT) DeleteByUUID(ctx context.Context, uuid string) error {
	query := `delete from jwt where uuid = $1;`
	_, err := obj.db.Exec(ctx, query, uuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return NothingInTableError()
		}
		fmt.Printf("Unable to delete token: %v\n", err)
		return err
	}
	return nil
}

func (obj *PostgresJWT) DeleteByUserID(ctx context.Context, userID int) error {
	query := `delete from jwt where user_id = $1;`
	_, err := obj.db.Exec(ctx, query, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return NothingInTableError()
		}
		fmt.Printf("Unable to delete token: %v\n", err)
		return err
	}
	return nil
}

func (obj *PostgresJWT) Get(ctx context.Context, uuid string) (models.RefreshTokenInfo, error) {
	query := `select user_id, expired_at from jwt where uuid = $1;`
	var userId pgtype.Int4
	var expiredAt pgtype.Timestamp
	err := obj.db.QueryRow(ctx, query, uuid).Scan(&userId, &expiredAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RefreshTokenInfo{}, NothingInTableError()
		} else if errors.Is(err, pgx.ErrTooManyRows) {
			return models.RefreshTokenInfo{}, TooManyRowsError()
		}
		fmt.Printf("Unable to get token: %v\n", err)
		return models.RefreshTokenInfo{}, err
	}
	token := models.RefreshTokenInfo{
		UserID:    int(userId.Int32),
		ExpiredAt: expiredAt.Time,
		DeviceID:  "",
	}
	return token, nil
}
