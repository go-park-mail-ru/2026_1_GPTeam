package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
)

type AuthHandler struct {
	authService auth.AuthenticationService
	userApp     application.UserUseCase
	accountApp  application.AccountUseCase
}

func NewAuthHandler(auth auth.AuthenticationService, userUseCase application.UserUseCase, accountUseCase application.AccountUseCase) *AuthHandler {
	return &AuthHandler{
		authService: auth,
		userApp:     userUseCase,
		accountApp:  accountUseCase,
	}
}

func (obj *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	obj.authService.ClearOld(r.Context(), w, r)
	response := web_helpers.NewLogoutSuccessResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	isAuth, userId := obj.authService.Refresh(r.Context(), w, r)
	authUser, ok := obj.userApp.IsAuthUserExists(r.Context(), isAuth, userId)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewLoginSuccessResponse(authUser)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var body web_helpers.SignupBodyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response := web_helpers.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []web_helpers.FieldError{
			web_helpers.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	if body.Username == "" || body.Password == "" || body.Email == "" || body.ConfirmPassword == "" {
		var fieldErrors []web_helpers.FieldError
		if body.Username == "" {
			fieldErrors = append(fieldErrors, web_helpers.NewFieldError("username", "Поле обязательно для заполнения"))
		}
		if body.Password == "" {
			fieldErrors = append(fieldErrors, web_helpers.NewFieldError("password", "Поле обязательно для заполнения"))
		}
		if body.Email == "" {
			fieldErrors = append(fieldErrors, web_helpers.NewFieldError("email", "Поле обязательно для заполнения"))
		}
		if body.ConfirmPassword == "" {
			fieldErrors = append(fieldErrors, web_helpers.NewFieldError("confirm_password", "Поле обязательно для заполнения"))
		}
		response := web_helpers.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", fieldErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	validationErrors := make([]web_helpers.FieldError, 0)
	err := validators.ValidateUsername(body.Username)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("username", err.Error()))
	}
	err = validators.ValidatePassword(body.Password)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", err.Error()))
	}
	err = validators.ValidateEmail(body.Email)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("email", err.Error()))
	}
	if body.Password != body.ConfirmPassword {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", "Пароли не совпадают"))
		validationErrors = append(validationErrors, web_helpers.NewFieldError("confirm_password", "Пароли не совпадают"))
	}
	if len(validationErrors) > 0 {
		response := web_helpers.NewValidationErrorResponse(validationErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	authUser, err := obj.userApp.Create(r.Context(), body)
	if err != nil {
		if errors.Is(err, application.HashPasswordError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
				web_helpers.NewFieldError("password", "Придумайте другой пароль"),
			})
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.DuplicatedDataError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Такой пользователь уже существует"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.ConstraintError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Введены некорректные данные"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewServerErrorResponse("req_id")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountModel := models.AccountModel{
		Name:      "base",
		Balance:   0,
		Currency:  "RUB",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	accountId, err := obj.accountApp.Create(r.Context(), accountModel)
	if err != nil {
		if errors.Is(err, repository.AccountDuplicatedDataError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Такой счёт уже существует"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.ConstraintError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Введены некорректные данные"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewServerErrorResponse("req_id")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	err = obj.accountApp.LinkAccountAndUser(r.Context(), accountId, accountId)
	if err != nil {
		if errors.Is(err, repository.ConstraintError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Введены некорректные данные"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.AccountForeignKeyError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Счёта не существует"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewServerErrorResponse("req_id")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewSignupSuccessResponse(authUser)
	obj.authService.GenerateNewAuth(r.Context(), w, authUser.Id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var userRequest web_helpers.LoginBodyRequest
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		fmt.Println(err)
		validationErrors := make([]web_helpers.FieldError, 0)
		validationErrors = append(validationErrors, web_helpers.NewFieldError("username", "Не удалось прочитать json"))
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", "Не удалось прочитать json"))
		response := web_helpers.NewLoginErrorResponse(validationErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	storedUser, err := obj.userApp.GetByCredentials(r.Context(), userRequest)
	if err != nil {
		validationErrors := make([]web_helpers.FieldError, 0)
		validationErrors = append(validationErrors, web_helpers.NewFieldError("username", "Неверный логин или пароль"))
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", "Неверный логин или пароль"))
		response := web_helpers.NewLoginErrorResponse(validationErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	user := web_helpers.User{
		Username:  storedUser.Username,
		Email:     storedUser.Email,
		CreatedAt: storedUser.CreatedAt,
		AvatarUrl: storedUser.AvatarUrl,
	}
	response := web_helpers.NewLoginSuccessResponse(user)
	obj.authService.GenerateNewAuth(r.Context(), w, storedUser.Id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
