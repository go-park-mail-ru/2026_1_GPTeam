package web

import (
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	base2 "github.com/go-park-mail-ru/2026_1_GPTeam/web/web_helpers"
)

type UserHandler struct {
	UseCase application.UserUseCaseInterface
}

func NewUserHandler(useCase application.UserUseCaseInterface) *UserHandler {
	return &UserHandler{UseCase: useCase}
}

func (obj *UserHandler) Balance(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserInfo)
	_ = authUser
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base2.NewUnauthorizedErrorResponse()
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}
	balance := 0.0
	currency := "RUB"
	response := base2.NewBalanceResponse(balance, currency, 0, 0)
	base2.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base2.NewUnauthorizedErrorResponse()
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := base2.User{
		Username:        authUser.Username,
		Email:           authUser.Email,
		CreatedAt:       authUser.CreatedAt,
		LastLogin:       authUser.LastLogin,
		AvatarUrl:       authUser.AvatarUrl,
		Balance:         0,
		BalanceCurrency: "RUB",
	}
	response := base2.NewLoginSuccessResponse(userResponse)
	base2.WriteResponseJSON(w, response.Code, response)
}
