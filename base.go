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
	Code    int    `json:"code"`
	Message string `json:"message"`
	User    User   `json:"user"`
}

type RegisterSuccessResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	User    User   `json:"user"`
}

type LogoutSuccessResponse struct {
	Code        int       `json:"code"`
	Message     string    `json:"message"`
	LoggedOutAt time.Time `json:"logged_out_at"`
}

type User struct {
	Id        int       `json:"id"`
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

type RegisterBodyRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginBodyRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RequestWithErrors struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Errors  []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"errors"`
}

type RegisterErrorResponse RequestWithErrors

func NewRegisterErrorResponse() RegisterErrorResponse {
	return RegisterErrorResponse{
		Code:    http.StatusConflict,
		Message: "User with this username or email already exists",
	}
}

type LoginErrorResponse RequestWithErrors

func NewLoginErrorResponse() LoginErrorResponse {
	return LoginErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: "Invalid username or password",
	}
}

type ValidationErrorResponse RequestWithErrors

func NewValidationErrorResponse() ValidationErrorResponse {
	return ValidationErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "Validation error",
	}
}

type UnauthorizedErrorResponse SimpleResponse

func NewUnauthorizedErrorResponse() UnauthorizedErrorResponse {
	return UnauthorizedErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized - invalid or missing token",
	}
}

type ForbiddenErrorResponse SimpleResponse

func NewForbiddenErrorResponse() ForbiddenErrorResponse {
	return ForbiddenErrorResponse{
		Code:    http.StatusForbidden,
		Message: "403 - Forbidden",
	}
}

type NotFoundErrorResponse SimpleResponse

func NewNotFoundErrorResponse() NotFoundErrorResponse {
	return NotFoundErrorResponse{
		Code:    http.StatusNotFound,
		Message: "404 - Not Found",
	}
}

type ServerErrorResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

func NewServerErrorResponse() ServerErrorResponse {
	return ServerErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
	}
}
