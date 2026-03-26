package web

import (
	"fmt"
	"net/http"
	"path/filepath"

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

func (obj *UserHandler) Balance(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	_ = authUser
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	balance := 0.0
	currency := "RUB"
	response := web_helpers.NewBalanceResponse(balance, currency, 0, 0)
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
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := web_helpers.User{
		Username:        authUser.Username,
		Email:           authUser.Email,
		CreatedAt:       authUser.CreatedAt,
		LastLogin:       authUser.LastLogin,
		AvatarUrl:       authUser.AvatarUrl,
		Balance:         0,
		BalanceCurrency: "RUB",
	}
	response := web_helpers.NewLoginSuccessResponse(userResponse)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
