package application

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_GPTeam/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
)

type BudgetUseCaseInterface interface {
	Create(ctx context.Context, budget models.BudgetInfo) (int, error)
	Delete(ctx context.Context, budgetID int, user models.UserInfo) error
	GetById(ctx context.Context, id int) (models.BudgetInfo, error)
	GetBudgetsOfUser(ctx context.Context, user models.UserInfo) ([]int, error)
	IsUserAuthor(budget models.BudgetInfo, user models.UserInfo) bool
}

var UserNotAuthorOfBudgetError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("user %v not author of budget %v\n", args[0], args[1])
}

type Budget struct {
	repo repository.BudgetRepositoryInterface
}

func NewBudget(repo repository.BudgetRepositoryInterface) *Budget {
	return &Budget{repo: repo}
}

func (obj *Budget) Create(ctx context.Context, budget models.BudgetInfo) (int, error) {
	id, err := obj.repo.Create(ctx, budget)
	return id, err
}

func (obj *Budget) Delete(ctx context.Context, budgetID int, user models.UserInfo) error {
	budget, err := obj.GetById(ctx, budgetID)
	if err != nil {
		return err
	}
	if !obj.IsUserAuthor(budget, user) {
		return UserNotAuthorOfBudgetError(user.Id, budget.Id)
	}
	err = obj.repo.Delete(ctx, budgetID)
	return err
}

func (obj *Budget) GetById(ctx context.Context, id int) (models.BudgetInfo, error) {
	budget, err := obj.repo.GetById(ctx, id)
	return budget, err
}

func (obj *Budget) GetBudgetsOfUser(ctx context.Context, user models.UserInfo) ([]int, error) {
	ids, err := obj.repo.GetIDsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Budget) IsUserAuthor(budget models.BudgetInfo, user models.UserInfo) bool {
	return user.Id == budget.Author
}
