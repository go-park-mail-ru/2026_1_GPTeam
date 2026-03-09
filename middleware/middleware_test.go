package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
	testhelper "github.com/go-park-mail-ru/2026_1_GPTeam/pkg"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"

	"github.com/stretchr/testify/require"
)

const (
	testUsername = "middlewareAdmin"
	testPassword = "Adm1n123"
	testUserID   = "0"
)

var once sync.Once

func setupStorage() {
	once.Do(func() {
		_ = jwt.NewRefreshTokenStore("secret123")
		storage.NewUserStore()
		storage.AddUser(storage.UserInfo{
			Id:        0,
			Username:  testUsername,
			Password:  testPassword,
			Email:     "middleware@test.com",
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		})
	})
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func makeAccessCookie(t *testing.T, userID string) *http.Cookie {
	t.Helper()
	tokenStr, err := jwt.GenerateToken(userID)
	require.NoError(t, err)
	return &http.Cookie{
		Name:     auth.TokenName,
		Value:    tokenStr,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(jwt.AccessTokenExpirationTime),
	}
}

func TestCORSMiddleware_SetsHeaders(t *testing.T) {
	t.Parallel()

	handler := CORSMiddleware(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	require.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
	require.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORSMiddleware_OptionsDoesNotCallNext(t *testing.T) {
	t.Parallel()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})

	handler := CORSMiddleware(next)
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.False(t, called, "next не должен вызываться при OPTIONS")
	require.NotEqual(t, http.StatusTeapot, w.Code)
}

func TestCORSMiddleware_NonOptionsCallsNext(t *testing.T) {
	t.Parallel()

	cases := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range cases {
		method := method
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			handler := CORSMiddleware(next)
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.True(t, called, "next должен вызываться для метода %s", method)
			require.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestMethodValidationMiddleware_AllowedMethod(t *testing.T) {
	t.Parallel()

	handler := MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestMethodValidationMiddleware_DisallowedMethod_Returns405(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		allowed string
		sent    string
	}{
		{"GET вместо POST", http.MethodPost, http.MethodGet},
		{"POST вместо GET", http.MethodGet, http.MethodPost},
		{"DELETE вместо PUT", http.MethodPut, http.MethodDelete},
		{"PATCH вместо DELETE", http.MethodDelete, http.MethodPatch},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			handler := MethodValidationMiddleware(c.allowed)(http.HandlerFunc(okHandler))
			req := httptest.NewRequest(c.sent, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, http.StatusMethodNotAllowed, w.Code)

			var resp map[string]any
			require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
			require.EqualValues(t, http.StatusMethodNotAllowed, resp["code"])
			require.NotEmpty(t, resp["message"])
		})
	}
}

func TestMethodValidationMiddleware_MultipleAllowedMethods(t *testing.T) {
	t.Parallel()

	handler := MethodValidationMiddleware(http.MethodGet, http.MethodPost)(http.HandlerFunc(okHandler))

	t.Run("GET разрешён", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST разрешён", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DELETE не разрешён", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestMethodValidationMiddleware_405ResponseBody(t *testing.T) {
	t.Parallel()

	handler := MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.EqualValues(t, http.StatusMethodNotAllowed, resp["code"])
	require.NotEmpty(t, resp["message"])
}

func TestAuthMiddleware_PublicPaths_NoCookieRequired(t *testing.T) {
	t.Parallel()

	cases := []string{
		"/",
		"/login",
		"/signup",
		"/auth/login",
		"/auth/refresh",
		"/auth/somethingelse",
	}

	for _, path := range cases {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			handler := AuthMiddleware(http.HandlerFunc(okHandler))
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code, "путь %q должен быть доступен без токена", path)
		})
	}
}

func TestAuthMiddleware_LogoutIsNotPublic(t *testing.T) {
	t.Parallel()

	handler := AuthMiddleware(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_PrivatePaths_NoCookie_Returns401(t *testing.T) {
	t.Parallel()

	privatePaths := []string{
		"/profile/balance",
		"/get_budgets",
		"/budget",
		"/auth/logout",
	}

	for _, path := range privatePaths {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			handler := AuthMiddleware(http.HandlerFunc(okHandler))
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, http.StatusUnauthorized, w.Code)

			var resp map[string]any
			require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
			require.EqualValues(t, http.StatusUnauthorized, resp["code"])
			require.NotEmpty(t, resp["message"])
		})
	}
}

func TestAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	t.Parallel()

	handler := AuthMiddleware(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/profile/balance", nil)
	req.AddCookie(&http.Cookie{
		Name:     auth.TokenName,
		Value:    "totally.invalid.token",
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_TokenWithNonNumericUserID_Returns401(t *testing.T) {
	t.Parallel()
	setupStorage()

	tokenStr, err := jwt.GenerateToken("not-a-number")
	require.NoError(t, err)

	handler := AuthMiddleware(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/profile/balance", nil)
	req.AddCookie(&http.Cookie{
		Name:     auth.TokenName,
		Value:    tokenStr,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(jwt.AccessTokenExpirationTime),
	})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidTokenUnknownUser_Returns401(t *testing.T) {
	t.Parallel()
	setupStorage()

	tokenStr, err := jwt.GenerateToken("9999")
	require.NoError(t, err)

	handler := AuthMiddleware(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/profile/balance", nil)
	req.AddCookie(&http.Cookie{
		Name:     auth.TokenName,
		Value:    tokenStr,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(jwt.AccessTokenExpirationTime),
	})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken_UserAddedToContext(t *testing.T) {
	t.Parallel()
	setupStorage()

	var gotUser any
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = r.Context().Value("user")
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/profile/balance", nil)
	req.AddCookie(makeAccessCookie(t, testUserID))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, gotUser, "user должен быть добавлен в контекст")

	user, ok := gotUser.(storage.UserInfo)
	require.True(t, ok, "user в контексте должен иметь тип storage.UserInfo")
	require.Equal(t, testUsername, user.Username)
}

func TestAuthMiddleware_PublicPath_NoUserInContext(t *testing.T) {
	t.Parallel()

	var gotUser any
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = r.Context().Value("user")
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(next)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Nil(t, gotUser, "user НЕ должен быть в контексте для публичного пути")
}

func TestAuthMiddleware_401ResponseBody(t *testing.T) {
	t.Parallel()

	handler := AuthMiddleware(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/budget", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.EqualValues(t, http.StatusUnauthorized, resp["code"])
	require.NotEmpty(t, resp["message"])
}

func TestAuthMiddleware_AuthPrefixEdgeCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path       string
		wantPublic bool
	}{
		{"/auth/login", true},
		{"/auth/refresh", true},
		{"/auth/other", true},
		{"/auth/logout", false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.path, func(t *testing.T) {
			t.Parallel()

			handler := AuthMiddleware(http.HandlerFunc(okHandler))
			req := httptest.NewRequest(http.MethodPost, c.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if c.wantPublic {
				require.Equal(t, http.StatusOK, w.Code, "путь %q должен быть публичным", c.path)
			} else {
				require.Equal(t, http.StatusUnauthorized, w.Code, "путь %q должен требовать авторизацию", c.path)
			}
		})
	}
}

func TestAuthMiddleware_PreservesExistingContextValues(t *testing.T) {
	t.Parallel()
	setupStorage()

	type ctxKey string
	const existingKey ctxKey = "existing"

	var gotExisting any
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotExisting = r.Context().Value(existingKey)
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/profile/balance", nil)
	req = req.WithContext(context.WithValue(req.Context(), existingKey, "hello"))
	req.AddCookie(makeAccessCookie(t, testUserID))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "hello", gotExisting, "существующие значения контекста должны сохраняться")
}

func TestAuthAndMethodMiddlewareComposition(t *testing.T) {
	t.Parallel()
	setupStorage()

	_ = testhelper.MustJSON

	mux := http.NewServeMux()
	mux.Handle("/profile/balance", MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(okHandler)))
	composed := AuthMiddleware(mux)

	cases := []struct {
		name         string
		method       string
		withToken    bool
		expectedCode int
	}{
		{"GET без токена → 401", http.MethodGet, false, http.StatusUnauthorized},
		{"POST с токеном → 405", http.MethodPost, true, http.StatusMethodNotAllowed},
		{"GET с токеном → 200", http.MethodGet, true, http.StatusOK},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(c.method, "/profile/balance", nil)
			if c.withToken {
				req.AddCookie(makeAccessCookie(t, testUserID))
			}
			w := httptest.NewRecorder()
			composed.ServeHTTP(w, req)
			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}
