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
		_ = jwt.NewRefreshTokenStore("secret123", "v1")
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

func TestCORSMiddlewareSetsHeaders(t *testing.T) {
	t.Parallel()

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		method := method
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			handler := CORSMiddleware(next)
			req := httptest.NewRequest(method, "/some/path", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.True(t, called, "next должен вызываться для метода %s", method)
			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
			require.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
			require.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
		})
	}
}

func TestCORSMiddlewareOptionsDoesNotCallNext(t *testing.T) {
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

func TestMethodValidationMiddlewareMultipleAllowedMethods(t *testing.T) {
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

		var resp map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		require.EqualValues(t, http.StatusMethodNotAllowed, resp["code"])
		require.NotEmpty(t, resp["message"])
	})
}

func TestMethodValidationMiddlewareDisallowedMethodReturns405(t *testing.T) {
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

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()
	setupStorage()

	validCookie := makeAccessCookie(t, testUserID)

	tokenNonNumeric, err := jwt.GenerateToken("not-a-number")
	require.NoError(t, err)
	cookieNonNumeric := &http.Cookie{
		Name:     auth.TokenName,
		Value:    tokenNonNumeric,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(jwt.AccessTokenExpirationTime),
	}

	tokenUnknown, err := jwt.GenerateToken("9999")
	require.NoError(t, err)
	cookieUnknown := &http.Cookie{
		Name:     auth.TokenName,
		Value:    tokenUnknown,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(jwt.AccessTokenExpirationTime),
	}

	cases := []struct {
		name       string
		method     string
		path       string
		cookie     *http.Cookie
		assertFunc func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any)
	}{
		{
			name:   "публичный путь / без токена",
			method: http.MethodGet,
			path:   "/",
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusOK, w.Code)
				require.Nil(t, gotCtxUser)
			},
		},
		{
			name:   "публичный путь /login без токена",
			method: http.MethodGet,
			path:   "/login",
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusOK, w.Code)
				require.Nil(t, gotCtxUser)
			},
		},
		{
			name:   "публичный путь /signup без токена",
			method: http.MethodGet,
			path:   "/signup",
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusOK, w.Code)
				require.Nil(t, gotCtxUser)
			},
		},
		{
			name:   "публичный путь /auth/login без токена",
			method: http.MethodPost,
			path:   "/auth/login",
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusOK, w.Code)
				require.Nil(t, gotCtxUser)
			},
		},
		{
			name:   "публичный путь /auth/refresh без токена",
			method: http.MethodPost,
			path:   "/auth/refresh",
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusOK, w.Code)
				require.Nil(t, gotCtxUser)
			},
		},
		{
			name:   "публичный путь /auth/other без токена",
			method: http.MethodPost,
			path:   "/auth/other",
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusOK, w.Code)
				require.Nil(t, gotCtxUser)
			},
		},
		{
			name:   "/auth/logout без токена → 401",
			method: http.MethodPost,
			path:   "/auth/logout",
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
				require.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var resp map[string]any
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.EqualValues(t, http.StatusUnauthorized, resp["code"])
				require.NotEmpty(t, resp["message"])
			},
		},
		{
			name:   "/profile/balance без токена → 401",
			method: http.MethodGet,
			path:   "/profile/balance",
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
				var resp map[string]any
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.EqualValues(t, http.StatusUnauthorized, resp["code"])
				require.NotEmpty(t, resp["message"])
			},
		},
		{
			name:   "/get_budgets без токена → 401",
			method: http.MethodGet,
			path:   "/get_budgets",
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name:   "/budget без токена → 401 с JSON телом",
			method: http.MethodGet,
			path:   "/budget",
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
				require.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var resp map[string]any
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.EqualValues(t, http.StatusUnauthorized, resp["code"])
				require.NotEmpty(t, resp["message"])
			},
		},
		{
			name:   "невалидный токен → 401",
			method: http.MethodGet,
			path:   "/profile/balance",
			cookie: &http.Cookie{
				Name:     auth.TokenName,
				Value:    "totally.invalid.token",
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			},
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name:   "токен с нечисловым userID → 401",
			method: http.MethodGet,
			path:   "/profile/balance",
			cookie: cookieNonNumeric,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name:   "валидный токен неизвестного пользователя → 401",
			method: http.MethodGet,
			path:   "/profile/balance",
			cookie: cookieUnknown,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name:   "валидный токен → пользователь добавлен в контекст",
			method: http.MethodGet,
			path:   "/profile/balance",
			cookie: validCookie,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, gotCtxUser any) {
				require.Equal(t, http.StatusOK, w.Code)
				require.NotNil(t, gotCtxUser)
				user, ok := gotCtxUser.(storage.UserInfo)
				require.True(t, ok)
				require.Equal(t, testUsername, user.Username)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			var gotCtxUser any
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotCtxUser = r.Context().Value("user")
				w.WriteHeader(http.StatusOK)
			})

			handler := AuthMiddleware(next)
			req := httptest.NewRequest(c.method, c.path, nil)
			if c.cookie != nil {
				req.AddCookie(c.cookie)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			c.assertFunc(t, w, req, gotCtxUser)
		})
	}
}

func TestAuthMiddlewarePreservesExistingContextValues(t *testing.T) {
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
