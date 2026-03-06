package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
	"github.com/go-park-mail-ru/2026_1_GPTeam/middleware"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"

	"github.com/stretchr/testify/require"
)

var once sync.Once

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/auth/login", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(loginHandler)))
	mux.Handle("/signup", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(signupHandler)))
	mux.Handle("/auth/logout", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(logoutHandler)))
	mux.Handle("/auth/refresh", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(refreshTokenHandler)))
	mux.Handle("/profile/balance", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(balanceHandler)))
	mux.Handle("/get_budgets", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(GetBudgetsHandler)))
	mux.Handle("/get_budget/{id}", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(GetBudgetHandler)))
	mux.Handle("/budget", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(CreateBudgetHandler)))
	mux.Handle("/budget/{id}", middleware.MethodValidationMiddleware(http.MethodDelete)(http.HandlerFunc(DeleteBudgetHandler)))
	return mux
}

func SetupStorage() {
	once.Do(func() {
		_ = jwt.NewRefreshTokenStore("secret123")
		storage.NewUserStore()
		storage.AddUser(storage.UserInfo{
			Id:              0,
			Username:        "admin",
			Password:        "Adm1n123",
			Email:           "email",
			CreatedAt:       time.Now(),
			LastLogin:       time.Now(),
			AvatarUrl:       "img/123.png",
			Balance:         100.5,
			BalanceCurrency: "RUB",
		})
		storage.NewBudgetStore()
	})
}

func TestLogin(t *testing.T) {
	testsCases := []struct {
		name         string
		method       string
		data         map[string]string
		expectedCode int
	}{
		{"get", http.MethodGet, map[string]string{}, http.StatusMethodNotAllowed},
		{"post not exist", http.MethodPost, map[string]string{"username": "", "password": ""}, http.StatusUnauthorized},
		{"post exist", http.MethodPost, map[string]string{"username": "admin", "password": "Adm1n123"}, http.StatusOK},
	}
	SetupStorage()
	mux := SetupRouter()
	url := "/auth/login"

	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body, err := json.Marshal(tc.data)
			require.NoError(t, err)
			request, err := http.NewRequest(tc.method, url, bytes.NewBuffer(body))
			require.NoError(t, err)
			r := httptest.NewRecorder()
			mux.ServeHTTP(r, request)
			require.Equal(t, tc.expectedCode, r.Code)
		})
	}
}
