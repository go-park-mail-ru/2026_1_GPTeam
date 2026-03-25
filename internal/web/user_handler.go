package web

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
	"go.uber.org/zap"
)

type UserHandler struct {
	userApp application.UserUseCase
	log     *zap.Logger
}

func NewUserHandler(useCase application.UserUseCase) *UserHandler {
	return &UserHandler{
		userApp: useCase,
		log:     logger.GetLogger(),
	}
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
	case http.MethodPatch:
		obj.UpdateProfile(w, r)
	}
}

func (obj *UserHandler) Balance(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("get balance request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		obj.log.Warn("user unauthorized",
			zap.String("request_id", r.Context().Value("request_id").(string)))
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	obj.log.Info("get balance success",
		zap.Int("user_id", authUser.Id),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewBalanceResponse(0.0, "RUB", 0, 0)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("get profile request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		obj.log.Warn("user unauthorized",
			zap.String("request_id", r.Context().Value("request_id").(string)))
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
	obj.log.Info("get profile success",
		zap.Int("user_id", authUser.Id),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewProfileSuccessResponse(userResponse)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("update profile request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		obj.log.Warn("user unauthorized",
			zap.String("request_id", r.Context().Value("request_id").(string)))
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var req web_helpers.UpdateUserProfileRequest
	if err := web_helpers.ReadRequestJSON(r, &req); err != nil {
		obj.log.Warn("failed to read body",
			zap.Int("user_id", authUser.Id),
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
		response := web_helpers.NewBadRequestErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if err := validateUpdateProfileRequest(req); err != nil {
		obj.log.Warn("validation error while updating profile",
			zap.Int("user_id", authUser.Id),
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
		response := web_helpers.NewBadRequestErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	updateProfile := models.UpdateUserProfile{
		Id:        authUser.Id,
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		AvatarUrl: req.AvatarUrl,
		UpdatedAt: time.Now(),
	}
	updatedUser, err := obj.userApp.Update(r.Context(), updateProfile)
	if err != nil {
		obj.log.Warn("failed to update profile",
			zap.Int("user_id", authUser.Id),
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
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
	obj.log.Info("update profile success",
		zap.Int("user_id", authUser.Id),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewUpdateProfileSuccessResponse(userResponse)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
