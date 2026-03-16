package application

import (
	"context"

	models2 "github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
)

type BudgetUseCaseInterface interface {
	Create(ctx context.Context, budget models2.BudgetInfo) (int, error)
	Delete(ctx context.Context, budgetID int, user models2.UserInfo) error
	GetById(ctx context.Context, id int) (models2.BudgetInfo, error)
	GetBudgetsOfUser(ctx context.Context, user models2.UserInfo) ([]int, error)
	IsUserAuthorOfBudget(budget models2.BudgetInfo, user models2.UserInfo) bool
	GetAllowedCurrencies() []string
}

type Budget struct {
	repo repository.BudgetRepositoryInterface
}

func NewBudget(repo repository.BudgetRepositoryInterface) *Budget {
	return &Budget{repo: repo}
}

func (obj *Budget) Create(ctx context.Context, budget models2.BudgetInfo) (int, error) {
	id, err := obj.repo.Create(ctx, budget)
	return id, err
}

func (obj *Budget) Delete(ctx context.Context, budgetID int, user models2.UserInfo) error {
	budget, err := obj.GetById(ctx, budgetID)
	if err != nil {
		return err
	}
	if !obj.IsUserAuthorOfBudget(budget, user) {
		return UserNotAuthorOfBudgetError(user.Id, budget.Id)
	}
	err = obj.repo.Delete(ctx, budgetID)
	return err
}

func (obj *Budget) GetById(ctx context.Context, id int) (models2.BudgetInfo, error) {
	budget, err := obj.repo.GetById(ctx, id)
	return budget, err
}

func (obj *Budget) GetBudgetsOfUser(ctx context.Context, user models2.UserInfo) ([]int, error) {
	ids, err := obj.repo.GetIDsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Budget) IsUserAuthorOfBudget(budget models2.BudgetInfo, user models2.UserInfo) bool {
	return user.Id == budget.Author
}

func (obj *Budget) GetAllowedCurrencies() []string {
	return obj.repo.GetCurrencies()
}
