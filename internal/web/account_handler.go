package web

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
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

func (obj *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get account request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
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
	log.Info("get account",
		zap.Int("account_id", accountId),
		zap.Int("user_id", authUser.Id))
	web_helpers.WriteResponseJSON(w, http.StatusOK, map[string]interface{}{
		"code":       http.StatusOK,
		"account_id": accountId,
	})
}

func (obj *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get short accounts request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	accounts, err := obj.accountApp.GetAllAccountsByUserId(r.Context(), authUser.Id)
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
	log.Info("get short accounts",
		zap.Any("accounts", accounts),
		zap.Int("user_id", authUser.Id))
	var shortAccounts []web_helpers.ShortAccount
	for _, account := range accounts {
		shortAccounts = append(shortAccounts, web_helpers.NewShortAccount(account.Id, account.Name, account.Balance))
	}
	response := web_helpers.NewShortAccountsResponse(shortAccounts)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
