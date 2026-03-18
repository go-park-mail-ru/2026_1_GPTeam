package web

import (
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

type TransactionHandler struct {
	transactionApp application.TransactionUseCase
}

func NewTransactionHandler(transactionApp application.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{transactionApp: transactionApp}
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
