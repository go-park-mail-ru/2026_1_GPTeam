package jwt

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var SECRET = []byte("myawesomesecret")

type RefreshTokenInfo struct {
	UserID    string
	ExpiredAt time.Time
	DeviceID  string
}

type RefreshTokenStore struct {
	Mu     sync.RWMutex
	Tokens map[string]RefreshTokenInfo // key: tokenID -> информация о токене
	NextID int
}

var TokenStore = RefreshTokenStore{
	Tokens: make(map[string]RefreshTokenInfo),
}

func CheckToken(tokenStr string) (bool, string) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return SECRET, nil
	})
	if err != nil {
		return false, ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["id"].(string)
		if ok {
			return true, userID
		}
	}
	return false, ""
}

func CheckRefreshToken(tokenStr string) (bool, string) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return SECRET, nil
	})
	if err != nil {
		return false, ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
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

	TokenStore.Mu.RLock()
	defer TokenStore.Mu.RUnlock()
	storedToken, exists := TokenStore.Tokens[tokenID]
	if !exists {
		return false, ""
	}
	if storedToken.UserID != userID {
		return false, ""
	}
	return true, userID
}

func GenerateToken(userID string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userID,
		"exp": expirationTime.Unix(),
	})

	tokenString, err := token.SignedString(SECRET)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateRefreshToken(userID string, deviceID string) (string, error) {
	expirationTime := time.Now().AddDate(0, 0, 7)
	TokenStore.Mu.RLock()
	tokenID := strconv.Itoa(TokenStore.NextID)
	TokenStore.Mu.RUnlock()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      tokenID,
		"exp":     expirationTime.Unix(),
		"user_id": userID,
	})
	refreshString, err := token.SignedString(SECRET)
	if err != nil {
		return "", err
	}

	TokenStore.Mu.Lock()
	defer TokenStore.Mu.Unlock()
	TokenStore.Tokens[tokenID] = RefreshTokenInfo{
		UserID:    userID,
		DeviceID:  deviceID,
		ExpiredAt: expirationTime,
	}
	TokenStore.NextID++

	return refreshString, nil
}

func DeleteRefreshToken(tokenStr string) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return SECRET, nil
	})
	if err != nil {
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["id"].(string)
		if !ok {
			return
		}
		TokenStore.Mu.Lock()
		delete(TokenStore.Tokens, id)
		TokenStore.Mu.Unlock()
	}
}
