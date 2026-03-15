package application

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_GPTeam/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
)

type BudgetUseCaseInterface interface {
	Create(ctx context.Context, budget storage.BudgetInfo) (int, error)
	Delete(ctx context.Context, budgetID int, user storage.UserInfo) error
	GetById(ctx context.Context, id int) (storage.BudgetInfo, error)
	GetBudgetsOfUser(ctx context.Context, user storage.UserInfo) ([]int, error)
	IsUserAuthor(ctx context.Context, budget storage.BudgetInfo, user storage.UserInfo) bool
}

type Budget struct {
	repo repository.BudgetRepositoryInterface
}

func NewBudget(repo repository.BudgetRepositoryInterface) *Budget {
	return &Budget{repo: repo}
}

func (obj *Budget) Create(ctx context.Context, budget storage.BudgetInfo) (int, error) {
	id, err := obj.repo.Create(ctx, budget)
	return id, err
}

func (obj *Budget) Delete(ctx context.Context, budgetID int, user storage.UserInfo) error {
	budget, err := obj.GetById(ctx, budgetID)
	if err != nil {
		return err
	}
	if !obj.IsUserAuthor(ctx, budget, user) {
		return fmt.Errorf("not author")
	}
	err = obj.repo.Delete(ctx, budgetID)
	return err
}

func (obj *Budget) GetById(ctx context.Context, id int) (storage.BudgetInfo, error) {
	budget, err := obj.repo.GetById(ctx, id)
	return budget, err
}

func (obj *Budget) GetBudgetsOfUser(ctx context.Context, user storage.UserInfo) ([]int, error) {
	ids, err := obj.repo.GetIDsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Budget) IsUserAuthor(ctx context.Context, budget storage.BudgetInfo, user storage.UserInfo) bool {
	return user.Id == budget.Author
}
