package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application"
	models2 "github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/application/validators"
	base2 "github.com/go-park-mail-ru/2026_1_GPTeam/web/base"
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
	authUser, ok := user.(models2.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base2.NewUnauthorizedErrorResponse()
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	ids, err := obj.useCase.GetBudgetsOfUser(ctx, authUser)
	if err != nil {
		fmt.Println(err)
		response := base2.NewValidationErrorResponse([]base2.FieldError{
			base2.NewFieldError("", err.Error()),
		})
		base2.WriteResponseJSON(w, response.Code, response)
	}
	response := base2.NewBudgetsIDsResponse(ids)
	base2.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(models2.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base2.NewUnauthorizedErrorResponse()
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		response := base2.NewNotFoundErrorResponse("Не указан ID бюджета")
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetID, err := strconv.Atoi(idStr)
	if err != nil {
		response := base2.NewNotFoundErrorResponse("Неверный ID бюджета")
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	budget, err := obj.useCase.GetById(ctx, budgetID)
	if err != nil {
		fmt.Println(err)
		response := base2.NewValidationErrorResponse([]base2.FieldError{
			base2.NewFieldError("", err.Error()),
		})
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}
	isAuthor := obj.useCase.IsUserAuthorOfBudget(budget, authUser)
	if !isAuthor {
		response := base2.NewNotFoundErrorResponse("Бюджет не найден")
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	result := base2.BudgetRequest{
		Title:       budget.Title,
		Description: budget.Description,
		CreatedAt:   budget.CreatedAt,
		StartAt:     budget.StartAt,
		EndAt:       budget.EndAt,
		Actual:      int(budget.Actual),
		Target:      int(budget.Target),
		Currency:    budget.Currency,
	}
	response := base2.NewBudgetGetSuccessResponse(result)
	base2.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(models2.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base2.NewUnauthorizedErrorResponse()
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	var body base2.BudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response := base2.NewBudgetErrorResponse(http.StatusBadRequest, "Неверный формат запроса", []base2.FieldError{
			base2.NewFieldError("", "Не удалось прочитать тело запроса"),
		})
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	var fieldErrors []base2.FieldError
	if body.Title == "" {
		fieldErrors = append(fieldErrors, base2.NewFieldError("title", "Поле обязательно для заполнения"))
	}
	if body.Description == "" {
		fieldErrors = append(fieldErrors, base2.NewFieldError("description", "Поле обязательно для заполнения"))
	}
	if body.Target == 0 {
		fieldErrors = append(fieldErrors, base2.NewFieldError("target", "Поле обязательно для заполнения"))
	}
	if body.Currency == "" {
		fieldErrors = append(fieldErrors, base2.NewFieldError("currency", "Поле обязательно для заполнения"))
	}
	err := validators.ValidateCurrency(body.Currency)
	if err != nil {
		fieldErrors = append(fieldErrors, base2.NewFieldError("currency", err.Error()))
	}
	err = validators.ValidateTargetBudget(body.Target)
	if err != nil {
		fieldErrors = append(fieldErrors, base2.NewFieldError("target", err.Error()))
	}
	err = validators.ValidateStartDate(body.StartAt)
	if err != nil {
		fieldErrors = append(fieldErrors, base2.NewFieldError("start_at", err.Error()))
	}
	err = validators.ValidateEndDate(body.StartAt, body.EndAt)
	if err != nil {
		fieldErrors = append(fieldErrors, base2.NewFieldError("end_at", err.Error()))
	}
	if len(fieldErrors) > 0 {
		response := base2.NewBudgetErrorResponse(http.StatusBadRequest, "Ошибка валидации", fieldErrors)
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	budget := models2.BudgetInfo{
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
	id, err := obj.useCase.Create(ctx, budget)
	if err != nil {
		fmt.Println(err)
		response := base2.NewValidationErrorResponse([]base2.FieldError{
			base2.NewFieldError("", err.Error()),
		})
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := base2.NewBudgetCreateSuccessResponse(id)
	base2.WriteResponseJSON(w, response.Code, response)
}

func (obj *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	user := r.Context().Value("user")
	authUser, ok := user.(models2.UserInfo)
	if !ok {
		fmt.Printf("user is a %T\n", user)
		response := base2.NewUnauthorizedErrorResponse()
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		response := base2.NewNotFoundErrorResponse("Не указан ID бюджета")
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}
	budgetID, err := strconv.Atoi(idStr)
	if err != nil {
		response := base2.NewNotFoundErrorResponse("Неверный ID бюджета")
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	err = obj.useCase.Delete(ctx, budgetID, authUser)
	if err != nil {
		fmt.Println(err)
		response := base2.NewNotFoundErrorResponse("Бюджет не найден")
		base2.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := base2.NewBudgetDeleteSuccessResponse()
	base2.WriteResponseJSON(w, response.Code, response)
}
