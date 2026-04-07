package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth"
	repomocks "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const authTestSecret = "testsecret123"
const authTestVersion = "v1"

func newJwtForAuthTests(t *testing.T, repo *repomocks.MockJwtRepository) *jwt_auth.Jwt {
	t.Helper()
	service, err := jwt_auth.NewJwt(repo, authTestSecret, authTestVersion)
	require.NoError(t, err)
	return service
}

func signRefreshTokenForAuthTests(t *testing.T, tokenID string, userID int) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      tokenID,
		"user_id": userID,
		"version": authTestVersion,
		"exp":     time.Now().Add(jwt_auth.RefreshTokenExpirationTime).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(authTestSecret))
	require.NoError(t, err)
	return tokenStr
}

func cookieMap(cookies []*http.Cookie) map[string]*http.Cookie {
	result := make(map[string]*http.Cookie, len(cookies))
	for _, cookie := range cookies {
		result[cookie.Name] = cookie
	}
	return result
}

func TestJwtAuthService_GenerateNewAuth_SetsBothCookies(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	jwtUC := newJwtForAuthTests(t, repo)
	service := NewJwtAuthService(jwtUC)

	repo.EXPECT().DeleteByUserId(gomock.Any(), 7).Return(nil)
	repo.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(models.RefreshTokenModel{})).DoAndReturn(
		func(_ context.Context, got models.RefreshTokenModel) error {
			require.Equal(t, 7, got.UserId)
			require.Equal(t, "pass", got.DeviceId)
			require.NotEmpty(t, got.Uuid)
			require.True(t, got.ExpiredAt.After(time.Now()))
			return nil
		},
	)

	rr := httptest.NewRecorder()
	service.GenerateNewAuth(context.Background(), rr, 7)

	cookies := cookieMap(rr.Result().Cookies())
	require.Contains(t, cookies, TokenName)
	require.Contains(t, cookies, RefreshTokenName)
	require.NotEmpty(t, cookies[TokenName].Value)
	require.NotEmpty(t, cookies[RefreshTokenName].Value)
	require.Equal(t, "/", cookies[TokenName].Path)
	require.Equal(t, "/auth/", cookies[RefreshTokenName].Path)
}

func TestJwtAuthService_IsAuth_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	jwtUC := newJwtForAuthTests(t, repo)
	service := NewJwtAuthService(jwtUC)

	token, err := jwtUC.GenerateToken(9)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.AddCookie(&http.Cookie{Name: TokenName, Value: token})

	isAuth, userID := service.IsAuth(req)
	require.True(t, isAuth)
	require.Equal(t, 9, userID)
}

func TestJwtAuthService_IsAuth_NoCookie(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	jwtUC := newJwtForAuthTests(t, repo)
	service := NewJwtAuthService(jwtUC)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)

	isAuth, userID := service.IsAuth(req)
	require.False(t, isAuth)
	require.Equal(t, -1, userID)
}

func TestJwtAuthService_ClearOld_DeletesRefreshTokenAndClearsCookies(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	jwtUC := newJwtForAuthTests(t, repo)
	service := NewJwtAuthService(jwtUC)

	refreshToken := signRefreshTokenForAuthTests(t, "rt-1", 7)
	repo.EXPECT().DeleteByUuid(gomock.Any(), "rt-1").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: RefreshTokenName, Value: refreshToken})
	rr := httptest.NewRecorder()

	service.ClearOld(context.Background(), rr, req)

	cookies := cookieMap(rr.Result().Cookies())
	require.Contains(t, cookies, TokenName)
	require.Contains(t, cookies, RefreshTokenName)
	require.Equal(t, "", cookies[TokenName].Value)
	require.Equal(t, "", cookies[RefreshTokenName].Value)
	require.Equal(t, "/", cookies[TokenName].Path)
	require.Equal(t, "/auth/", cookies[RefreshTokenName].Path)
}

func TestJwtAuthService_Refresh_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	jwtUC := newJwtForAuthTests(t, repo)
	service := NewJwtAuthService(jwtUC)

	refreshToken := signRefreshTokenForAuthTests(t, "rt-1", 7)
	gomock.InOrder(
		repo.EXPECT().Get(gomock.Any(), "rt-1").Return(models.RefreshTokenModel{Uuid: "rt-1", UserId: 7}, nil),
		repo.EXPECT().DeleteByUuid(gomock.Any(), "rt-1").Return(nil),
		repo.EXPECT().DeleteByUserId(gomock.Any(), 7).Return(nil),
		repo.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(models.RefreshTokenModel{})).Return(nil),
	)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: RefreshTokenName, Value: refreshToken})
	rr := httptest.NewRecorder()

	isAuth, userID := service.Refresh(context.Background(), rr, req)

	require.True(t, isAuth)
	require.Equal(t, 7, userID)
	cookies := cookieMap(rr.Result().Cookies())
	require.Contains(t, cookies, TokenName)
	require.Contains(t, cookies, RefreshTokenName)
}

func TestJwtAuthService_Refresh_NoCookie(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockJwtRepository(ctrl)
	jwtUC := newJwtForAuthTests(t, repo)
	service := NewJwtAuthService(jwtUC)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	rr := httptest.NewRecorder()

	isAuth, userID := service.Refresh(context.Background(), rr, req)
	require.False(t, isAuth)
	require.Equal(t, -1, userID)
}
