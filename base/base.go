package base

import (
	"fmt"
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
	User User `json:"user"`
}

func NewLoginSuccessResponse(user User) LoginSuccessResponse {
	return LoginSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Вход в систему прошел успешно",
		},
		User: user,
	}
}

type SignupSuccessResponse struct {
	SimpleResponse
	User AuthUser `json:"user"`
}

func NewSignupSuccessResponse(user AuthUser) SignupSuccessResponse {
	return SignupSuccessResponse{
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
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	CreatedAt       time.Time `json:"created_at"`
	LastLogin       time.Time `json:"last_login,omitempty"`
	AvatarUrl       string    `json:"avatar_url"`
	Balance         float64   `json:"balance"`
	BalanceCurrency string    `json:"currency"`
}

type BudgetRequest struct {
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
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Date     string  `json:"date"`
}

func NewBalanceResponse(balance float64, currency string, income float64, expenses float64) BalanceResponse {
	months := map[time.Month]string{
		time.January:   "Январь",
		time.February:  "Февраль",
		time.March:     "Март",
		time.April:     "Апрель",
		time.May:       "Май",
		time.June:      "Июнь",
		time.July:      "Июль",
		time.August:    "Август",
		time.September: "Сентябрь",
		time.October:   "Октябрь",
		time.November:  "Ноябрь",
		time.December:  "Декабрь",
	}
	curMonth := time.Now().Month()
	curYear := time.Now().Year()
	curTime := fmt.Sprintf("%s %d", months[curMonth], curYear)
	return BalanceResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		Balance:  balance,
		Currency: currency,
		Income:   income,
		Expenses: expenses,
		Date:     curTime,
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

type SignupBodyRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Email           string `json:"email"`
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

type SignupErrorResponse RequestWithErrors

func NewSignupErrorResponse(code int, message string, errors []FieldError) SignupErrorResponse {
	return SignupErrorResponse{
		Code:    code,
		Message: message,
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

func NewNotFoundErrorResponse(message string) NotFoundErrorResponse {
	return NotFoundErrorResponse{
		Code:    http.StatusNotFound,
		Message: message,
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

type BudgetErrorResponse RequestWithErrors

func NewBudgetErrorResponse(code int, message string, errors []FieldError) BudgetErrorResponse {
	return BudgetErrorResponse{
		Code:    code,
		Message: message,
		Errors:  errors,
	}
}

type BudgetDeleteSuccessResponse SimpleResponse

func NewBudgetDeleteSuccessResponse() BudgetDeleteSuccessResponse {
	return BudgetDeleteSuccessResponse{
		Code:    http.StatusOK,
		Message: "Бюджет успешно удалён",
	}
}

type BudgetUpdateSuccessResponse SimpleResponse

func NewBudgetUpdateSuccessResponse() BudgetUpdateSuccessResponse {
	return BudgetUpdateSuccessResponse{
		Code:    http.StatusOK,
		Message: "Бюджет успешно обновлён",
	}
}

type BudgetCreateSuccessResponse struct {
	SimpleResponse
	BudgetID int `json:"budget_id"`
}

func NewBudgetCreateSuccessResponse(budgetID int) BudgetCreateSuccessResponse {
	return BudgetCreateSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Бюджет успешно создан",
		},
		BudgetID: budgetID,
	}
}

type BudgetGetSuccessResponse struct {
	SimpleResponse
	Budget BudgetRequest `json:"budget"`
}

func NewBudgetGetSuccessResponse(budget BudgetRequest) BudgetGetSuccessResponse {
	return BudgetGetSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Бюджет успешно получен",
		},
		Budget: budget,
	}
}
