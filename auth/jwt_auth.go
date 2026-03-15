package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
)

type AuthenticationServiceInterface interface {
	GenerateNewAuth(w http.ResponseWriter, userID int)
	IsAuth(r *http.Request) (bool, int)
	ClearOld(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request) (bool, int)
}

type JWTAuthService struct {
	jwtUseCase jwt.JWTUseCaseInterface
}

const TokenName = "token"
const RefreshTokenName = "refresh_token"

func NewJWTAuth(useCase jwt.JWTUseCaseInterface) *JWTAuthService {
	return &JWTAuthService{jwtUseCase: useCase}
}

func (obj *JWTAuthService) GenerateNewAuth(w http.ResponseWriter, userID int) {
	fmt.Printf("set cookie to user id %v\n", userID)
	token, err := obj.jwtUseCase.GenerateToken(userID)
	if err != nil {
		return
	}
	cookie := &http.Cookie{
		Name:     TokenName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(jwt.AccessTokenExpirationTime),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	token, err = obj.jwtUseCase.GenerateRefreshToken(context.Background(), userID, "pass") // ToDo: general context
	if err != nil {
		return
	}
	cookie = &http.Cookie{
		Name:     RefreshTokenName,
		Value:    token,
		Path:     "/auth/",
		Expires:  time.Now().Add(jwt.RefreshTokenExpirationTime),
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
	isValid, userID := obj.jwtUseCase.CheckToken(token)
	fmt.Println("---", isValid, userID)
	return isValid, userID
}

func (obj *JWTAuthService) ClearOld(w http.ResponseWriter, r *http.Request) {
	cookie, err := obj.GetRefreshToken(r)
	if err != nil {
		fmt.Println(err)
	} else {
		refreshToken := cookie.Value
		err = obj.jwtUseCase.DeleteRefreshToken(context.Background(), refreshToken) // ToDo: general context
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

func (obj *JWTAuthService) Refresh(w http.ResponseWriter, r *http.Request) (bool, int) {
	cookie, err := obj.GetRefreshToken(r)
	if err != nil {
		fmt.Println(err)
		return false, -1
	}
	token := cookie.Value
	isValid, userID := obj.jwtUseCase.CheckRefreshToken(context.Background(), token) // ToDo: general context
	if !isValid {
		return false, -1
	}
	obj.ClearOld(w, r)
	obj.GenerateNewAuth(w, userID)
	return isValid, userID
}
