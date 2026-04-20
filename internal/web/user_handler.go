package web

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/context_helper"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
	"go.uber.org/zap"
)

type UserHandler struct {
	userApp    application.UserUseCase
	accountApp application.AccountUseCase
}

func NewUserHandler(userApp application.UserUseCase, accountApp application.AccountUseCase) *UserHandler {
	return &UserHandler{
		userApp:    userApp,
		accountApp: accountApp,
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
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("get balance request")

	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	accounts, incomes, expenses, err := obj.accountApp.GetAllAccountsByUserIdWithBalance(r.Context(), authUser.Id)
	if err != nil {
		log.Error("failed to calculate user balance", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var balances []web_helpers.CurrencyBalance
	for i := range len(accounts) {
		balances = append(balances, web_helpers.CurrencyBalance{
			Currency: accounts[i].Currency,
			Balance:  accounts[i].Balance,
			Income:   incomes[i],
			Expenses: expenses[i],
		})
	}

	response := web_helpers.NewBalanceResponse(balances)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("changing avatar")

	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		log.Warn("failed to read body",
			zap.Error(err))
		response := web_helpers.NewBadRequestErrorResponse("Слишком большой файл")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		log.Warn("failed to change avatar (no file)",
			zap.Error(err))
		response := web_helpers.NewBadRequestErrorResponse("Нет файла аватара")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	defer file.Close()

	buff := make([]byte, 512)
	if _, err = file.Read(buff); err != nil {
		log.Warn("failed to read buff",
			zap.Error(err))
		response := web_helpers.NewServerErrorResponse("Ошибка чтения")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	fileType := http.DetectContentType(buff)
	if fileType != "image/jpeg" && fileType != "image/png" {
		log.Warn("file type not supported",
			zap.String("file type", fileType),
			zap.Error(err))
		response := web_helpers.NewBadRequestErrorResponse("Допустимы только форматы JPEG и PNG")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	if _, err = file.Seek(0, 0); err != nil {
		log.Warn("failed to seek file",
			zap.Error(err))
		response := web_helpers.NewServerErrorResponse("Внутренняя ошибка")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	userContext := r.Context().Value("user")
	authUser, ok := userContext.(*models.UserModel)
	if !ok {
		log.Warn("user unauthorized")
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

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		log.Warn("SERVER_URL not set, using default value")
		serverURL = "http://localhost:8081"
	}

	finalUrl, err := url.JoinPath(serverURL, "img", avatarName)
	if err != nil {
		log.Error("failed to build final url", zap.Error(err))
		response := web_helpers.NewServerErrorResponse("Внутренняя ошибка при формировании ссылки")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	log.Info("upload avatar success",
		zap.Int("user_id", authUser.Id))
	response := web_helpers.NewAvatarUploadSuccessResponse(finalUrl)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("get profile request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := web_helpers.User{
		Username:  secure.SanitizeXss(authUser.Username),
		Email:     authUser.Email,
		CreatedAt: authUser.CreatedAt,
		AvatarUrl: authUser.AvatarUrl,
	}
	log.Info("get profile success",
		zap.Int("user_id", authUser.Id))
	response := web_helpers.NewProfileSuccessResponse(userResponse)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWIthRequestId(r.Context())
	log.Info("update profile request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var req web_helpers.UpdateUserProfileRequest
	if err := web_helpers.ReadRequestJSON(r, &req); err != nil {
		log.Warn("failed to read body",
			zap.Int("user_id", authUser.Id),
			zap.Error(err))
		response := web_helpers.NewBadRequestErrorResponse("невозможно прочитать тело запроса")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	if req.Username != nil {
		sanitized := secure.SanitizeXss(*req.Username)
		req.Username = &sanitized
	}
	validationErrors := validators.ValidateUpdateUser(req)
	if len(validationErrors) > 0 {
		log.Warn("validation error while updating profile",
			zap.Int("user_id", authUser.Id),
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
	log.Info("update profile success",
		zap.Int("user_id", authUser.Id))
	response := web_helpers.NewUpdateProfileSuccessResponse(userResponse)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
