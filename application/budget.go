package application

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
)

type Budget struct {
	repository repository.BudgetRepositoryInterface
}

func NewBudget(repository repository.BudgetRepositoryInterface) *Budget {
	return &Budget{repository: repository}
}

func (obj *Budget) Create(ctx context.Context, budget models.BudgetModel) (int, error) {
	id, err := obj.repository.Create(ctx, budget)
	return id, err
}

func (obj *Budget) Delete(ctx context.Context, budgetId int, user models.UserModel) error {
	budget, err := obj.GetById(ctx, budgetId)
	if err != nil {
		return err
	}
	if !obj.IsUserAuthorOfBudget(budget, user) {
		return UserNotAuthorOfBudgetError(user.Id, budget.Id)
	}
	err = obj.repository.Delete(ctx, budgetId)
	return err
}

func (obj *Budget) GetById(ctx context.Context, id int) (models.BudgetModel, error) {
	budget, err := obj.repository.GetById(ctx, id)
	return budget, err
}

func (obj *Budget) GetBudgetsOfUser(ctx context.Context, user models.UserModel) ([]int, error) {
	ids, err := obj.repository.GetIdsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Budget) IsUserAuthorOfBudget(budget models.BudgetModel, user models.UserModel) bool {
	return user.Id == budget.Author
}

func (obj *Budget) GetAllowedCurrencies() []string {
	return obj.repository.GetCurrencies()
}
