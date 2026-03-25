package web

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService auth.AuthenticationService
	userApp     application.UserUseCase
	accountApp  application.AccountUseCase
	log         *zap.Logger
}

func NewAuthHandler(auth auth.AuthenticationService, userUseCase application.UserUseCase, accountUseCase application.AccountUseCase) *AuthHandler {
	return &AuthHandler{
		authService: auth,
		userApp:     userUseCase,
		accountApp:  accountUseCase,
		log:         logger.GetLogger(),
	}
}

func (obj *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("logout request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	obj.authService.ClearOld(r.Context(), w, r)
	obj.log.Info("logout success",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewLogoutSuccessResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("refresh token request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	isAuth, userId := obj.authService.Refresh(r.Context(), w, r)
	authUser, ok := obj.userApp.IsAuthUserExists(r.Context(), isAuth, userId)
	if !ok {
		obj.log.Warn("user unauthorized",
			zap.Int("user_id", userId),
			zap.String("request_id", r.Context().Value("request_id").(string)))
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	obj.log.Info("refresh token success",
		zap.Int("user_id", userId),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewLoginSuccessResponse(authUser)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("sign up request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	var body web_helpers.SignupBodyRequest
	if err := web_helpers.ReadRequestJSON(r, &body); err != nil {
		obj.log.Warn("unable to read body",
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
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
		obj.log.Warn("validation error",
			zap.Any("fieldErrors", fieldErrors),
			zap.String("request_id", r.Context().Value("request_id").(string)))
		response := web_helpers.NewSignupErrorResponse(http.StatusBadRequest, "Неверный формат запроса", fieldErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	validationErrors := make([]web_helpers.FieldError, 0)
	if err := validators.ValidateUsername(body.Username); err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("username", err.Error()))
	}
	if err := validators.ValidatePassword(body.Password); err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", err.Error()))
	}
	if err := validators.ValidateEmail(body.Email); err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("email", err.Error()))
	}
	if body.Password != body.ConfirmPassword {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", "Пароли не совпадают"))
		validationErrors = append(validationErrors, web_helpers.NewFieldError("confirm_password", "Пароли не совпадают"))
	}
	if len(validationErrors) > 0 {
		obj.log.Warn("validation error",
			zap.Any("validationErrors", validationErrors),
			zap.String("request_id", r.Context().Value("request_id").(string)))
		response := web_helpers.NewValidationErrorResponse(validationErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	authUser, err := obj.userApp.Create(r.Context(), body)
	if err != nil {
		obj.log.Warn("failed to create user",
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
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
		response := web_helpers.NewServerErrorResponse(r.Context().Value("request_id").(string))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	obj.log.Info("user created",
		zap.Int("user_id", authUser.Id),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	obj.log.Info("new account creating for user",
		zap.Int("user_id", authUser.Id),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	accountModel := models.AccountModel{
		Name:      "base",
		Balance:   0,
		Currency:  "RUB",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	accountId, err := obj.accountApp.Create(r.Context(), accountModel)
	if err != nil {
		obj.log.Warn("failed to create account",
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
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
		response := web_helpers.NewServerErrorResponse(r.Context().Value("request_id").(string))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	obj.log.Info("account created",
		zap.Int("user_id", authUser.Id),
		zap.Int("account_id", accountId),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	if err = obj.accountApp.LinkAccountAndUser(r.Context(), accountId, authUser.Id); err != nil {
		obj.log.Warn("failed to link account and user",
			zap.Int("user_id", authUser.Id),
			zap.Int("account_id", accountId),
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
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
		response := web_helpers.NewServerErrorResponse(r.Context().Value("request_id").(string))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	obj.log.Info("account linked",
		zap.Int("user_id", authUser.Id),
		zap.Int("account_id", accountId),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewSignupSuccessResponse(authUser)
	obj.authService.GenerateNewAuth(r.Context(), w, authUser.Id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("login request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	var userRequest web_helpers.LoginBodyRequest
	if err := web_helpers.ReadRequestJSON(r, &userRequest); err != nil {
		obj.log.Warn("failed to read body",
			zap.String("request_id", r.Context().Value("request_id").(string)),
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
		obj.log.Warn("failed to get user by credentials",
			zap.String("username", userRequest.Username),
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
		response := web_helpers.NewLoginErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("username", "Неверный логин или пароль"),
			web_helpers.NewFieldError("password", "Неверный логин или пароль"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	user := web_helpers.User{
		Username:  storedUser.Username,
		Email:     storedUser.Email,
		CreatedAt: storedUser.CreatedAt,
		AvatarUrl: storedUser.AvatarUrl,
	}
	obj.log.Info("login success",
		zap.Int("user_id", storedUser.Id),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewLoginSuccessResponse(user)
	obj.authService.GenerateNewAuth(r.Context(), w, storedUser.Id)
	obj.log.Info("generated new auth cookie",
		zap.Int("user_id", storedUser.Id),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
