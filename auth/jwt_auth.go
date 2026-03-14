package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
)

const TokenName = "token"
const RefreshTokenName = "refresh_token"

func GenerateNewAuthCookie(repo repository.JWTRepositoryInterface, w http.ResponseWriter, userID string) {
	token, err := jwt.GenerateToken(repo, userID)
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
	token, err = jwt.GenerateRefreshToken(repo, userID, "pass")
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

func GetAuthCookie(r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(TokenName)
	return cookie, err
}

func GetRefreshToken(r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(RefreshTokenName)
	return cookie, err
}

func IsAuth(repo repository.JWTRepositoryInterface, r *http.Request) (bool, string) {
	cookie, err := GetAuthCookie(r)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	token := cookie.Value
	isValid, userID := jwt.CheckToken(repo, token)
	return isValid, userID
}

func ClearOldToken(repo repository.JWTRepositoryInterface, w http.ResponseWriter, r *http.Request) {
	cookie, err := GetRefreshToken(r)
	if err != nil {
		fmt.Println(err)
	} else {
		refreshToken := cookie.Value
		err = jwt.DeleteRefreshToken(repo, refreshToken)
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

func RefreshToken(repo repository.JWTRepositoryInterface, w http.ResponseWriter, r *http.Request) (bool, string) {
	cookie, err := GetRefreshToken(r)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	token := cookie.Value
	isValid, userID := jwt.CheckRefreshToken(repo, token)
	if !isValid {
		return false, ""
	}
	ClearOldToken(repo, w, r)
	GenerateNewAuthCookie(repo, w, userID)
	return isValid, userID
}
