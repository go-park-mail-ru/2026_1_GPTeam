package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
	"github.com/go-park-mail-ru/2026_1_GPTeam/middleware"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"

	"github.com/joho/godotenv"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
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
	var body base.SignupBodyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println(err)
		response := base.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []base.FieldError{
			{Field: "", Message: "Не удалось прочитать тело запроса"},
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	if body.Username == "" || body.Password == "" || body.Email == "" || body.ConfirmPassword == "" {
		var fieldErrors []base.FieldError
		if body.Username == "" {
			fieldErrors = append(fieldErrors, base.FieldError{Field: "username", Message: "Поле обязательно для заполнения"})
		}
		if body.Password == "" {
			fieldErrors = append(fieldErrors, base.FieldError{Field: "password", Message: "Поле обязательно для заполнения"})
		}
		if body.Email == "" {
			fieldErrors = append(fieldErrors, base.FieldError{Field: "email", Message: "Поле обязательно для заполнения"})
		}
		if body.ConfirmPassword == "" {
			fieldErrors = append(fieldErrors, base.FieldError{Field: "confirm_password", Message: "Поле обязательно для заполнения"})
		}
		response := base.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", fieldErrors)
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
		response := base.NewSignupErrorResponse(http.StatusConflict, "Пользователь с таким именем уже существует", []base.FieldError{
			{
				Field:   "username",
				Message: "Пользователь с таким именем уже существует",
			},
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	if storage.EmailExists(body.Email) {
		response := base.NewSignupErrorResponse(http.StatusConflict, "Пользователь с таким email уже существует", []base.FieldError{
			{
				Field:   "email",
				Message: "Пользователь с таким email уже существует",
			},
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	if body.Password != body.ConfirmPassword {
		response := base.NewSignupErrorResponse(http.StatusBadRequest, "Пароли не совпадают", []base.FieldError{
			{
				Field:   "password",
				Message: "Пароли не совпадают",
			},
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

func GetBudgetsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromCookie(r)
	if err != nil {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	ids := storage.GetBudgetIDsByUserID(userID)
	response := base.NewBudgetsIDsResponse(ids)
	base.WriteResponseJSON(w, response.Code, response)
}

func GetBudgetHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromCookie(r)
	if err != nil {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		response := base.NewNotFoundErrorResponse("Не указан ID бюджета")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetID, err := strconv.Atoi(idStr)
	if err != nil {
		response := base.NewNotFoundErrorResponse("Неверный ID бюджета")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	budget, ok := storage.GetBudgetByIDAndUserID(budgetID, userID)
	if !ok {
		response := base.NewNotFoundErrorResponse("Бюджет не найден")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	result := base.BudgetRequest{
		Title:       budget.Title,
		Description: budget.Description,
		CreatedAt:   budget.CreatedAt,
		StartAt:     budget.StartAt,
		EndAt:       budget.EndAt,
		Actual:      budget.Actual,
		Target:      budget.Target,
		Currency:    budget.Currency,
	}
	response := base.NewBudgetGetSuccessResponse(result)
	base.WriteResponseJSON(w, response.Code, response)
}

func CreateBudgetHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromCookie(r)
	if err != nil {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	var body base.BudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response := base.NewBudgetErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []base.FieldError{
			{Field: "", Message: "Не удалось прочитать тело запроса"},
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	var fieldErrors []base.FieldError
	if body.Title == "" {
		fieldErrors = append(fieldErrors, base.FieldError{Field: "title", Message: "Поле обязательно для заполнения"})
	}
	if body.Target == 0 {
		fieldErrors = append(fieldErrors, base.FieldError{Field: "target", Message: "Поле обязательно для заполнения"})
	}
	if body.Currency == "" {
		fieldErrors = append(fieldErrors, base.FieldError{Field: "currency", Message: "Поле обязательно для заполнения"})
	}
	if len(fieldErrors) > 0 {
		response := base.NewBudgetErrorResponse(http.StatusBadRequest, "Ошибка валидации", fieldErrors)
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	budget := storage.BudgetInfo{
		Title:       body.Title,
		Description: body.Description,
		CreatedAt:   time.Now(),
		StartAt:     body.StartAt,
		EndAt:       body.EndAt,
		Actual:      0,
		Target:      body.Target,
		Currency:    body.Currency,
		Author:      userID,
	}
	id := storage.AddBudget(budget)
	response := base.NewBudgetCreateSuccessResponse(id)
	base.WriteResponseJSON(w, response.Code, response)
}

func DeleteBudgetHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromCookie(r)
	if err != nil {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		response := base.NewNotFoundErrorResponse("Не указан ID бюджета")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetID, err := strconv.Atoi(idStr)
	if err != nil {
		response := base.NewNotFoundErrorResponse("Неверный ID бюджета")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	ok := storage.DeleteBudgetByIDAndUserID(budgetID, userID)
	if !ok {
		response := base.NewNotFoundErrorResponse("Бюджет не найден")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := base.NewBudgetDeleteSuccessResponse()
	base.WriteResponseJSON(w, response.Code, response)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearOldToken(w, r)
	w.Header().Set("Content-Type", "application/json")
}

func balanceHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
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
	mux.Handle("/auth/login", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(loginHandler)))
	mux.Handle("/signup", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(signupHandler)))
	mux.Handle("/auth/logout", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(logoutHandler)))
	mux.Handle("/auth/refresh", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(refreshTokenHandler)))
	mux.Handle("/profile/balance", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(balanceHandler)))
	mux.Handle("/get_budgets", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(GetBudgetsHandler)))
	mux.Handle("/get_budget/{id}", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(GetBudgetHandler)))
	mux.Handle("/budget", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(CreateBudgetHandler)))
	mux.Handle("/budget/{id}", middleware.MethodValidationMiddleware(http.MethodDelete)(http.HandlerFunc(DeleteBudgetHandler)))

	handler := middleware.AuthMiddleware(mux)
	handler = middleware.CORSMiddleware(handler)

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
