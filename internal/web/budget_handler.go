package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
	"go.uber.org/zap"
)

type BudgetHandler struct {
	budgetApp application.BudgetUseCase
	enumsApp  application.EnumsUseCase
	log       *zap.Logger
}

func NewBudgetHandler(useCase application.BudgetUseCase, enumsApp application.EnumsUseCase) *BudgetHandler {
	return &BudgetHandler{
		budgetApp: useCase,
		enumsApp:  enumsApp,
		log:       logger.GetLogger(),
	}
}

func (obj *BudgetHandler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("get budgets request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		obj.log.Warn("user unauthorized",
			zap.String("request_id", r.Context().Value("request_id").(string)))
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
	obj.log.Info("get budgets success",
		zap.Int("user_id", authUser.Id),
		zap.Ints("budget ids", ids),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewBudgetsIdsResponse(ids)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("get budget request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		obj.log.Warn("user unauthorized",
			zap.String("request_id", r.Context().Value("request_id").(string)))
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	idStr := r.PathValue("id")
	if idStr == "" {
		obj.log.Warn("budget id required",
			zap.Int("user_id", authUser.Id),
			zap.String("request_id", r.Context().Value("request_id").(string)))
		response := web_helpers.NewNotFoundErrorResponse("Не указан ID бюджета")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetId, err := strconv.Atoi(idStr)
	if err != nil {
		obj.log.Warn("budget id required",
			zap.Int("user_id", authUser.Id),
			zap.String("request_id", r.Context().Value("request_id").(string)),
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
		Title:       budget.Title,
		Description: budget.Description,
		CreatedAt:   budget.CreatedAt,
		StartAt:     budget.StartAt,
		EndAt:       budget.EndAt,
		Actual:      int(budget.Actual),
		Target:      int(budget.Target),
		Currency:    budget.Currency,
	}
	obj.log.Info("get budget success",
		zap.Int("user_id", authUser.Id),
		zap.Int("budget_id", budgetId),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewBudgetGetSuccessResponse(result)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("create budget request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		obj.log.Warn("user unauthorized",
			zap.String("request_id", r.Context().Value("request_id").(string)))
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	var body web_helpers.BudgetRequest
	if err := web_helpers.ReadRequestJSON(r, &body); err != nil {
		obj.log.Warn("failed to read body",
			zap.Int("user_id", authUser.Id),
			zap.String("request_id", r.Context().Value("request_id").(string)),
			zap.Error(err))
		response := web_helpers.NewBudgetErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []web_helpers.FieldError{
			web_helpers.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	validationErrors := validators.ValidateBudget(body, obj.enumsApp.GetCurrencyCodes())
	if len(validationErrors) > 0 {
		obj.log.Warn("validation error when budget creating",
			zap.Int("user_id", authUser.Id),
			zap.Any("validationErrors", validationErrors),
			zap.String("request_id", r.Context().Value("request_id").(string)))
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
	obj.log.Info("budget created success",
		zap.Int("user_id", authUser.Id),
		zap.Int("budget_id", id),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewBudgetCreateSuccessResponse(id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	obj.log.Info("delete budget request",
		zap.String("request_id", r.Context().Value("request_id").(string)))
	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		obj.log.Warn("user unauthorized",
			zap.String("request_id", r.Context().Value("request_id").(string)))
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	idStr := r.PathValue("id")
	if idStr == "" {
		obj.log.Warn("id required",
			zap.Int("user_id", authUser.Id),
			zap.String("request_id", r.Context().Value("request_id").(string)))
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
	obj.log.Info("budget deleted success",
		zap.Int("user_id", authUser.Id),
		zap.Int("budget_id", budgetId),
		zap.String("request_id", r.Context().Value("request_id").(string)))
	response := web_helpers.NewBudgetDeleteSuccessResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
