package auth

import (
	"fmt"
	"main/jwt"
	"net/http"
	"time"
)

const TokenName = "token"
const RefreshTokenName = "refresh_token"

func GenerateNewAuthCookie(w http.ResponseWriter, userID string) {
	token, err := jwt.GenerateToken(userID)
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
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
	token, err = jwt.GenerateRefreshToken(userID, "pass")
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
		SameSite: http.SameSiteStrictMode,
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

func IsAuth(r *http.Request) (bool, string) {
	cookie, err := GetAuthCookie(r)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	token := cookie.Value
	isValid, userID := jwt.CheckToken(token)
	return isValid, userID
}

func ClearOldToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := GetRefreshToken(r)
	if err != nil {
		fmt.Println(err)
		return
	}
	refreshToken := cookie.Value
	jwt.DeleteRefreshToken(refreshToken)

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

func RefreshToken(w http.ResponseWriter, r *http.Request) (bool, string) {
	cookie, err := GetRefreshToken(r)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	token := cookie.Value
	isValid, userID := jwt.CheckRefreshToken(token)
	if !isValid {
		return false, ""
	}
	ClearOldToken(w, r)
	GenerateNewAuthCookie(w, userID)
	return isValid, userID
}
