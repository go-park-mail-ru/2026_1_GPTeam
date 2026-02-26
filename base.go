package main

import (
	"net/http"
	"time"
)

type SimpleResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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

type RegisterSuccessResponse struct {
	SimpleResponse
	User AuthUser `json:"user"`
}

type LogoutSuccessResponse struct {
	SimpleResponse
	LoggedOutAt time.Time `json:"logged_out_at"`
}

// юзер для авторизации
type AuthUser struct {
	Id        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	LastLogin time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// юзер для профиля
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

type BudgetsIdsResponse struct {
	SimpleResponse
	Len int   `json:"len"`
	Ids []int `json:"ids"`
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

type RequestWithErrors struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors"`
}

type RegisterErrorResponse RequestWithErrors

func NewRegisterErrorResponse() RegisterErrorResponse {
	return RegisterErrorResponse{
		Code:    http.StatusConflict,
		Message: "Пользователь с таким именем или email уже существует",
	}
}

type LoginErrorResponse RequestWithErrors

func NewLoginErrorResponse() LoginErrorResponse {
	return LoginErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: "Неверный логин или пароль",
	}
}

type ValidationErrorResponse RequestWithErrors

func NewValidationErrorResponse() ValidationErrorResponse {
	return ValidationErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "Ошибка валидации",
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
	RequestId string `json:"request_id"`
}

func NewServerErrorResponse(requestId string) ServerErrorResponse {
	return ServerErrorResponse{
		Code:      http.StatusInternalServerError,
		Message:   "Внутренняя ошибка сервера",
		RequestId: requestId,
	}
}
