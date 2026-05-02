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
	Id        int       `json:"id,omitempty"`
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
	Category    []string  `json:"category"`
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

// TransactionWithCurrency — модель транзакции с валютой, полученной из связанного счёта.
// Используется только на уровне хэндлера, модель TransactionModel не меняется.
type TransactionWithCurrency struct {
	models.TransactionModel
	Currency string
}

type TransactionResponse struct {
	Id              int       `json:"id"`
	UserId          int       `json:"user_id"`
	AccountId       int       `json:"account_id"`
	Value           float64   `json:"value"`
	Type            string    `json:"type"`
	Category        string    `json:"category"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
	TransactionDate time.Time `json:"transaction_date"`
}

func newTransactionResponse(t TransactionWithCurrency) TransactionResponse {
	return TransactionResponse{
		Id:              t.Id,
		UserId:          t.UserId,
		AccountId:       t.AccountId,
		Value:           t.Value,
		Type:            t.Type,
		Category:        t.Category,
		Title:           t.Title,
		Description:     t.Description,
		Currency:        t.Currency,
		CreatedAt:       t.CreatedAt,
		TransactionDate: t.TransactionDate,
	}
}

type TransactionDetailSuccessResponse struct {
	SimpleResponse
	Transaction TransactionResponse `json:"transaction"`
}

func NewTransactionDetailSuccessResponse(t TransactionWithCurrency) *TransactionDetailSuccessResponse {
	return &TransactionDetailSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		Transaction: newTransactionResponse(t),
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
	Title       string    `json:"title"`
	Description string    `json:"description"`
	RecordedAt  time.Time `json:"recorded_at"`
	Date        time.Time `json:"date"`
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

type TooManyRequestsResponse SimpleResponse

func NewTooManyRequestsResponse() TooManyRequestsResponse {
	return TooManyRequestsResponse{
		Code:    http.StatusTooManyRequests,
		Message: "Слишком много запросов",
	}
}

type AccountRequest struct {
	Name     string   `json:"name"`
	Balance  *float64 `json:"balance,omitempty"`
	Currency string   `json:"currency"`
}

type AccountPatchRequest struct {
	Name     *string  `json:"name,omitempty"`
	Balance  *float64 `json:"balance,omitempty"`
	Currency *string  `json:"currency,omitempty"`
}

type AccountResponse struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccountListResponse struct {
	SimpleResponse
	Accounts []AccountResponse `json:"accounts"`
}

func NewAccountListResponse(accounts []AccountResponse) AccountListResponse {
	return AccountListResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		Accounts: accounts,
	}
}

type AccountGetSuccessResponse struct {
	SimpleResponse
	Account AccountResponse `json:"account"`
}

func NewAccountGetSuccessResponse(account AccountResponse) AccountGetSuccessResponse {
	return AccountGetSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Счёт успешно получен",
		},
		Account: account,
	}
}

type AccountCreateSuccessResponse struct {
	SimpleResponse
	Account AccountResponse `json:"account"`
}

func NewAccountCreateSuccessResponse(account AccountResponse) AccountCreateSuccessResponse {
	return AccountCreateSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Счёт успешно создан",
		},
		Account: account,
	}
}

type AccountUpdateSuccessResponse struct {
	SimpleResponse
	Account AccountResponse `json:"account"`
}

func NewAccountUpdateSuccessResponse(account AccountResponse) AccountUpdateSuccessResponse {
	return AccountUpdateSuccessResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Счёт успешно обновлён",
		},
		Account: account,
	}
}

type AccountDeleteSuccessResponse SimpleResponse

func NewAccountDeleteSuccessResponse() AccountDeleteSuccessResponse {
	return AccountDeleteSuccessResponse{
		Code:    http.StatusOK,
		Message: "Счёт успешно удалён",
	}
}

type SupportRequest struct {
	Category string `json:"category"`
	Message  string `json:"message"`
}

type ShortSupport struct {
	Category string `json:"category"`
	Message  string `json:"message"`
}

type SupportsResponse struct {
	SimpleResponse
	Supports []SupportResponse `json:"supports"`
}

func NewSupportsResponse(supports []SupportResponse) SupportsResponse {
	return SupportsResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "OK",
		},
		Supports: supports,
	}
}

type SupportResponse struct {
	Id        int       `json:"id"`
	Category  string    `json:"category"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user"`
}

func NewSupportResponse(user User, support models.SupportModel) SupportResponse {
	return SupportResponse{
		Id:        support.Id,
		Category:  support.Category,
		Message:   support.Message,
		Status:    support.Status,
		CreatedAt: support.CreatedAt,
		User:      user,
	}
}

type IsStaffResponse struct {
	SimpleResponse
	IsStaff bool `json:"is_staff"`
}

func NewIsStaffResponse(isStaff bool) IsStaffResponse {
	return IsStaffResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "OK",
		},
		IsStaff: isStaff,
	}
}

type UpdateSupportStatusRequest struct {
	Status string `json:"status"`
}

type TransactionsSearchResponse struct {
	SimpleResponse
	Transactions []TransactionResponse `json:"transactions"`
}

func NewTransactionsSearchResponse(transactions []TransactionWithCurrency) TransactionsSearchResponse {
	items := make([]TransactionResponse, 0, len(transactions))
	for _, t := range transactions {
		items = append(items, newTransactionResponse(t))
	}
	return TransactionsSearchResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		Transactions: items,
	}
}

type MethodNotAllowedErrorResponse SimpleResponse

func NewMethodNotAllowedErrorResponse() MethodNotAllowedErrorResponse {
	return MethodNotAllowedErrorResponse{
		Code:    http.StatusMethodNotAllowed,
		Message: "Метод не поддерживается",
	}
}
