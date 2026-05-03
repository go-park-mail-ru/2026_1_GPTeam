package application

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/currency_converter"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=budget.go -destination=mocks/mock_budget.go -package=mocks
type BudgetUseCase interface {
	Create(ctx context.Context, budget models.BudgetModel, categories []string) (int, error)
	Delete(ctx context.Context, budgetId int, user models.UserModel) error
	GetById(ctx context.Context, id int, user models.UserModel) (models.BudgetModel, []string, error)
	GetBudgetsOfUser(ctx context.Context, user models.UserModel) ([]int, error)
	IsUserAuthorOfBudget(budget models.BudgetModel, user models.UserModel) bool
	Update(ctx context.Context, budget models.BudgetModel) error
}

type Budget struct {
	repository     repository.BudgetRepository
	transactionApp TransactionUseCase
	accountApp     AccountUseCase
}

func NewBudget(repository repository.BudgetRepository, transactionApp TransactionUseCase, accountApp AccountUseCase) *Budget {
	return &Budget{
		repository:     repository,
		transactionApp: transactionApp,
		accountApp:     accountApp,
	}
}

func (obj *Budget) Create(ctx context.Context, budget models.BudgetModel, categories []string) (int, error) {
	loc := time.FixedZone("UTC+3", 3*60*60)
	budget.CreatedAt = time.Date(budget.CreatedAt.Year(), budget.CreatedAt.Month(), budget.CreatedAt.Day(), 0, 0, 0, 0, loc)
	actual, err := calculateActual(ctx, budget, categories, obj.transactionApp, obj.accountApp)
	if err != nil {
		return 0, err
	}
	budget.Actual = actual
	id, err := obj.repository.Create(ctx, budget)
	if err != nil {
		return 0, err
	}
	for _, category := range categories {
		err = obj.repository.LinkBudgetAndCategory(ctx, id, category)
		if err != nil {
			return 0, err
		}
	}
	return id, nil
}

func (obj *Budget) Delete(ctx context.Context, budgetId int, user models.UserModel) error {
	_, _, err := obj.GetById(ctx, budgetId, user)
	if err != nil {
		return err
	}
	err = obj.repository.Delete(ctx, budgetId)
	return err
}

func (obj *Budget) GetById(ctx context.Context, id int, user models.UserModel) (models.BudgetModel, []string, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	budget, err := obj.repository.GetById(ctx, id)
	if err != nil {
		return models.BudgetModel{}, []string{}, err
	}
	if !obj.IsUserAuthorOfBudget(budget, user) {
		log.Warn("user is not author of budget",
			zap.Int("user_id", user.Id),
			zap.Int("budget_id", id))
		return models.BudgetModel{}, []string{}, UserNotAuthorOfBudgetError
	}
	category, err := obj.repository.GetCategoryOfBudget(ctx, budget.Id)
	return budget, category, err
}

func (obj *Budget) GetBudgetsOfUser(ctx context.Context, user models.UserModel) ([]int, error) {
	ids, err := obj.repository.GetIdsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Budget) IsUserAuthorOfBudget(budget models.BudgetModel, user models.UserModel) bool {
	return user.Id == budget.Author
}

func (obj *Budget) Update(ctx context.Context, budget models.BudgetModel) error {
	err := obj.repository.Update(ctx, budget)
	return err
}

func calculateActual(ctx context.Context, budget models.BudgetModel, categories []string, transactionApp TransactionUseCase, accountApp AccountUseCase) (float64, error) {
	var actual float64
	for _, category := range categories {
		filter := repository.TransactionFilters{
			StartDate: &budget.StartAt,
			Category:  &category,
		}
		if !budget.EndAt.IsZero() {
			filter.EndDate = &budget.EndAt
		}
		transactions, err := transactionApp.Search(ctx, budget.Author, filter)
		if err != nil {
			if !errors.Is(err, repository.NothingInTableError) {
				return 0, err
			}
		}
		for _, transaction := range transactions {
			currency, err := accountApp.GetCurrencyByAccountId(ctx, transaction.AccountId)
			if err != nil {
				return 0, err
			}
			if transaction.Type == "INCOME" {
				actual -= currency_converter.ConvertToRub(transaction.Value, currency)
			} else {
				actual += currency_converter.ConvertToRub(transaction.Value, currency)
			}
		}
	}
	return max(0, actual), nil
}
