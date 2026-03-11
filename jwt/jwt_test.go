package jwt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
)

func setupJWTTest(t *testing.T) {
	t.Helper()
	err := jwt.NewRefreshTokenStore("testsecret123", "0")
	require.NoError(t, err)
}

func TestNewRefreshTokenStore(t *testing.T) {
	err := jwt.NewRefreshTokenStore("testsecret123", "0")
	require.NoError(t, err)
}

func TestGenerateTokenReturnsNonEmptyToken(t *testing.T) {
	setupJWTTest(t)

	token, err := jwt.GenerateToken("42")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestCheckTokenInvalidToken(t *testing.T) {
	ok, userID := jwt.CheckToken("invalid-token")

	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestGenerateTokenThenCheckTokenReturnsSameUserID(t *testing.T) {
	setupJWTTest(t)

	token, err := jwt.GenerateToken("42")
	require.NoError(t, err)

	ok, userID := jwt.CheckToken(token)
	assert.True(t, ok)
	assert.Equal(t, "42", userID)
}

func TestGenerateRefreshTokenThenCheckRefreshTokenReturnsSameUserID(t *testing.T) {
	setupJWTTest(t)

	token, err := jwt.GenerateRefreshToken("42", "test-device")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	ok, userID := jwt.CheckRefreshToken(token)
	assert.True(t, ok)
	assert.Equal(t, "42", userID)
}

func TestCheckRefreshTokenInvalidToken(t *testing.T) {
	setupJWTTest(t)

	ok, userID := jwt.CheckRefreshToken("invalid-refresh-token")

	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestDeleteRefreshTokenInvalidatesToken(t *testing.T) {
	setupJWTTest(t)

	token, err := jwt.GenerateRefreshToken("42", "test-device")
	require.NoError(t, err)

	jwt.DeleteRefreshToken(token)

	ok, userID := jwt.CheckRefreshToken(token)
	assert.False(t, ok)
	assert.Empty(t, userID)
}
