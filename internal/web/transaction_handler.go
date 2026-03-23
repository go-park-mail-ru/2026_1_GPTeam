package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
)

type TransactionHandler struct {
	transactionApp application.TransactionUseCase
	enumsApp       application.EnumsUseCase
	accountApp     application.AccountUseCase
}

func NewTransactionHandler(transactionApp application.TransactionUseCase, enumsApp application.EnumsUseCase, accountUseCase application.AccountUseCase) *TransactionHandler {
	return &TransactionHandler{
		transactionApp: transactionApp,
		enumsApp:       enumsApp,
		accountApp:     accountUseCase,
	}
}

func (obj *TransactionHandler) Transactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		obj.getTransactions(w, r)
	case http.MethodPost:
		obj.create(w, r)
	}
}

func (obj *TransactionHandler) Transaction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		obj.detail(w, r)
	case http.MethodDelete:
		obj.delete(w, r)
	}
}

func (obj *TransactionHandler) create(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	var body web_helpers.TransactionRequest
	if err := web_helpers.ReadRequestJSON(r, &body); err != nil {
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	validationErrors := validators.ValidateTransaction(body, obj.enumsApp.GetTransactionTypes(), obj.enumsApp.GetCategoryTypes())
	if len(validationErrors) > 0 {
		response := web_helpers.NewValidationErrorResponse(validationErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	transaction := models.TransactionModel{
		Id:              0,
		UserId:          authUser.Id,
		AccountId:       body.AccountId,
		Value:           body.Value,
		Type:            body.Type,
		Category:        body.Category,
		Title:           body.Title,
		Description:     body.Description,
		CreatedAt:       time.Now(),
		TransactionDate: body.TransactionDate,
	}
	if !obj.accountApp.IsUserAuthorOfAccount(r.Context(), authUser.Id, transaction.AccountId) {
		response := web_helpers.NewForbiddenErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	id, err := obj.transactionApp.Create(r.Context(), transaction)
	if err != nil {
		if errors.Is(err, repository.DuplicatedDataError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Такая транзакция уже существует"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.ConstraintError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Введены некорректные данные"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.TransactionAccountForeignKeyError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
				web_helpers.NewFieldError("account", "Счёта не существует"),
			})
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewServerErrorResponse("req_id")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewTransactionCreateSuccessResponse(id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *TransactionHandler) getTransactions(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	ids, err := obj.transactionApp.GetTransactionIdsOfUser(r.Context(), authUser)
	if err != nil {
		if errors.Is(err, repository.NothingInTableError) {
			response := web_helpers.NewNotFoundErrorResponse("Транзакции не найдены")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewServerErrorResponse("req_id")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewTransactionsIdsResponse(ids)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *TransactionHandler) delete(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	transactionId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("id", "Некорректный ID транзакции"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	id, err := obj.transactionApp.Delete(r.Context(), transactionId, authUser.Id)
	if err != nil {
		if errors.Is(err, repository.NothingInTableError) {
			response := web_helpers.NewNotFoundErrorResponse("Транзакция не найдена")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, application.ForbiddenError) {
			response := web_helpers.NewForbiddenErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewServerErrorResponse("req_id")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewTransactionDeleteSuccessResponse(id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *TransactionHandler) detail(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	transactionId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("id", "Некорректный ID транзакции"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	transaction, err := obj.transactionApp.Detail(r.Context(), transactionId, authUser.Id)
	if err != nil {
		if errors.Is(err, repository.NothingInTableError) {
			response := web_helpers.NewNotFoundErrorResponse("Транзакция не найдена")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, application.ForbiddenError) {
			response := web_helpers.NewForbiddenErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewServerErrorResponse("req_id")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewTransactionDetailSuccessResponse(transaction)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
