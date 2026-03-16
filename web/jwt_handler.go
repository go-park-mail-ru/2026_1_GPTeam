package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/application/validators"
	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/auth/jwt"
	base2 "github.com/go-park-mail-ru/2026_1_GPTeam/web/base"
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
	}
}

func (obj *JWTHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	obj.auth.ClearOld(w, r)
	response := base2.NewLogoutSuccessResponse()
	base2.WriteResponseJSON(w, response.Code, response)
}

func (obj *JWTHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	isAuth, userID := obj.auth.Refresh(w, r)
	authUser, ok := obj.userUseCase.IsAuthUserExists(ctx, isAuth, userID)
	if !ok {
		response := base2.NewUnauthorizedErrorResponse()
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := base2.NewLoginSuccessResponse(authUser)
	base2.WriteResponseJSON(w, response.Code, response)
}

func (obj *JWTHandlers) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var body base2.SignupBodyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println(err)
		response := base2.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []base2.FieldError{
			base2.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}
	if body.Username == "" || body.Password == "" || body.Email == "" || body.ConfirmPassword == "" {
		var fieldErrors []base2.FieldError
		if body.Username == "" {
			fieldErrors = append(fieldErrors, base2.NewFieldError("username", "Поле обязательно для заполнения"))
		}
		if body.Password == "" {
			fieldErrors = append(fieldErrors, base2.NewFieldError("password", "Поле обязательно для заполнения"))
		}
		if body.Email == "" {
			fieldErrors = append(fieldErrors, base2.NewFieldError("email", "Поле обязательно для заполнения"))
		}
		if body.ConfirmPassword == "" {
			fieldErrors = append(fieldErrors, base2.NewFieldError("confirm_password", "Поле обязательно для заполнения"))
		}
		response := base2.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", fieldErrors)
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	errors := make([]base2.FieldError, 0)
	err := validators.ValidateUsername(body.Username)
	if err != nil {
		errors = append(errors, base2.NewFieldError("username", err.Error()))
	}
	err = validators.ValidatePassword(body.Password)
	if err != nil {
		errors = append(errors, base2.NewFieldError("password", err.Error()))
	}
	err = validators.ValidateEmail(body.Email)
	if err != nil {
		errors = append(errors, base2.NewFieldError("email", err.Error()))
	}
	if body.Password != body.ConfirmPassword {
		errors = append(errors, base2.NewFieldError("password", "Пароли не совпадают"))
		errors = append(errors, base2.NewFieldError("confirm_password", "Пароли не совпадают"))
	}
	if len(errors) > 0 {
		response := base2.NewValidationErrorResponse(errors)
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	authUser, err := obj.userUseCase.Create(ctx, body)
	if err != nil {
		response := base2.NewValidationErrorResponse([]base2.FieldError{
			base2.NewFieldError("field", err.Error()),
		})
		base2.WriteResponseJSON(w, response.Code, response)
		return // ToDo: add err check
	}
	response := base2.NewSignupSuccessResponse(authUser)
	obj.auth.GenerateNewAuth(w, authUser.ID)
	base2.WriteResponseJSON(w, response.Code, response)
}

func (obj *JWTHandlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var userRequest base2.LoginBodyRequest
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		fmt.Println(err)
		errors := make([]base2.FieldError, 0)
		errors = append(errors, base2.NewFieldError("username", "Не удалось прочитать json"))
		errors = append(errors, base2.NewFieldError("password", "Не удалось прочитать json"))
		response := base2.NewLoginErrorResponse(errors)
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	storedUser, err := obj.userUseCase.GetByCredentials(ctx, userRequest)
	if err != nil {
		errors := make([]base2.FieldError, 0)
		errors = append(errors, base2.NewFieldError("username", "Неверный логин или пароль"))
		errors = append(errors, base2.NewFieldError("password", "Неверный логин или пароль"))
		response := base2.NewLoginErrorResponse(errors)
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}
	user := base2.User{
		Username:        storedUser.Username,
		Email:           storedUser.Email,
		LastLogin:       time.Now(),
		CreatedAt:       storedUser.CreatedAt,
		AvatarUrl:       storedUser.AvatarUrl,
		Balance:         0,
		BalanceCurrency: "RUB",
	}
	response := base2.NewLoginSuccessResponse(user)
	obj.auth.GenerateNewAuth(w, storedUser.Id)
	base2.WriteResponseJSON(w, response.Code, response)
}
