package application

import (
	"context"

	models2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

type BudgetUseCaseInterface interface {
	Create(ctx context.Context, budget models2.BudgetModel) (int, error)
	Delete(ctx context.Context, budgetId int, user models2.UserModel) error
	GetById(ctx context.Context, id int) (models2.BudgetModel, error)
	GetBudgetsOfUser(ctx context.Context, user models2.UserModel) ([]int, error)
	IsUserAuthorOfBudget(budget models2.BudgetModel, user models2.UserModel) bool
	GetAllowedCurrencies() []string
}

type UserUseCaseInterface interface {
	Create(ctx context.Context, user web_helpers.SignupBodyRequest) (web_helpers.AuthUser, error)
	GetById(ctx context.Context, id int) (models2.UserModel, error)
	GetByCredentials(ctx context.Context, user web_helpers.LoginBodyRequest) (models2.UserModel, error)
	IsAuthUserExists(ctx context.Context, isAuth bool, userId int) (web_helpers.User, bool)
}
