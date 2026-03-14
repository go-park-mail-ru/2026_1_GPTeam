package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type JWTRepositoryInterface interface {
	Create(ctx context.Context, uuid string, token jwt.RefreshTokenInfo) error
	Delete(ctx context.Context, uuid string) error
	Get(ctx context.Context, uuid string) (jwt.RefreshTokenInfo, error)
}

type JWTPostgresqlRepository struct {
	db *pgx.Conn
}

func NewJWTPostgresqlRepository(db *pgx.Conn) *JWTPostgresqlRepository {
	return &JWTPostgresqlRepository{db: db}
}

func (obj *JWTPostgresqlRepository) Create(ctx context.Context, uuid string, token jwt.RefreshTokenInfo) error {
	query := `insert into jwt (uuid, user_id, expired_at) values ($1, $2, $3);`
	pk := pgtype.Text{
		String: uuid,
		Valid:  true,
	}
	intUserID, err := strconv.Atoi(token.UserID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	userID := pgtype.Int4{
		Int32: int32(intUserID), // ToDo: change to int
		Valid: true,
	}
	expiredAt := pgtype.Timestamp{
		Time:  token.ExpiredAt,
		Valid: true,
	}
	_, err = obj.db.Exec(ctx, query, pk, userID, expiredAt)
	if err != nil {
		fmt.Printf("Unable to create token: %v\n", err)
		return err
	}
	return nil
}

func (obj *JWTPostgresqlRepository) Delete(ctx context.Context, uuid string) error {
	query := `delete from jwt where uuid = $1;`
	_, err := obj.db.Exec(ctx, query, uuid)
	if err != nil {
		fmt.Printf("Unable to delete token: %v\n", err)
		return err
	}
	return nil
}

func (obj *JWTPostgresqlRepository) Get(ctx context.Context, uuid string) (jwt.RefreshTokenInfo, error) {
	query := `select user_id, expired_at from jwt where uuid = $1;`
	var userId pgtype.Int4
	var expiredAt pgtype.Timestamp
	err := obj.db.QueryRow(ctx, query, uuid).Scan(&userId, &expiredAt)
	if err != nil {
		fmt.Printf("Unable to get token: %v\n", err)
		return jwt.RefreshTokenInfo{}, err
	}
	token := jwt.RefreshTokenInfo{
		UserID:    strconv.Itoa(int(userId.Int32)),
		ExpiredAt: expiredAt.Time,
		DeviceID:  "",
	}
	return token, nil
}
