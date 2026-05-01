package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/context_helper"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type AccountHandler struct {
	accountApp application.AccountUseCase
}

func NewAccountHandler(accountApp application.AccountUseCase) *AccountHandler {
	return &AccountHandler{
		accountApp: accountApp,
	}
}

func accountToResponse(account models.AccountModel) web_helpers.AccountResponse {
	return web_helpers.AccountResponse{
		Id:        account.Id,
		Name:      account.Name,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}
}

func (obj *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get base account request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountId, err := obj.accountApp.GetAccountIdByUserId(r.Context(), authUser.Id)
	if err != nil {
		if errors.Is(err, application.ErrAccountNotFound) {
			response := web_helpers.NewNotFoundErrorResponse("Счёт не найден")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewInternalServerErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	web_helpers.WriteResponseJSON(w, http.StatusOK, map[string]interface{}{
		"code":       http.StatusOK,
		"account_id": accountId,
	})
}

func (obj *AccountHandler) Accounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		obj.ListAccounts(w, r)
	case http.MethodPost:
		obj.CreateAccount(w, r)
	default:
		response := web_helpers.NewMethodError()
		web_helpers.WriteResponseJSON(w, response.Code, response)
	}
}

func (obj *AccountHandler) Account(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		obj.ReadAccount(w, r)
	case http.MethodPut, http.MethodPatch:
		obj.UpdateAccount(w, r)
	case http.MethodDelete:
		obj.DeleteAccount(w, r)
	default:
		response := web_helpers.NewMethodError()
		web_helpers.WriteResponseJSON(w, response.Code, response)
	}
}

func getAccountIdFromPath(r *http.Request) (int, error) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/accounts/")
	return strconv.Atoi(idStr)
}

func normalizeAccountCurrency(value string) (string, bool) {
	currency := strings.ToUpper(strings.TrimSpace(value))

	switch currency {
	case "RUB", "USD", "EUR":
		return currency, true
	default:
		return currency, false
	}
}

func writeAccountValidationError(w http.ResponseWriter, field string, message string) {
	response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
		web_helpers.NewFieldError(field, message),
	})
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	accounts, err := obj.accountApp.GetByUserId(r.Context(), authUser.Id)
	if err != nil {
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	items := make([]web_helpers.AccountResponse, 0, len(accounts))
	for _, account := range accounts {
		items = append(items, accountToResponse(account))
	}
	response := web_helpers.NewAccountListResponse(items)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var body web_helpers.AccountRequest
	if err := web_helpers.ReadRequestJSON(r, &body); err != nil {
		response := web_helpers.NewBadRequestErrorResponse("Неверный формат запроса")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	body.Name = strings.TrimSpace(secure.SanitizeXss(body.Name))
	if body.Name == "" {
		writeAccountValidationError(w, "name", "Название счёта обязательно")
		return
	}

	currency, ok := normalizeAccountCurrency(body.Currency)
	if !ok {
		writeAccountValidationError(w, "currency", "Валюта должна быть одной из: RUB, USD, EUR")
		return
	}

	balance := 0.0
	if body.Balance != nil {
		balance = *body.Balance
	}

	account, err := obj.accountApp.CreateForUser(r.Context(), authUser.Id, models.AccountCreateModel{
		Name:     body.Name,
		Balance:  balance,
		Currency: currency,
	})
	if err != nil {
		obj.writeAccountError(w, r, err)
		return
	}

	response := web_helpers.NewAccountCreateSuccessResponse(accountToResponse(account))
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountHandler) ReadAccount(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	accountId, err := getAccountIdFromPath(r)
	if err != nil {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id счёта")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	account, err := obj.accountApp.GetById(r.Context(), authUser.Id, accountId)
	if err != nil {
		obj.writeAccountError(w, r, err)
		return
	}
	response := web_helpers.NewAccountGetSuccessResponse(accountToResponse(account))
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountId, err := getAccountIdFromPath(r)
	if err != nil {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id счёта")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var body web_helpers.AccountPatchRequest
	if err = web_helpers.ReadRequestJSON(r, &body); err != nil {
		response := web_helpers.NewBadRequestErrorResponse("Неверный формат запроса")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if body.Name != nil {
		name := strings.TrimSpace(secure.SanitizeXss(*body.Name))
		if name == "" {
			writeAccountValidationError(w, "name", "Название счёта не может быть пустым")
			return
		}
		body.Name = &name
	}

	if body.Currency != nil {
		currency, ok := normalizeAccountCurrency(*body.Currency)
		if !ok {
			writeAccountValidationError(w, "currency", "Валюта должна быть одной из: RUB, USD, EUR")
			return
		}
		body.Currency = &currency
	}

	account, err := obj.accountApp.Update(r.Context(), authUser.Id, accountId, models.AccountUpdateModel{
		Name:     body.Name,
		Balance:  body.Balance,
		Currency: body.Currency,
	})
	if err != nil {
		obj.writeAccountError(w, r, err)
		return
	}

	response := web_helpers.NewAccountUpdateSuccessResponse(accountToResponse(account))
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	accountId, err := getAccountIdFromPath(r)
	if err != nil {
		response := web_helpers.NewBadRequestErrorResponse("Некорректный id счёта")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	err = obj.accountApp.Delete(r.Context(), authUser.Id, accountId)
	if err != nil {
		obj.writeAccountError(w, r, err)
		return
	}
	response := web_helpers.NewAccountDeleteSuccessResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AccountHandler) writeAccountError(w http.ResponseWriter, r *http.Request, err error) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Warn("account request failed", zap.Error(err))

	if errors.Is(err, application.ErrAccountNotFound) {
		response := web_helpers.NewNotFoundErrorResponse("Счёт не найден")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if errors.Is(err, application.AllFieldsEmptyError) {
		response := web_helpers.NewBadRequestErrorResponse("Передайте хотя бы одно поле для обновления")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
