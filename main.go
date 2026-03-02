package main

import (
	"encoding/json"
	"fmt"
	"main/auth"
	"main/jwt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func SetCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONT_URL"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Origin, Cache-Control, X-Requested-With")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	auth.GenerateNewAuthCookie(w, "1")
	w.Header().Set("Content-Type", "application/json")
	user := AuthUser{
		ID:        1,
		Username:  "username",
		Email:     "email",
		LastLogin: time.Time{},
		CreatedAt: time.Time{},
	}
	response := NewLoginSuccessResponse(user)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println(err)
	}
}

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	w.Header().Set("Content-Type", "application/json")
	isAuth, userID := auth.RefreshToken(w, r)
	data := make(map[string]interface{})
	data["is_auth"] = isAuth
	data["user_id"] = userID
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Println(err)
	}
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "<h1>Страница регистрации</h1>")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearOldToken(w, r)
	w.Header().Set("Content-Type", "application/json")
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	w.Header().Set("Content-Type", "application/json")
	isAuth, userID := auth.IsAuth(r)
	var response interface{}
	if !isAuth {
		response = NewUnauthorizedErrorResponse()
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	storedUser := User{
		Username:  "username",
		Email:     "email",
		CreatedAt: time.Now(),
		AvatarUrl: "img/123.png",
	}
	_ = userID
	err := json.NewEncoder(w).Encode(storedUser)

	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	jwt.TokenStore, err = jwt.NewRefreshTokenStore(os.Getenv("JWT_SECRET"))
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/auth/logout", logoutHandler)
	http.HandleFunc("/auth/refresh", refreshTokenHandler)
	http.HandleFunc("/profile", profileHandler)

	fmt.Println("starting server at :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
