package jwt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ErrorFunc func(args ...interface{}) error

var WrongSigningMethodError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("unexpected signing method: %v\n", args)
}
var InvalidTokenID ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("invalid token id %v\n", args)
}
var JWTSecretError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("secret must be at least 8 bytes")
}
var JWTVersionError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("JWT_VERSION env variable not set")
}

type JWTUseCaseInterface interface {
	parseToken(tokenStr string) (*jwt.Token, error)
	CheckToken(tokenStr string) (bool, int)
	CheckRefreshToken(ctx context.Context, tokenStr string) (bool, int)
	GenerateToken(userID int) (string, error)
	GenerateRefreshToken(ctx context.Context, userID int, deviceID string) (string, error)
	DeleteRefreshToken(ctx context.Context, tokenStr string) error
	GetJWTSecret() []byte
	GetVersion() string
}

const AccessTokenExpirationTime = time.Minute * 15
const RefreshTokenExpirationTime = time.Hour * 24 * 7

type Jwt struct {
	repo    repository.JWTRepositoryInterface
	mu      sync.RWMutex
	secret  []byte
	version string
}

func NewJWT(repo repository.JWTRepositoryInterface, secret string, version string) (*Jwt, error) {
	if len(secret) < 8 {
		return &Jwt{}, JWTSecretError()
	}
	if version == "" {
		return &Jwt{}, JWTVersionError()
	}
	return &Jwt{
		repo:    repo,
		secret:  []byte(secret),
		version: version,
	}, nil
}

func (obj *Jwt) parseToken(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, WrongSigningMethodError(token.Header["alg"])
		}
		return obj.GetJWTSecret(), nil
	})
	if err != nil {
		return &jwt.Token{}, err
	}
	return token, nil
}

func (obj *Jwt) CheckToken(tokenStr string) (bool, int) {
	token, err := obj.parseToken(tokenStr)
	if err != nil {
		fmt.Println(err)
		return false, -1
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return false, -1
	}

	version, ok := claims["version"].(string)
	if !ok {
		return false, -1
	}
	curVersion := obj.GetVersion()
	if version != curVersion {
		return false, -1
	}

	userIDFloat, ok := claims["user_id"].(float64)
	userID := int(userIDFloat)
	if !ok {
		return false, -1
	}
	return true, userID
}

func (obj *Jwt) CheckRefreshToken(ctx context.Context, tokenStr string) (bool, int) {
	token, err := obj.parseToken(tokenStr)
	if err != nil {
		fmt.Println(err)
		return false, -1
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return false, -1
	}

	version, ok := claims["version"].(string)
	if !ok {
		return false, -1
	}
	curVersion := obj.GetVersion()
	if version != curVersion {
		return false, -1
	}

	tokenID, ok := claims["id"].(string)
	if !ok {
		return false, -1
	}

	userIDFloat, ok := claims["user_id"].(float64)
	userID := int(userIDFloat)
	if !ok {
		return false, -1
	}

	storedToken, err := obj.repo.Get(ctx, tokenID)
	if err != nil {
		return false, -1
	}
	if storedToken.UserID != userID {
		return false, -1
	}
	return true, userID
}

func (obj *Jwt) GenerateToken(userID int) (string, error) {
	expirationTime := time.Now().Add(AccessTokenExpirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
		"version": obj.GetVersion(),
	})

	tokenString, err := token.SignedString(obj.GetJWTSecret())
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return tokenString, nil
}

func (obj *Jwt) GenerateRefreshToken(ctx context.Context, userID int, deviceID string) (string, error) {
	expirationTime := time.Now().Add(RefreshTokenExpirationTime)

	tokenID := uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      tokenID,
		"exp":     expirationTime.Unix(),
		"user_id": userID,
		"version": obj.GetVersion(),
	})
	refreshString, err := token.SignedString(obj.GetJWTSecret())
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	err = obj.repo.DeleteByUserID(ctx, userID)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	err = obj.repo.Create(ctx, tokenID, models.RefreshTokenInfo{
		Uuid:      tokenID,
		UserID:    userID,
		DeviceID:  deviceID,
		ExpiredAt: expirationTime,
	})
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return refreshString, nil
}

func (obj *Jwt) DeleteRefreshToken(ctx context.Context, tokenStr string) error {
	token, err := obj.parseToken(tokenStr)
	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["id"].(string)
		if !ok {
			return InvalidTokenID(claims["id"])
		}
		err = obj.repo.DeleteByUUID(ctx, id)
		if err != nil {
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
