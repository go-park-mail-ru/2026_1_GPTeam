package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
	"go.uber.org/zap"
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

// Transactions /transactions — GET (список) и POST (создать)
func (obj *TransactionHandler) Transactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		obj.getTransactions(w, r)
	case http.MethodPost:
		obj.create(w, r)
	}
}

// Transaction /transactions/{id} — GET (детали), DELETE (удалить), PUT (обновить)
func (obj *TransactionHandler) Transaction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		obj.detail(w, r)
	case http.MethodDelete:
		obj.delete(w, r)
	case http.MethodPut:
		obj.update(w, r)
	}
}

func (obj *TransactionHandler) create(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("create transaction request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	var body web_helpers.TransactionRequest
	if err := web_helpers.ReadRequestJSON(r, &body); err != nil {
		log.Warn("failed to read body",
			zap.Error(err))
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	body.Title = secure.SanitizeXss(body.Title)
	body.Description = secure.SanitizeXss(body.Description)
	validationErrors := validators.ValidateTransaction(
		body,
		obj.enumsApp.GetTransactionTypes(),
		obj.enumsApp.GetCategoryTypes(),
		obj.enumsApp.GetCurrencyCodes(),
	)
	if len(validationErrors) > 0 {
		log.Warn("validation error while creating transaction",
			zap.Any("validationErrors", validationErrors))
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
		Currency:        body.Currency,
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
		response := web_helpers.NewServerErrorResponse(r.Context().Value("request_id").(string))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("created transaction",
		zap.Int("user_id", authUser.Id),
		zap.Int("transaction_id", id))
	response := web_helpers.NewTransactionCreateSuccessResponse(id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *TransactionHandler) getTransactions(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("get transactions request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
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
		response := web_helpers.NewServerErrorResponse(r.Context().Value("request_id").(string))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("get transaction ids success",
		zap.Int("user_id", authUser.Id),
		zap.Ints("transaction_ids", ids))
	response := web_helpers.NewTransactionsIdsResponse(ids)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *TransactionHandler) update(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("update transaction request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	idStr := r.PathValue("id")
	transactionId, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn("invalid transaction id",
			zap.Error(err))
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("id", "Некорректный ID транзакции"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	var body web_helpers.TransactionRequest
	err = web_helpers.ReadRequestJSON(r, &body)
	if err != nil {
		log.Warn("invalid request body",
			zap.Error(err))
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	body.Title = secure.SanitizeXss(body.Title)
	body.Description = secure.SanitizeXss(body.Description)
	validationErrors := validators.ValidateTransaction(
		body,
		obj.enumsApp.GetTransactionTypes(),
		obj.enumsApp.GetCategoryTypes(),
		obj.enumsApp.GetCurrencyCodes(),
	)
	if len(validationErrors) > 0 {
		log.Warn("validation error while updating transaction",
			zap.Int("user_id", authUser.Id),
			zap.Any("validationErrors", validationErrors))
		response := web_helpers.NewValidationErrorResponse(validationErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	if !obj.accountApp.IsUserAuthorOfAccount(r.Context(), authUser.Id, body.AccountId) {
		response := web_helpers.NewForbiddenErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	transaction := models.TransactionModel{
		Id:              transactionId,
		UserId:          authUser.Id,
		AccountId:       body.AccountId,
		Value:           body.Value,
		Type:            body.Type,
		Category:        body.Category,
		Currency:        body.Currency,
		Title:           body.Title,
		Description:     body.Description,
		TransactionDate: body.TransactionDate,
	}
	err = obj.transactionApp.Update(r.Context(), transaction)
	if err != nil {
		if errors.Is(err, repository.IncorrectRowsAffectedError) || errors.Is(err, repository.NothingInTableError) {
			response := web_helpers.NewNotFoundErrorResponse("Транзакция не найдена")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, application.ForbiddenError) {
			response := web_helpers.NewForbiddenErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.ConstraintError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = err.Error()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewServerErrorResponse("req_id")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("update transaction success",
		zap.Int("user_id", authUser.Id),
		zap.Int("transaction_id", transactionId))
	response := web_helpers.NewTransactionUpdateSuccessResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *TransactionHandler) delete(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("delete transaction request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	transactionId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Warn("failed to read body",
			zap.Int("user_id", authUser.Id),
			zap.Error(err))
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
		response := web_helpers.NewServerErrorResponse(r.Context().Value("request_id").(string))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("transaction deleted success",
		zap.Int("user_id", authUser.Id),
		zap.Int("transaction_id", transactionId))
	response := web_helpers.NewTransactionDeleteSuccessResponse(id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *TransactionHandler) detail(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("get transaction request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	transactionId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Warn("failed to read body",
			zap.Int("user_id", authUser.Id),
			zap.Error(err))
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
		response := web_helpers.NewServerErrorResponse(r.Context().Value("request_id").(string))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	transaction.Title = secure.SanitizeXss(transaction.Title)
	transaction.Description = secure.SanitizeXss(transaction.Description)
	log.Info("get transaction success",
		zap.Int("user_id", authUser.Id),
		zap.Int("transaction_id", transactionId))
	response := web_helpers.NewTransactionDetailSuccessResponse(transaction)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
