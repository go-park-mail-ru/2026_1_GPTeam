package jwt_auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	repomocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const jwtTestSecret = "testsecret123"
const jwtTestVersion = "v1"

func newJwtForTests(t *testing.T, repo *repomocks.MockJwtRepository) *Jwt {
	t.Helper()
	obj, err := NewJwt(repo, jwtTestSecret, jwtTestVersion)
	require.NoError(t, err)
	return obj
}

func signTokenForJwtTests(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(jwtTestSecret))
	require.NoError(t, err)
	return tokenStr
}

func TestNewJwt_Validation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)

	_, err := NewJwt(repo, "short", jwtTestVersion)
	require.ErrorIs(t, err, JwtSecretError)

	_, err = NewJwt(repo, jwtTestSecret, "")
	require.ErrorIs(t, err, JwtVersionError)
}

func TestJwt_CheckToken_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	obj := newJwtForTests(t, repo)

	token := signTokenForJwtTests(t, jwt.MapClaims{
		"user_id": 9,
		"version": jwtTestVersion,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	isValid, userID := obj.CheckToken(token)
	require.True(t, isValid)
	require.Equal(t, 9, userID)
}

func TestJwt_CheckToken_WrongVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	obj := newJwtForTests(t, repo)

	token := signTokenForJwtTests(t, jwt.MapClaims{
		"user_id": 9,
		"version": "v2",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	isValid, userID := obj.CheckToken(token)
	require.False(t, isValid)
	require.Equal(t, -1, userID)
}

func TestJwt_CheckRefreshToken_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	obj := newJwtForTests(t, repo)

	token := signTokenForJwtTests(t, jwt.MapClaims{
		"id":      "rt-1",
		"user_id": 11,
		"version": jwtTestVersion,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	repo.EXPECT().Get(gomock.Any(), "rt-1").Return(models.RefreshTokenModel{Uuid: "rt-1", UserId: 11}, nil)

	isValid, userID := obj.CheckRefreshToken(context.Background(), token)
	require.True(t, isValid)
	require.Equal(t, 11, userID)
}

func TestJwt_CheckRefreshToken_UserMismatch(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	obj := newJwtForTests(t, repo)

	token := signTokenForJwtTests(t, jwt.MapClaims{
		"id":      "rt-1",
		"user_id": 11,
		"version": jwtTestVersion,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	repo.EXPECT().Get(gomock.Any(), "rt-1").Return(models.RefreshTokenModel{Uuid: "rt-1", UserId: 12}, nil)

	isValid, userID := obj.CheckRefreshToken(context.Background(), token)
	require.False(t, isValid)
	require.Equal(t, -1, userID)
}

func TestJwt_GenerateRefreshToken_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	obj := newJwtForTests(t, repo)

	repo.EXPECT().DeleteByUserId(gomock.Any(), 7).Return(nil)
	repo.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(models.RefreshTokenModel{})).DoAndReturn(
		func(_ context.Context, got models.RefreshTokenModel) error {
			require.Equal(t, 7, got.UserId)
			require.Equal(t, "device-1", got.DeviceId)
			require.NotEmpty(t, got.Uuid)
			require.True(t, got.ExpiredAt.After(time.Now()))
			return nil
		},
	)

	token, err := obj.GenerateRefreshToken(context.Background(), 7, "device-1")
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestJwt_GenerateRefreshToken_DeleteOldFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	obj := newJwtForTests(t, repo)
	genericErr := errors.New("delete old token failed")

	repo.EXPECT().DeleteByUserId(gomock.Any(), 7).Return(genericErr)

	token, err := obj.GenerateRefreshToken(context.Background(), 7, "device-1")
	require.ErrorIs(t, err, genericErr)
	require.Empty(t, token)
}

func TestJwt_DeleteRefreshToken_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	obj := newJwtForTests(t, repo)

	token := signTokenForJwtTests(t, jwt.MapClaims{
		"id":      "rt-1",
		"user_id": 7,
		"version": jwtTestVersion,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	repo.EXPECT().DeleteByUuid(gomock.Any(), "rt-1").Return(nil)

	obj.DeleteRefreshToken(context.Background(), token)
}

func TestJwt_DeleteRefreshToken_InvalidTokenID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	obj := newJwtForTests(t, repo)

	token := signTokenForJwtTests(t, jwt.MapClaims{
		"user_id": 7,
		"version": jwtTestVersion,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	obj.DeleteRefreshToken(context.Background(), token)
}
