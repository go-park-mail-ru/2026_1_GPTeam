package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	models2 "github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	jwt2 "github.com/go-park-mail-ru/2026_1_GPTeam/auth/jwt_auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
	"github.com/go-park-mail-ru/2026_1_GPTeam/middleware"
	"github.com/go-park-mail-ru/2026_1_GPTeam/models"
	testhelper "github.com/go-park-mail-ru/2026_1_GPTeam/pkg"
	"github.com/go-park-mail-ru/2026_1_GPTeam/web/web_helpers"

	"github.com/stretchr/testify/require"
)

const (
	testUsername = "admin"
	testPassword = "Adm1n123"
)

var once sync.Once

func SetupStorage() {
	once.Do(func() {
		_ = jwt.NewRefreshTokenStore("secret123", "0")
		models.NewUserStore()
		models.AddUser(models2.UserModel{
			Id:              0,
			Username:        testUsername,
			Password:        testPassword,
			Email:           "email@example.com",
			CreatedAt:       time.Now(),
			LastLogin:       time.Now(),
			AvatarUrl:       "img/123.png",
			Balance:         100.5,
			BalanceCurrency: "RUB",
		})
		models.NewBudgetStore()
	})
}

func loginAndGetCookies(t *testing.T, handler http.Handler) []*http.Cookie {
	t.Helper()
	body, err := json.Marshal(map[string]string{"username": testUsername, "password": testPassword})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	return w.Result().Cookies()
}

func assertAuthCookies(t *testing.T, cookies []*http.Cookie) {
	t.Helper()
	cookieMap := make(map[string]*http.Cookie)
	for _, c := range cookies {
		cookieMap[c.Name] = c
	}
	for name, expectedPath := range map[string]string{
		auth.TokenName:        "/",
		auth.RefreshTokenName: "/auth/",
	} {
		c, ok := cookieMap[name]
		require.True(t, ok, "кука %q отсутствует в ответе", name)
		require.Equal(t, expectedPath, c.Path)
		require.NotEmpty(t, c.Value)
		require.True(t, c.HttpOnly)
		require.True(t, c.Secure)
		require.Equal(t, http.SameSiteLaxMode, c.SameSite)
	}
}

func newMux(routes map[string]struct {
	method  string
	handler http.HandlerFunc
}) http.Handler {
	mux := http.NewServeMux()
	for path, r := range routes {
		mux.Handle(path, middleware.MethodValidationMiddleware(r.method)(r.handler))
	}
	return middleware.AuthMiddleware(mux)
}

func TestLogin(t *testing.T) {
	SetupStorage()

	type tc struct {
		name         string
		method       string
		body         []byte
		expectedCode int
		assertFunc   func(*testing.T, *httptest.ResponseRecorder)
	}

	cases := []tc{
		{
			name:         "GET метод не разрешён",
			method:       http.MethodGet,
			body:         nil,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "невалидный JSON",
			method:       http.MethodPost,
			body:         []byte(`не валидный json`),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "неверные credentials",
			method:       http.MethodPost,
			body:         testhelper.MustJSON(t, map[string]string{"username": "", "password": ""}),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "успешный логин куки выставлены",
			method:       http.MethodPost,
			body:         testhelper.MustJSON(t, map[string]string{"username": testUsername, "password": testPassword}),
			expectedCode: http.StatusOK,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assertAuthCookies(t, w.Result().Cookies())
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			handler := middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(loginHandler))
			req := httptest.NewRequest(c.method, "/auth/login", bytes.NewBuffer(c.body))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, c.expectedCode, w.Code)
			if c.assertFunc != nil {
				c.assertFunc(t, w)
			}
		})
	}
}

