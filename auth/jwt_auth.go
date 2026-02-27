package auth

import (
	"main/jwt"
	"net/http"
	"time"
)

const TokenName = "token"

func GenerateNewAuthCookie(userID string) (http.Cookie, error) {
	token, err := jwt.GenerateToken(userID)
	return http.Cookie{
		Name:    TokenName,
		Value:   token,
		Path:    "/",
		Expires: time.Now().AddDate(0, 1, 0),
	}, err
}

func GetAuthCookie(r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(TokenName)
	return cookie, err
}

func IsAuth(r *http.Request) (bool, string) {
	cookie, err := GetAuthCookie(r)
	if err != nil {
		return false, ""
	}
	token := cookie.Value
	return jwt.CheckToken(token)
}

func ClearOldToken(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:    TokenName,
		Value:   "",
		Path:    "/",
		Expires: time.Now().AddDate(0, -1, 0),
	}
	http.SetCookie(w, cookie)
}

func RefreshToken() {

}
