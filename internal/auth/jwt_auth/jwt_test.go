package jwt_auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
)

const (
	testSecret  = "testsecret123"
	testVersion = "v1"
	testUserID  = "42"
	testDevice  = "test-device"
)

func setupJWTTest(t *testing.T) {
	t.Helper()
	err := jwt.NewRefreshTokenStore(testSecret, testVersion)
	require.NoError(t, err)
}

func TestNewRefreshTokenStore(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		secret  string
		version string
		wantErr bool
	}{
		{
			name:    "валидные параметры",
			secret:  testSecret,
			version: testVersion,
			wantErr: false,
		},
		{
			name:    "секрет меньше 8 байт",
			secret:  "short",
			version: testVersion,
			wantErr: true,
		},
		{
			name:    "пустая версия",
			secret:  testSecret,
			version: "",
			wantErr: true,
		},
		{
			name:    "пустой секрет",
			secret:  "",
			version: testVersion,
			wantErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := jwt.NewRefreshTokenStore(c.secret, c.version)
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	t.Parallel()
	setupJWTTest(t)

	cases := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "обычный userID",
			userID:  testUserID,
			wantErr: false,
		},
		{
			name:    "пустой userID",
			userID:  "",
			wantErr: false,
		},
		{
			name:    "длинный userID",
			userID:  "99999999999999999999",
			wantErr: false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			token, err := jwt.GenerateToken(c.userID)
			if c.wantErr {
				require.Error(t, err)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestCheckToken(t *testing.T) {
	t.Parallel()
	setupJWTTest(t)

	validToken, err := jwt.GenerateToken(testUserID)
	require.NoError(t, err)

	cases := []struct {
		name       string
		token      string
		wantOK     bool
		wantUserID string
	}{
		{
			name:       "валидный токен",
			token:      validToken,
			wantOK:     true,
			wantUserID: testUserID,
		},
		{
			name:       "невалидная строка",
			token:      "invalid-token",
			wantOK:     false,
			wantUserID: "",
		},
		{
			name:       "пустая строка",
			token:      "",
			wantOK:     false,
			wantUserID: "",
		},
		{
			name:       "случайный JWT-подобный токен",
			token:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.abc",
			wantOK:     false,
			wantUserID: "",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ok, userID := jwt.CheckToken(c.token)
			assert.Equal(t, c.wantOK, ok)
			assert.Equal(t, c.wantUserID, userID)
		})
	}
}

func TestGenerateAndCheckToken_RoundTrip(t *testing.T) {
	t.Parallel()
	setupJWTTest(t)

	cases := []struct {
		name   string
		userID string
	}{
		{"обычный userID", testUserID},
		{"пустой userID", ""},
		{"userID с символами", "user-abc_123"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			token, err := jwt.GenerateToken(c.userID)
			require.NoError(t, err)
			require.NotEmpty(t, token)

			ok, gotUserID := jwt.CheckToken(token)
			assert.True(t, ok)
			assert.Equal(t, c.userID, gotUserID)
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	t.Parallel()
	setupJWTTest(t)

	cases := []struct {
		name     string
		userID   string
		deviceID string
		wantErr  bool
	}{
		{
			name:     "обычные параметры",
			userID:   testUserID,
			deviceID: testDevice,
			wantErr:  false,
		},
		{
			name:     "пустой deviceID",
			userID:   testUserID,
			deviceID: "",
			wantErr:  false,
		},
		{
			name:     "пустой userID",
			userID:   "",
			deviceID: testDevice,
			wantErr:  false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			token, err := jwt.GenerateRefreshToken(c.userID, c.deviceID)
			if c.wantErr {
				require.Error(t, err)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestCheckRefreshToken(t *testing.T) {
	t.Parallel()
	setupJWTTest(t)

	validToken, err := jwt.GenerateRefreshToken(testUserID, testDevice)
	require.NoError(t, err)

	cases := []struct {
		name       string
		token      string
		wantOK     bool
		wantUserID string
	}{
		{
			name:       "валидный refresh-токен",
			token:      validToken,
			wantOK:     true,
			wantUserID: testUserID,
		},
		{
			name:       "невалидная строка",
			token:      "invalid-refresh-token",
			wantOK:     false,
			wantUserID: "",
		},
		{
			name:       "пустая строка",
			token:      "",
			wantOK:     false,
			wantUserID: "",
		},
		{
			name:       "access-токен вместо refresh",
			token:      func() string { t2, _ := jwt.GenerateToken(testUserID); return t2 }(),
			wantOK:     false,
			wantUserID: "",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ok, userID := jwt.CheckRefreshToken(c.token)
			assert.Equal(t, c.wantOK, ok)
			assert.Equal(t, c.wantUserID, userID)
		})
	}
}

func TestGenerateAndCheckRefreshToken_RoundTrip(t *testing.T) {
	t.Parallel()
	setupJWTTest(t)

	cases := []struct {
		name     string
		userID   string
		deviceID string
	}{
		{"стандартные параметры", testUserID, testDevice},
		{"пустой deviceID", testUserID, ""},
		{"разные устройства одного пользователя", testUserID, "other-device"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			token, err := jwt.GenerateRefreshToken(c.userID, c.deviceID)
			require.NoError(t, err)
			require.NotEmpty(t, token)

			ok, gotUserID := jwt.CheckRefreshToken(token)
			assert.True(t, ok)
			assert.Equal(t, c.userID, gotUserID)
		})
	}
}

func TestDeleteRefreshToken(t *testing.T) {
	t.Parallel()
	setupJWTTest(t)

	cases := []struct {
		name        string
		setup       func(t *testing.T) string
		assertAfter func(t *testing.T, token string)
	}{
		{
			name: "удаление валидного токена инвалидирует его",
			setup: func(t *testing.T) string {
				token, err := jwt.GenerateRefreshToken(testUserID, testDevice)
				require.NoError(t, err)
				return token
			},
			assertAfter: func(t *testing.T, token string) {
				ok, userID := jwt.CheckRefreshToken(token)
				assert.False(t, ok)
				assert.Empty(t, userID)
			},
		},
		{
			name: "удаление невалидного токена не паникует",
			setup: func(t *testing.T) string {
				return "invalid-token"
			},
			assertAfter: func(t *testing.T, token string) {
				ok, userID := jwt.CheckRefreshToken(token)
				assert.False(t, ok)
				assert.Empty(t, userID)
			},
		},
		{
			name: "удаление пустой строки не паникует",
			setup: func(t *testing.T) string {
				return ""
			},
			assertAfter: func(t *testing.T, token string) {
				ok, userID := jwt.CheckRefreshToken(token)
				assert.False(t, ok)
				assert.Empty(t, userID)
			},
		},
		{
			name: "повторное удаление уже удалённого токена",
			setup: func(t *testing.T) string {
				token, err := jwt.GenerateRefreshToken(testUserID, testDevice)
				require.NoError(t, err)
				jwt.DeleteRefreshToken(token)
				return token
			},
			assertAfter: func(t *testing.T, token string) {
				assert.NotPanics(t, func() {
					jwt.DeleteRefreshToken(token)
				})
				ok, userID := jwt.CheckRefreshToken(token)
				assert.False(t, ok)
				assert.Empty(t, userID)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			token := c.setup(t)
			jwt.DeleteRefreshToken(token)
			c.assertAfter(t, token)
		})
	}
}

func TestMultipleRefreshTokens_IndependentInvalidation(t *testing.T) {
	t.Parallel()
	setupJWTTest(t)

	token1, err := jwt.GenerateRefreshToken(testUserID, "device-1")
	require.NoError(t, err)

	token2, err := jwt.GenerateRefreshToken(testUserID, "device-2")
	require.NoError(t, err)

	jwt.DeleteRefreshToken(token1)

	ok1, _ := jwt.CheckRefreshToken(token1)
	assert.False(t, ok1, "token1 должен быть инвалидирован")

	ok2, userID2 := jwt.CheckRefreshToken(token2)
	assert.True(t, ok2, "token2 должен оставаться валидным")
	assert.Equal(t, testUserID, userID2)
}
