package web

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/context_helper"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
	"go.uber.org/zap"
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
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("logout request")
	obj.authService.ClearOld(r.Context(), w, r)
	log.Info("logout success")
	response := web_helpers.NewLogoutSuccessResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("refresh token request")
	isAuth, userId := obj.authService.Refresh(r.Context(), w, r)
	authUser, ok := obj.userApp.IsAuthUserExists(r.Context(), isAuth, userId)
	if !ok {
		log.Warn("user unauthorized",
			zap.Int("user_id", userId))
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("refresh token success",
		zap.Int("user_id", userId))
	response := web_helpers.NewLoginSuccessResponse(authUser)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("sign up request")
	var body web_helpers.SignupBodyRequest
	if err := web_helpers.ReadRequestJSON(r, &body); err != nil {
		log.Warn("unable to read body",
			zap.Error(err))
		response := web_helpers.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []web_helpers.FieldError{
			web_helpers.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	body.Username = secure.SanitizeXss(body.Username)
	body.Email = secure.SanitizeXss(body.Email)
	validationErrors := validators.ValidateSignUpUser(body)
	if len(validationErrors) > 0 {
		log.Warn("validation error",
			zap.Any("validationErrors", validationErrors))
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
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("user created",
		zap.Int("user_id", authUser.Id))
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
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("account created",
		zap.Int("user_id", authUser.Id),
		zap.Int("account_id", accountId))
	if err = obj.accountApp.LinkAccountAndUser(r.Context(), accountId, authUser.Id); err != nil {
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
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("account linked",
		zap.Int("user_id", authUser.Id),
		zap.Int("account_id", accountId))
	response := web_helpers.NewSignupSuccessResponse(authUser)
	obj.authService.GenerateNewAuth(r.Context(), w, authUser.Id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("login request")
	var userRequest web_helpers.LoginBodyRequest
	if err := web_helpers.ReadRequestJSON(r, &userRequest); err != nil {
		log.Warn("failed to read body",
			zap.Error(err))
		response := web_helpers.NewLoginErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("username", "Не удалось прочитать json"),
			web_helpers.NewFieldError("password", "Не удалось прочитать json"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	storedUser, err := obj.userApp.GetByCredentials(r.Context(), userRequest)
	if err != nil {
		response := web_helpers.NewLoginErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("username", "Неверный логин или пароль"),
			web_helpers.NewFieldError("password", "Неверный логин или пароль"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	_ = obj.userApp.UpdateLastLogin(r.Context(), storedUser.Id)
	user := web_helpers.User{
		Username:  storedUser.Username,
		Email:     storedUser.Email,
		CreatedAt: storedUser.CreatedAt,
		AvatarUrl: storedUser.AvatarUrl,
	}
	log.Info("login success",
		zap.Int("user_id", storedUser.Id))
	response := web_helpers.NewLoginSuccessResponse(user)
	obj.authService.GenerateNewAuth(r.Context(), w, storedUser.Id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
