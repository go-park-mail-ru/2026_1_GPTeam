package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
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

func TestNoDirListing(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		path         string
		expectedCode int
		nextCalled   bool
	}{
		{
			name:         "/img/ — блокируется",
			path:         "/img/",
			expectedCode: http.StatusNotFound,
			nextCalled:   false,
		},
		{
			name:         "/img/photo.png — пропускается",
			path:         "/img/photo.png",
			expectedCode: http.StatusOK,
			nextCalled:   true,
		},
		{
			name:         "/static/ — пропускается (не /img/)",
			path:         "/static/",
			expectedCode: http.StatusOK,
			nextCalled:   true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			handler := NoDirListing(next)
			req := httptest.NewRequest(http.MethodGet, c.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			require.Equal(t, c.expectedCode, w.Code)
			require.Equal(t, c.nextCalled, called)
		})
	}
}

func TestSetUserCtx_GetUserCtx(t *testing.T) {
	t.Parallel()

	t.Run("set и get возвращают того же пользователя", func(t *testing.T) {
		t.Parallel()

		user := models.UserModel{Id: 1, Username: "testuser", Email: "test@example.com"}
		ctx := SetUserCtx(context.Background(), user)
		got, ok := GetUserCtx(ctx)

		require.True(t, ok)
		require.Equal(t, user.Id, got.Id)
		require.Equal(t, user.Username, got.Username)
	})

	t.Run("get из пустого контекста возвращает false", func(t *testing.T) {
		t.Parallel()

		_, ok := GetUserCtx(context.Background())
		require.False(t, ok)
	})
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

func TestPanicMiddleware_WithRequestId(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic with id")
	})

	handler := PanicMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), "request_id", "test-req-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
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
		setupMocks   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase)
		expectedCode int
		checkCtx     bool
	}{
		{
			name:   "/auth/login — публичный путь",
			path:   "/auth/login",
			method: http.MethodPost,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "/auth/refresh — публичный путь",
			path:   "/auth/refresh",
			method: http.MethodPost,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "/enums/types — публичный путь",
			path:   "/enums/types",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "/healthz — публичный путь",
			path:   "/healthz",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "/auth/logout — без токена → 401",
			path:   "/auth/logout",
			method: http.MethodPost,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any(), gomock.Any()).Return(false, -1)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "/profile — без токена → 401",
			path:   "/profile",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any(), gomock.Any()).Return(false, -1)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "/profile — токен валиден, пользователь не найден → 401",
			path:   "/profile",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any(), gomock.Any()).Return(true, 1)
				userApp.EXPECT().GetById(gomock.Any(), 1).Return(nil, errors.New("not found"))
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "/profile — валидный токен → 200, пользователь в контексте",
			path:   "/profile",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any(), gomock.Any()).Return(true, 1)
				userApp.EXPECT().GetById(gomock.Any(), 1).Return(testUser, nil)
			},
			expectedCode: http.StatusOK,
			checkCtx:     true,
		},
		{
			name:   "/get_budgets — без токена → 401",
			path:   "/get_budgets",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any(), gomock.Any()).Return(false, -1)
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
			userApp := appmocks.NewMockUserUseCase(ctrl)
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
	userApp := appmocks.NewMockUserUseCase(ctrl)
	authSvc.EXPECT().IsAuth(gomock.Any(), gomock.Any()).Return(true, 1)
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
		setupMocks   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase)
		expectedCode int
	}{
		{
			name:   "GET без токена → 401",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any(), gomock.Any()).Return(false, -1)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "POST с токеном → 405",
			method: http.MethodPost,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any(), gomock.Any()).Return(true, 1)
				userApp.EXPECT().GetById(gomock.Any(), 1).Return(testUser, nil)
			},
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "GET с токеном → 200",
			method: http.MethodGet,
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				authSvc.EXPECT().IsAuth(gomock.Any(), gomock.Any()).Return(true, 1)
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
			userApp := appmocks.NewMockUserUseCase(ctrl)
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

func TestAccessLogMiddleware(t *testing.T) {
	t.Parallel()

	var gotReqId string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqId, ok := r.Context().Value("request_id").(string); ok {
			gotReqId = reqId
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	handler := AccessLogMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NotEmpty(t, gotReqId)
	require.Equal(t, "created", w.Body.String())
}

func TestAccessLogMiddleware_ImplicitStatus(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("implicit ok"))
	})

	handler := AccessLogMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// responseWriter должен корректно перехватить код 200 OK при прямом Write
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "implicit ok", w.Body.String())
}

// Заглушка для тестирования CSRFMiddleware без необходимости генерации мока
type fakeCsrf struct {
	validSite bool
	accToken  string
	accErr    error
	cookie    string
	validCsrf bool
	csrfErr   error
	header    string
}

func (m *fakeCsrf) ValidateSecFetchSite(r *http.Request) bool {
	return m.validSite
}
func (m *fakeCsrf) GetAccessToken(ctx context.Context, r *http.Request) (string, error) {
	return m.accToken, m.accErr
}
func (m *fakeCsrf) GetCsrfFromCookie(ctx context.Context, r *http.Request) string {
	return m.cookie
}
func (m *fakeCsrf) ValidateCsrf(ctx context.Context, csrfToken string, accessToken string) (bool, error) {
	return m.validCsrf, m.csrfErr
}
func (m *fakeCsrf) GetCsrfFromHeader(r *http.Request) string {
	return m.header
}
func (m *fakeCsrf) SetCsrfCookie(ctx context.Context, w http.ResponseWriter, r *http.Request) {}

