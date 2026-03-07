package auth

import (
	"main/jwt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func setupAuthJWT(t *testing.T) {
	t.Helper()

	if err := jwt.NewRefreshTokenStore("testsecret123"); err != nil {
		t.Fatalf("failed to init jwt storage: %v", err)
	}
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
	GenerateNewAuthCookie(rec, userID)

	cookies := rec.Result().Cookies()
	accessCookie := findCookie(cookies, TokenName)
	refreshCookie := findCookie(cookies, RefreshTokenName)

	if accessCookie == nil {
		t.Fatal("expected access cookie to be set")
	}
	if refreshCookie == nil {
		t.Fatal("expected refresh cookie to be set")
	}

	return accessCookie, refreshCookie
}

func TestGenerateNewAuthCookie_SetsAccessCookie(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "123")

	if accessCookie.Name != TokenName {
		t.Fatalf("expected cookie name %q, got %q", TokenName, accessCookie.Name)
	}
	if accessCookie.Value == "" {
		t.Fatal("expected access cookie value to be non-empty")
	}
	if accessCookie.Path != "/" {
		t.Fatalf("expected access cookie path '/', got %q", accessCookie.Path)
	}
	if !accessCookie.HttpOnly {
		t.Fatal("expected access cookie to be HttpOnly")
	}
}

func TestGenerateNewAuthCookie_SetsRefreshCookie(t *testing.T) {
	setupAuthJWT(t)

	_, refreshCookie := issueAuthCookies(t, "123")

	if refreshCookie.Name != RefreshTokenName {
		t.Fatalf("expected cookie name %q, got %q", RefreshTokenName, refreshCookie.Name)
	}
	if refreshCookie.Value == "" {
		t.Fatal("expected refresh cookie value to be non-empty")
	}
	if refreshCookie.Path != "/auth/" {
		t.Fatalf("expected refresh cookie path '/auth/', got %q", refreshCookie.Path)
	}
	if !refreshCookie.HttpOnly {
		t.Fatal("expected refresh cookie to be HttpOnly")
	}
}

func TestGetAuthCookie_ReturnsCookie(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "44")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(accessCookie)

	cookie, err := GetAuthCookie(req)
	if err != nil {
		t.Fatalf("GetAuthCookie returned error: %v", err)
	}
	if cookie.Value != accessCookie.Value {
		t.Fatal("returned cookie value does not match original")
	}
}

func TestIsAuth_ReturnsAuthorizedUser(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "44")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(accessCookie)

	ok, userID := IsAuth(req)
	if !ok {
		t.Fatal("expected request to be authorized")
	}
	if userID != "44" {
		t.Fatalf("expected userID %q, got %q", "44", userID)
	}
}

func TestIsAuth_RejectsMissingCookie(t *testing.T) {
	setupAuthJWT(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	ok, userID := IsAuth(req)
	if ok {
		t.Fatal("expected request without cookie to be unauthorized")
	}
	if userID != "" {
		t.Fatalf("expected empty userID, got %q", userID)
	}
}

func TestGetUserIDFromCookie_ReturnsIntUserID(t *testing.T) {
	setupAuthJWT(t)

	accessCookie, _ := issueAuthCookies(t, "77")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(accessCookie)

	userID, err := GetUserIDFromCookie(req)
	if err != nil {
		t.Fatalf("GetUserIDFromCookie returned error: %v", err)
	}
	if userID != 77 {
		t.Fatalf("expected userID 77, got %d", userID)
	}
}

func TestGetUserIDFromCookie_RejectsInvalidToken(t *testing.T) {
	setupAuthJWT(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  TokenName,
		Value: "broken-token",
	})

	_, err := GetUserIDFromCookie(req)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestGetUserIDFromCookie_RejectsNonNumericUserID(t *testing.T) {
	setupAuthJWT(t)

	token, err := jwt.GenerateToken("abc")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  TokenName,
		Value: token,
	})

	_, err = GetUserIDFromCookie(req)
	if err == nil {
		t.Fatal("expected error for non-numeric userID")
	}
	if !strings.Contains(err.Error(), strconv.ErrSyntax.Error()) {
		t.Fatalf("expected strconv syntax error, got %v", err)
	}
}

func TestRefreshToken_RotatesCookies(t *testing.T) {
	setupAuthJWT(t)

	oldAccessCookie, oldRefreshCookie := issueAuthCookies(t, "501")

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(oldRefreshCookie)

	rec := httptest.NewRecorder()
	ok, userID := RefreshToken(rec, req)

	if !ok {
		t.Fatal("expected refresh to succeed")
	}
	if userID != "501" {
		t.Fatalf("expected userID %q, got %q", "501", userID)
	}

	newCookies := rec.Result().Cookies()
	newAccessCookie := findCookie(newCookies, TokenName)
	newRefreshCookie := findCookie(newCookies, RefreshTokenName)

	if newAccessCookie == nil {
		t.Fatal("expected new access cookie after refresh")
	}
	if newRefreshCookie == nil {
		t.Fatal("expected new refresh cookie after refresh")
	}
	if newAccessCookie.Value == oldAccessCookie.Value {
		t.Fatal("expected access token to rotate")
	}
	if newRefreshCookie.Value == oldRefreshCookie.Value {
		t.Fatal("expected refresh token to rotate")
	}
}

func TestRefreshToken_OldRefreshTokenBecomesInvalid(t *testing.T) {
	setupAuthJWT(t)

	_, oldRefreshCookie := issueAuthCookies(t, "501")

	req1 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req1.AddCookie(oldRefreshCookie)

	rec1 := httptest.NewRecorder()
	ok, userID := RefreshToken(rec1, req1)
	if !ok || userID != "501" {
		t.Fatalf("expected first refresh to succeed, got ok=%v userID=%q", ok, userID)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req2.AddCookie(oldRefreshCookie)

	rec2 := httptest.NewRecorder()
	ok, userID = RefreshToken(rec2, req2)
	if ok {
		t.Fatal("expected old refresh token to be invalid after rotation")
	}
	if userID != "" {
		t.Fatalf("expected empty userID, got %q", userID)
	}
}

func TestClearOldToken_ExpiresCookies(t *testing.T) {
	setupAuthJWT(t)

	_, refreshCookie := issueAuthCookies(t, "808")

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(refreshCookie)

	rec := httptest.NewRecorder()
	ClearOldToken(rec, req)

	cookies := rec.Result().Cookies()
	clearedAccessCookie := findCookie(cookies, TokenName)
	clearedRefreshCookie := findCookie(cookies, RefreshTokenName)

	if clearedAccessCookie == nil {
		t.Fatal("expected cleared access cookie")
	}
	if clearedRefreshCookie == nil {
		t.Fatal("expected cleared refresh cookie")
	}
	if clearedAccessCookie.Value != "" {
		t.Fatal("expected cleared access cookie value to be empty")
	}
	if clearedRefreshCookie.Value != "" {
		t.Fatal("expected cleared refresh cookie value to be empty")
	}
	if !clearedAccessCookie.Expires.Before(refreshCookie.Expires) {
		t.Fatal("expected cleared access cookie to expire before original refresh cookie")
	}
}
