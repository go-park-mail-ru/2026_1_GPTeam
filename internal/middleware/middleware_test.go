package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	usermocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	authmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/mocks"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func TestCORSMiddleware_SetsHeaders(t *testing.T) {
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

			require.True(t, called)
			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
			require.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
			require.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
		})
	}
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

	require.False(t, called)
	require.NotEqual(t, http.StatusTeapot, w.Code)
}

func TestMethodValidationMiddleware_AllowedMethods(t *testing.T) {
	t.Parallel()

	handler := MethodValidationMiddleware(http.MethodGet, http.MethodPost)(http.HandlerFunc(okHandler))

	cases := []struct {
		name         string
		method       string
		expectedCode int
	}{
		{"GET разрешён", http.MethodGet, http.StatusOK},
		{"POST разрешён", http.MethodPost, http.StatusOK},
		{"DELETE не разрешён", http.MethodDelete, http.StatusMethodNotAllowed},
		{"PUT не разрешён", http.MethodPut, http.StatusMethodNotAllowed},
		{"PATCH не разрешён", http.MethodPatch, http.StatusMethodNotAllowed},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(c.method, "/", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, c.expectedCode, w.Code)
			if c.expectedCode == http.StatusMethodNotAllowed {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.EqualValues(t, http.StatusMethodNotAllowed, resp["code"])
				require.NotEmpty(t, resp["message"])
			}
		})
	}
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

func TestPanicMiddleware_RecoversPanic(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := PanicMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.EqualValues(t, http.StatusInternalServerError, resp["code"])
}

func TestPanicMiddleware_NoPanic(t *testing.T) {
	t.Parallel()

	handler := PanicMiddleware(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	testUser := &models.UserModel{
		Id:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	cases := []struct {
		name         string
		path         string
		method       string
		setupMocks   func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase)
		expectedCode int
		checkCtx     bool
	}{
		{
			name:   "/auth/login — публичный путь",
			path:   "/auth/login",
			method: http.MethodPost,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "/auth/refresh — публичный путь",
			path:   "/auth/refresh",
			method: http.MethodPost,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "/enums/types — публичный путь",
			path:   "/enums/types",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "/auth/logout — без токена → 401",
			path:   "/auth/logout",
			method: http.MethodPost,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any()).Return(false, -1)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "/profile — без токена → 401",
			path:   "/profile",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any()).Return(false, -1)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "/profile — токен валиден, пользователь не найден → 401",
			path:   "/profile",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any()).Return(true, 1)
				userApp.EXPECT().GetById(gomock.Any(), 1).Return(nil, errors.New("not found"))
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "/profile — валидный токен → 200, пользователь в контексте",
			path:   "/profile",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any()).Return(true, 1)
				userApp.EXPECT().GetById(gomock.Any(), 1).Return(testUser, nil)
			},
			expectedCode: http.StatusOK,
			checkCtx:     true,
		},
		{
			name:   "/get_budgets — без токена → 401",
			path:   "/get_budgets",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any()).Return(false, -1)
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authSvc := authmocks.NewMockAuthenticationService(ctrl)
			userApp := usermocks.NewMockUserUseCase(ctrl)
			c.setupMocks(authSvc, userApp)

			var gotCtxUser any
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotCtxUser = r.Context().Value("user")
				w.WriteHeader(http.StatusOK)
			})

			handler := AuthMiddleware(next, authSvc, userApp)
			req := httptest.NewRequest(c.method, c.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			require.Equal(t, c.expectedCode, w.Code)

			if c.expectedCode == http.StatusUnauthorized {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.EqualValues(t, http.StatusUnauthorized, resp["code"])
				require.NotEmpty(t, resp["message"])
			}

			if c.checkCtx {
				require.NotNil(t, gotCtxUser)
				user, ok := gotCtxUser.(*models.UserModel)
				require.True(t, ok)
				require.Equal(t, testUser.Username, user.Username)
			}
		})
	}
}

func TestAuthMiddleware_PreservesContext(t *testing.T) {
	t.Parallel()

	testUser := &models.UserModel{Id: 1, Username: "testuser"}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authSvc := authmocks.NewMockAuthenticationService(ctrl)
	userApp := usermocks.NewMockUserUseCase(ctrl)
	authSvc.EXPECT().IsAuth(gomock.Any()).Return(true, 1)
	userApp.EXPECT().GetById(gomock.Any(), 1).Return(testUser, nil)

	type ctxKey string
	const existingKey ctxKey = "existing"

	var gotExisting any
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotExisting = r.Context().Value(existingKey)
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(next, authSvc, userApp)
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req = req.WithContext(context.WithValue(req.Context(), existingKey, "hello"))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "hello", gotExisting)
}

func TestAuthAndMethodMiddlewareComposition(t *testing.T) {
	t.Parallel()

	testUser := &models.UserModel{Id: 1, Username: "testuser"}

	cases := []struct {
		name         string
		method       string
		setupMocks   func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase)
		expectedCode int
	}{
		{
			name:   "GET без токена → 401",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any()).Return(false, -1)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "POST с токеном → 405",
			method: http.MethodPost,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any()).Return(true, 1)
				userApp.EXPECT().GetById(gomock.Any(), 1).Return(testUser, nil)
			},
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "GET с токеном → 200",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *usermocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any()).Return(true, 1)
				userApp.EXPECT().GetById(gomock.Any(), 1).Return(testUser, nil)
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authSvc := authmocks.NewMockAuthenticationService(ctrl)
			userApp := usermocks.NewMockUserUseCase(ctrl)
			c.setupMocks(authSvc, userApp)

			mux := http.NewServeMux()
			mux.Handle("/profile", MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(okHandler)))
			handler := AuthMiddleware(mux, authSvc, userApp)

			req := httptest.NewRequest(c.method, "/profile", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}
