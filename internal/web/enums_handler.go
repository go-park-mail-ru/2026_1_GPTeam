package web

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type EnumsHandler struct {
	enumApp application.EnumsUseCase
	log     *zap.Logger
}

func NewEnumsHandler(enumApp application.EnumsUseCase) *EnumsHandler {
	return &EnumsHandler{
		enumApp: enumApp,
		log:     logger.GetLogger(),
	}
}

func (obj *EnumsHandler) CurrencyCodes(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("get currency codes request")
	codes := obj.enumApp.GetCurrencyCodes()
	obj.log.Info("currency codes get success")
	response := web_helpers.NewCurrencyCodesResponse(codes)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *EnumsHandler) TransactionTypes(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("get transaction types request")
	types := obj.enumApp.GetTransactionTypes()
	obj.log.Info("transaction types get success")
	response := web_helpers.NewEnumListResponse(types)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *EnumsHandler) CategoryTypes(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("get category types request")
	categories := obj.enumApp.GetCategoryTypes()
	obj.log.Info("category types get success")
	response := web_helpers.NewEnumListResponse(categories)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
