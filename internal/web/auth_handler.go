package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	web_helpers2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
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
	ctx := context.Background()
	obj.authService.ClearOld(ctx, w, r)
	response := web_helpers2.NewLogoutSuccessResponse()
	web_helpers2.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	isAuth, userId := obj.authService.Refresh(ctx, w, r)
	authUser, ok := obj.userApp.IsAuthUserExists(ctx, isAuth, userId)
	if !ok {
		response := web_helpers2.NewUnauthorizedErrorResponse()
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers2.NewLoginSuccessResponse(authUser)
	web_helpers2.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandlers) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var body web_helpers2.SignupBodyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println(err)
		response := web_helpers2.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []web_helpers2.FieldError{
			web_helpers2.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}
	if body.Username == "" || body.Password == "" || body.Email == "" || body.ConfirmPassword == "" {
		var fieldErrors []web_helpers2.FieldError
		if body.Username == "" {
			fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("username", "Поле обязательно для заполнения"))
		}
		if body.Password == "" {
			fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("password", "Поле обязательно для заполнения"))
		}
		if body.Email == "" {
			fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("email", "Поле обязательно для заполнения"))
		}
		if body.ConfirmPassword == "" {
			fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("confirm_password", "Поле обязательно для заполнения"))
		}
		response := web_helpers2.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", fieldErrors)
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	errors := make([]web_helpers2.FieldError, 0)
	err := validators.ValidateUsername(body.Username)
	if err != nil {
		errors = append(errors, web_helpers2.NewFieldError("username", err.Error()))
	}
	err = validators.ValidatePassword(body.Password)
	if err != nil {
		errors = append(errors, web_helpers2.NewFieldError("password", err.Error()))
	}
	err = validators.ValidateEmail(body.Email)
	if err != nil {
		errors = append(errors, web_helpers2.NewFieldError("email", err.Error()))
	}
	if body.Password != body.ConfirmPassword {
		errors = append(errors, web_helpers2.NewFieldError("password", "Пароли не совпадают"))
		errors = append(errors, web_helpers2.NewFieldError("confirm_password", "Пароли не совпадают"))
	}
	if len(errors) > 0 {
		response := web_helpers2.NewValidationErrorResponse(errors)
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	authUser, err := obj.userApp.Create(ctx, body)
	if err != nil {
		response := web_helpers2.NewValidationErrorResponse([]web_helpers2.FieldError{
			web_helpers2.NewFieldError("field", err.Error()),
		})
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return // ToDo: add err check
	}
	response := web_helpers2.NewSignupSuccessResponse(authUser)
	obj.authService.GenerateNewAuth(ctx, w, authUser.ID)
	web_helpers2.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var userRequest web_helpers2.LoginBodyRequest
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		fmt.Println(err)
		errors := make([]web_helpers2.FieldError, 0)
		errors = append(errors, web_helpers2.NewFieldError("username", "Не удалось прочитать json"))
		errors = append(errors, web_helpers2.NewFieldError("password", "Не удалось прочитать json"))
		response := web_helpers2.NewLoginErrorResponse(errors)
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	storedUser, err := obj.userApp.GetByCredentials(ctx, userRequest)
	if err != nil {
		errors := make([]web_helpers2.FieldError, 0)
		errors = append(errors, web_helpers2.NewFieldError("username", "Неверный логин или пароль"))
		errors = append(errors, web_helpers2.NewFieldError("password", "Неверный логин или пароль"))
		response := web_helpers2.NewLoginErrorResponse(errors)
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}
	user := web_helpers2.User{
		Username:        storedUser.Username,
		Email:           storedUser.Email,
		LastLogin:       time.Now(),
		CreatedAt:       storedUser.CreatedAt,
		AvatarUrl:       storedUser.AvatarUrl,
		Balance:         0,
		BalanceCurrency: "RUB",
	}
	response := web_helpers2.NewLoginSuccessResponse(user)
	obj.authService.GenerateNewAuth(ctx, w, storedUser.Id)
	web_helpers2.WriteResponseJSON(w, response.Code, response)
}
