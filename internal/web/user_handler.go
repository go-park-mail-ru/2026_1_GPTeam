package web

import (
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	web_helpers2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

type UserHandler struct {
	userApp application.UserUseCaseInterface
}

func NewUserHandler(useCase application.UserUseCaseInterface) *UserHandler {
	return &UserHandler{userApp: useCase}
}

func (obj *UserHandler) Balance(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	_ = authUser
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers2.NewUnauthorizedErrorResponse()
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}
	balance := 0.0
	currency := "RUB"
	response := web_helpers2.NewBalanceResponse(balance, currency, 0, 0)
	web_helpers2.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers2.NewUnauthorizedErrorResponse()
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := web_helpers2.User{
		Username:        authUser.Username,
		Email:           authUser.Email,
		CreatedAt:       authUser.CreatedAt,
		LastLogin:       authUser.LastLogin,
		AvatarUrl:       authUser.AvatarUrl,
		Balance:         0,
		BalanceCurrency: "RUB",
	}
	response := web_helpers2.NewLoginSuccessResponse(userResponse)
	web_helpers2.WriteResponseJSON(w, response.Code, response)
}
