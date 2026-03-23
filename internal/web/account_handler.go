package web

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

type AccountHandler struct {
	accountApp application.AccountUseCase
}

func NewAccountHandler(accountApp application.AccountUseCase) *AccountHandler {
	return &AccountHandler{accountApp: accountApp}
}

func (obj *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
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