func TestSignup(t *testing.T) {
	SetupStorage()

	type tc struct {
		name         string
		method       string
		body         []byte
		expectedCode int
		assertFunc   func(*testing.T, *httptest.ResponseRecorder)
	}

	cases := []tc{
		{
			name:         "GET метод не разрешён",
			method:       http.MethodGet,
			body:         nil,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "невалидный JSON",
			method:       http.MethodPost,
			body:         []byte(`не валидный json`),
			expectedCode: http.StatusBadRequest,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp web_helpers.SignupErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.Equal(t, http.StatusBadRequest, resp.Code)
				require.Equal(t, "Неверный формат запроса", resp.Message)
				require.Equal(t, []web_helpers.FieldError{
					{Field: "", Message: "Не удалось прочитать тело запроса"},
				}, resp.Errors)
			},
		},
		{
			name:         "все поля пустые",
			method:       http.MethodPost,
			body:         testhelper.MustJSON(t, map[string]string{"username": "", "password": "", "email": "", "confirm_password": ""}),
			expectedCode: http.StatusBadRequest,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp web_helpers.SignupErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.Equal(t, http.StatusBadRequest, resp.Code)
				require.Equal(t, "Неверный формат запроса", resp.Message)
				require.Equal(t, []web_helpers.FieldError{
					{Field: "username", Message: "Поле обязательно для заполнения"},
					{Field: "password", Message: "Поле обязательно для заполнения"},
					{Field: "email", Message: "Поле обязательно для заполнения"},
					{Field: "confirm_password", Message: "Поле обязательно для заполнения"},
				}, resp.Errors)
			},
		},
		{
			name:         "username уже занят",
			method:       http.MethodPost,
			body:         testhelper.MustJSON(t, map[string]string{"username": testUsername, "password": testPassword, "email": "new@email.com", "confirm_password": testPassword}),
			expectedCode: http.StatusConflict,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp web_helpers.SignupErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.Equal(t, http.StatusConflict, resp.Code)
				require.Equal(t, "Пользователь с таким логином уже существует", resp.Message)
				require.Equal(t, []web_helpers.FieldError{
					{Field: "username", Message: "Пользователь с таким логином уже существует"},
				}, resp.Errors)
			},
		},
		{
			name:         "email уже занят",
			method:       http.MethodPost,
			body:         testhelper.MustJSON(t, map[string]string{"username": "admin2", "password": testPassword, "email": "email@example.com", "confirm_password": testPassword}),
			expectedCode: http.StatusConflict,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp web_helpers.SignupErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.Equal(t, http.StatusConflict, resp.Code)
				require.Equal(t, "Пользователь с таким email уже существует", resp.Message)
				require.Equal(t, []web_helpers.FieldError{
					{Field: "email", Message: "Пользователь с таким email уже существует"},
				}, resp.Errors)
			},
		},
		{
			name:         "пароли не совпадают",
			method:       http.MethodPost,
			body:         testhelper.MustJSON(t, map[string]string{"username": "admin2", "password": testPassword, "email": "email2@example.com", "confirm_password": "Adm1n123456"}),
			expectedCode: http.StatusBadRequest,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp web_helpers.SignupErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.Equal(t, http.StatusBadRequest, resp.Code)
				require.Equal(t, "Ошибка валидации", resp.Message)
				require.Equal(t, []web_helpers.FieldError{
					{Field: "password", Message: "Пароли не совпадают"},
					{Field: "confirm_password", Message: "Пароли не совпадают"},
				}, resp.Errors)
			},
		},
		{
			name:         "успешная регистрация куки выставлены",
			method:       http.MethodPost,
			body:         testhelper.MustJSON(t, map[string]string{"username": "newuser", "password": "Admin123", "email": "new@example.com", "confirm_password": "Admin123"}),
			expectedCode: http.StatusOK,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assertAuthCookies(t, w.Result().Cookies())
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			handler := middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(signupHandler))
			req := httptest.NewRequest(c.method, "/auth/signup", bytes.NewBuffer(c.body))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, c.expectedCode, w.Code)
			if c.assertFunc != nil {
				c.assertFunc(t, w)
			}
		})
	}
}

