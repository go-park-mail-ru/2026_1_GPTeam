package web_helpers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
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

func NewLogoutSuccessResponse() LogoutSuccessResponse {
	loggedOutAt := time.Now()
	return LogoutSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Успешный выход из системы",
		},
		LoggedOutAt: loggedOutAt,
	}
}

type AuthUser struct {
	Id        int       `json:"id"`
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

type CurrencyBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
}

type BalanceResponse struct {
	SimpleResponse
	Balances []CurrencyBalance `json:"balances"`
	Date     string            `json:"date"`
}

func NewBalanceResponse(balances []CurrencyBalance) BalanceResponse {
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
		Balances: balances,
		Date:     curTime,
	}
}

type BudgetsIdsResponse struct {
	SimpleResponse
	Len int   `json:"len"`
	Ids []int `json:"ids"`
}

func NewBudgetsIdsResponse(ids []int) BudgetsIdsResponse {
	return BudgetsIdsResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		Len: len(ids),
		Ids: ids,
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
	RequestId string `json:"request_id"`
}

func NewServerErrorResponse(requestId string) ServerErrorResponse {
	return ServerErrorResponse{
		Code:      http.StatusInternalServerError,
		Message:   "Внутренняя ошибка сервера",
		RequestId: requestId,
	}
}

type MethodError SimpleResponse

func NewMethodError() MethodError {
	return MethodError{
		Code:    http.StatusMethodNotAllowed,
		Message: "Метод не поддерживается",
	}
}

func NewBadRequestErrorResponse(message string) SimpleResponse {
	return SimpleResponse{
		Code:    http.StatusBadRequest,
		Message: message,
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
	BudgetId int `json:"budget_id"`
}

func NewBudgetCreateSuccessResponse(budgetId int) BudgetCreateSuccessResponse {
	return BudgetCreateSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Бюджет успешно создан",
		},
		BudgetId: budgetId,
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

type CurrencyCodesResponse struct {
	SimpleResponse
	CurrencyCodes []string `json:"currency_codes"`
}

func NewCurrencyCodesResponse(codes []string) CurrencyCodesResponse {
	return CurrencyCodesResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		CurrencyCodes: codes,
	}
}

type TransactionsIdsResponse struct {
	SimpleResponse
	Len int   `json:"len"`
	Ids []int `json:"ids"`
}

func NewTransactionsIdsResponse(ids []int) TransactionsIdsResponse {
	return TransactionsIdsResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		Len: len(ids),
		Ids: ids,
	}
}

type TransactionRequest struct {
	AccountId       int       `json:"account_id"`
	Value           float64   `json:"value"`
	Type            string    `json:"type"`
	Category        string    `json:"category"`
	Currency        string    `json:"currency"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
}

type TransactionCreateSuccessResponse struct {
	SimpleResponse
	TransactionId int `json:"transaction_id"`
}

func NewTransactionCreateSuccessResponse(id int) *TransactionCreateSuccessResponse {
	return &TransactionCreateSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		TransactionId: id,
	}
}

type TransactionDeleteSuccessResponse struct {
	SimpleResponse
	TransactionId int `json:"transaction_id"`
}

func NewTransactionDeleteSuccessResponse(id int) *TransactionDeleteSuccessResponse {
	return &TransactionDeleteSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		TransactionId: id,
	}
}

type TransactionResponse struct {
	Id              int       `json:"id"`
	UserId          int       `json:"user_id"`
	AccountId       int       `json:"account_id"`
	Value           float64   `json:"value"`
	Type            string    `json:"type"`
	Category        string    `json:"category"`
	Currency        string    `json:"currency"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	TransactionDate time.Time `json:"transaction_date"`
}

type TransactionDetailSuccessResponse struct {
	SimpleResponse
	Transaction TransactionResponse `json:"transaction"`
}

func NewTransactionDetailSuccessResponse(transaction models.TransactionModel) *TransactionDetailSuccessResponse {
	return &TransactionDetailSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		Transaction: TransactionResponse{
			Id:              transaction.Id,
			UserId:          transaction.UserId,
			AccountId:       transaction.AccountId,
			Value:           transaction.Value,
			Type:            transaction.Type,
			Category:        transaction.Category,
			Currency:        transaction.Currency,
			Title:           transaction.Title,
			Description:     transaction.Description,
			CreatedAt:       transaction.CreatedAt,
			TransactionDate: transaction.TransactionDate,
		},
	}
}

type TransactionUpdateSuccessResponse SimpleResponse

func NewTransactionUpdateSuccessResponse() TransactionUpdateSuccessResponse {
	return TransactionUpdateSuccessResponse{
		Code:    http.StatusOK,
		Message: "Транзакция успешно обновлена",
	}
}

type UpdateUserProfileRequest struct {
	Username  *string `json:"username"`
	Email     *string `json:"email"`
	Password  *string `json:"password"`
	AvatarUrl *string `json:"avatar_url"`
}

type BadRequestErrorResponse SimpleResponse

type InternalServerErrorResponse SimpleResponse

func NewInternalServerErrorResponse() InternalServerErrorResponse {
	return InternalServerErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "Внутренняя ошибка сервера",
	}
}

type UpdateProfileSuccessResponse struct {
	SimpleResponse
	User User `json:"user"`
}

func NewUpdateProfileSuccessResponse(user User) UpdateProfileSuccessResponse {
	return UpdateProfileSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Профиль успешно обновлён",
		},
		User: user,
	}
}

type ProfileSuccessResponse struct {
	SimpleResponse
	User User `json:"user"`
}

func NewProfileSuccessResponse(user User) ProfileSuccessResponse {
	return ProfileSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Профиль получен",
		},
		User: user,
	}
}

type EnumListResponse struct {
	Code  int      `json:"code"`
	Items []string `json:"items"`
}

func NewEnumListResponse(items []string) EnumListResponse {
	return EnumListResponse{Code: 200, Items: items}
}

type AvatarUploadSuccessResponse struct {
	SimpleResponse
	AvatarUrl string `json:"avatar_url"`
}

func NewAvatarUploadSuccessResponse(avatarUrl string) AvatarUploadSuccessResponse {
	return AvatarUploadSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Аватар успешно загружен",
		},
		AvatarUrl: avatarUrl,
	}
}

type TransactionDraftData struct {
	RawText     string    `json:"raw_text"`
	Value       float64   `json:"value"`
	Type        string    `json:"type"`
	Category    string    `json:"category"`
	Currency    string    `json:"currency"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	RecordedAt  time.Time `json:"recorded_at"`
}

type VoiceTransactionDraftResponse struct {
	SimpleResponse
	Draft TransactionDraftData `json:"draft"`
}

func NewVoiceTransactionDraftResponse(draft TransactionDraftData) VoiceTransactionDraftResponse {
	return VoiceTransactionDraftResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Транзакция распознана",
		},
		Draft: draft,
	}
}

type VoiceErrorResponse SimpleResponse

func NewVoiceErrorResponse(code int, message string) VoiceErrorResponse {
	return VoiceErrorResponse{
		Code:    code,
		Message: message,
	}
}
