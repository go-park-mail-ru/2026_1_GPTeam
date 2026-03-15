package application

import (
	"context"
	"fmt"
	"time"

	jwtStore "github.com/go-park-mail-ru/2026_1_GPTeam/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTUseCaseInterface interface {
	parseToken(tokenStr string) (*jwt.Token, error)
	CheckToken(tokenStr string) (bool, string)
	CheckRefreshToken(ctx context.Context, tokenStr string) (bool, string)
	GenerateToken(userID string) (string, error)
	GenerateRefreshToken(ctx context.Context, userID string, deviceID string) (string, error)
	DeleteRefreshToken(ctx context.Context, tokenStr string) error
}

const AccessTokenExpirationTime = time.Minute * 15
const RefreshTokenExpirationTime = time.Hour * 24 * 7

type ErrorFunc func(args ...interface{}) error

var WrongSigningMethodError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("unexpected signing method: %v\n", args)
}
var InvalidTokenID ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("invalid token id %v\n", args)
}

type Jwt struct {
	repo repository.JWTRepositoryInterface
}

func NewJWT(repo repository.JWTRepositoryInterface) *Jwt {
	return &Jwt{
		repo: repo,
	}
}

func (obj *Jwt) parseToken(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, WrongSigningMethodError(token.Header["alg"])
		}
		return obj.repo.GetJWTSecret(), nil
	})
	if err != nil {
		return &jwt.Token{}, err
	}
	return token, nil
}

func (obj *Jwt) CheckToken(tokenStr string) (bool, string) {
	token, err := obj.parseToken(tokenStr)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return false, ""
	}

	version, ok := claims["version"].(string)
	if !ok {
		return false, ""
	}
	curVersion := obj.repo.GetVersion()
	if version != curVersion {
		return false, ""
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return false, ""
	}
	return true, userID
}

func (obj *Jwt) CheckRefreshToken(ctx context.Context, tokenStr string) (bool, string) {
	token, err := obj.parseToken(tokenStr)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return false, ""
	}

	version, ok := claims["version"].(string)
	if !ok {
		return false, ""
	}
	curVersion := obj.repo.GetVersion()
	if version != curVersion {
		return false, ""
	}

	tokenID, ok := claims["id"].(string)
	if !ok {
		return false, ""
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return false, ""
	}

	storedToken, err := obj.repo.Get(ctx, tokenID)
	if err != nil {
		return false, ""
	}
	if storedToken.UserID != userID {
		return false, ""
	}
	return true, userID
}

func (obj *Jwt) GenerateToken(userID string) (string, error) {
	expirationTime := time.Now().Add(AccessTokenExpirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
		"version": obj.repo.GetVersion(),
	})

	tokenString, err := token.SignedString(obj.repo.GetJWTSecret())
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return tokenString, nil
}

func (obj *Jwt) GenerateRefreshToken(ctx context.Context, userID string, deviceID string) (string, error) {
	expirationTime := time.Now().Add(RefreshTokenExpirationTime)

	tokenID := uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      tokenID,
		"exp":     expirationTime.Unix(),
		"user_id": userID,
		"version": obj.repo.GetVersion(),
	})
	refreshString, err := token.SignedString(obj.repo.GetJWTSecret())
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	err = obj.repo.Create(ctx, tokenID, jwtStore.RefreshTokenInfo{
		UserID:    userID,
		DeviceID:  deviceID,
		ExpiredAt: expirationTime,
	})
	if err != nil {
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
		err = obj.repo.Delete(ctx, id)
		if err != nil {
			return err
		}
	}
	return nil
}
