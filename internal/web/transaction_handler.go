package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

type TransactionHandler struct {
	transactionApp application.TransactionUseCase
}

func NewTransactionHandler(transactionApp application.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{transactionApp: transactionApp}
}

func (obj *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var body web_helpers.TransactionRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var validationErrors []web_helpers.FieldError
	// ToDo: validators
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
	id, err := obj.transactionApp.Create(r.Context(), transaction)
	if err != nil {
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		if errors.Is(err, repository.DuplicatedDataError) {
			response.Message = "Такой бюджет уже существует"
		} else if errors.Is(err, repository.ConstraintError) {
			response.Message = "Введены некорректные данные"
		} else if errors.Is(err, repository.ForeignKeyError) {
			response.Message = "Счёта не существует"
		} else {
			response.Message = err.Error()
		}
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewTransactionCreateSuccessResponse(id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	ids, err := obj.transactionApp.GetTransactionsOfUser(r.Context(), authUser)
	if err != nil {
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		response.Message = err.Error()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewTransactionsIdsResponse(ids)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
