package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type AuthenticationService interface {
	GenerateNewAuth(ctx context.Context, w http.ResponseWriter, userId int)
	IsAuth(ctx context.Context, r *http.Request) (bool, int)
	ClearOld(ctx context.Context, w http.ResponseWriter, r *http.Request)
	Refresh(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, int)
}

type JwtAuthService struct {
	jwt jwt_auth.JwtUseCase
	log *zap.Logger
}

const TokenName = "token"
const RefreshTokenName = "refresh_token"

func NewJwtAuthService(useCase jwt_auth.JwtUseCase) *JwtAuthService {
	return &JwtAuthService{
		jwt: useCase,
		log: logger.GetLogger(),
	}
}

func (obj *JwtAuthService) GenerateNewAuth(ctx context.Context, w http.ResponseWriter, userId int) {
	obj.log.Info("generating new auth for user",
		zap.Int("user_id", userId),
		zap.String("request_id", ctx.Value("request_id").(string)))
	token, err := obj.jwt.GenerateToken(userId)
	if err != nil {
		return
	}
	cookie := &http.Cookie{
		Name:     TokenName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(jwt_auth.AccessTokenExpirationTime),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	obj.log.Info("set access token cookie",
		zap.Int("user_id", userId),
		zap.String("request_id", ctx.Value("request_id").(string)))
	token, err = obj.jwt.GenerateRefreshToken(ctx, userId, "pass")
	if err != nil {
		return
	}
	cookie = &http.Cookie{
		Name:     RefreshTokenName,
		Value:    token,
		Path:     "/auth/",
		Expires:  time.Now().Add(jwt_auth.RefreshTokenExpirationTime),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	obj.log.Info("set refresh token cookie",
		zap.Int("user_id", userId),
		zap.String("request_id", ctx.Value("request_id").(string)))
}

func (obj *JwtAuthService) GetAuthCookie(ctx context.Context, r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(TokenName)
	if err != nil {
		obj.log.Warn("failed to get token cookie",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
	}
	return cookie, err
}

func (obj *JwtAuthService) GetRefreshToken(ctx context.Context, r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(RefreshTokenName)
	if err != nil {
		obj.log.Warn("failed to get refresh token cookie",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
	}
	return cookie, err
}

func (obj *JwtAuthService) IsAuth(ctx context.Context, r *http.Request) (bool, int) {
	obj.log.Info("checking if user authenticated",
		zap.String("request_id", ctx.Value("request_id").(string)))
	cookie, err := obj.GetAuthCookie(ctx, r)
	if err != nil {
		return false, -1
	}
	token := cookie.Value
	isValid, userId := obj.jwt.CheckToken(token)
	return isValid, userId
}

func (obj *JwtAuthService) ClearOld(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	obj.log.Info("clear old token cookie",
		zap.String("request_id", ctx.Value("request_id").(string)))
	cookie, err := obj.GetRefreshToken(ctx, r)
	if err == nil {
		refreshToken := cookie.Value
		obj.jwt.DeleteRefreshToken(ctx, refreshToken)
	}

	cookie = &http.Cookie{
		Name:     TokenName,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().AddDate(0, -1, 0),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
	obj.log.Info("set empty old token cookie",
		zap.String("request_id", ctx.Value("request_id").(string)))
	cookie = &http.Cookie{
		Name:     RefreshTokenName,
		Value:    "",
		Path:     "/auth/",
		Expires:  time.Now().AddDate(0, -1, 0),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
	obj.log.Info("set empty old refresh token cookie",
		zap.String("request_id", ctx.Value("request_id").(string)))
}

func (obj *JwtAuthService) Refresh(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, int) {
	obj.log.Info("refresh token cookie",
		zap.String("request_id", ctx.Value("request_id").(string)))
	cookie, err := obj.GetRefreshToken(ctx, r)
	if err != nil {
		return false, -1
	}
	token := cookie.Value
	isValid, userId := obj.jwt.CheckRefreshToken(ctx, token)
	if !isValid {
		return false, -1
	}
	obj.ClearOld(ctx, w, r)
	obj.GenerateNewAuth(ctx, w, userId)
	return isValid, userId
}
