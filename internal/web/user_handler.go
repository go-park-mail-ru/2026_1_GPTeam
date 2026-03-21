package web

import (
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

type UserHandler struct {
	userApp application.UserUseCase
}

func NewUserHandler(useCase application.UserUseCase) *UserHandler {
	return &UserHandler{userApp: useCase}
}

func getAuthUser(r *http.Request) (models.UserModel, bool) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
	}
	return authUser, ok
}

func (obj *UserHandler) Balance(w http.ResponseWriter, r *http.Request) {
	_, ok := getAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewBalanceResponse(0.0, "RUB", 0, 0)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	authUser, ok := getAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := web_helpers.User{
		Username:  authUser.Username,
		Email:     authUser.Email,
		CreatedAt: authUser.CreatedAt,
		AvatarUrl: authUser.AvatarUrl,
	}
	response := web_helpers.NewProfileSuccessResponse(userResponse)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	authUser, ok := getAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	var updateProfileRequest web_helpers.UpdateUserProfileRequest
	if err := web_helpers.ReadRequestJSON(r, &updateProfileRequest); err != nil {
		response := web_helpers.NewBadRequestErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	if updateProfileRequest.Username == "" &&
		updateProfileRequest.Email == "" &&
		updateProfileRequest.Password == "" &&
		updateProfileRequest.AvatarUrl == "" {
		response := web_helpers.NewBadRequestErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	userModel := models.UserModel{
		Id:        authUser.Id,
		Username:  updateProfileRequest.Username,
		Email:     updateProfileRequest.Email,
		Password:  updateProfileRequest.Password,
		AvatarUrl: updateProfileRequest.AvatarUrl,
	}
	updatedUser, err := obj.userApp.Update(r.Context(), userModel)
	if err != nil {
		response := web_helpers.NewInternalServerErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := web_helpers.User{
		Username:  updatedUser.Username,
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
		AvatarUrl: updatedUser.AvatarUrl,
	}
	response := web_helpers.NewUpdateProfileSuccessResponse(userResponse)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
