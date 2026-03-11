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
	"github.com/go-park-mail-ru/2026_1_GPTeam/validators"

	"github.com/joho/godotenv"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var userRequest base.LoginBodyRequest
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		fmt.Println(err)
		errors := make([]base.FieldError, 0)
		errors = append(errors, base.NewFieldError("username", "Не удалось прочитать json"))
		errors = append(errors, base.NewFieldError("password", "Не удалось прочитать json"))
		response := base.NewLoginErrorResponse(errors)
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	storedUser, exists := storage.FindUserByCredentials(userRequest)
	if !exists {
		errors := make([]base.FieldError, 0)
		errors = append(errors, base.NewFieldError("username", "Неверный логин или пароль"))
		errors = append(errors, base.NewFieldError("password", "Неверный логин или пароль"))
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
			base.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	if body.Username == "" || body.Password == "" || body.Email == "" || body.ConfirmPassword == "" {
		var fieldErrors []base.FieldError
		if body.Username == "" {
			fieldErrors = append(fieldErrors, base.NewFieldError("username", "Поле обязательно для заполнения"))
		}
		if body.Password == "" {
			fieldErrors = append(fieldErrors, base.NewFieldError("password", "Поле обязательно для заполнения"))
		}
		if body.Email == "" {
			fieldErrors = append(fieldErrors, base.NewFieldError("email", "Поле обязательно для заполнения"))
		}
		if body.ConfirmPassword == "" {
			fieldErrors = append(fieldErrors, base.NewFieldError("confirm_password", "Поле обязательно для заполнения"))
		}
		response := base.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", fieldErrors)
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	errors := make([]base.FieldError, 0)
	err := validators.ValidateUsername(body.Username)
	if err != nil {
		errors = append(errors, base.NewFieldError("username", err.Error()))
	}
	err = validators.ValidatePassword(body.Password)
	if err != nil {
		errors = append(errors, base.NewFieldError("password", err.Error()))
	}
	err = validators.ValidateEmail(body.Email)
	if err != nil {
		errors = append(errors, base.NewFieldError("email", err.Error()))
	}
	if body.Password != body.ConfirmPassword {
		errors = append(errors, base.NewFieldError("password", "Пароли не совпадают"))
		errors = append(errors, base.NewFieldError("confirm_password", "Пароли не совпадают"))
	}
	if len(errors) > 0 {
		response := base.NewValidationErrorResponse(errors)
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
		response := base.NewSignupErrorResponse(http.StatusConflict, "Пользователь с таким логином уже существует", []base.FieldError{
			base.NewFieldError("username", "Пользователь с таким логином уже существует"),
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	if storage.EmailExists(body.Email) {
		response := base.NewSignupErrorResponse(http.StatusConflict, "Пользователь с таким email уже существует", []base.FieldError{
			base.NewFieldError("email", "Пользователь с таким email уже существует"),
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

func getBudgetsHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	ids := storage.GetBudgetIDsByUserID(authUser.Id)
	response := base.NewBudgetsIDsResponse(ids)
	base.WriteResponseJSON(w, response.Code, response)
}

func getBudgetHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
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

	budget, ok := storage.GetBudgetByIDAndUserID(budgetID, authUser.Id)
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

func createBudgetHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	var body base.BudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response := base.NewBudgetErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []base.FieldError{
			base.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	var fieldErrors []base.FieldError
	if body.Title == "" {
		fieldErrors = append(fieldErrors, base.NewFieldError("title", "Поле обязательно для заполнения"))
	}
	if body.Description == "" {
		fieldErrors = append(fieldErrors, base.NewFieldError("description", "Поле обязательно для заполнения"))
	}
	if body.Target == 0 {
		fieldErrors = append(fieldErrors, base.NewFieldError("target", "Поле обязательно для заполнения"))
	}
	if body.Currency == "" {
		fieldErrors = append(fieldErrors, base.NewFieldError("currency", "Поле обязательно для заполнения"))
	}
	err := validators.ValidateCurrency(body.Currency)
	if err != nil {
		fieldErrors = append(fieldErrors, base.NewFieldError("currency", err.Error()))
	}
	err = validators.ValidateTargetBudget(body.Target)
	if err != nil {
		fieldErrors = append(fieldErrors, base.NewFieldError("target", err.Error()))
	}
	err = validators.ValidateStartDate(body.StartAt)
	if err != nil {
		fieldErrors = append(fieldErrors, base.NewFieldError("start_at", err.Error()))
	}
	err = validators.ValidateEndDate(body.StartAt, body.EndAt)
	if err != nil {
		fieldErrors = append(fieldErrors, base.NewFieldError("end_at", err.Error()))
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
		Author:      authUser.Id,
	}
	id := storage.AddBudget(budget)
	response := base.NewBudgetCreateSuccessResponse(id)
	base.WriteResponseJSON(w, response.Code, response)
}

func DeleteBudgetHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
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

	ok = storage.DeleteBudgetByIDAndUserID(budgetID, authUser.Id)
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
	response := base.NewLogoutSuccessResponse()
	base.WriteResponseJSON(w, response.Code, response)
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
	response := base.NewBalanceResponse(balance, currency, 0, 0)
	base.WriteResponseJSON(w, response.Code, response)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := base.User{
		Username:        authUser.Username,
		Email:           authUser.Email,
		CreatedAt:       authUser.CreatedAt,
		LastLogin:       authUser.LastLogin,
		AvatarUrl:       authUser.AvatarUrl,
		Balance:         authUser.Balance,
		BalanceCurrency: authUser.BalanceCurrency,
	}
	response := base.NewLoginSuccessResponse(userResponse)
	base.WriteResponseJSON(w, response.Code, response)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	err = jwt.NewRefreshTokenStore(os.Getenv("JWT_SECRET"), os.Getenv("JWT_VERSION"))
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
	mux.Handle("/auth/signup", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(signupHandler)))
	mux.Handle("/auth/logout", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(logoutHandler)))
	mux.Handle("/auth/refresh", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(refreshTokenHandler)))
	mux.Handle("/profile/balance", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(balanceHandler)))
	mux.Handle("/get_budgets", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(getBudgetsHandler)))
	mux.Handle("/get_budget/{id}", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(getBudgetHandler)))
	mux.Handle("/budget", middleware.MethodValidationMiddleware(http.MethodPost)(http.HandlerFunc(createBudgetHandler)))
	mux.Handle("/budget/{id}", middleware.MethodValidationMiddleware(http.MethodDelete)(http.HandlerFunc(DeleteBudgetHandler)))
	mux.Handle("/profile", middleware.MethodValidationMiddleware(http.MethodGet)(http.HandlerFunc(profileHandler)))

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
