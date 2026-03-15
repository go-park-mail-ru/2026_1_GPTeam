package repository

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type JWTRepositoryInterface interface {
	Create(ctx context.Context, uuid string, token storage.RefreshTokenInfo) error
	Delete(ctx context.Context, uuid string) error
	Get(ctx context.Context, uuid string) (storage.RefreshTokenInfo, error)
	GetJWTSecret() []byte
	GetVersion() string
}

type JWTPostgresqlRepository struct {
	db      *pgx.Conn
	mu      sync.RWMutex
	secret  []byte
	version string // ToDo: move to auth packet
}

func NewJWTPostgresqlRepository(db *pgx.Conn, secret string, version string) (*JWTPostgresqlRepository, error) {
	if len(secret) < 8 {
		return &JWTPostgresqlRepository{}, fmt.Errorf("secret must be at least 8 bytes")
	}
	if version == "" {
		return &JWTPostgresqlRepository{}, fmt.Errorf("JWT_VERSION env variable not set")
	}
	return &JWTPostgresqlRepository{
		db:      db,
		secret:  []byte(secret),
		version: version,
	}, nil
}

func (obj *JWTPostgresqlRepository) Create(ctx context.Context, uuid string, token storage.RefreshTokenInfo) error {
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

func (obj *JWTPostgresqlRepository) Get(ctx context.Context, uuid string) (storage.RefreshTokenInfo, error) {
	query := `select user_id, expired_at from jwt where uuid = $1;`
	var userId pgtype.Int4
	var expiredAt pgtype.Timestamp
	err := obj.db.QueryRow(ctx, query, uuid).Scan(&userId, &expiredAt)
	if err != nil {
		fmt.Printf("Unable to get token: %v\n", err)
		return storage.RefreshTokenInfo{}, err
	}
	token := storage.RefreshTokenInfo{
		UserID:    strconv.Itoa(int(userId.Int32)),
		ExpiredAt: expiredAt.Time,
		DeviceID:  "",
	}
	return token, nil
}

func (obj *JWTPostgresqlRepository) GetJWTSecret() []byte {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.secret
}

func (obj *JWTPostgresqlRepository) GetVersion() string {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.version
}
