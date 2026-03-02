package main

import (
	"encoding/json"
	"fmt"
	"main/auth"
	"main/base"
	"main/storage"
	"net/http"
	"os"
	"strconv"
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
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		response := base.NewMethodError()
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	var userRequest base.LoginBodyRequest
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	storedUser, exists := storage.FindUserByCredentials(userRequest)
	var response interface{}
	if exists {
		user := base.AuthUser{
			ID:        storedUser.Id,
			Username:  storedUser.Username,
			Email:     storedUser.Email,
			LastLogin: time.Now(),
			CreatedAt: time.Time{},
		}
		response = base.NewLoginSuccessResponse(user)
		auth.GenerateNewAuthCookie(w, strconv.Itoa(storedUser.Id))
	} else {
		errors := make([]base.FieldError, 0)
		errors = append(errors, base.FieldError{
			Field:   "username",
			Message: "Такого пользователя нет",
		})
		errors = append(errors, base.FieldError{
			Field:   "password",
			Message: "Такого пользователя нет",
		})
		response = base.NewLoginErrorResponse(errors)
	}
	err = json.NewEncoder(w).Encode(response)
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

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	err = storage.NewRefreshTokenStore(os.Getenv("JWT_SECRET"))
	if err != nil {
		fmt.Println(err)
		return
	}
	storage.NewUserStore()
	storage.NewBudgetStore()

	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/auth/logout", logoutHandler)
	http.HandleFunc("/auth/refresh", refreshTokenHandler)

	fmt.Println("starting server at :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
