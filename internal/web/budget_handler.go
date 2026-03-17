package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	models2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	web_helpers2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
)

type BudgetHandler struct {
	budgetApp application.BudgetUseCase
}

func NewBudgetHandler(useCase application.BudgetUseCase) *BudgetHandler {
	return &BudgetHandler{budgetApp: useCase}
}

func (obj *BudgetHandler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(models2.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers2.NewUnauthorizedErrorResponse()
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	ids, err := obj.budgetApp.GetBudgetsOfUser(ctx, authUser)
	if err != nil {
		response := web_helpers2.NewValidationErrorResponse([]web_helpers2.FieldError{
			web_helpers2.NewFieldError("", err.Error()),
		})
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers2.NewBudgetsIDsResponse(ids)
	web_helpers2.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(models2.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers2.NewUnauthorizedErrorResponse()
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		response := web_helpers2.NewNotFoundErrorResponse("Не указан ID бюджета")
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetId, err := strconv.Atoi(idStr)
	if err != nil {
		response := web_helpers2.NewNotFoundErrorResponse("Неверный ID бюджета")
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	budget, err := obj.budgetApp.GetById(ctx, budgetId)
	if err != nil {
		fmt.Println(err)
		response := web_helpers2.NewValidationErrorResponse([]web_helpers2.FieldError{
			web_helpers2.NewFieldError("", err.Error()),
		})
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}
	isAuthor := obj.budgetApp.IsUserAuthorOfBudget(budget, authUser)
	if !isAuthor {
		response := web_helpers2.NewNotFoundErrorResponse("Бюджет не найден")
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	result := web_helpers2.BudgetRequest{
		Title:       budget.Title,
		Description: budget.Description,
		CreatedAt:   budget.CreatedAt,
		StartAt:     budget.StartAt,
		EndAt:       budget.EndAt,
		Actual:      int(budget.Actual),
		Target:      int(budget.Target),
		Currency:    budget.Currency,
	}
	response := web_helpers2.NewBudgetGetSuccessResponse(result)
	web_helpers2.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(models2.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers2.NewUnauthorizedErrorResponse()
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	var body web_helpers2.BudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response := web_helpers2.NewBudgetErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []web_helpers2.FieldError{
			web_helpers2.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	var fieldErrors []web_helpers2.FieldError
	if body.Title == "" {
		fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("title", "Поле обязательно для заполнения"))
	}
	if body.Description == "" {
		fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("description", "Поле обязательно для заполнения"))
	}
	if body.Target == 0 {
		fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("target", "Поле обязательно для заполнения"))
	}
	if body.Currency == "" {
		fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("currency", "Поле обязательно для заполнения"))
	}
	err := validators.ValidateCurrency(body.Currency, obj.budgetApp.GetAllowedCurrencies())
	if err != nil {
		fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("currency", err.Error()))
	}
	err = validators.ValidateTargetBudget(body.Target)
	if err != nil {
		fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("target", err.Error()))
	}
	err = validators.ValidateStartDate(body.StartAt)
	if err != nil {
		fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("start_at", err.Error()))
	}
	err = validators.ValidateEndDate(body.StartAt, body.EndAt)
	if err != nil {
		fieldErrors = append(fieldErrors, web_helpers2.NewFieldError("end_at", err.Error()))
	}
	if len(fieldErrors) > 0 {
		response := web_helpers2.NewBudgetErrorResponse(http.StatusBadRequest, "Ошибка валидации", fieldErrors)
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	budget := models2.BudgetModel{
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
	id, err := obj.budgetApp.Create(ctx, budget)
	if err != nil {
		fmt.Println(err)
		response := web_helpers2.NewValidationErrorResponse([]web_helpers2.FieldError{
			web_helpers2.NewFieldError("", err.Error()),
		})
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := web_helpers2.NewBudgetCreateSuccessResponse(id)
	web_helpers2.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(models2.UserModel)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := web_helpers2.NewUnauthorizedErrorResponse()
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		response := web_helpers2.NewNotFoundErrorResponse("Не указан ID бюджета")
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetId, err := strconv.Atoi(idStr)
	if err != nil {
		response := web_helpers2.NewNotFoundErrorResponse("Неверный ID бюджета")
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	err = obj.budgetApp.Delete(ctx, budgetId, authUser)
	if err != nil {
		fmt.Println(err)
		response := web_helpers2.NewNotFoundErrorResponse("Бюджет не найден")
		web_helpers2.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := web_helpers2.NewBudgetDeleteSuccessResponse()
	web_helpers2.WriteResponseJSON(w, response.Code, response)
}
