package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/metrics"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
	"go.uber.org/zap"
)

type SupportHandler struct {
	supportApp application.SupportUseCase
	userApp    application.UserUseCase
}

func NewSupportHandler(supportApp application.SupportUseCase, userApp application.UserUseCase) *SupportHandler {
	return &SupportHandler{
		supportApp: supportApp,
		userApp:    userApp,
	}
}

func (obj *SupportHandler) Create(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("create support request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	var body web_helpers.SupportRequest
	if err := web_helpers.ReadRequestJSON(r, &body); err != nil {
		log.Warn("failed to read body",
			zap.Error(err))
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	body.Category = secure.SanitizeXss(body.Category)
	body.Message = secure.SanitizeXss(body.Message)
	validationErrors := validators.ValidateSupport(body, authUser)
	if len(validationErrors) > 0 {
		response := web_helpers.NewValidationErrorResponse(validationErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	id, err := obj.supportApp.Create(r.Context(), body, authUser.Id)
	if err != nil {
		if errors.Is(err, repository.DuplicatedDataError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Сообщение уже существует"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.ConstraintError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Введены некорректные данные"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewInternalServerErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("created support", zap.Int("support_id", id), zap.Int("user_id", authUser.Id))
	response := web_helpers.NewOkResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
	appMetrics := metrics.GetMetrics()
	appMetrics.SupportCreationsTotal.WithLabelValues().Inc()
}

func (obj *SupportHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get all supports request")

	supports, err := obj.supportApp.GetAll(r.Context())
	if err != nil {
		response := web_helpers.NewInternalServerErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var fullSupports []web_helpers.SupportResponse
	for _, support := range supports {
		user, err := obj.userApp.GetById(r.Context(), support.UserId)
		if err != nil {
			response := web_helpers.NewInternalServerErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		userResponse := web_helpers.User{
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			AvatarUrl: user.AvatarUrl,
		}
		fullSupports = append(fullSupports, web_helpers.NewSupportResponse(userResponse, support))
	}

	log.Info("get all supports success", zap.Int("len", len(supports)))
	response := web_helpers.NewSupportsResponse(fullSupports)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *SupportHandler) Detail(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("detail support request")
	idStr := r.PathValue("id")
	supportId, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn("invalid support id",
			zap.Error(err))
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("id", "Некорректный ID"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	support, err := obj.supportApp.GetById(r.Context(), supportId)
	if err != nil {
		if errors.Is(err, repository.NothingInTableError) {
			response := web_helpers.NewNotFoundErrorResponse("Не найдено")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewInternalServerErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	user, err := obj.userApp.GetById(r.Context(), support.UserId)
	if err != nil {
		if errors.Is(err, repository.NothingInTableError) {
			response := web_helpers.NewNotFoundErrorResponse("Не найдено")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewInternalServerErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	userResponse := web_helpers.User{
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		AvatarUrl: user.AvatarUrl,
	}
	response := web_helpers.NewSupportResponse(userResponse, support)
	web_helpers.WriteResponseJSON(w, http.StatusOK, response)
}

func (obj *SupportHandler) GetAllByUser(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get all supports by user request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	supports, err := obj.supportApp.GetAllByUser(r.Context(), authUser.Id)
	if err != nil {
		response := web_helpers.NewInternalServerErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	var fullSupports []web_helpers.SupportResponse
	for _, support := range supports {
		fullSupports = append(fullSupports, web_helpers.SupportResponse{
			Id:        support.Id,
			Category:  secure.SanitizeXss(support.Category),
			Message:   secure.SanitizeXss(support.Message),
			Status:    support.Status,
			CreatedAt: support.CreatedAt,
		})
	}
	response := web_helpers.NewSupportsResponse(fullSupports)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *SupportHandler) Update(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("update support request")
	idStr := r.PathValue("id")
	supportId, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn("invalid support id",
			zap.Error(err))
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{
			web_helpers.NewFieldError("id", "Некорректный ID"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	var body web_helpers.UpdateSupportStatusRequest
	if err = web_helpers.ReadRequestJSON(r, &body); err != nil {
		log.Warn("failed to read body",
			zap.Error(err))
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	err = obj.supportApp.Update(r.Context(), supportId, body.Status)
	if err != nil {
		if errors.Is(err, repository.NothingInTableError) {
			response := web_helpers.NewNotFoundErrorResponse("Заявка не найдена")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewInternalServerErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers.NewOkResponse()
	web_helpers.WriteResponseJSON(w, http.StatusOK, response)
}
