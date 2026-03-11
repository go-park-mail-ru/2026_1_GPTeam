package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
)

func setupAuthJWT(t *testing.T) {
	t.Helper()
	err := jwt.NewRefreshTokenStore("testsecret123", "0")
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
	accessCookie := findCookie(cookies, auth.TokenName)
	refreshCookie := findCookie(cookies, auth.RefreshTokenName)

	require.NotNil(t, accessCookie)
	require.NotNil(t, refreshCookie)

	return accessCookie, refreshCookie
}

func TestGenerateNewAuthCookieSetsAccessCookie(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "123")

	assert.Equal(t, auth.TokenName, accessCookie.Name)
	assert.NotEmpty(t, accessCookie.Value)
	assert.Equal(t, "/", accessCookie.Path)
	assert.True(t, accessCookie.HttpOnly)
}

func TestGenerateNewAuthCookieSetsRefreshCookie(t *testing.T) {
	setupAuthJWT(t)

	_, refreshCookie := issueAuthCookies(t, "123")

	assert.Equal(t, auth.RefreshTokenName, refreshCookie.Name)
	assert.NotEmpty(t, refreshCookie.Value)
	assert.Equal(t, "/auth/", refreshCookie.Path)
	assert.True(t, refreshCookie.HttpOnly)
}

func TestGetAuthCookieReturnsCookie(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "44")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(accessCookie)

	cookie, err := auth.GetAuthCookie(req)
	require.NoError(t, err)
	require.NotNil(t, cookie)

	assert.Equal(t, accessCookie.Value, cookie.Value)
}

func TestIsAuthReturnsAuthorizedUser(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "44")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(accessCookie)

	ok, userID := auth.IsAuth(req)
	assert.True(t, ok)
	assert.Equal(t, "44", userID)
}

func TestIsAuthRejectsMissingCookie(t *testing.T) {
	setupAuthJWT(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	ok, userID := auth.IsAuth(req)
	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestRefreshTokenRotatesCookies(t *testing.T) {
	setupAuthJWT(t)

	oldAccessCookie, oldRefreshCookie := issueAuthCookies(t, "501")

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(oldRefreshCookie)

	rec := httptest.NewRecorder()
	ok, userID := auth.RefreshToken(rec, req)

	assert.True(t, ok)
	assert.Equal(t, "501", userID)

	newCookies := rec.Result().Cookies()
	newAccessCookie := findCookie(newCookies, auth.TokenName)
	newRefreshCookie := findCookie(newCookies, auth.RefreshTokenName)

	require.NotNil(t, newAccessCookie)
	require.NotNil(t, newRefreshCookie)

	assert.NotEqual(t, oldAccessCookie.Value, newAccessCookie.Value)
	assert.NotEqual(t, oldRefreshCookie.Value, newRefreshCookie.Value)
}

func TestRefreshTokenOldRefreshTokenBecomesInvalid(t *testing.T) {
	setupAuthJWT(t)

	_, oldRefreshCookie := issueAuthCookies(t, "501")

	req1 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req1.AddCookie(oldRefreshCookie)

	rec1 := httptest.NewRecorder()
	ok, userID := auth.RefreshToken(rec1, req1)
	assert.True(t, ok)
	assert.Equal(t, "501", userID)

	req2 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req2.AddCookie(oldRefreshCookie)

	rec2 := httptest.NewRecorder()
	ok, userID = auth.RefreshToken(rec2, req2)
	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestClearOldTokenExpiresCookies(t *testing.T) {
	setupAuthJWT(t)

	_, refreshCookie := issueAuthCookies(t, "808")

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(refreshCookie)

	rec := httptest.NewRecorder()
	auth.ClearOldToken(rec, req)

	cookies := rec.Result().Cookies()
	clearedAccessCookie := findCookie(cookies, auth.TokenName)
	clearedRefreshCookie := findCookie(cookies, auth.RefreshTokenName)

	require.NotNil(t, clearedAccessCookie)
	require.NotNil(t, clearedRefreshCookie)

	assert.Empty(t, clearedAccessCookie.Value)
	assert.Empty(t, clearedRefreshCookie.Value)
	assert.Equal(t, "/", clearedAccessCookie.Path)
	assert.Equal(t, "/auth/", clearedRefreshCookie.Path)
}
