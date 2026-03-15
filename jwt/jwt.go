package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const AccessTokenExpirationTime = time.Minute * 15
const RefreshTokenExpirationTime = time.Hour * 24 * 7

func parseToken(repo repository.JWTRepositoryInterface, tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v\n", token.Header["alg"])
		}
		return repo.GetJWTSecret(), nil
	})
	if err != nil {
		return &jwt.Token{}, err
	}
	return token, nil
}

func CheckToken(repo repository.JWTRepositoryInterface, tokenStr string) (bool, string) {
	token, err := parseToken(repo, tokenStr)
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
	curVersion := repo.GetVersion()
	if version != curVersion {
		return false, ""
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return false, ""
	}
	return true, userID
}

func CheckRefreshToken(repo repository.JWTRepositoryInterface, tokenStr string) (bool, string) {
	token, err := parseToken(repo, tokenStr)
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
	curVersion := repo.GetVersion()
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

	storedToken, err := repo.Get(context.Background(), tokenID)
	if err != nil {
		return false, ""
	}
	if storedToken.UserID != userID {
		return false, ""
	}
	return true, userID
}

func GenerateToken(repo repository.JWTRepositoryInterface, userID string) (string, error) {
	expirationTime := time.Now().Add(AccessTokenExpirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
		"version": repo.GetVersion(),
	})

	tokenString, err := token.SignedString(repo.GetJWTSecret())
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return tokenString, nil
}

func GenerateRefreshToken(repo repository.JWTRepositoryInterface, userID string, deviceID string) (string, error) {
	expirationTime := time.Now().Add(RefreshTokenExpirationTime)

	tokenID := uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      tokenID,
		"exp":     expirationTime.Unix(),
		"user_id": userID,
		"version": repo.GetVersion(),
	})
	refreshString, err := token.SignedString(repo.GetJWTSecret())
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	err = repo.Create(context.Background(), tokenID, models.RefreshTokenInfo{
		UserID:    userID,
		DeviceID:  deviceID,
		ExpiredAt: expirationTime,
	})
	if err != nil {
		return "", err
	}

	return refreshString, nil
}

func DeleteRefreshToken(repo repository.JWTRepositoryInterface, tokenStr string) error {
	token, err := parseToken(repo, tokenStr)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["id"].(string)
		if !ok {
			return fmt.Errorf("invalid token id %v", claims["id"])
		}
		err = repo.Delete(context.Background(), id) // ToDo: general context
		if err != nil {
			return err
		}
	}
	return nil
}
