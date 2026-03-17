package jwt_auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	internalTestSecret  = "internalsecret99"
	internalTestVersion = "internal-v1"
)

func setupInternalJWT(t *testing.T) {
	t.Helper()
	initTokenStorage(internalTestSecret, internalTestVersion)
}

func makeTokenWithClaims(t *testing.T, claims gojwt.MapClaims, secret string) string {
	t.Helper()
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	str, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return str
}

func makeRS256Token(t *testing.T, claims gojwt.MapClaims) string {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	str, err := token.SignedString(privateKey)
	require.NoError(t, err)
	return str
}

func TestCheckToken_Internal(t *testing.T) {
	setupInternalJWT(t)

	cases := []struct {
		name       string
		token      func() string
		wantOK     bool
		wantUserID string
	}{
		{
			name: "неверный алгоритм подписи RS256",
			token: func() string {
				return makeRS256Token(t, gojwt.MapClaims{
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				})
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "неверный секрет",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				}, "wrong-secret-xyz")
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "истёкший токен",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"user_id": "42",
					"exp":     time.Now().Add(-time.Hour).Unix(),
					"version": internalTestVersion,
				}, internalTestSecret)
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "версия не совпадает",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": "wrong-version",
				}, internalTestSecret)
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "отсутствует user_id",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				}, internalTestSecret)
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "валидный токен",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				}, internalTestSecret)
			},
			wantOK:     true,
			wantUserID: "42",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			ok, userID := CheckToken(c.token())
			assert.Equal(t, c.wantOK, ok)
			assert.Equal(t, c.wantUserID, userID)
		})
	}
}

func TestCheckRefreshToken_Internal(t *testing.T) {
	setupInternalJWT(t)

	mismatchTokenID := "mismatch-token-id"
	addToken(RefreshTokenInfo{
		UserID:    "99",
		DeviceID:  "device",
		ExpiredAt: time.Now().Add(time.Hour),
	}, mismatchTokenID)

	cases := []struct {
		name       string
		token      func() string
		wantOK     bool
		wantUserID string
	}{
		{
			name: "неверный алгоритм подписи RS256",
			token: func() string {
				return makeRS256Token(t, gojwt.MapClaims{
					"id":      "some-id",
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				})
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "неверный секрет",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"id":      "some-id",
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				}, "wrong-secret-xyz")
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "истёкший токен",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"id":      "expired-token-id",
					"user_id": "42",
					"exp":     time.Now().Add(-time.Hour).Unix(),
					"version": internalTestVersion,
				}, internalTestSecret)
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "версия не совпадает",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"id":      "some-id",
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": "wrong-version",
				}, internalTestSecret)
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "отсутствует id",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				}, internalTestSecret)
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "отсутствует user_id",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"id":      "some-id",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				}, internalTestSecret)
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "токен не найден в store",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"id":      "nonexistent-uuid-9999",
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				}, internalTestSecret)
			},
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "userID не совпадает с хранилищем",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"id":      mismatchTokenID,
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				}, internalTestSecret)
			},
			wantOK:     false,
			wantUserID: "",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			ok, userID := CheckRefreshToken(c.token())
			assert.Equal(t, c.wantOK, ok)
			assert.Equal(t, c.wantUserID, userID)
		})
	}
}

func TestDeleteRefreshToken_Internal(t *testing.T) {
	setupInternalJWT(t)

	cases := []struct {
		name  string
		token func() string
	}{
		{
			name: "неверный алгоритм подписи RS256",
			token: func() string {
				return makeRS256Token(t, gojwt.MapClaims{
					"id":  "some-id",
					"exp": time.Now().Add(time.Hour).Unix(),
				})
			},
		},
		{
			name: "токен без поля id",
			token: func() string {
				return makeTokenWithClaims(t, gojwt.MapClaims{
					"user_id": "42",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"version": internalTestVersion,
				}, internalTestSecret)
			},
		},
		{
			name: "невалидная строка",
			token: func() string {
				return "invalid-token"
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				DeleteRefreshToken(c.token())
			})
		})
	}
}
