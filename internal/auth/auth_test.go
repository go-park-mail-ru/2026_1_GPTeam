package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	auth2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
)

const (
	testSecret  = "testsecret123"
	testVersion = "v1"
)

func setupAuthJWT(t *testing.T) {
	t.Helper()
	err := jwt.NewRefreshTokenStore(testSecret, testVersion)
	require.NoError(t, err)
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func issueAuthCookies(t *testing.T, userID string) (*http.Cookie, *http.Cookie) {
	t.Helper()

	rec := httptest.NewRecorder()
	auth.GenerateNewAuthCookie(rec, userID)

	cookies := rec.Result().Cookies()
	accessCookie := findCookie(cookies, auth2.TokenName)
	refreshCookie := findCookie(cookies, auth2.RefreshTokenName)

	require.NotNil(t, accessCookie)
	require.NotNil(t, refreshCookie)

	return accessCookie, refreshCookie
}

func TestGenerateNewAuthCookie(t *testing.T) {
	t.Parallel()
	setupAuthJWT(t)

	cases := []struct {
		name          string
		userID        string
		assertAccess  func(t *testing.T, c *http.Cookie)
		assertRefresh func(t *testing.T, c *http.Cookie)
	}{
		{
			name:   "устанавливает access-куку",
			userID: "123",
			assertAccess: func(t *testing.T, c *http.Cookie) {
				require.NotNil(t, c)
				assert.Equal(t, auth2.TokenName, c.Name)
				assert.NotEmpty(t, c.Value)
				assert.Equal(t, "/", c.Path)
				assert.True(t, c.HttpOnly)
				assert.True(t, c.Secure)
			},
			assertRefresh: nil,
		},
		{
			name:         "устанавливает refresh-куку",
			userID:       "123",
			assertAccess: nil,
			assertRefresh: func(t *testing.T, c *http.Cookie) {
				require.NotNil(t, c)
				assert.Equal(t, auth2.RefreshTokenName, c.Name)
				assert.NotEmpty(t, c.Value)
				assert.Equal(t, "/auth/", c.Path)
				assert.True(t, c.HttpOnly)
				assert.True(t, c.Secure)
			},
		},
		{
			name:   "разные userID дают разные токены",
			userID: "999",
			assertAccess: func(t *testing.T, c *http.Cookie) {
				require.NotNil(t, c)
				assert.NotEmpty(t, c.Value)
			},
			assertRefresh: func(t *testing.T, c *http.Cookie) {
				require.NotNil(t, c)
				assert.NotEmpty(t, c.Value)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			auth.GenerateNewAuthCookie(rec, c.userID)
			cookies := rec.Result().Cookies()

			if c.assertAccess != nil {
				c.assertAccess(t, findCookie(cookies, auth2.TokenName))
			}
			if c.assertRefresh != nil {
				c.assertRefresh(t, findCookie(cookies, auth2.RefreshTokenName))
			}
		})
	}
}

func TestGetAuthCookie(t *testing.T) {
	t.Parallel()
	setupAuthJWT(t)

	validAccess, _ := issueAuthCookies(t, "44")

	cases := []struct {
		name    string
		cookie  *http.Cookie
		wantErr bool
		wantVal string
	}{
		{
			name:    "валидная кука возвращается без ошибки",
			cookie:  validAccess,
			wantErr: false,
			wantVal: validAccess.Value,
		},
		{
			name:    "без куки → ошибка",
			cookie:  nil,
			wantErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if c.cookie != nil {
				req.AddCookie(c.cookie)
			}

			got, err := auth.GetAuthCookie(req)
			if c.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, c.wantVal, got.Value)
			}
		})
	}
}

