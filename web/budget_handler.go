package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
	"github.com/go-park-mail-ru/2026_1_GPTeam/validators"
)

type BudgetHandler struct {
	useCase application.BudgetUseCaseInterface
}

func NewBudgetHandler(useCase application.BudgetUseCaseInterface) *BudgetHandler {
	return &BudgetHandler{useCase: useCase}
}

func (obj *BudgetHandler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	ids, err := obj.useCase.GetBudgetsOfUser(ctx, authUser)
	if err != nil {
		fmt.Println(err)
		response := base.NewValidationErrorResponse([]base.FieldError{
			base.NewFieldError("", err.Error()),
		})
		base.WriteResponseJSON(w, response.Code, response)
	}
	response := base.NewBudgetsIDsResponse(ids)
	base.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		response := base.NewNotFoundErrorResponse("Не указан ID бюджета")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetID, err := strconv.Atoi(idStr)
	if err != nil {
		response := base.NewNotFoundErrorResponse("Неверный ID бюджета")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	budget, err := obj.useCase.GetById(ctx, budgetID)
	if err != nil {
		fmt.Println(err)
		response := base.NewValidationErrorResponse([]base.FieldError{
			base.NewFieldError("", err.Error()),
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	isAuthor := obj.useCase.IsUserAuthor(ctx, budget, authUser)
	if !isAuthor {
		response := base.NewNotFoundErrorResponse("Бюджет не найден")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	result := base.BudgetRequest{
		Title:       budget.Title,
		Description: budget.Description,
		CreatedAt:   budget.CreatedAt,
		StartAt:     budget.StartAt,
		EndAt:       budget.EndAt,
		Actual:      budget.Actual,
		Target:      budget.Target,
		Currency:    budget.Currency,
	}
	response := base.NewBudgetGetSuccessResponse(result)
	base.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	var body base.BudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response := base.NewBudgetErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []base.FieldError{
			base.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	var fieldErrors []base.FieldError
	if body.Title == "" {
		fieldErrors = append(fieldErrors, base.NewFieldError("title", "Поле обязательно для заполнения"))
	}
	if body.Description == "" {
		fieldErrors = append(fieldErrors, base.NewFieldError("description", "Поле обязательно для заполнения"))
	}
	if body.Target == 0 {
		fieldErrors = append(fieldErrors, base.NewFieldError("target", "Поле обязательно для заполнения"))
	}
	if body.Currency == "" {
		fieldErrors = append(fieldErrors, base.NewFieldError("currency", "Поле обязательно для заполнения"))
	}
	err := validators.ValidateCurrency(body.Currency)
	if err != nil {
		fieldErrors = append(fieldErrors, base.NewFieldError("currency", err.Error()))
	}
	err = validators.ValidateTargetBudget(body.Target)
	if err != nil {
		fieldErrors = append(fieldErrors, base.NewFieldError("target", err.Error()))
	}
	err = validators.ValidateStartDate(body.StartAt)
	if err != nil {
		fieldErrors = append(fieldErrors, base.NewFieldError("start_at", err.Error()))
	}
	err = validators.ValidateEndDate(body.StartAt, body.EndAt)
	if err != nil {
		fieldErrors = append(fieldErrors, base.NewFieldError("end_at", err.Error()))
	}
	if len(fieldErrors) > 0 {
		response := base.NewBudgetErrorResponse(http.StatusBadRequest, "Ошибка валидации", fieldErrors)
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	budget := storage.BudgetInfo{
		Title:       body.Title,
		Description: body.Description,
		CreatedAt:   time.Now(),
		StartAt:     body.StartAt,
		EndAt:       body.EndAt,
		Actual:      0,
		Target:      body.Target,
		Currency:    body.Currency,
		Author:      authUser.Id,
	}
	id, err := obj.useCase.Create(ctx, budget)
	if err != nil {
		fmt.Println(err)
		response := base.NewValidationErrorResponse([]base.FieldError{
			base.NewFieldError("", err.Error()),
		})
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := base.NewBudgetCreateSuccessResponse(id)
	base.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(storage.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base.NewUnauthorizedErrorResponse()
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		response := base.NewNotFoundErrorResponse("Не указан ID бюджета")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetID, err := strconv.Atoi(idStr)
	if err != nil {
		response := base.NewNotFoundErrorResponse("Неверный ID бюджета")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	err = obj.useCase.Delete(ctx, budgetID, authUser)
	if err != nil {
		fmt.Println(err)
		response := base.NewNotFoundErrorResponse("Бюджет не найден")
		base.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := base.NewBudgetDeleteSuccessResponse()
	base.WriteResponseJSON(w, response.Code, response)
}
