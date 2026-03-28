package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	authmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

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
			name:         "невалидный json → 400",
			body:         "{not json}",
			setupMocks:   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {},
			expectedCode: http.StatusBadRequest,
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
			setupMocks:   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "невалидный пароль → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "weak",
				ConfirmPassword: "weak",
				Email:           "new@example.com",
			},
			setupMocks:   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "невалидный email → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "notanemail",
			},
			setupMocks:   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {},
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
			setupMocks:   func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {},
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
			name: "ConstraintError при создании юзера → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(web_helpers.AuthUser{}, repository.ConstraintError)
			},
			expectedCode: http.StatusBadRequest,
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
		{
			name: "неизвестная ошибка при создании юзера → 500",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(web_helpers.AuthUser{}, errors.New("db error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "AccountDuplicatedDataError при создании аккаунта → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authUser, nil)
				accountApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(-1, repository.AccountDuplicatedDataError)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ConstraintError при создании аккаунта → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authUser, nil)
				accountApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(-1, repository.ConstraintError)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "неизвестная ошибка при создании аккаунта → 500",
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
			name: "ConstraintError при линковке аккаунта → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authUser, nil)
				accountApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(1, nil)
				accountApp.EXPECT().LinkAccountAndUser(gomock.Any(), 1, authUser.Id).Return(repository.ConstraintError)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "AccountForeignKeyError при линковке аккаунта → 400",
			body: web_helpers.SignupBodyRequest{
				Username:        "newuser",
				Password:        "Admin123",
				ConfirmPassword: "Admin123",
				Email:           "new@example.com",
			},
			setupMocks: func(authSvc *authmocks.MockAuthenticationService, userApp *appmocks.MockUserUseCase, accountApp *appmocks.MockAccountUseCase) {
				userApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authUser, nil)
				accountApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(1, nil)
				accountApp.EXPECT().LinkAccountAndUser(gomock.Any(), 1, authUser.Id).Return(repository.AccountForeignKeyError)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "неизвестная ошибка при линковке аккаунта → 500",
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

			var bodyBytes []byte
			if s, ok := c.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(c.body)
			}
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.SignUp(w, req)

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}

func TestAuthHandler_GetProfile(t *testing.T) {
	t.Parallel()

	testUser := models.UserModel{Id: 1, Username: "testuser", Email: "test@example.com"}

	cases := []struct {
		name         string
		buildReq     func() *http.Request
		expectedCode int
	}{
		{
			name: "пользователь в контексте → 200",
			buildReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/profile", nil)
				return req.WithContext(middleware.SetUserCtx(req.Context(), testUser))
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "нет пользователя в контексте → 401",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/profile", nil)
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

			handler := NewAuthHandler(authSvc, userApp, accountApp)
			w := httptest.NewRecorder()
			handler.GetProfile(w, c.buildReq())

			require.Equal(t, c.expectedCode, w.Code)
		})
	}
}