package jwt

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenExpirationTime = time.Minute * 15
const RefreshTokenExpirationTime = time.Hour * 24 * 7

type RefreshTokenInfo struct {
	UserID    string
	ExpiredAt time.Time
	DeviceID  string
}

type RefreshTokenStore struct {
	Mu     sync.RWMutex
	Tokens map[string]RefreshTokenInfo // key: tokenID -> информация о токене
	NextID int
	secret []byte
}

func NewRefreshTokenStore(secret string) (*RefreshTokenStore, error) {
	if len(secret) < 8 {
		return nil, fmt.Errorf("secret must be at least 8 bytes")
	}
	return &RefreshTokenStore{
		Tokens: make(map[string]RefreshTokenInfo),
		NextID: 0,
		secret: []byte(secret),
	}, nil
}

var TokenStore *RefreshTokenStore

func CheckToken(tokenStr string) (bool, string) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return TokenStore.secret, nil
	})
	if err != nil {
		fmt.Println(err)
		return false, ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
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
		return TokenStore.secret, nil
	})
	if err != nil {
		fmt.Println(err)
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
	expirationTime := time.Now().Add(AccessTokenExpirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
	})

	tokenString, err := token.SignedString(TokenStore.secret)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return tokenString, nil
}

func GenerateRefreshToken(userID string, deviceID string) (string, error) {
	expirationTime := time.Now().Add(RefreshTokenExpirationTime)
	TokenStore.Mu.Lock()
	defer TokenStore.Mu.Unlock()
	tokenID := strconv.Itoa(TokenStore.NextID)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      tokenID,
		"exp":     expirationTime.Unix(),
		"user_id": userID,
	})
	refreshString, err := token.SignedString(TokenStore.secret)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

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
		return TokenStore.secret, nil
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
		TokenStore.Mu.Lock()
		delete(TokenStore.Tokens, id)
		TokenStore.Mu.Unlock()
	}
}
