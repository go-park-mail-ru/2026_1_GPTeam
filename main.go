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
	user := base.User{
		Username:        storedUser.Username,
		Email:           storedUser.Email,
		LastLogin:       time.Now(),
		CreatedAt:       storedUser.CreatedAt,
		AvatarUrl:       storedUser.AvatarUrl,
		Balance:         storedUser.Balance,
		BalanceCurrency: storedUser.BalanceCurrency,
	}
	response := base.NewLoginSuccessResponse(user)
	auth.GenerateNewAuthCookie(w, strconv.Itoa(storedUser.Id))
	base.WriteResponseJSON(w, response.Code, response)
}

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := base.NewMethodError()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	isAuth, userID := auth.RefreshToken(w, r)
	authUser, ok := storage.IsAuthUserInDatabase(isAuth, userID)
	if !ok {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := base.NewLoginSuccessResponse(authUser)
	base.WriteResponseJSON(w, response.Code, response)
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

func isLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := base.NewMethodError()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	isAuth, userID := auth.IsAuth(r)
	authUser, ok := storage.IsAuthUserInDatabase(isAuth, userID)
	if !ok {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := base.NewLoginSuccessResponse(authUser)
	base.WriteResponseJSON(w, response.Code, response)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, userID := auth.IsAuth(r)
	if !isAuth {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	storedUser := storage.UserInfo{
		Id:        0,
		Username:  "admin",
		Password:  "Adm1n123",
		Email:     "email",
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
		AvatarUrl: "img/123.png",
		Balance:   0,
	}
	_ = userID
	base.WriteResponseJSON(w, http.StatusOK, storedUser)
}

func balanceHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, userID := auth.IsAuth(r)
	authUser, ok := storage.IsAuthUserInDatabase(isAuth, userID)
	if !ok {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	balance := authUser.Balance
	currency := authUser.BalanceCurrency
	response := base.NewBalanceResponse(balance, currency, 100, 46)
	base.WriteResponseJSON(w, response.Code, response)
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
		Id:              0,
		Username:        "admin",
		Password:        "Adm1n123",
		Email:           "email",
		CreatedAt:       time.Now(),
		LastLogin:       time.Now(),
		AvatarUrl:       "img/123.png",
		Balance:         100.5,
		BalanceCurrency: "RUB",
	})
	storage.NewBudgetStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login", loginHandler)
	mux.HandleFunc("/signup", signupHandler)
	mux.HandleFunc("/auth/logout", logoutHandler)
	mux.HandleFunc("/auth/refresh", refreshTokenHandler)
	mux.HandleFunc("/profile", profileHandler)
	mux.HandleFunc("/is_login", isLogin)
	mux.HandleFunc("/balance", balanceHandler)

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
