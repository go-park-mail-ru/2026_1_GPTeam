package web

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

type EnumsHandler struct {
	enumApp application.EnumsUseCase
}

func NewEnumsHandler(enumApp application.EnumsUseCase) *EnumsHandler {
	return &EnumsHandler{enumApp: enumApp}
}

func (obj *EnumsHandler) CurrencyCodes(w http.ResponseWriter, r *http.Request) {
	codes := obj.enumApp.GetCurrencyCodes()
	response := web_helpers.NewCurrencyCodesResponse(codes)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