func TestCSRFMiddleware(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		path         string
		method       string
		mockSetup    fakeCsrf
		expectedCode int
	}{
		{
			name:         "Путь /auth/ пропускается",
			path:         "/auth/login",
			method:       http.MethodPost,
			mockSetup:    fakeCsrf{validSite: false}, // Игнорируется
			expectedCode: http.StatusOK,
		},
		{
			name:         "Путь /healthz пропускается",
			path:         "/healthz",
			method:       http.MethodGet,
			mockSetup:    fakeCsrf{validSite: false},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Невалидный Sec-Fetch-Site -> 403",
			path:         "/api/data",
			method:       http.MethodPost,
			mockSetup:    fakeCsrf{validSite: false},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "GET запрос -> пропускается проверка токенов, устанавливается кука",
			path:         "/api/data",
			method:       http.MethodGet,
			mockSetup:    fakeCsrf{validSite: true},
			expectedCode: http.StatusOK,
		},
		{
			name:         "POST без Access Token -> 403",
			path:         "/api/data",
			method:       http.MethodPost,
			mockSetup:    fakeCsrf{validSite: true, accErr: errors.New("no token")},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "POST без CSRF Cookie -> 403",
			path:         "/api/data",
			method:       http.MethodPost,
			mockSetup:    fakeCsrf{validSite: true, accToken: "token", cookie: ""},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "POST невалидный CSRF -> 403",
			path:         "/api/data",
			method:       http.MethodPost,
			mockSetup:    fakeCsrf{validSite: true, accToken: "token", cookie: "cookie", validCsrf: false},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "POST без CSRF Header -> 403",
			path:         "/api/data",
			method:       http.MethodPost,
			mockSetup:    fakeCsrf{validSite: true, accToken: "token", cookie: "cookie", validCsrf: true, header: ""},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "POST Header не совпадает с Cookie -> 403",
			path:         "/api/data",
			method:       http.MethodPost,
			mockSetup:    fakeCsrf{validSite: true, accToken: "token", cookie: "cookie", validCsrf: true, header: "wrong"},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "POST успешный -> 200",
			path:         "/api/data",
			method:       http.MethodPost,
			mockSetup:    fakeCsrf{validSite: true, accToken: "token", cookie: "cookie", validCsrf: true, header: "cookie"},
			expectedCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			csrfSvc := &c.mockSetup
			handler := CSRFMiddleware(http.HandlerFunc(okHandler), csrfSvc)
			req := httptest.NewRequest(c.method, c.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestCSPMiddleware(t *testing.T) {
	testCases := []struct {
		name   string
		method string
		path   string
		code   int
	}{
		{
			name:   "simple",
			method: http.MethodGet,
			path:   "/",
			code:   http.StatusOK,
		},
		{
			name:   "post",
			method: http.MethodPost,
			path:   "/",
			code:   http.StatusOK,
		},
		{
			name:   "/auth/",
			method: http.MethodGet,
			path:   "/auth/login",
			code:   http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			handler := CSPMiddleware(http.HandlerFunc(okHandler))
			r := httptest.NewRequest(testCase.method, testCase.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			require.Equal(t, testCase.code, w.Result().StatusCode)
			require.NotEmpty(t, w.Result().Header.Get("Content-Security-Policy"))
		})
	}
}

func TestOnlyStaffMiddleware(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(userApp *appmocks.MockUserUseCase)
		ctx    context.Context
		method string
		path   string
		code   int
	}{
		{
			name: "staff get",
			setup: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().IsStaff(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			ctx:    context.WithValue(context.Background(), "user", models.UserModel{IsStaff: true}),
			method: http.MethodGet,
			path:   "/staff",
			code:   http.StatusOK,
		},
		{
			name: "staff post",
			setup: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().IsStaff(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			ctx:    context.WithValue(context.Background(), "user", models.UserModel{IsStaff: true}),
			method: http.MethodPost,
			path:   "/staff",
			code:   http.StatusOK,
		},
		{
			name: "not staff get",
			setup: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().IsStaff(gomock.Any(), gomock.Any()).Return(false, nil)
			},
			ctx:    context.WithValue(context.Background(), "user", models.UserModel{IsStaff: false}),
			method: http.MethodGet,
			path:   "/staff",
			code:   http.StatusNotFound,
		},
		{
			name:   "unauthorized",
			setup:  func(userApp *appmocks.MockUserUseCase) {},
			ctx:    context.Background(),
			method: http.MethodGet,
			path:   "/staff",
			code:   http.StatusUnauthorized,
		},
		{
			name: "empty user",
			setup: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().IsStaff(gomock.Any(), gomock.Any()).Return(false, nil)
			},
			ctx:    context.WithValue(context.Background(), "user", models.UserModel{}),
			method: http.MethodGet,
			path:   "/staff",
			code:   http.StatusNotFound,
		},
		{
			name: "fail",
			setup: func(userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().IsStaff(gomock.Any(), gomock.Any()).Return(false, repository.NothingInTableError)
			},
			ctx:    context.WithValue(context.Background(), "user", models.UserModel{IsStaff: true}),
			method: http.MethodGet,
			path:   "/staff",
			code:   http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			userApp := appmocks.NewMockUserUseCase(ctrl)
			testCase.setup(userApp)
			handler := OnlyStaffMiddleware(http.HandlerFunc(okHandler), userApp)
			r := httptest.NewRequest(testCase.method, testCase.path, nil)
			r = r.WithContext(testCase.ctx)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			require.Equal(t, testCase.code, w.Result().StatusCode)
		})
	}
}
