package jwt

import (
    "sync"
    "testing"
    timepkg "time"
)

func setupJWTTest(t *testing.T) {
    t.Helper()
    onceJWT = sync.Once{}
    if err := NewRefreshTokenStore("testsecret123"); err != nil {
        t.Fatalf("NewRefreshTokenStore returned error: %v", err)
    }
    tokenStore.mu.Lock()
    tokenStore.tokens = make(map[string]RefreshTokenInfo)
    tokenStore.secret = []byte("testsecret123")
    tokenStore.mu.Unlock()
}

func TestNewRefreshTokenStoreRejectsShortSecret(t *testing.T) {
    onceJWT = sync.Once{}
    err := NewRefreshTokenStore("short")
    if err == nil {
        t.Fatal("expected error for short secret")
    }
}

func TestGenerateAndCheckToken(t *testing.T) {
    setupJWTTest(t)

    token, err := GenerateToken("42")
    if err != nil {
        t.Fatalf("GenerateToken returned error: %v", err)
    }

    ok, userID := CheckToken(token)
    if !ok {
        t.Fatal("expected token to be valid")
    }
    if userID != "42" {
        t.Fatalf("expected user id 42, got %q", userID)
    }
}

func TestCheckTokenRejectsInvalidString(t *testing.T) {
    setupJWTTest(t)

    ok, userID := CheckToken("not-a-jwt")
    if ok || userID != "" {
        t.Fatalf("expected invalid token, got ok=%v userID=%q", ok, userID)
    }
}

func TestGenerateAndCheckRefreshToken(t *testing.T) {
    setupJWTTest(t)

    token, err := GenerateRefreshToken("7", "device-1")
    if err != nil {
        t.Fatalf("GenerateRefreshToken returned error: %v", err)
    }

    ok, userID := CheckRefreshToken(token)
    if !ok {
        t.Fatal("expected refresh token to be valid")
    }
    if userID != "7" {
        t.Fatalf("expected user id 7, got %q", userID)
    }
    if len(tokenStore.tokens) != 1 {
        t.Fatalf("expected 1 stored refresh token, got %d", len(tokenStore.tokens))
    }
}

func TestDeleteRefreshTokenInvalidatesIt(t *testing.T) {
    setupJWTTest(t)

    token, err := GenerateRefreshToken("8", "device-1")
    if err != nil {
        t.Fatalf("GenerateRefreshToken returned error: %v", err)
    }

    DeleteRefreshToken(token)

    ok, userID := CheckRefreshToken(token)
    if ok || userID != "" {
        t.Fatalf("expected deleted refresh token to be invalid, got ok=%v userID=%q", ok, userID)
    }
}

func TestCheckRefreshTokenFailsWhenStoreRecordRemoved(t *testing.T) {
    setupJWTTest(t)

    token, err := GenerateRefreshToken("15", "device-9")
    if err != nil {
        t.Fatalf("GenerateRefreshToken returned error: %v", err)
    }

    tokenStore.mu.Lock()
    for id := range tokenStore.tokens {
        delete(tokenStore.tokens, id)
    }
    tokenStore.mu.Unlock()

    ok, userID := CheckRefreshToken(token)
    if ok || userID != "" {
        t.Fatalf("expected token without storage record to be invalid, got ok=%v userID=%q", ok, userID)
    }
}

func TestStoredRefreshTokenMetadata(t *testing.T) {
    setupJWTTest(t)

    before := timepkg.Now()
    _, err := GenerateRefreshToken("99", "phone")
    if err != nil {
        t.Fatalf("GenerateRefreshToken returned error: %v", err)
    }

    tokenStore.mu.RLock()
    defer tokenStore.mu.RUnlock()
    if len(tokenStore.tokens) != 1 {
        t.Fatalf("expected exactly one token in store, got %d", len(tokenStore.tokens))
    }
    for _, info := range tokenStore.tokens {
        if info.UserID != "99" {
            t.Fatalf("expected stored user id 99, got %q", info.UserID)
        }
        if info.DeviceID != "phone" {
            t.Fatalf("expected device id phone, got %q", info.DeviceID)
        }
        if !info.ExpiredAt.After(before) {
            t.Fatalf("expected expiration time after test start, got %v", info.ExpiredAt)
        }
    }
}
