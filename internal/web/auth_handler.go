package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
)

type AuthHandlers struct {
	authService auth.AuthenticationService
	userApp     application.UserUseCase
}

func NewJWTHandler(auth auth.AuthenticationService, userUseCase application.UserUseCase) *AuthHandlers {
	return &AuthHandlers{
		authService: auth,
		userApp:     userUseCase,
	}
}

func (obj *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	obj.authService.ClearOld(r.Context(), w, r)
	response := web_helpers.NewLogoutSuccessResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
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

func (obj *AuthHandlers) SignUp(w http.ResponseWriter, r *http.Request) {
	var body web_helpers.SignupBodyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println(err)
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

	errors := make([]web_helpers.FieldError, 0)
	err := validators.ValidateUsername(body.Username)
	if err != nil {
		errors = append(errors, web_helpers.NewFieldError("username", err.Error()))
	}
	err = validators.ValidatePassword(body.Password)
	if err != nil {
		errors = append(errors, web_helpers.NewFieldError("password", err.Error()))
	}
	err = validators.ValidateEmail(body.Email)
	if err != nil {
		errors = append(errors, web_helpers.NewFieldError("email", err.Error()))
	}
	if body.Password != body.ConfirmPassword {
		errors = append(errors, web_helpers.NewFieldError("password", "Пароли не совпадают"))
		errors = append(errors, web_helpers.NewFieldError("confirm_password", "Пароли не совпадают"))
	}
	if len(errors) > 0 {
		response := web_helpers.NewValidationErrorResponse(errors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	authUser, err := obj.userApp.Create(r.Context(), body)
	if err != nil {
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("field", err.Error()),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return // ToDo: add err check
	}
	response := web_helpers.NewSignupSuccessResponse(authUser)
	obj.authService.GenerateNewAuth(r.Context(), w, authUser.ID)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var userRequest web_helpers.LoginBodyRequest
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		fmt.Println(err)
		errors := make([]web_helpers.FieldError, 0)
		errors = append(errors, web_helpers.NewFieldError("username", "Не удалось прочитать json"))
		errors = append(errors, web_helpers.NewFieldError("password", "Не удалось прочитать json"))
		response := web_helpers.NewLoginErrorResponse(errors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	storedUser, err := obj.userApp.GetByCredentials(r.Context(), userRequest)
	if err != nil {
		errors := make([]web_helpers.FieldError, 0)
		errors = append(errors, web_helpers.NewFieldError("username", "Неверный логин или пароль"))
		errors = append(errors, web_helpers.NewFieldError("password", "Неверный логин или пароль"))
		response := web_helpers.NewLoginErrorResponse(errors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	user := web_helpers.User{
		Username:        storedUser.Username,
		Email:           storedUser.Email,
		LastLogin:       time.Now(),
		CreatedAt:       storedUser.CreatedAt,
		AvatarUrl:       storedUser.AvatarUrl,
		Balance:         0,
		BalanceCurrency: "RUB",
	}
	response := web_helpers.NewLoginSuccessResponse(user)
	obj.authService.GenerateNewAuth(r.Context(), w, storedUser.Id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
