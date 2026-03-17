package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
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
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
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
	response := web_helpers.NewBudgetsIDsResponse(ids)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
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
	response := web_helpers.NewBudgetGetSuccessResponse(result)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var body web_helpers.BudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response := web_helpers.NewBudgetErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []web_helpers.FieldError{
			web_helpers.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	var validationErrors []web_helpers.FieldError
	if body.Title == "" {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("title", "Поле обязательно для заполнения"))
	}
	if body.Description == "" {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("description", "Поле обязательно для заполнения"))
	}
	if body.Target == 0 {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("target", "Поле обязательно для заполнения"))
	}
	if body.Currency == "" {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("currency", "Поле обязательно для заполнения"))
	}
	err := validators.ValidateCurrency(body.Currency, obj.enumsApp.GetCurrencyCodes())
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("currency", err.Error()))
	}
	err = validators.ValidateTargetBudget(body.Target)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("target", err.Error()))
	}
	err = validators.ValidateStartDate(body.StartAt)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("start_at", err.Error()))
	}
	err = validators.ValidateEndDate(body.StartAt, body.EndAt)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("end_at", err.Error()))
	}
	if len(validationErrors) > 0 {
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
	response := web_helpers.NewBudgetCreateSuccessResponse(id)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user")
	authUser, ok := user.(models.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
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

	err = obj.budgetApp.Delete(r.Context(), budgetId, authUser)
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
		response := web_helpers.NewNotFoundErrorResponse(err.Error())
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := web_helpers.NewBudgetDeleteSuccessResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
