package auth

import (
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
		Name:    TokenName,
		Value:   token,
		Path:    "/",
		Expires: time.Now().AddDate(0, 1, 0),
	}
	http.SetCookie(w, cookie)
	token, err = jwt.GenerateRefreshToken(userID, "pass")
	if err != nil {
		return
	}
	cookie = &http.Cookie{
		Name:    RefreshTokenName,
		Value:   token,
		Path:    "/",
		Expires: time.Now().AddDate(0, 1, 0),
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

func IsAuth(w http.ResponseWriter, r *http.Request) (bool, string) {
	cookie, err := GetAuthCookie(r)
	if err != nil {
		return false, ""
	}
	token := cookie.Value
	isValid, userID := jwt.CheckToken(token)
	if isValid {
		return true, userID
	}
	cookie, err = GetRefreshToken(r)
	if err != nil {
		return false, ""
	}
	token = cookie.Value
	isValid, userID = jwt.CheckRefreshToken(token)
	if isValid {
		RefreshToken(w, r, userID)
	}
	return isValid, userID
}

func ClearOldToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := GetRefreshToken(r)
	if err != nil {
		return
	}
	refreshToken := cookie.Value
	jwt.DeleteRefreshToken(refreshToken)

	cookie = &http.Cookie{
		Name:    TokenName,
		Value:   "",
		Path:    "/",
		Expires: time.Now().AddDate(0, -1, 0),
	}
	http.SetCookie(w, cookie)
	cookie = &http.Cookie{
		Name:    RefreshTokenName,
		Value:   "",
		Path:    "/",
		Expires: time.Now().AddDate(0, -1, 0),
	}
	http.SetCookie(w, cookie)
}

func RefreshToken(w http.ResponseWriter, r *http.Request, userID string) {
	ClearOldToken(w, r)
	GenerateNewAuthCookie(w, userID)
}
