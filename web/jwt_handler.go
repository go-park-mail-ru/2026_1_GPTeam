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
	auth    auth.AuthInterface
}

func NewJWTHandler(useCase application.JWTUseCaseInterface, auth auth.AuthInterface) *JWTHandlers {
	return &JWTHandlers{
		useCase: useCase,
		auth:    auth,
	} // ToDo: get auth packet; auth packet creates in main
}

func (obj *JWTHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	obj.auth.ClearOld(w, r)
	response := base.NewLogoutSuccessResponse()
	base.WriteResponseJSON(w, response.Code, response)
}

func (obj *JWTHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	isAuth, userID := obj.auth.Refresh(w, r)
	authUser, ok := storage.IsAuthUserInDatabase(isAuth, userID)
	if !ok {
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := base.NewLoginSuccessResponse(authUser)
	base.WriteResponseJSON(w, response.Code, response)
}