func TestIsAuth(t *testing.T) {
	t.Parallel()
	setupAuthJWT(t)

	validAccess, _ := issueAuthCookies(t, "44")

	cases := []struct {
		name       string
		cookie     *http.Cookie
		wantOK     bool
		wantUserID string
	}{
		{
			name:       "валидный токен → авторизован",
			cookie:     validAccess,
			wantOK:     true,
			wantUserID: "44",
		},
		{
			name:       "нет куки → не авторизован",
			cookie:     nil,
			wantOK:     false,
			wantUserID: "",
		},
		{
			name: "сломанный токен → не авторизован",
			cookie: &http.Cookie{
				Name:  auth2.TokenName,
				Value: "broken-token",
			},
			wantOK:     false,
			wantUserID: "",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if c.cookie != nil {
				req.AddCookie(c.cookie)
			}

			ok, userID := auth.IsAuth(req)
			assert.Equal(t, c.wantOK, ok)
			assert.Equal(t, c.wantUserID, userID)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	t.Parallel()
	setupAuthJWT(t)

	t.Run("ротирует куки и возвращает userID", func(t *testing.T) {
		t.Parallel()

		oldAccess, oldRefresh := issueAuthCookies(t, "501")

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(oldRefresh)

		rec := httptest.NewRecorder()
		ok, userID := auth.RefreshToken(rec, req)

		assert.True(t, ok)
		assert.Equal(t, "501", userID)

		newCookies := rec.Result().Cookies()
		newAccess := findCookie(newCookies, auth2.TokenName)
		newRefresh := findCookie(newCookies, auth2.RefreshTokenName)

		require.NotNil(t, newAccess)
		require.NotNil(t, newRefresh)
		assert.NotEqual(t, oldAccess.Value, newAccess.Value)
		assert.NotEqual(t, oldRefresh.Value, newRefresh.Value)
	})

	t.Run("старый refresh-токен становится невалидным", func(t *testing.T) {
		t.Parallel()

		_, oldRefresh := issueAuthCookies(t, "502")

		req1 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req1.AddCookie(oldRefresh)
		rec1 := httptest.NewRecorder()
		ok, _ := auth.RefreshToken(rec1, req1)
		require.True(t, ok)

		req2 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req2.AddCookie(oldRefresh)
		rec2 := httptest.NewRecorder()
		ok, userID := auth.RefreshToken(rec2, req2)
		assert.False(t, ok)
		assert.Empty(t, userID)
	})

	t.Run("нет refresh-куки → false", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		rec := httptest.NewRecorder()
		ok, userID := auth.RefreshToken(rec, req)
		assert.False(t, ok)
		assert.Empty(t, userID)
	})

	t.Run("сломанный refresh-токен → false", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{
			Name:  auth2.RefreshTokenName,
			Value: "broken-refresh-token",
		})
		rec := httptest.NewRecorder()
		ok, userID := auth.RefreshToken(rec, req)
		assert.False(t, ok)
		assert.Empty(t, userID)
	})
}

func TestClearOldToken(t *testing.T) {
	t.Parallel()
	setupAuthJWT(t)

	t.Run("экспайрит обе куки", func(t *testing.T) {
		t.Parallel()

		_, refreshCookie := issueAuthCookies(t, "808")

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.AddCookie(refreshCookie)

		rec := httptest.NewRecorder()
		auth.ClearOldToken(rec, req)

		cookies := rec.Result().Cookies()
		clearedAccess := findCookie(cookies, auth2.TokenName)
		clearedRefresh := findCookie(cookies, auth2.RefreshTokenName)

		require.NotNil(t, clearedAccess)
		require.NotNil(t, clearedRefresh)
		assert.Empty(t, clearedAccess.Value)
		assert.Empty(t, clearedRefresh.Value)
		assert.Equal(t, "/", clearedAccess.Path)
		assert.Equal(t, "/auth/", clearedRefresh.Path)
	})

	t.Run("нет refresh-куки → не паникует", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		rec := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			auth.ClearOldToken(rec, req)
		})
	})

	t.Run("после ClearOldToken refresh-токен становится невалидным", func(t *testing.T) {
		t.Parallel()

		_, refreshCookie := issueAuthCookies(t, "909")

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.AddCookie(refreshCookie)
		rec := httptest.NewRecorder()
		auth.ClearOldToken(rec, req)

		req2 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req2.AddCookie(refreshCookie)
		rec2 := httptest.NewRecorder()
		ok, userID := auth.RefreshToken(rec2, req2)
		assert.False(t, ok)
		assert.Empty(t, userID)
	})
}
