package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/secure"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
	"go.uber.org/zap"
)

type BudgetHandler struct {
	budgetApp application.BudgetUseCase
	enumsApp  application.EnumsUseCase
}

func NewBudgetHandler(useCase application.BudgetUseCase, enumsApp application.EnumsUseCase) *BudgetHandler {
	return &BudgetHandler{
		budgetApp: useCase,
		enumsApp:  enumsApp,
	}
}

func (obj *BudgetHandler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get budgets request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	ids, err := obj.budgetApp.GetBudgetsOfUser(r.Context(), authUser)
	if err != nil {
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		response.Message = err.Error()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("get budgets success",
		zap.Int("user_id", authUser.Id),
		zap.Ints("budget ids", ids))
	response := web_helpers.NewBudgetsIdsResponse(ids)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get budget request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	idStr := r.PathValue("id")
	if idStr == "" {
		log.Warn("budget id required",
			zap.Int("user_id", authUser.Id))
		response := web_helpers.NewNotFoundErrorResponse("Не указан ID бюджета")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetId, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn("budget id required",
			zap.Int("user_id", authUser.Id),
			zap.Error(err))
		response := web_helpers.NewNotFoundErrorResponse("Неверный ID бюджета")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	budget, err := obj.budgetApp.GetById(r.Context(), budgetId, authUser)
	if err != nil {
		if errors.Is(err, application.UserNotAuthorOfBudgetError) {
			response := web_helpers.NewForbiddenErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.NothingInTableError) {
			response := web_helpers.NewNotFoundErrorResponse("Бюджет не найден")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		response.Message = err.Error()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	result := web_helpers.BudgetRequest{
		Title:       secure.SanitizeXss(budget.Title),
		Description: secure.SanitizeXss(budget.Description),
		CreatedAt:   budget.CreatedAt,
		StartAt:     budget.StartAt,
		EndAt:       budget.EndAt,
		Actual:      int(budget.Actual),
		Target:      int(budget.Target),
		Currency:    budget.Currency,
	}
	log.Info("get budget success",
		zap.Int("user_id", authUser.Id),
		zap.Int("budget_id", budgetId))
	response := web_helpers.NewBudgetGetSuccessResponse(result)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("create budget request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	var body web_helpers.BudgetRequest
	if err := web_helpers.ReadRequestJSON(r, &body); err != nil {
		log.Warn("failed to read body",
			zap.Int("user_id", authUser.Id),
			zap.Error(err))
		response := web_helpers.NewBudgetErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []web_helpers.FieldError{
			web_helpers.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	body.Title = secure.SanitizeXss(body.Title)
	body.Description = secure.SanitizeXss(body.Description)
	validationErrors := validators.ValidateBudget(body, obj.enumsApp.GetCurrencyCodes())
	if len(validationErrors) > 0 {
		log.Warn("validation error when budget creating",
			zap.Int("user_id", authUser.Id),
			zap.Any("validationErrors", validationErrors))
		response := web_helpers.NewBudgetErrorResponse(http.StatusBadRequest, "Ошибка валидации", validationErrors)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	budget := models.BudgetModel{
		Title:       body.Title,
		Description: body.Description,
		CreatedAt:   time.Now(),
		StartAt:     body.StartAt,
		EndAt:       body.EndAt,
		Actual:      0,
		Target:      float64(body.Target),
		Currency:    body.Currency,
		Author:      authUser.Id,
	}
	id, err := obj.budgetApp.Create(r.Context(), budget)
	if err != nil {
		if errors.Is(err, repository.DuplicatedDataError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Такой бюджет уже существует"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.ConstraintError) {
			response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
			response.Message = "Введены некорректные данные"
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewValidationErrorResponse([]web_helpers.FieldError{})
		response.Message = err.Error()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("budget created success",
		zap.Int("user_id", authUser.Id),
		zap.Int("budget_id", id))
	response := web_helpers.NewBudgetCreateSuccessResponse(id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("delete budget request")
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	idStr := r.PathValue("id")
	if idStr == "" {
		log.Warn("id required",
			zap.Int("user_id", authUser.Id))
		response := web_helpers.NewNotFoundErrorResponse("Не указан ID бюджета")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetId, err := strconv.Atoi(idStr)
	if err != nil {
		response := web_helpers.NewNotFoundErrorResponse("Неверный ID бюджета")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	if err = obj.budgetApp.Delete(r.Context(), budgetId, authUser); err != nil {
		if errors.Is(err, application.UserNotAuthorOfBudgetError) {
			response := web_helpers.NewForbiddenErrorResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		if errors.Is(err, repository.NothingInTableError) {
			response := web_helpers.NewNotFoundErrorResponse("Бюджет не найден")
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		response := web_helpers.NewNotFoundErrorResponse(err.Error())
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	log.Info("budget deleted success",
		zap.Int("user_id", authUser.Id),
		zap.Int("budget_id", budgetId))
	response := web_helpers.NewBudgetDeleteSuccessResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
