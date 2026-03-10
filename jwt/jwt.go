package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const AccessTokenExpirationTime = time.Minute * 15
const RefreshTokenExpirationTime = time.Hour * 24 * 7

func CheckToken(tokenStr string) (bool, string) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v\n", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})
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
	curVersion := getVersion()
	if version != curVersion {
		return false, ""
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return false, ""
	}
	return true, userID
}

func CheckRefreshToken(tokenStr string) (bool, string) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v\n", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})
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
	curVersion := getVersion()
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

	storedToken, exists := getToken(tokenID)
	if !exists {
		return false, ""
	}
	if storedToken.UserID != userID {
		return false, ""
	}
	return true, userID
}

func GenerateToken(userID string) (string, error) {
	expirationTime := time.Now().Add(AccessTokenExpirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
		"version": getVersion(),
	})

	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return tokenString, nil
}

func GenerateRefreshToken(userID string, deviceID string) (string, error) {
	expirationTime := time.Now().Add(RefreshTokenExpirationTime)

	tokenID := uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      tokenID,
		"exp":     expirationTime.Unix(),
		"user_id": userID,
		"version": getVersion(),
	})
	refreshString, err := token.SignedString(getJWTSecret())
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	addToken(RefreshTokenInfo{
		UserID:    userID,
		DeviceID:  deviceID,
		ExpiredAt: expirationTime,
	}, tokenID)

	return refreshString, nil
}

func DeleteRefreshToken(tokenStr string) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v\n", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["id"].(string)
		if !ok {
			return
		}
		deleteToken(id)
	}
}
