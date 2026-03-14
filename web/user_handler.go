package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
	"github.com/go-park-mail-ru/2026_1_GPTeam/validators"
)

type UserHandler struct {
	UseCase *application.UserUseCase
}

func NewUserHandler(useCase *application.UserUseCase) *UserHandler {
	return &UserHandler{UseCase: useCase}
}

func (obj *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
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

	authUser, err := obj.UseCase.Create(ctx, body)
	if err != nil {
		response := base.NewValidationErrorResponse([]base.FieldError{
			base.NewFieldError("field", err.Error()),
		})
		base.WriteResponseJSON(w, response.Code, response)
		return // ToDo: add err check
	}
	response := base.NewSignupSuccessResponse(authUser)
	auth.GenerateNewAuthCookie(w, strconv.Itoa(authUser.ID))
	base.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
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

	storedUser, err := obj.UseCase.GetByCredentials(ctx, userRequest)
	if err != nil {
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

func (obj *UserHandler) Balance(w http.ResponseWriter, r *http.Request) {
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

func (obj *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
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
