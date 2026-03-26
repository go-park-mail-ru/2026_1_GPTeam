package jwt_auth

import (
	"context"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const AccessTokenExpirationTime = time.Minute * 15
const RefreshTokenExpirationTime = time.Hour * 24 * 7

type JwtUseCase interface {
	parseToken(tokenStr string) (*jwt.Token, error)
	CheckToken(tokenStr string) (bool, int)
	CheckRefreshToken(ctx context.Context, tokenStr string) (bool, int)
	GenerateToken(userId int) (string, error)
	GenerateRefreshToken(ctx context.Context, userId int, deviceId string) (string, error)
	DeleteRefreshToken(ctx context.Context, tokenStr string) error
	GetJWTSecret() []byte
	GetVersion() string
}

type Jwt struct {
	repository repository.JwtRepository
	mu         sync.RWMutex
	secret     []byte
	version    string
	log        *zap.Logger
}

func NewJwt(repository repository.JwtRepository, secret string, version string) (*Jwt, error) {
	if len(secret) < 8 {
		return &Jwt{}, JwtSecretError
	}
	if version == "" {
		return &Jwt{}, JwtVersionError
	}
	return &Jwt{
		repository: repository,
		secret:     []byte(secret),
		version:    version,
		log:        logger.GetLogger(),
	}, nil
}

func (obj *Jwt) parseToken(tokenStr string) (*jwt.Token, error) {
	obj.log.Info("parsing token")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			obj.log.Error("unexpected signing method")
			return nil, WrongSigningMethodError
		}
		return obj.GetJWTSecret(), nil
	})
	if err != nil {
		obj.log.Error("failed to parse token")
		return &jwt.Token{}, err
	}
	return token, nil
}

func (obj *Jwt) CheckToken(tokenStr string) (bool, int) {
	obj.log.Info("checking token")
	token, err := obj.parseToken(tokenStr)
	if err != nil {
		return false, -1
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		obj.log.Warn("invalid token (unable to claim payload)")
		return false, -1
	}

	version, ok := claims["version"].(string)
	if !ok {
		obj.log.Warn("invalid token (unable to claim version)")
		return false, -1
	}
	curVersion := obj.GetVersion()
	if version != curVersion {
		obj.log.Error("invalid token (invalid version)")
		return false, -1
	}

	userIdFloat, ok := claims["user_id"].(float64)
	userId := int(userIdFloat)
	if !ok {
		obj.log.Error("invalid token (unable to claim user_id)")
		return false, -1
	}
	return true, userId
}

func (obj *Jwt) CheckRefreshToken(ctx context.Context, tokenStr string) (bool, int) {
	obj.log.Info("checking refresh token",
		zap.String("request_id", ctx.Value("request_id").(string)))
	token, err := obj.parseToken(tokenStr)
	if err != nil {
		return false, -1
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		obj.log.Warn("invalid token (unable to claim payload)",
			zap.String("request_id", ctx.Value("request_id").(string)))
		return false, -1
	}

	version, ok := claims["version"].(string)
	if !ok {
		obj.log.Warn("invalid token (unable to claim version)",
			zap.String("request_id", ctx.Value("request_id").(string)))
		return false, -1
	}
	curVersion := obj.GetVersion()
	if version != curVersion {
		obj.log.Error("invalid token (invalid version)",
			zap.String("request_id", ctx.Value("request_id").(string)))
		return false, -1
	}

	tokenId, ok := claims["id"].(string)
	if !ok {
		obj.log.Warn("invalid token (unable to claim token_id)",
			zap.String("request_id", ctx.Value("request_id").(string)))
		return false, -1
	}

	userIdFloat, ok := claims["user_id"].(float64)
	userId := int(userIdFloat)
	if !ok {
		obj.log.Warn("invalid token (unable to claim user_id)",
			zap.String("request_id", ctx.Value("request_id").(string)))
		return false, -1
	}

	storedToken, err := obj.repository.Get(ctx, tokenId)
	if err != nil {
		obj.log.Error("failed to get token from db",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return false, -1
	}
	if storedToken.UserId != userId {
		obj.log.Error("invalid token (invalid userId)",
			zap.String("request_id", ctx.Value("request_id").(string)))
		return false, -1
	}
	return true, userId
}

func (obj *Jwt) GenerateToken(userID int) (string, error) {
	obj.log.Info("generating new token",
		zap.Int("user_id", userID))
	expirationTime := time.Now().Add(AccessTokenExpirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
		"version": obj.GetVersion(),
	})

	tokenStr, err := token.SignedString(obj.GetJWTSecret())
	if err != nil {
		obj.log.Error("failed to sign token",
			zap.Int("user_id", userID),
			zap.Error(err))
		return "", err
	}

	return tokenStr, nil
}

func (obj *Jwt) GenerateRefreshToken(ctx context.Context, userId int, deviceId string) (string, error) {
	obj.log.Info("generating new refresh token",
		zap.Int("user_id", userId),
		zap.String("request_id", ctx.Value("request_id").(string)))
	expirationTime := time.Now().Add(RefreshTokenExpirationTime)

	tokenId := uuid.New().String()
	obj.log.Info("generated uuid for token",
		zap.String("token_id", tokenId),
		zap.Int("user_id", userId),
		zap.String("request_id", ctx.Value("request_id").(string)))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      tokenId,
		"exp":     expirationTime.Unix(),
		"user_id": userId,
		"version": obj.GetVersion(),
	})
	refreshStr, err := token.SignedString(obj.GetJWTSecret())
	if err != nil {
		obj.log.Error("failed to sign token",
			zap.Int("user_id", userId),
			zap.String("token_id", tokenId),
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return "", err
	}

	err = obj.repository.DeleteByUserId(ctx, userId)
	if err != nil {
		obj.log.Error("failed to delete old refresh token from db",
			zap.Int("user_id", userId),
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return "", err
	}
	err = obj.repository.Create(ctx, models.RefreshTokenModel{
		Uuid:      tokenId,
		UserId:    userId,
		DeviceId:  deviceId,
		ExpiredAt: expirationTime,
	})
	if err != nil {
		obj.log.Error("failed to create new refresh token",
			zap.String("token_id", tokenId),
			zap.Int("user_id", userId),
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return "", err
	}

	return refreshStr, nil
}

func (obj *Jwt) DeleteRefreshToken(ctx context.Context, tokenStr string) error {
	obj.log.Info("deleting refresh token",
		zap.String("request_id", ctx.Value("request_id").(string)))
	token, err := obj.parseToken(tokenStr)
	if err != nil {
		obj.log.Error("failed to parse token",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["id"].(string)
		if !ok {
			obj.log.Error("invalid token (unable to claim token_id)",
				zap.String("request_id", ctx.Value("request_id").(string)))
			return InvalidTokenId
		}
		err = obj.repository.DeleteByUuid(ctx, id)
		if err != nil {
			obj.log.Error("failed to delete refresh token from db",
				zap.String("request_id", ctx.Value("request_id").(string)),
				zap.Error(err))
			return err
		}
	}
	return nil
}

func (obj *Jwt) GetJWTSecret() []byte {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.secret
}

func (obj *Jwt) GetVersion() string {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.version
}
