package web

import (
	"net/http"
	"path/filepath"
	"time"

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
	case http.MethodPatch:
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

func (obj *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		response := web_helpers.NewBadRequestErrorResponse("Слишком большой файл")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		response := web_helpers.NewBadRequestErrorResponse("Нет файла аватара")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	defer file.Close()

	buff := make([]byte, 512)
	if _, err = file.Read(buff); err != nil {
		response := web_helpers.NewServerErrorResponse("Ошибка чтения")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	fileType := http.DetectContentType(buff)
	if fileType != "image/jpeg" && fileType != "image/png" {
		response := web_helpers.NewBadRequestErrorResponse("Допустимы только форматы JPEG и PNG")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if _, err = file.Seek(0, 0); err != nil {
		response := web_helpers.NewServerErrorResponse("Внутренняя ошибка")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	userContext := r.Context().Value("user")
	authUser, ok := userContext.(models.UserModel)
	if !ok {
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	ext := filepath.Ext(header.Filename)
	avatarName, err := obj.userApp.UploadAvatar(r.Context(), authUser.Id, file, ext)
	if err != nil {
		response := web_helpers.NewServerErrorResponse("Не удалось сохранить аватар")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	finalUrl := "/img/" + avatarName
	response := web_helpers.NewAvatarUploadSuccessResponse(finalUrl)
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
		Username:  authUser.Username,
		Email:     authUser.Email,
		CreatedAt: authUser.CreatedAt,
		AvatarUrl: authUser.AvatarUrl,
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
		response := web_helpers.NewBadRequestErrorResponse("невозможно прочитать тело запроса")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if err := validateUpdateProfileRequest(req); err != nil {
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