func TestRefreshMethodNotAllowed(t *testing.T) {
	SetupStorage()

	cases := []struct {
		name   string
		method string
	}{
		{"GET", http.MethodGet},
		{"PUT", http.MethodPut},
		{"DELETE", http.MethodDelete},
	}

	for _, c := range cases {
		t.Run(c.name+" метод не разрешён", func(t *testing.T) {
			t.Parallel()
			handler := middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(refreshTokenHandler))
			req := httptest.NewRequest(c.method, "/auth/refresh", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

func TestRefreshWithValidToken(t *testing.T) {
	SetupStorage()

	handler := newMux(map[string]struct {
		method  string
		handler http.HandlerFunc
	}{
		"/auth/login":   {http.MethodPost, loginHandler},
		"/auth/refresh": {http.MethodPost, refreshTokenHandler},
	})

	cookies := loginAndGetCookies(t, handler)
	assertAuthCookies(t, cookies)

	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == auth.RefreshTokenName {
			refreshCookie = c
		}
	}
	require.NotNil(t, refreshCookie)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(refreshCookie)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assertAuthCookies(t, w.Result().Cookies())
}

func TestRefreshWithInvalidToken(t *testing.T) {
	SetupStorage()

	handler := middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(refreshTokenHandler))
	incorrectToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzM0ODg2ODQsImlkIjoiYTc3OGQ1YzktZjY1MS00NjQ5LWI3MGQtY2QxNWFiZmJhYmEwIiwidXNlcl9pZCI6IjEifQ.pIIPP8Mb_yqg_37OmZVZERJxMDKboau0xsgYVsfpgxs"

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:     auth.RefreshTokenName,
		Value:    incorrectToken,
		Path:     "/auth/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(jwt2.AccessTokenExpirationTime),
	})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogout(t *testing.T) {
	SetupStorage()

	handler := newMux(map[string]struct {
		method  string
		handler http.HandlerFunc
	}{
		"/auth/login":  {http.MethodPost, loginHandler},
		"/auth/logout": {http.MethodPost, logoutHandler},
	})

	cases := []struct {
		name         string
		method       string
		withAuth     bool
		expectedCode int
	}{
		{"GET без авторизации", http.MethodGet, false, http.StatusUnauthorized},
		{"POST без авторизации", http.MethodPost, false, http.StatusUnauthorized},
		{"POST с авторизацией", http.MethodPost, true, http.StatusOK},
	}

	cookies := loginAndGetCookies(t, handler)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(c.method, "/auth/logout", nil)
			if c.withAuth {
				for _, cookie := range cookies {
					req.AddCookie(cookie)
				}
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestBalance(t *testing.T) {
	SetupStorage()

	handler := newMux(map[string]struct {
		method  string
		handler http.HandlerFunc
	}{
		"/auth/login":      {http.MethodPost, loginHandler},
		"/profile/balance": {http.MethodGet, balanceHandler},
	})

	cookies := loginAndGetCookies(t, handler)

	cases := []struct {
		name         string
		method       string
		withAuth     bool
		expectedCode int
		assertFunc   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "POST метод не разрешён",
			method:       http.MethodPost,
			withAuth:     true,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "GET без авторизации",
			method:       http.MethodGet,
			withAuth:     false,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "GET с авторизацией",
			method:       http.MethodGet,
			withAuth:     true,
			expectedCode: http.StatusOK,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp web_helpers.BalanceResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.Equal(t, http.StatusOK, resp.Code)
				require.Equal(t, 100.5, resp.Balance)
				require.Equal(t, "RUB", resp.Currency)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(c.method, "/profile/balance", nil)
			if c.withAuth {
				for _, cookie := range cookies {
					req.AddCookie(cookie)
				}
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, c.expectedCode, w.Code)
			if c.assertFunc != nil {
				c.assertFunc(t, w)
			}
		})
	}
}

func TestGetBudgets(t *testing.T) {
	SetupStorage()

	handler := newMux(map[string]struct {
		method  string
		handler http.HandlerFunc
	}{
		"/auth/login":  {http.MethodPost, loginHandler},
		"/get_budgets": {http.MethodGet, getBudgetsHandler},
	})

	cookies := loginAndGetCookies(t, handler)

	cases := []struct {
		name         string
		method       string
		withAuth     bool
		expectedCode int
	}{
		{"GET без авторизации", http.MethodGet, false, http.StatusUnauthorized},
		{"POST без авторизации", http.MethodPost, false, http.StatusUnauthorized},
		{"GET с авторизацией", http.MethodGet, true, http.StatusOK},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(c.method, "/get_budgets", nil)
			if c.withAuth {
				for _, cookie := range cookies {
					req.AddCookie(cookie)
				}
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestGetBudget(t *testing.T) {
	SetupStorage()

	handler := newMux(map[string]struct {
		method  string
		handler http.HandlerFunc
	}{
		"/auth/login":      {http.MethodPost, loginHandler},
		"/budget":          {http.MethodPost, createBudgetHandler},
		"/get_budget/{id}": {http.MethodGet, getBudgetHandler},
	})

	cookies := loginAndGetCookies(t, handler)

	budgetBody := testhelper.MustJSON(t, map[string]any{"title": "Тестовый бюджет", "description": "text", "target": 5000, "currency": "RUB", "start_at": time.Now()})
	reqCreate := httptest.NewRequest(http.MethodPost, "/budget", bytes.NewBuffer(budgetBody))
	for _, c := range cookies {
		reqCreate.AddCookie(c)
	}
	wCreate := httptest.NewRecorder()
	handler.ServeHTTP(wCreate, reqCreate)
	require.Equal(t, http.StatusOK, wCreate.Code)

	var createResp web_helpers.BudgetCreateSuccessResponse
	require.NoError(t, json.NewDecoder(wCreate.Body).Decode(&createResp))
	budgetID := createResp.BudgetID

	cases := []struct {
		name         string
		id           string
		withAuth     bool
		expectedCode int
	}{
		{"без авторизации", "0", false, http.StatusUnauthorized},
		{"несуществующий ID", "999", true, http.StatusNotFound},
		{"невалидный ID", "abc", true, http.StatusNotFound},
		{"существующий бюджет", fmt.Sprintf("%d", budgetID), true, http.StatusOK},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "/get_budget/"+c.id, nil)
			if c.withAuth {
				for _, cookie := range cookies {
					req.AddCookie(cookie)
				}
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestCreateBudget(t *testing.T) {
	SetupStorage()

	handler := newMux(map[string]struct {
		method  string
		handler http.HandlerFunc
	}{
		"/auth/login": {http.MethodPost, loginHandler},
		"/budget":     {http.MethodPost, createBudgetHandler},
	})

	cookies := loginAndGetCookies(t, handler)

	cases := []struct {
		name         string
		method       string
		body         []byte
		withAuth     bool
		expectedCode int
		assertFunc   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "GET метод не разрешён",
			method:       http.MethodGet,
			body:         nil,
			withAuth:     true,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "без авторизации",
			method:       http.MethodPost,
			body:         testhelper.MustJSON(t, map[string]any{"title": "test", "target": 100, "currency": "RUB"}),
			withAuth:     false,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "невалидный JSON",
			method:       http.MethodPost,
			body:         []byte(`не валидный json`),
			withAuth:     true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "пустые обязательные поля",
			method:       http.MethodPost,
			body:         testhelper.MustJSON(t, map[string]any{"title": "", "description": "", "target": 0, "currency": "", "start_at": time.Time{}}),
			withAuth:     true,
			expectedCode: http.StatusBadRequest,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp web_helpers.BudgetErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.Equal(t, "Ошибка валидации", resp.Message)
				fieldNames := make([]string, 0, len(resp.Errors))
				for _, e := range resp.Errors {
					fieldNames = append(fieldNames, e.Field)
				}
				require.ElementsMatch(t, []string{"title", "description", "target", "target", "currency", "currency", "start_at"}, fieldNames)
			},
		},
		{
			name:         "успешное создание",
			method:       http.MethodPost,
			body:         testhelper.MustJSON(t, map[string]any{"title": "Отпуск", "description": "text", "target": 10000, "currency": "RUB", "start_at": time.Now().AddDate(0, 0, 1)}),
			withAuth:     true,
			expectedCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(c.method, "/budget", bytes.NewBuffer(c.body))
			if c.withAuth {
				for _, cookie := range cookies {
					req.AddCookie(cookie)
				}
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, c.expectedCode, w.Code)
			if c.assertFunc != nil {
				c.assertFunc(t, w)
			}
		})
	}
}

func TestDeleteBudget(t *testing.T) {
	SetupStorage()

	handler := newMux(map[string]struct {
		method  string
		handler http.HandlerFunc
	}{
		"/auth/login":  {http.MethodPost, loginHandler},
		"/budget":      {http.MethodPost, createBudgetHandler},
		"/budget/{id}": {http.MethodDelete, DeleteBudgetHandler},
	})

	cookies := loginAndGetCookies(t, handler)

	budgetBody := testhelper.MustJSON(t, map[string]any{"title": "Бюджет для удаления", "description": "text", "target": 1000, "currency": "RUB", "start_at": time.Now()})
	reqCreate := httptest.NewRequest(http.MethodPost, "/budget", bytes.NewBuffer(budgetBody))
	for _, c := range cookies {
		reqCreate.AddCookie(c)
	}
	wCreate := httptest.NewRecorder()
	handler.ServeHTTP(wCreate, reqCreate)
	require.Equal(t, http.StatusOK, wCreate.Code)

	var createResp web_helpers.BudgetCreateSuccessResponse
	require.NoError(t, json.NewDecoder(wCreate.Body).Decode(&createResp))
	budgetID := fmt.Sprintf("%d", createResp.BudgetID)

	cases := []struct {
		name         string
		id           string
		withAuth     bool
		expectedCode int
	}{
		{"без авторизации", budgetID, false, http.StatusUnauthorized},
		{"невалидный ID", "abc", true, http.StatusNotFound},
		{"несуществующий бюджет", "999", true, http.StatusNotFound},
		{"успешное удаление", budgetID, true, http.StatusOK},
		{"повторное удаление", budgetID, true, http.StatusNotFound},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/budget/"+c.id, nil)
			if c.withAuth {
				for _, cookie := range cookies {
					req.AddCookie(cookie)
				}
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestHandlersNoUserInContext(t *testing.T) {
	SetupStorage()

	handlers := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		body    []byte
	}{
		{"getBudgetsHandler", getBudgetsHandler, http.MethodGet, nil},
		{"getBudgetHandler", getBudgetHandler, http.MethodGet, nil},
		{"createBudgetHandler", createBudgetHandler, http.MethodPost, []byte(`{"title":"t","target":1,"currency":"RUB"}`)},
		{"DeleteBudgetHandler", DeleteBudgetHandler, http.MethodDelete, nil},
		{"balanceHandler", balanceHandler, http.MethodGet, nil},
	}

	for _, h := range handlers {
		t.Run(h.name+" без юзера в контексте", func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(h.method, "/", bytes.NewBuffer(h.body))
			w := httptest.NewRecorder()
			h.handler(w, req)
			require.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestHandlersEmptyPathID(t *testing.T) {
	SetupStorage()

	user := models2.UserModel{
		Id:       0,
		Username: testUsername,
	}

	withUser := func(r *http.Request) *http.Request {
		ctx := context.WithValue(r.Context(), "user", user)
		return r.WithContext(ctx)
	}

	t.Run("getBudgetHandler пустой id", func(t *testing.T) {
		t.Parallel()
		req := withUser(httptest.NewRequest(http.MethodGet, "/get_budget/", nil))
		w := httptest.NewRecorder()
		getBudgetHandler(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("DeleteBudgetHandler пустой id", func(t *testing.T) {
		t.Parallel()
		req := withUser(httptest.NewRequest(http.MethodDelete, "/budget/", nil))
		w := httptest.NewRecorder()
		DeleteBudgetHandler(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	})
}
