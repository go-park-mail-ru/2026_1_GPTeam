package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type JWTRepositoryInterface interface {
	Create(ctx context.Context, uuid string, token models.RefreshTokenInfo) error
	Delete(ctx context.Context, uuid string) error
	DeleteByUserID(ctx context.Context, userID int) error
	Get(ctx context.Context, uuid string) (models.RefreshTokenInfo, error)
	GetJWTSecret() []byte
	GetVersion() string
}

type ErrorFunc func(args ...interface{}) error

var JWTSecretError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("secret must be at least 8 bytes")
}
var JWTVersionError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("JWT_VERSION env variable not set")
}

type PostgresJWT struct {
	db      *pgx.Conn
	mu      sync.RWMutex
	secret  []byte
	version string // ToDo: move to auth packet
}

func NewPostgresJWT(db *pgx.Conn, secret string, version string) (*PostgresJWT, error) {
	if len(secret) < 8 {
		return &PostgresJWT{}, JWTSecretError()
	}
	if version == "" {
		return &PostgresJWT{}, JWTVersionError()
	}
	return &PostgresJWT{
		db:      db,
		secret:  []byte(secret),
		version: version,
	}, nil
}

func (obj *PostgresJWT) Create(ctx context.Context, uuid string, token models.RefreshTokenInfo) error {
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
	if err != nil {
		fmt.Printf("Unable to create token: %v\n", err)
		return err
	}
	return nil
}

func (obj *PostgresJWT) Delete(ctx context.Context, uuid string) error {
	query := `delete from jwt where uuid = $1;`
	_, err := obj.db.Exec(ctx, query, uuid)
	if err != nil {
		fmt.Printf("Unable to delete token: %v\n", err)
		return err
	}
	return nil
}

func (obj *PostgresJWT) DeleteByUserID(ctx context.Context, userID int) error {
	query := `delete from jwt where user_id = $1;`
	_, err := obj.db.Exec(ctx, query, userID)
	if err != nil {
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

func (obj *PostgresJWT) GetJWTSecret() []byte {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.secret
}

func (obj *PostgresJWT) GetVersion() string {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.version
}
