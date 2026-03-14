package storage

import (
	"fmt"
	"sync"
	"time"
)

var onceJWT sync.Once
var tokenStore RefreshTokenStore

type RefreshTokenInfo struct {
	UserID    string
	ExpiredAt time.Time
	DeviceID  string
}

type RefreshTokenStore struct {
	mu      sync.RWMutex
	tokens  map[string]RefreshTokenInfo // key: tokenID -> информация о токене
	secret  []byte
	version string
}

func initTokenStorage(secret string, version string) {
	tokenStore = RefreshTokenStore{
		tokens:  make(map[string]RefreshTokenInfo),
		secret:  []byte(secret),
		version: version,
	}
}

func NewRefreshTokenStore(secret string, version string) error {
	if len(secret) < 8 {
		return fmt.Errorf("secret must be at least 8 bytes")
	}
	if version == "" {
		return fmt.Errorf("JWT_VERSION env variable not set")
	}
	onceJWT.Do(func() {
		initTokenStorage(secret, version)
	})
	return nil
}

func getJWTSecret() []byte {
	tokenStore.mu.RLock()
	defer tokenStore.mu.RUnlock()
	return tokenStore.secret
}

func getVersion() string {
	tokenStore.mu.RLock()
	defer tokenStore.mu.RUnlock()
	return tokenStore.version
}

func getToken(tokenID string) (RefreshTokenInfo, bool) {
	tokenStore.mu.RLock()
	defer tokenStore.mu.RUnlock()
	storedToken, exists := tokenStore.tokens[tokenID]
	return storedToken, exists
}

func addToken(obj RefreshTokenInfo, id string) {
	tokenStore.mu.Lock()
	defer tokenStore.mu.Unlock()
	tokenStore.tokens[id] = obj
}

func deleteToken(id string) {
	tokenStore.mu.Lock()
	defer tokenStore.mu.Unlock()
	delete(tokenStore.tokens, id)
}
