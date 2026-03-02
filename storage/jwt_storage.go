package storage

import (
	"fmt"
	"sync"
	"time"
)

var once sync.Once
var tokenStore RefreshTokenStore

type RefreshTokenInfo struct {
	UserID    string
	ExpiredAt time.Time
	DeviceID  string
}

type RefreshTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]RefreshTokenInfo // key: tokenID -> информация о токене
	secret []byte
}

func initTokenStorage(secret string) {
	tokenStore = RefreshTokenStore{
		tokens: make(map[string]RefreshTokenInfo),
		secret: []byte(secret),
	}
}

func NewRefreshTokenStore(secret string) error {
	if len(secret) < 8 {
		return fmt.Errorf("secret must be at least 8 bytes")
	}
	once.Do(func() {
		initTokenStorage(secret)
	})
	return nil
}

func GetJWTSecret() []byte {
	return tokenStore.secret
}

func GetToken(tokenID string) (RefreshTokenInfo, bool) {
	tokenStore.mu.RLock()
	defer tokenStore.mu.RUnlock()
	storedToken, exists := tokenStore.tokens[tokenID]
	return storedToken, exists
}

func DoWithLock(f func()) {
	tokenStore.mu.Lock()
	defer tokenStore.mu.Unlock()
	f()
}

func AddToken(obj RefreshTokenInfo, id string) {
	tokenStore.mu.Lock()
	defer tokenStore.mu.Unlock()
	tokenStore.tokens[id] = obj
}

func DeleteToken(id string) {
	tokenStore.mu.Lock()
	defer tokenStore.mu.Unlock()
	delete(tokenStore.tokens, id)
}
