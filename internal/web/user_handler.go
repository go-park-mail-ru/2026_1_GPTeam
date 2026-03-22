package web

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
)

type UserHandler struct {
	userApp application.UserUseCase
}

func NewUserHandler(useCase application.UserUseCase) *UserHandler {
	return &UserHandler{userApp: useCase}
}

func validateUpdateProfileRequest(req web_helpers.UpdateUserProfileRequest) error {
	if req.Username != nil {
		if err := validators.ValidateUsername(*req.Username); err != nil {
			return err
		}
	}
	if req.Email != nil {
		if err := validators.ValidateEmail(*req.Email); err != nil {
			return err
		}
	}
	if req.Password != nil {
		if err := validators.ValidatePassword(*req.Password); err != nil {
			return err
		}
	}
	return nil
}

func (obj *UserHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		obj.Profile(w, r)
	case http.MethodPut:
		obj.UpdateProfile(w, r)
	}
}

func (obj *UserHandler) Balance(w http.ResponseWriter, r *http.Request) {
	_, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewBalanceResponse(0.0, "RUB", 0, 0)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := web_helpers.User{
		Username:  *authUser.Username,
		Email:     *authUser.Email,
		CreatedAt: authUser.CreatedAt,
		AvatarUrl: *authUser.AvatarUrl,
	}
	response := web_helpers.NewProfileSuccessResponse(userResponse)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var req web_helpers.UpdateUserProfileRequest
	if err := web_helpers.ReadRequestJSON(r, &req); err != nil {
		response := web_helpers.NewBadRequestErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if err := validateUpdateProfileRequest(req); err != nil {
		response := web_helpers.NewBadRequestErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	userModel := models.UserModel{
		Id:        authUser.Id,
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		AvatarUrl: req.AvatarUrl,
	}
	updatedUser, err := obj.userApp.Update(r.Context(), userModel)
	if err != nil {
		response := web_helpers.NewInternalServerErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := web_helpers.User{
		Username:  *updatedUser.Username,
		Email:     *updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
		AvatarUrl: *updatedUser.AvatarUrl,
	}
	response := web_helpers.NewUpdateProfileSuccessResponse(userResponse)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
