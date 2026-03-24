package application

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type BudgetUseCase interface {
	Create(ctx context.Context, budget models.BudgetModel) (int, error)
	Delete(ctx context.Context, budgetId int, user models.UserModel) error
	GetById(ctx context.Context, id int, user models.UserModel) (models.BudgetModel, error)
	GetBudgetsOfUser(ctx context.Context, user models.UserModel) ([]int, error)
	IsUserAuthorOfBudget(budget models.BudgetModel, user models.UserModel) bool
}

type Budget struct {
	repository repository.BudgetRepository
	log        *zap.Logger
}

func NewBudget(repository repository.BudgetRepository) *Budget {
	return &Budget{
		repository: repository,
		log:        logger.GetLogger(),
	}
}

func (obj *Budget) Create(ctx context.Context, budget models.BudgetModel) (int, error) {
	obj.log.Info("creating budget", zap.String("budget_title", budget.Title), zap.Int("user_id", budget.Author))
	budget.CreatedAt = time.Date(budget.CreatedAt.Year(), budget.CreatedAt.Month(), budget.CreatedAt.Day(), 0, 0, 0, 0, time.UTC)
	id, err := obj.repository.Create(ctx, budget)
	return id, err
}

func (obj *Budget) Delete(ctx context.Context, budgetId int, user models.UserModel) error {
	obj.log.Info("deleting budget", zap.Int("budget_id", budgetId))
	_, err := obj.GetById(ctx, budgetId, user)
	if err != nil {
		return err
	}
	err = obj.repository.Delete(ctx, budgetId)
	return err
}

func (obj *Budget) GetById(ctx context.Context, id int, user models.UserModel) (models.BudgetModel, error) {
	obj.log.Info("getting budget", zap.Int("budget_id", id))
	budget, err := obj.repository.GetById(ctx, id)
	if !obj.IsUserAuthorOfBudget(budget, user) {
		return models.BudgetModel{}, UserNotAuthorOfBudgetError
	}
	return budget, err
}

func (obj *Budget) GetBudgetsOfUser(ctx context.Context, user models.UserModel) ([]int, error) {
	obj.log.Info("getting budgets of user", zap.Int("user_id", user.Id))
	ids, err := obj.repository.GetIdsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Budget) IsUserAuthorOfBudget(budget models.BudgetModel, user models.UserModel) bool {
	obj.log.Info("checking if user author of budget", zap.Int("user_id", budget.Author), zap.Int("budget_id", budget.Id), zap.Bool("res", user.Id == budget.Author))
	return user.Id == budget.Author
}
