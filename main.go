package main

import (
	"encoding/json"
	"fmt"
	"main/auth"
	"main/base"
	"main/jwt"
	"main/middleware"
	"main/storage"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := base.NewMethodError()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	var userRequest base.LoginBodyRequest
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		fmt.Println(err)
		errors := []base.FieldError{
			{
				Field:   "username",
				Message: "Не удалось прочитать json",
			},
			{
				Field:   "password",
				Message: "Не удалось прочитать json",
			},
		}
		response := base.NewLoginErrorResponse(errors)
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	storedUser, exists := storage.FindUserByCredentials(userRequest)
	if !exists {
		errors := make([]base.FieldError, 0)
		errors = append(errors, base.FieldError{
			Field:   "username",
			Message: "Неверный логин или пароль",
		})
		errors = append(errors, base.FieldError{
			Field:   "password",
			Message: "Неверный логин или пароль",
		})
		response := base.NewLoginErrorResponse(errors)
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	user := base.AuthUser{
		ID:        storedUser.Id,
		Username:  storedUser.Username,
		Email:     storedUser.Email,
		LastLogin: time.Now(),
		CreatedAt: time.Time{},
	}
	response := base.NewLoginSuccessResponse(user)
	auth.GenerateNewAuthCookie(w, strconv.Itoa(storedUser.Id))
	base.WriteResponseJSON(w, response.Code, response)
}

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, userID := auth.RefreshToken(w, r)
	data := make(map[string]interface{})
	data["is_auth"] = isAuth
	data["user_id"] = userID
	base.WriteResponseJSON(w, http.StatusOK, data)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := base.NewMethodError()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	var body base.RegisterBodyRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(base.NewValidationErrorResponse(nil))
		return
	}
	auth.GenerateNewAuthCookie(w, "1")
	user := base.AuthUser{
		ID:        1,
		Username:  "username",
		Email:     "email",
		CreatedAt: time.Time{},
	}
	response := base.NewRegisterSuccessResponse(user)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println(err)
	}
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

	err = jwt.NewRefreshTokenStore(os.Getenv("JWT_SECRET"))
	if err != nil {
		fmt.Println(err)
		return
	}
	storage.NewUserStore()
	storage.AddUser(storage.UserInfo{
		Id:        0,
		Username:  "admin",
		Password:  "Adm1n123",
		Email:     "email",
		CreatedAt: time.Now(),
		LastLogin: time.Time{},
		AvatarUrl: "img/123.png",
	})
	storage.NewBudgetStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login", loginHandler)
	mux.HandleFunc("/signup", signupHandler)
	mux.HandleFunc("/auth/logout", logoutHandler)
	mux.HandleFunc("/auth/refresh", refreshTokenHandler)

	handler := middleware.CORSMiddleware(mux)

	server := http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("starting server at :8080")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}
