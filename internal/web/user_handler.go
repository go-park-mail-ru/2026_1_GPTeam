package web

import (
	"net/http"
	"os"
	"path/filepath"
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

func (obj *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("changing avatar",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		obj.log.Warn("failed to read body",
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
		response := web_helpers.NewBadRequestErrorResponse("Слишком большой файл")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		obj.log.Warn("failed to change avatar (no file)",
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
		response := web_helpers.NewBadRequestErrorResponse("Нет файла аватара")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	defer file.Close()

	buff := make([]byte, 512)
	if _, err = file.Read(buff); err != nil {
		obj.log.Warn("failed to read buff",
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
		response := web_helpers.NewServerErrorResponse("Ошибка чтения")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	fileType := http.DetectContentType(buff)
	if fileType != "image/jpeg" && fileType != "image/png" {
		obj.log.Warn("file type not supported",
			zap.String("file type", fileType),
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
		response := web_helpers.NewBadRequestErrorResponse("Допустимы только форматы JPEG и PNG")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if _, err = file.Seek(0, 0); err != nil {
		obj.log.Warn("failed to seek file",
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
		response := web_helpers.NewServerErrorResponse("Внутренняя ошибка")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	userContext := r.Context().Value("user")
	authUser, ok := userContext.(*models.UserModel)
	if !ok {
		obj.log.Warn("user unauthorized",
			zap.String("request_id", r.Context().Value("request_id").(string)))
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	ext := filepath.Ext(header.Filename)
	avatarName, err := obj.userApp.UploadAvatar(r.Context(), authUser.Id, file, ext)
	if err != nil {
		obj.log.Warn("failed to upload avatar",
			zap.Int("user_id", authUser.Id),
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
		response := web_helpers.NewServerErrorResponse("Не удалось сохранить аватар")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	finalUrl := os.Getenv("SERVER_URL") + "/img/" + avatarName
	obj.log.Info("upload avatar success",
		zap.Int("user_id", authUser.Id),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewAvatarUploadSuccessResponse(finalUrl)
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
		response := web_helpers.NewBadRequestErrorResponse("невозможно прочитать тело запроса")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	validationErrors := validators.ValidateUpdateUser(req)
	if len(validationErrors) > 0 {
		obj.log.Warn("validation error while updating profile",
			zap.Int("user_id", authUser.Id),
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Any("validationErrors", validationErrors))
		response := web_helpers.NewBadRequestErrorResponse("ошибка валидации")
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
