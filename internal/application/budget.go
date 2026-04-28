package application

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=budget.go -destination=mocks/budget.go -package=mocks
type BudgetUseCase interface {
	Create(ctx context.Context, budget models.BudgetModel, categories []string) (int, error)
	Delete(ctx context.Context, budgetId int, user models.UserModel) error
	GetById(ctx context.Context, id int, user models.UserModel) (models.BudgetModel, []string, error)
	GetBudgetsOfUser(ctx context.Context, user models.UserModel) ([]int, error)
	IsUserAuthorOfBudget(budget models.BudgetModel, user models.UserModel) bool
}

type Budget struct {
	repository repository.BudgetRepository
}

func NewBudget(repository repository.BudgetRepository) *Budget {
	return &Budget{
		repository: repository,
	}
}

func (obj *Budget) Create(ctx context.Context, budget models.BudgetModel, categories []string) (int, error) {
	loc := time.FixedZone("UTC+3", 3*60*60)
	budget.CreatedAt = time.Date(budget.CreatedAt.Year(), budget.CreatedAt.Month(), budget.CreatedAt.Day(), 0, 0, 0, 0, loc)
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
	if !obj.IsUserAuthorOfBudget(budget, user) {
		log.Warn("user is not author of budget",
			zap.Int("user_id", user.Id),
			zap.Int("budget_id", id))
		return models.BudgetModel{}, []string{}, UserNotAuthorOfBudgetError
	}
	if err != nil {
		return models.BudgetModel{}, []string{}, err
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
