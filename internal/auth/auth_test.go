package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	jwtmocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth/mocks"
)

func TestJwtAuthService_GenerateNewAuth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		setupMocks func(jwtUseCase *jwtmocks.MockJwtUseCase)
		checkReq   func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			setupMocks: func(jwtUseCase *jwtmocks.MockJwtUseCase) {
				jwtUseCase.EXPECT().GenerateToken(42).Return("access_token", nil)
				jwtUseCase.EXPECT().GenerateRefreshToken(gomock.Any(), 42, "pass").Return("refresh_token", nil)
			},
			checkReq: func(t *testing.T, w *httptest.ResponseRecorder) {
				cookies := w.Result().Cookies()
				require.Len(t, cookies, 2)

				accessCookie := cookies[0]
				require.Equal(t, TokenName, accessCookie.Name)
				require.Equal(t, "access_token", accessCookie.Value)
				require.Equal(t, "/", accessCookie.Path)
				require.True(t, accessCookie.HttpOnly)

				refreshCookie := cookies[1]
				require.Equal(t, RefreshTokenName, refreshCookie.Name)
				require.Equal(t, "refresh_token", refreshCookie.Value)
				require.Equal(t, "/auth/", refreshCookie.Path)
				require.True(t, refreshCookie.HttpOnly)
			},
		},
		{
			name: "generate access token fails",
			setupMocks: func(jwtUseCase *jwtmocks.MockJwtUseCase) {
				jwtUseCase.EXPECT().GenerateToken(42).Return("", errors.New("error"))
			},
			checkReq: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Empty(t, w.Result().Cookies())
			},
		},
		{
			name: "generate refresh token fails",
			setupMocks: func(jwtUseCase *jwtmocks.MockJwtUseCase) {
				jwtUseCase.EXPECT().GenerateToken(42).Return("access_token", nil)
				jwtUseCase.EXPECT().GenerateRefreshToken(gomock.Any(), 42, "pass").Return("", errors.New("error"))
			},
			checkReq: func(t *testing.T, w *httptest.ResponseRecorder) {
				cookies := w.Result().Cookies()
				require.Len(t, cookies, 1)
				require.Equal(t, TokenName, cookies[0].Name)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			jwtUseCase := jwtmocks.NewMockJwtUseCase(ctrl)
			c.setupMocks(jwtUseCase)

			authService := NewJwtAuthService(jwtUseCase)
			w := httptest.NewRecorder()

			authService.GenerateNewAuth(context.Background(), w, 42)
			c.checkReq(t, w)
		})
	}
}

func TestJwtAuthService_IsAuth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		setupReq     func() *http.Request
		setupMocks   func(jwtUseCase *jwtmocks.MockJwtUseCase)
		expectedBool bool
		expectedId   int
	}{
		{
			name: "no cookie",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			setupMocks:   func(jwtUseCase *jwtmocks.MockJwtUseCase) {},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "invalid token",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{Name: TokenName, Value: "invalid_token"})
				return req
			},
			setupMocks: func(jwtUseCase *jwtmocks.MockJwtUseCase) {
				jwtUseCase.EXPECT().CheckToken("invalid_token").Return(false, -1)
			},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "valid token",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{Name: TokenName, Value: "valid_token"})
				return req
			},
			setupMocks: func(jwtUseCase *jwtmocks.MockJwtUseCase) {
				jwtUseCase.EXPECT().CheckToken("valid_token").Return(true, 42)
			},
			expectedBool: true,
			expectedId:   42,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			jwtUseCase := jwtmocks.NewMockJwtUseCase(ctrl)
			c.setupMocks(jwtUseCase)

			authService := NewJwtAuthService(jwtUseCase)
			req := c.setupReq()

			isValid, userId := authService.IsAuth(context.Background(), req)
			require.Equal(t, c.expectedBool, isValid)
			require.Equal(t, c.expectedId, userId)
		})
	}
}

func TestJwtAuthService_ClearOld(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		setupReq   func() *http.Request
		setupMocks func(jwtUseCase *jwtmocks.MockJwtUseCase)
	}{
		{
			name: "without old refresh token",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/logout", nil)
			},
			setupMocks: func(jwtUseCase *jwtmocks.MockJwtUseCase) {},
		},
		{
			name: "with old refresh token",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/logout", nil)
				req.AddCookie(&http.Cookie{Name: RefreshTokenName, Value: "old_refresh"})
				return req
			},
			setupMocks: func(jwtUseCase *jwtmocks.MockJwtUseCase) {
				jwtUseCase.EXPECT().DeleteRefreshToken(gomock.Any(), "old_refresh")
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			jwtUseCase := jwtmocks.NewMockJwtUseCase(ctrl)
			c.setupMocks(jwtUseCase)

			authService := NewJwtAuthService(jwtUseCase)
			w := httptest.NewRecorder()
			req := c.setupReq()

			authService.ClearOld(context.Background(), w, req)

			cookies := w.Result().Cookies()
			require.Len(t, cookies, 2)

			accessCookie := cookies[0]
			require.Equal(t, TokenName, accessCookie.Name)
			require.Empty(t, accessCookie.Value)
			require.True(t, accessCookie.Expires.Before(time.Now()))

			refreshCookie := cookies[1]
			require.Equal(t, RefreshTokenName, refreshCookie.Name)
			require.Empty(t, refreshCookie.Value)
			require.True(t, refreshCookie.Expires.Before(time.Now()))
		})
	}
}

func TestJwtAuthService_Refresh(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		setupReq     func() *http.Request
		setupMocks   func(jwtUseCase *jwtmocks.MockJwtUseCase)
		expectedBool bool
		expectedId   int
	}{
		{
			name: "no refresh cookie",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/refresh", nil)
			},
			setupMocks:   func(jwtUseCase *jwtmocks.MockJwtUseCase) {},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "invalid refresh token",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
				req.AddCookie(&http.Cookie{Name: RefreshTokenName, Value: "invalid_refresh"})
				return req
			},
			setupMocks: func(jwtUseCase *jwtmocks.MockJwtUseCase) {
				jwtUseCase.EXPECT().CheckRefreshToken(gomock.Any(), "invalid_refresh").Return(false, -1)
			},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "valid refresh token",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
				req.AddCookie(&http.Cookie{Name: RefreshTokenName, Value: "valid_refresh"})
				return req
			},
			setupMocks: func(jwtUseCase *jwtmocks.MockJwtUseCase) {
				jwtUseCase.EXPECT().CheckRefreshToken(gomock.Any(), "valid_refresh").Return(true, 42)
				jwtUseCase.EXPECT().DeleteRefreshToken(gomock.Any(), "valid_refresh")
				jwtUseCase.EXPECT().GenerateToken(42).Return("new_access", nil)
				jwtUseCase.EXPECT().GenerateRefreshToken(gomock.Any(), 42, "pass").Return("new_refresh", nil)
			},
			expectedBool: true,
			expectedId:   42,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			jwtUseCase := jwtmocks.NewMockJwtUseCase(ctrl)
			c.setupMocks(jwtUseCase)

			authService := NewJwtAuthService(jwtUseCase)
			w := httptest.NewRecorder()
			req := c.setupReq()

			isValid, userId := authService.Refresh(context.Background(), w, req)
			require.Equal(t, c.expectedBool, isValid)
			require.Equal(t, c.expectedId, userId)
		})
	}
}
