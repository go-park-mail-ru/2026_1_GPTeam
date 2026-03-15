package application

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
	jwtStore "github.com/go-park-mail-ru/2026_1_GPTeam/storage"
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

type JWTUseCase struct {
	repo repository.JWTRepositoryInterface
}

func NewJWTUseCase(repo repository.JWTRepositoryInterface) *JWTUseCase {
	return &JWTUseCase{
		repo: repo,
	}
}

func (obj *JWTUseCase) parseToken(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v\n", token.Header["alg"])
		}
		return obj.repo.GetJWTSecret(), nil
	})
	if err != nil {
		return &jwt.Token{}, err
	}
	return token, nil
}

func (obj *JWTUseCase) CheckToken(tokenStr string) (bool, string) {
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

func (obj *JWTUseCase) CheckRefreshToken(ctx context.Context, tokenStr string) (bool, string) {
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

func (obj *JWTUseCase) GenerateToken(userID string) (string, error) {
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

func (obj *JWTUseCase) GenerateRefreshToken(ctx context.Context, userID string, deviceID string) (string, error) {
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

func (obj *JWTUseCase) DeleteRefreshToken(ctx context.Context, tokenStr string) error {
	token, err := obj.parseToken(tokenStr)
	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["id"].(string)
		if !ok {
			return fmt.Errorf("invalid token id %v", claims["id"])
		}
		err = obj.repo.Delete(ctx, id)
		if err != nil {
			return err
		}
	}
	return nil
}
