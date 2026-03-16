package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/jwt"
	"github.com/go-park-mail-ru/2026_1_GPTeam/validators"
)

type JWTHandlers struct {
	useCase     jwt.JWTUseCaseInterface
	auth        auth.AuthenticationServiceInterface
	userUseCase application.UserUseCaseInterface
}

func NewJWTHandler(auth auth.AuthenticationServiceInterface, userUseCase application.UserUseCaseInterface) *JWTHandlers {
	return &JWTHandlers{
		auth:        auth,
		userUseCase: userUseCase,
	} // ToDo: get auth packet; auth packet creates in main
}

func (obj *JWTHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	obj.auth.ClearOld(w, r)
	response := base.NewLogoutSuccessResponse()
	base.WriteResponseJSON(w, response.Code, response)
}

func (obj *JWTHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	isAuth, userID := obj.auth.Refresh(w, r)
	authUser, ok := obj.userUseCase.IsAuthUserExists(ctx, isAuth, userID)
	if !ok {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := base.NewLoginSuccessResponse(authUser)
	base.WriteResponseJSON(w, response.Code, response)
}

func (obj *JWTHandlers) SignUp(w http.ResponseWriter, r *http.Request) {
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

	authUser, err := obj.userUseCase.Create(ctx, body)
	if err != nil {
		response := base.NewValidationErrorResponse([]base.FieldError{
			base.NewFieldError("field", err.Error()),
		})
		base.WriteResponseJSON(w, response.Code, response)
		return // ToDo: add err check
	}
	response := base.NewSignupSuccessResponse(authUser)
	obj.auth.GenerateNewAuth(w, authUser.ID)
	base.WriteResponseJSON(w, response.Code, response)
}

func (obj *JWTHandlers) Login(w http.ResponseWriter, r *http.Request) {
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

	storedUser, err := obj.userUseCase.GetByCredentials(ctx, userRequest)
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
		Balance:         0,
		BalanceCurrency: "RUB",
	}
	response := base.NewLoginSuccessResponse(user)
	obj.auth.GenerateNewAuth(w, storedUser.Id)
	base.WriteResponseJSON(w, response.Code, response)
}
