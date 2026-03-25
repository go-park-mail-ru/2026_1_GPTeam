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
	log        *zap.Logger
}

func NewAccountHandler(accountApp application.AccountUseCase) *AccountHandler {
	return &AccountHandler{
		accountApp: accountApp,
		log:        logger.GetLogger(),
	}
}

func (obj *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("get account request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		obj.log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accountId, err := obj.accountApp.GetAccountIdByUserId(r.Context(), authUser.Id)
	if err != nil {
		obj.log.Warn("failed to get account", zap.Error(err))
		if errors.Is(err, application.ErrAccountNotFound) {
			response := web_helpers.NewNotFoundErrorResponse("Счёт не найден")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}

		response := web_helpers.NewInternalServerErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	obj.log.Info("get account", zap.Int("account_id", accountId), zap.Int("user_id", authUser.Id))
	web_helpers.WriteResponseJSON(w, http.StatusOK, map[string]interface{}{
		"code":       http.StatusOK,
		"account_id": accountId,
	})
}
