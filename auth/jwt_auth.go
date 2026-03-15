package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
)

type AuthenticationServiceInterface interface {
	GenerateNewAuth(w http.ResponseWriter, userID int)
	IsAuth(r *http.Request) (bool, int)
	ClearOld(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request) (bool, int)
}

type JWTAuthService struct {
	repo repository.JWTRepositoryInterface
}

const TokenName = "token"
const RefreshTokenName = "refresh_token"

func NewJWTAuth(repo repository.JWTRepositoryInterface) *JWTAuthService {
	return &JWTAuthService{repo: repo}
}

func (obj *JWTAuthService) GenerateNewAuth(w http.ResponseWriter, userID int) {
	token, err := jwt.GenerateToken(obj.repo, userID)
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
	token, err = jwt.GenerateRefreshToken(obj.repo, userID, "pass")
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
	isValid, userID := jwt.CheckToken(obj.repo, token)
	return isValid, userID
}

func (obj *JWTAuthService) ClearOld(w http.ResponseWriter, r *http.Request) {
	cookie, err := obj.GetRefreshToken(r)
	if err != nil {
		fmt.Println(err)
	} else {
		refreshToken := cookie.Value
		err = jwt.DeleteRefreshToken(obj.repo, refreshToken)
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
	isValid, userID := jwt.CheckRefreshToken(obj.repo, token)
	if !isValid {
		return false, -1
	}
	obj.ClearOld(w, r)
	obj.GenerateNewAuth(w, userID)
	return isValid, userID
}
