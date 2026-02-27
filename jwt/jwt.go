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

func GenerateToken(userID string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
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

func GenerateRefreshToken(userID string, deviseID string) (string, error) {
	expirationTime := time.Now().AddDate(0, 0, 7)
	tokenID := strconv.Itoa(TokenStore.NextID)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  tokenID,
		"exp": expirationTime.Unix(),
	})
	refreshString, err := token.SignedString(SECRET)
	if err != nil {
		return "", err
	}

	TokenStore.Mu.Lock()
	defer TokenStore.Mu.Unlock()
	TokenStore.Tokens[tokenID] = RefreshTokenInfo{
		UserID:    userID,
		ExpiredAt: expirationTime,
		DeviceID:  deviseID,
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
		fmt.Println(TokenStore.Tokens)
		delete(TokenStore.Tokens, id)
		fmt.Println(TokenStore.Tokens)
		TokenStore.Mu.Unlock()
	}
}
