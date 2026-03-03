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
	var body base.SignupBodyRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Println(err)
		response := base.NewSignupErrorResponse([]base.FieldError{
			{
				Field:   "username",
				Message: "Не удалось прочитать json",
			},
			{
				Field:   "password",
				Message: "Не удалось прочитать json",
			},
			{
				Field:   "confirm_password",
				Message: "Не удалось прочитать json",
			},
			{
				Field:   "email",
				Message: "Не удалось прочитать json",
			},
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	user := storage.UserInfo{
		Username:  body.Username,
		Password:  body.Password,
		Email:     body.Email,
		CreatedAt: time.Now(),
		LastLogin: time.Time{},
	}
	if storage.UserExists(body.Username) {
		response := base.NewSignupErrorResponse([]base.FieldError{
			{
				Field:   "username",
				Message: "Пользователь с таким именем уже существует",
			},
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	if storage.EmailExists(body.Email) {
		response := base.NewSignupErrorResponse([]base.FieldError{
			{
				Field:   "email",
				Message: "Пользователь с таким email уже существует",
			},
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	if body.Password != body.ConfirmPassword {
		response := base.NewSignupErrorResponse([]base.FieldError{
			{
				Field:   "confirm_password",
				Message: "Пароли не совпадают",
			},
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	id := storage.AddUser(user)
	authUser := base.AuthUser{
		ID:        id,
		Username:  user.Username,
		Email:     user.Email,
		LastLogin: time.Now(),
	}
	response := base.NewSignupSuccessResponse(authUser)
	auth.GenerateNewAuthCookie(w, strconv.Itoa(id))
	base.WriteResponseJSON(w, response.Code, response)
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
