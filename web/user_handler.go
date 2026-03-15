package web

import (
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
)

type UserHandler struct {
	UseCase application.UserUseCaseInterface
}

func NewUserHandler(useCase application.UserUseCaseInterface) *UserHandler {
	return &UserHandler{UseCase: useCase}
}

func (obj *UserHandler) Balance(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	balance := authUser.Balance
	currency := authUser.BalanceCurrency
	response := base.NewBalanceResponse(balance, currency, 0, 0)
	base.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := base.User{
		Username:        authUser.Username,
		Email:           authUser.Email,
		CreatedAt:       authUser.CreatedAt,
		LastLogin:       authUser.LastLogin,
		AvatarUrl:       authUser.AvatarUrl,
		Balance:         authUser.Balance,
		BalanceCurrency: authUser.BalanceCurrency,
	}
	response := base.NewLoginSuccessResponse(userResponse)
	base.WriteResponseJSON(w, response.Code, response)
}
