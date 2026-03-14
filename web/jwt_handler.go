package web

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/auth"
	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
)

type JWTHandlers struct {
	useCase application.JWTUseCaseInterface
}

func NewJWTHandler(useCase application.JWTUseCaseInterface) *JWTHandlers {
	return &JWTHandlers{useCase: useCase} // ToDo: get auth packet; auth packet creates in main
}

func (obj *JWTHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearOldToken(w, r)
	response := base.NewLogoutSuccessResponse()
	base.WriteResponseJSON(w, response.Code, response)
}

func (obj *JWTHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	isAuth, userID := auth.RefreshToken(w, r)
	authUser, ok := storage.IsAuthUserInDatabase(isAuth, userID)
	if !ok {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := base.NewLoginSuccessResponse(authUser)
	base.WriteResponseJSON(w, response.Code, response)
}
