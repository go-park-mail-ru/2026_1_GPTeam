package base

import (
	"net/http"
	"time"
)

type SimpleResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type OkResponse SimpleResponse

func NewOkResponse() OkResponse {
	return OkResponse{
		Code:    http.StatusOK,
		Message: "Ok",
	}
}

type LoginSuccessResponse struct {
	SimpleResponse
	User AuthUser `json:"user"`
}

func NewLoginSuccessResponse(user AuthUser) LoginSuccessResponse {
	return LoginSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Вход в систему прошел успешно",
		},
		User: user,
	}
}

type RegisterSuccessResponse struct {
	SimpleResponse
	User AuthUser `json:"user"`
}

func NewRegisterSuccessResponse(user AuthUser) RegisterSuccessResponse {
	return RegisterSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Регистрация прошла успешно",
		},
		User: user,
	}
}

type LogoutSuccessResponse struct {
	SimpleResponse
	LoggedOutAt time.Time `json:"logged_out_at"`
}

func NewLogoutSuccessResponse(loggedOutAt time.Time) LogoutSuccessResponse {
	return LogoutSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Успешный выход из системы",
		},
		LoggedOutAt: loggedOutAt,
	}
}

type AuthUser struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	LastLogin time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type User struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	AvatarUrl string    `json:"avatar_url"`
}

type Budget struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Actual      int       `json:"actual"`
	Target      int       `json:"target"`
	Currency    string    `json:"currency"`
}

type BalanceResponse struct {
	SimpleResponse
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

func NewBalanceResponse(balance float64, currency string) BalanceResponse {
	return BalanceResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		Balance:  balance,
		Currency: currency,
	}
}

type BudgetsIDsResponse struct {
	SimpleResponse
	Len int   `json:"len"`
	IDs []int `json:"ids"`
}

func NewBudgetsIDsResponse(ids []int) BudgetsIDsResponse {
	return BudgetsIDsResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		Len: len(ids),
		IDs: ids,
	}
}

type RegisterBodyRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginBodyRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewFieldError(field string, message string) FieldError {
	return FieldError{
		Field:   field,
		Message: message,
	}
}

type RequestWithErrors struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors"`
}

type RegisterErrorResponse RequestWithErrors

func NewRegisterErrorResponse(errors []FieldError) RegisterErrorResponse {
	return RegisterErrorResponse{
		Code:    http.StatusConflict,
		Message: "Пользователь с таким именем или email уже существует",
		Errors:  errors,
	}
}

type LoginErrorResponse RequestWithErrors

func NewLoginErrorResponse(errors []FieldError) LoginErrorResponse {
	return LoginErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: "Неверный логин или пароль",
		Errors:  errors,
	}
}

type ValidationErrorResponse RequestWithErrors

func NewValidationErrorResponse(errors []FieldError) ValidationErrorResponse {
	return ValidationErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "Ошибка валидации",
		Errors:  errors,
	}
}

type UnauthorizedErrorResponse SimpleResponse

func NewUnauthorizedErrorResponse() UnauthorizedErrorResponse {
	return UnauthorizedErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: "Не авторизован — невалидный или отсутствующий токен",
	}
}

type ForbiddenErrorResponse SimpleResponse

func NewForbiddenErrorResponse() ForbiddenErrorResponse {
	return ForbiddenErrorResponse{
		Code:    http.StatusForbidden,
		Message: "Доступ запрещён",
	}
}

type NotFoundErrorResponse SimpleResponse

func NewNotFoundErrorResponse() NotFoundErrorResponse {
	return NotFoundErrorResponse{
		Code:    http.StatusNotFound,
		Message: "Не найдено",
	}
}

type ServerErrorResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func NewServerErrorResponse(requestID string) ServerErrorResponse {
	return ServerErrorResponse{
		Code:      http.StatusInternalServerError,
		Message:   "Внутренняя ошибка сервера",
		RequestID: requestID,
	}
}

type MethodError SimpleResponse

func NewMethodError() MethodError {
	return MethodError{
		Code:    http.StatusMethodNotAllowed,
		Message: "Метод не поддерживается",
	}
}
