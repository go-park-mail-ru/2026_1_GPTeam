package web

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
)

type EnumsHandler struct {
	enumApp application.EnumsUseCase
}

func NewEnumsHandler(enumApp application.EnumsUseCase) *EnumsHandler {
	return &EnumsHandler{
		enumApp: enumApp,
	}
}

func (obj *EnumsHandler) CurrencyCodes(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get currency codes request")
	codes := obj.enumApp.GetCurrencyCodes()
	log.Info("currency codes get success")
	response := web_helpers.NewCurrencyCodesResponse(codes)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *EnumsHandler) TransactionTypes(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get transaction types request")
	types := obj.enumApp.GetTransactionTypes()
	log.Info("transaction types get success")
	response := web_helpers.NewEnumListResponse(types)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *EnumsHandler) CategoryTypes(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get category types request")
	categories := obj.enumApp.GetCategoryTypes()
	log.Info("category types get success")
	response := web_helpers.NewEnumListResponse(categories)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
