package jwt_auth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type JwtUseCaseInterface interface {
	parseToken(tokenStr string) (*jwt.Token, error)
	CheckToken(tokenStr string) (bool, int)
	CheckRefreshToken(ctx context.Context, tokenStr string) (bool, int)
	GenerateToken(userId int) (string, error)
	GenerateRefreshToken(ctx context.Context, userId int, deviceId string) (string, error)
	DeleteRefreshToken(ctx context.Context, tokenStr string) error
	GetJWTSecret() []byte
	GetVersion() string
}
