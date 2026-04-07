package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	authmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

func TestMain(m *testing.M) {
	secure.XssSanitizerInit()
	os.Exit(m.Run())
}

func authUserCtx(user models.UserModel) context.Context {
	return context.WithValue(context.Background(), "user", user)
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authSvc := authmocks.NewMockAuthenticationService(ctrl)
	userApp := appmocks.NewMockUserUseCase(ctrl)
	accountApp := appmocks.NewMockAccountUseCase(ctrl)

	authSvc.EXPECT().ClearOld(gomock.Any(), gomock.Any(), gomock.Any())

	handler := NewAuthHandler(authSvc, userApp, accountApp)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

	ctx := context.WithValue(req.Context(), "request_id", "test-req-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.Logout(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.EqualValues(t, http.StatusOK, resp["code"])
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Parallel()

	testUser := web_helpers.User{Username: "testuser", Email: "test@example.com"}

	cases := []struct {
		name         string
		setupMocks   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase)
		expectedCode int
	}{
		{
			name: "успешный refresh",
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				authSvc.EXPECT().Refresh(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, 1)
				userApp.EXPECT().IsAuthUserExists(gomock.Any(), true, 1).Return(testUser, true)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "refresh не удался → 401",
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				authSvc.EXPECT().Refresh(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, -1)
				userApp.EXPECT().IsAuthUserExists(gomock.Any(), false, -1).Return(web_helpers.User{}, false)
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
			accountApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(authSvc, userApp)

			handler := NewAuthHandler(authSvc, userApp, accountApp)
			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)

			ctx := context.WithValue(req.Context(), "request_id", "test-req-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.RefreshToken(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()

	storedUser := &models.UserModel{
		Id:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	cases := []struct {
		name         string
		body         any
		setupMocks   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase)
		expectedCode int
	}{
		{
			name: "успешный логин",
			body: web_helpers.LoginBodyRequest{Username: "testuser", Password: "Admin123"},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().GetByCredentials(gomock.Any(), gomock.Any()).Return(storedUser, nil)
				userApp.EXPECT().UpdateLastLogin(gomock.Any(), gomock.Any()).Return(nil)
				authSvc.EXPECT().GenerateNewAuth(gomock.Any(), gomock.Any(), storedUser.Id)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "неверные credentials → 401",
			body: web_helpers.LoginBodyRequest{Username: "testuser", Password: "wrong"},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {
				userApp.EXPECT().GetByCredentials(gomock.Any(), gomock.Any()).Return(nil, errors.New("invalid credentials"))
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "невалидный json → 401",
			body:         "not json",
			setupMocks:   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase) {},
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
			accountApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(authSvc, userApp)

			handler := NewAuthHandler(authSvc, userApp, accountApp)

			bodyBytes, _ := json.Marshal(c.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), "request_id", "test-req-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.Login(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestAuthHandler_SignUp(t *testing.T) {
	t.Parallel()

	authUser := web_helpers.AuthUser{Id: 1, Username: "newuser", Email: "new@example.com"}

	cases := []struct {
		name         string
		body         any
		setupMocks   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase)
		expectedCode int
	}{
		{
			name: "успешная регистрация",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authUser, nil)
				accountApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(1, nil)
				accountApp.EXPECT().LinkAccountAndUser(gomock.Any(), 1, authUser.Id).Return(nil)
				authSvc.EXPECT().GenerateNewAuth(gomock.Any(), gomock.Any(), authUser.Id)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "пустые поля → 400",
			body: web_helpers.SignupBodyRequest{},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "невалидный username → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "ab",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "пароли не совпадают → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin456",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "дубликат пользователя → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(web_helpers.AuthUser{}, repository.DuplicatedDataError)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ошибка создания аккаунта → 500",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authUser, nil)
				accountApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(-1, errors.New("db error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "ошибка линковки аккаунта → 500",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authUser, nil)
				accountApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(1, nil)
				accountApp.EXPECT().LinkAccountAndUser(gomock.Any(), 1, authUser.Id).Return(errors.New("db error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "ошибка хэширования пароля → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(web_helpers.AuthUser{}, application.HashPasswordError)
			},
			expectedCode: http.StatusBadRequest,
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
			accountApp := appmocks.NewMockAccountUseCase(ctrl)
			c.setupMocks(authSvc, userApp, accountApp)

			handler := NewAuthHandler(authSvc, userApp, accountApp)

			bodyBytes, _ := json.Marshal(c.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), "request_id", "test-req-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.SignUp(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}
