package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	appmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	authmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/mocks"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newAuthRequest(t *testing.T, method, target string, body any) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		err := json.NewEncoder(&buf).Encode(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, target, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestAuthHandler_SignUp_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	authService := authmocks.NewMockAuthenticationService(ctrl)
	userUC := appmocks.NewMockUserUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewAuthHandler(authService, userUC, accountUC)

	body := web_helpers.SignupBodyRequest{
		Username:        "Mike123",
		Password:        "Admin123",
		ConfirmPassword: "Admin123",
		Email:           "mike@example.com",
	}
	createdAt := time.Date(2026, 4, 7, 12, 0, 0, 0, time.UTC)
	authUser := web_helpers.AuthUser{
		Id:        17,
		Username:  body.Username,
		Email:     body.Email,
		CreatedAt: createdAt,
	}

	userUC.EXPECT().Create(gomock.Any(), body).Return(authUser, nil)
	accountUC.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(models.AccountModel{})).DoAndReturn(
		func(_ any, got models.AccountModel) (int, error) {
			require.Equal(t, "base", got.Name)
			require.Equal(t, 0.0, got.Balance)
			require.Equal(t, "RUB", got.Currency)
			require.False(t, got.CreatedAt.IsZero())
			require.False(t, got.UpdatedAt.IsZero())
			return 101, nil
		},
	)
	accountUC.EXPECT().LinkAccountAndUser(gomock.Any(), 101, 17).Return(nil)
	authService.EXPECT().GenerateNewAuth(gomock.Any(), gomock.Any(), 17)

	req := newAuthRequest(t, http.MethodPost, "/auth/signup", body)
	rr := httptest.NewRecorder()

	handler.SignUp(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		User struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"user"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 200, resp.Code)
	require.Equal(t, 17, resp.User.ID)
	require.Equal(t, "Mike123", resp.User.Username)
	require.Equal(t, "mike@example.com", resp.User.Email)
}

func TestAuthHandler_SignUp_MissingFields(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	authService := authmocks.NewMockAuthenticationService(ctrl)
	userUC := appmocks.NewMockUserUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewAuthHandler(authService, userUC, accountUC)

	body := web_helpers.SignupBodyRequest{}

	req := newAuthRequest(t, http.MethodPost, "/auth/signup", body)
	rr := httptest.NewRecorder()

	handler.SignUp(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	var resp struct {
		Code   int                      `json:"code"`
		Errors []web_helpers.FieldError `json:"errors"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Len(t, resp.Errors, 4)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	authService := authmocks.NewMockAuthenticationService(ctrl)
	userUC := appmocks.NewMockUserUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewAuthHandler(authService, userUC, accountUC)

	body := web_helpers.LoginBodyRequest{Username: "Mike123", Password: "Admin123"}
	createdAt := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	storedUser := &models.UserModel{
		Id:        17,
		Username:  "Mike123",
		Email:     "mike@example.com",
		CreatedAt: createdAt,
		AvatarUrl: "avatar.png",
	}

	userUC.EXPECT().GetByCredentials(gomock.Any(), body).Return(storedUser, nil)
	authService.EXPECT().GenerateNewAuth(gomock.Any(), gomock.Any(), 17)

	req := newAuthRequest(t, http.MethodPost, "/auth/login", body)
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		User struct {
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"user"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "Mike123", resp.User.Username)
	require.Equal(t, "mike@example.com", resp.User.Email)
}

func TestAuthHandler_Login_Unauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	authService := authmocks.NewMockAuthenticationService(ctrl)
	userUC := appmocks.NewMockUserUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewAuthHandler(authService, userUC, accountUC)

	body := web_helpers.LoginBodyRequest{Username: "Mike123", Password: "Wrong123"}
	userUC.EXPECT().GetByCredentials(gomock.Any(), body).Return(nil, errors.New("bad credentials"))

	req := newAuthRequest(t, http.MethodPost, "/auth/login", body)
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	authService := authmocks.NewMockAuthenticationService(ctrl)
	userUC := appmocks.NewMockUserUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewAuthHandler(authService, userUC, accountUC)

	user := web_helpers.User{
		Username:  "Mike123",
		Email:     "mike@example.com",
		CreatedAt: time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
		AvatarUrl: "avatar.png",
	}

	req := newAuthRequest(t, http.MethodPost, "/auth/refresh", nil)
	rr := httptest.NewRecorder()

	authService.EXPECT().Refresh(gomock.Any(), gomock.Any(), req).Return(true, 17)
	userUC.EXPECT().IsAuthUserExists(gomock.Any(), true, 17).Return(user, true)

	handler.RefreshToken(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		User struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "Mike123", resp.User.Username)
}

func TestAuthHandler_RefreshToken_Unauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	authService := authmocks.NewMockAuthenticationService(ctrl)
	userUC := appmocks.NewMockUserUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewAuthHandler(authService, userUC, accountUC)

	req := newAuthRequest(t, http.MethodPost, "/auth/refresh", nil)
	rr := httptest.NewRecorder()

	authService.EXPECT().Refresh(gomock.Any(), gomock.Any(), req).Return(false, -1)
	userUC.EXPECT().IsAuthUserExists(gomock.Any(), false, -1).Return(web_helpers.User{}, false)

	handler.RefreshToken(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	authService := authmocks.NewMockAuthenticationService(ctrl)
	userUC := appmocks.NewMockUserUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewAuthHandler(authService, userUC, accountUC)

	req := newAuthRequest(t, http.MethodPost, "/auth/logout", nil)
	rr := httptest.NewRecorder()

	authService.EXPECT().ClearOld(gomock.Any(), gomock.Any(), req)

	handler.Logout(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 200, resp.Code)
}

func TestAuthHandler_SignUp_DuplicateUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	authService := authmocks.NewMockAuthenticationService(ctrl)
	userUC := appmocks.NewMockUserUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewAuthHandler(authService, userUC, accountUC)

	body := web_helpers.SignupBodyRequest{
		Username:        "Mike123",
		Password:        "Admin123",
		ConfirmPassword: "Admin123",
		Email:           "mike@example.com",
	}
	userUC.EXPECT().Create(gomock.Any(), body).Return(web_helpers.AuthUser{}, repository.DuplicatedDataError)

	req := newAuthRequest(t, http.MethodPost, "/auth/signup", body)
	rr := httptest.NewRecorder()

	handler.SignUp(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthHandler_SignUp_HashPasswordError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	authService := authmocks.NewMockAuthenticationService(ctrl)
	userUC := appmocks.NewMockUserUseCase(ctrl)
	accountUC := appmocks.NewMockAccountUseCase(ctrl)
	handler := NewAuthHandler(authService, userUC, accountUC)

	body := web_helpers.SignupBodyRequest{
		Username:        "Mike123",
		Password:        "Admin123",
		ConfirmPassword: "Admin123",
		Email:           "mike@example.com",
	}
	userUC.EXPECT().Create(gomock.Any(), body).Return(web_helpers.AuthUser{}, application.HashPasswordError)

	req := newAuthRequest(t, http.MethodPost, "/auth/signup", body)
	rr := httptest.NewRecorder()

	handler.SignUp(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}
