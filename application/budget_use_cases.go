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

type BudgetUseCase struct {
	repo repository.BudgetRepositoryInterface
}

func NewBudgetUseCase(repo repository.BudgetRepositoryInterface) *BudgetUseCase {
	return &BudgetUseCase{repo: repo}
}

func (obj *BudgetUseCase) Create(ctx context.Context, budget storage.BudgetInfo) (int, error) {
	id, err := obj.repo.Create(ctx, budget)
	return id, err
}

func (obj *BudgetUseCase) Delete(ctx context.Context, budgetID int, user storage.UserInfo) error {
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

func (obj *BudgetUseCase) GetById(ctx context.Context, id int) (storage.BudgetInfo, error) {
	budget, err := obj.repo.GetById(ctx, id)
	return budget, err
}

func (obj *BudgetUseCase) GetBudgetsOfUser(ctx context.Context, user storage.UserInfo) ([]int, error) {
	ids, err := obj.repo.GetIDsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *BudgetUseCase) IsUserAuthor(ctx context.Context, budget storage.BudgetInfo, user storage.UserInfo) bool {
	return user.Id == budget.Author
}
