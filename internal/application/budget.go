package application

import (
	"context"

	models2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
)

type BudgetUseCase interface {
	Create(ctx context.Context, budget models2.BudgetModel) (int, error)
	Delete(ctx context.Context, budgetId int, user models2.UserModel) error
	GetById(ctx context.Context, id int) (models2.BudgetModel, error)
	GetBudgetsOfUser(ctx context.Context, user models2.UserModel) ([]int, error)
	IsUserAuthorOfBudget(budget models2.BudgetModel, user models2.UserModel) bool
	GetAllowedCurrencies() []string
}

type Budget struct {
	repository repository.BudgetRepository
}

func NewBudget(repository repository.BudgetRepository) *Budget {
	return &Budget{repository: repository}
}

func (obj *Budget) Create(ctx context.Context, budget models2.BudgetModel) (int, error) {
	id, err := obj.repository.Create(ctx, budget)
	return id, err
}

func (obj *Budget) Delete(ctx context.Context, budgetId int, user models2.UserModel) error {
	budget, err := obj.GetById(ctx, budgetId)
	if err != nil {
		return err
	}
	if !obj.IsUserAuthorOfBudget(budget, user) {
		return UserNotAuthorOfBudgetError
	}
	err = obj.repository.Delete(ctx, budgetId)
	return err
}

func (obj *Budget) GetById(ctx context.Context, id int) (models2.BudgetModel, error) {
	budget, err := obj.repository.GetById(ctx, id)
	return budget, err
}

func (obj *Budget) GetBudgetsOfUser(ctx context.Context, user models2.UserModel) ([]int, error) {
	ids, err := obj.repository.GetIdsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Budget) IsUserAuthorOfBudget(budget models2.BudgetModel, user models2.UserModel) bool {
	return user.Id == budget.Author
}

func (obj *Budget) GetAllowedCurrencies() []string {
	return obj.repository.GetCurrencies()
}
