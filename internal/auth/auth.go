package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth/jwt_auth"
)

type AuthenticationService interface {
	GenerateNewAuth(ctx context.Context, w http.ResponseWriter, userId int)
	IsAuth(r *http.Request) (bool, int)
	ClearOld(ctx context.Context, w http.ResponseWriter, r *http.Request)
	Refresh(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, int)
}

type JWTAuthService struct {
	jwt jwt_auth.JwtUseCase
}

const TokenName = "token"
const RefreshTokenName = "refresh_token"

func NewJWTAuth(useCase jwt_auth.JwtUseCase) *JWTAuthService {
	return &JWTAuthService{jwt: useCase}
}

func (obj *JWTAuthService) GenerateNewAuth(ctx context.Context, w http.ResponseWriter, userId int) {
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
}

func (obj *JWTAuthService) GetAuthCookie(r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(TokenName)
	return cookie, err
}

func (obj *JWTAuthService) GetRefreshToken(r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(RefreshTokenName)
	return cookie, err
}

func (obj *JWTAuthService) IsAuth(r *http.Request) (bool, int) {
	cookie, err := obj.GetAuthCookie(r)
	if err != nil {
		fmt.Println(err)
		return false, -1
	}
	token := cookie.Value
	isValid, userId := obj.jwt.CheckToken(token)
	return isValid, userId
}

func (obj *JWTAuthService) ClearOld(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	cookie, err := obj.GetRefreshToken(r)
	if err != nil {
		fmt.Println(err)
	} else {
		refreshToken := cookie.Value
		err = obj.jwt.DeleteRefreshToken(ctx, refreshToken)
		if err != nil {
			fmt.Println(err)
		}
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
}

func (obj *JWTAuthService) Refresh(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, int) {
	cookie, err := obj.GetRefreshToken(r)
	if err != nil {
		fmt.Println(err)
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
