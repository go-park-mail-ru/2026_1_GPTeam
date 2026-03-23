package web

import (
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
		response := web_helpers.NewNotFoundErrorResponse("Счёт не найден")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	web_helpers.WriteResponseJSON(w, http.StatusOK, map[string]interface{}{
		"code":       200,
		"account_id": accountId,
	})
}
