package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"main/auth"
	"main/jwt"
)

func setupAuthJWT(t *testing.T) {
	t.Helper()
	err := jwt.NewRefreshTokenStore("testsecret123")
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

func TestGenerateNewAuthCookie_SetsAccessCookie(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "123")

	assert.Equal(t, auth.TokenName, accessCookie.Name)
	assert.NotEmpty(t, accessCookie.Value)
	assert.Equal(t, "/", accessCookie.Path)
	assert.True(t, accessCookie.HttpOnly)
}

func TestGenerateNewAuthCookie_SetsRefreshCookie(t *testing.T) {
	setupAuthJWT(t)

	_, refreshCookie := issueAuthCookies(t, "123")

	assert.Equal(t, auth.RefreshTokenName, refreshCookie.Name)
	assert.NotEmpty(t, refreshCookie.Value)
	assert.Equal(t, "/auth/", refreshCookie.Path)
	assert.True(t, refreshCookie.HttpOnly)
}

func TestGetAuthCookie_ReturnsCookie(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "44")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(accessCookie)

	cookie, err := auth.GetAuthCookie(req)
	require.NoError(t, err)
	require.NotNil(t, cookie)

	assert.Equal(t, accessCookie.Value, cookie.Value)
}

func TestIsAuth_ReturnsAuthorizedUser(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "44")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(accessCookie)

	ok, userID := auth.IsAuth(req)
	assert.True(t, ok)
	assert.Equal(t, "44", userID)
}

func TestIsAuth_RejectsMissingCookie(t *testing.T) {
	setupAuthJWT(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	ok, userID := auth.IsAuth(req)
	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestGetUserIDFromCookie_ReturnsIntUserID(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "77")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(accessCookie)

	userID, err := auth.GetUserIDFromCookie(req)
	require.NoError(t, err)

	assert.Equal(t, 77, userID)
}

func TestGetUserIDFromCookie_RejectsInvalidToken(t *testing.T) {
	setupAuthJWT(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.TokenName,
		Value: "broken-token",
	})

	_, err := auth.GetUserIDFromCookie(req)
	require.Error(t, err)
}

func TestGetUserIDFromCookie_RejectsNonNumericUserID(t *testing.T) {
	setupAuthJWT(t)

	token, err := jwt.GenerateToken("abc")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.TokenName,
		Value: token,
	})

	_, err = auth.GetUserIDFromCookie(req)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), strconv.ErrSyntax.Error()))
}

func TestRefreshToken_RotatesCookies(t *testing.T) {
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

func TestRefreshToken_OldRefreshTokenBecomesInvalid(t *testing.T) {
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

func TestClearOldToken_ExpiresCookies(t *testing.T) {
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
