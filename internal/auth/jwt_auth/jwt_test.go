package jwt_auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	repomocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
)

const (
	testSecret  = "super_secret_key_123"
	testVersion = "1.0.0"
)

func createTestJwt(t *testing.T, ctrl *gomock.Controller) (*Jwt, *repomocks.MockJwtRepository) {
	repo := repomocks.NewMockJwtRepository(ctrl)
	jwtSvc, err := NewJwt(repo, testSecret, testVersion)
	require.NoError(t, err)
	return jwtSvc, repo
}

func generateCustomToken(claims jwt.MapClaims, secret string, method jwt.SigningMethod) string {
	token := jwt.NewWithClaims(method, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestNewJwt(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockJwtRepository(ctrl)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		jwtSvc, err := NewJwt(repo, testSecret, testVersion)
		require.NoError(t, err)
		require.NotNil(t, jwtSvc)
		require.Equal(t, testVersion, jwtSvc.GetVersion())
		require.Equal(t, []byte(testSecret), jwtSvc.GetJWTSecret())
	})
}

func TestJwt_CheckToken(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		setupToken   func() string
		expectedBool bool
		expectedId   int
	}{
		{
			name: "valid token",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"user_id": float64(42),
					"version": testVersion,
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			expectedBool: true,
			expectedId:   42,
		},
		{
			name: "invalid token string",
			setupToken: func() string {
				return "invalid.token.string"
			},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "wrong signing method",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"user_id": float64(42),
					"version": testVersion,
				}, testSecret, jwt.SigningMethodNone)
			},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "wrong version",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"user_id": float64(42),
					"version": "0.0.1",
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "missing version",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"user_id": float64(42),
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "missing user_id",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"version": testVersion,
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "expired token",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"user_id": float64(42),
					"version": testVersion,
					"exp":     time.Now().Add(-time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			expectedBool: false,
			expectedId:   -1,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			jwtSvc, _ := createTestJwt(t, ctrl)
			tokenStr := c.setupToken()

			isValid, userId := jwtSvc.CheckToken(tokenStr)
			require.Equal(t, c.expectedBool, isValid)
			require.Equal(t, c.expectedId, userId)
		})
	}
}

func TestJwt_CheckRefreshToken(t *testing.T) {
	t.Parallel()

	validTokenId := uuid.New().String()

	cases := []struct {
		name         string
		setupToken   func() string
		setupMocks   func(repo *repomocks.MockJwtRepository)
		expectedBool bool
		expectedId   int
	}{
		{
			name: "valid refresh token",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"id":      validTokenId,
					"user_id": float64(42),
					"version": testVersion,
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			setupMocks: func(repo *repomocks.MockJwtRepository) {
				repo.EXPECT().Get(gomock.Any(), validTokenId).Return(models.RefreshTokenModel{
					Uuid:   validTokenId,
					UserId: 42,
				}, nil)
			},
			expectedBool: true,
			expectedId:   42,
		},
		{
			name: "token not found in db",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"id":      validTokenId,
					"user_id": float64(42),
					"version": testVersion,
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			setupMocks: func(repo *repomocks.MockJwtRepository) {
				repo.EXPECT().Get(gomock.Any(), validTokenId).Return(models.RefreshTokenModel{}, errors.New("not found"))
			},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "user_id mismatch in db",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"id":      validTokenId,
					"user_id": float64(42),
					"version": testVersion,
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			setupMocks: func(repo *repomocks.MockJwtRepository) {
				repo.EXPECT().Get(gomock.Any(), validTokenId).Return(models.RefreshTokenModel{
					Uuid:   validTokenId,
					UserId: 99,
				}, nil)
			},
			expectedBool: false,
			expectedId:   -1,
		},
		{
			name: "missing id claim",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"user_id": float64(42),
					"version": testVersion,
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			setupMocks:   func(repo *repomocks.MockJwtRepository) {},
			expectedBool: false,
			expectedId:   -1,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			jwtSvc, repo := createTestJwt(t, ctrl)
			tokenStr := c.setupToken()
			c.setupMocks(repo)

			isValid, userId := jwtSvc.CheckRefreshToken(context.Background(), tokenStr)
			require.Equal(t, c.expectedBool, isValid)
			require.Equal(t, c.expectedId, userId)
		})
	}
}

func TestJwt_GenerateToken(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtSvc, _ := createTestJwt(t, ctrl)

	tokenStr, err := jwtSvc.GenerateToken(42)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	isValid, userId := jwtSvc.CheckToken(tokenStr)
	require.True(t, isValid)
	require.Equal(t, 42, userId)
}

func TestJwt_GenerateRefreshToken(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		setupMocks  func(repo *repomocks.MockJwtRepository)
		expectedErr bool
	}{
		{
			name: "success",
			setupMocks: func(repo *repomocks.MockJwtRepository) {
				repo.EXPECT().DeleteByUserId(gomock.Any(), 42).Return(nil)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "delete fails",
			setupMocks: func(repo *repomocks.MockJwtRepository) {
				repo.EXPECT().DeleteByUserId(gomock.Any(), 42).Return(errors.New("db error"))
			},
			expectedErr: true,
		},
		{
			name: "create fails",
			setupMocks: func(repo *repomocks.MockJwtRepository) {
				repo.EXPECT().DeleteByUserId(gomock.Any(), 42).Return(nil)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			expectedErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			jwtSvc, repo := createTestJwt(t, ctrl)
			c.setupMocks(repo)

			tokenStr, err := jwtSvc.GenerateRefreshToken(context.Background(), 42, "device1")
			if c.expectedErr {
				require.Error(t, err)
				require.Empty(t, tokenStr)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, tokenStr)
			}
		})
	}
}

func TestJwt_DeleteRefreshToken(t *testing.T) {
	t.Parallel()

	validTokenId := uuid.New().String()

	cases := []struct {
		name       string
		setupToken func() string
		setupMocks func(repo *repomocks.MockJwtRepository)
	}{
		{
			name: "success",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"id":      validTokenId,
					"user_id": float64(42),
					"version": testVersion,
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			setupMocks: func(repo *repomocks.MockJwtRepository) {
				repo.EXPECT().DeleteByUuid(gomock.Any(), validTokenId).Return(nil)
			},
		},
		{
			name: "invalid token string",
			setupToken: func() string {
				return "invalid"
			},
			setupMocks: func(repo *repomocks.MockJwtRepository) {},
		},
		{
			name: "missing id claim",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"user_id": float64(42),
					"version": testVersion,
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			setupMocks: func(repo *repomocks.MockJwtRepository) {},
		},
		{
			name: "db delete error",
			setupToken: func() string {
				return generateCustomToken(jwt.MapClaims{
					"id":      validTokenId,
					"user_id": float64(42),
					"version": testVersion,
					"exp":     time.Now().Add(time.Hour).Unix(),
				}, testSecret, jwt.SigningMethodHS256)
			},
			setupMocks: func(repo *repomocks.MockJwtRepository) {
				repo.EXPECT().DeleteByUuid(gomock.Any(), validTokenId).Return(errors.New("db error"))
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			jwtSvc, repo := createTestJwt(t, ctrl)
			tokenStr := c.setupToken()
			c.setupMocks(repo)

			jwtSvc.DeleteRefreshToken(context.Background(), tokenStr)
		})
	}
}
