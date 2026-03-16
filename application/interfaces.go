package application

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/web/web_helpers"
)

type BudgetUseCaseInterface interface {
	Create(ctx context.Context, budget models.BudgetInfo) (int, error)
	Delete(ctx context.Context, budgetID int, user models.UserInfo) error
	GetById(ctx context.Context, id int) (models.BudgetInfo, error)
	GetBudgetsOfUser(ctx context.Context, user models.UserInfo) ([]int, error)
	IsUserAuthorOfBudget(budget models.BudgetInfo, user models.UserInfo) bool
	GetAllowedCurrencies() []string
}

type UserUseCaseInterface interface {
	Create(ctx context.Context, user web_helpers.SignupBodyRequest) (web_helpers.AuthUser, error)
	GetById(ctx context.Context, id int) (models.UserInfo, error)
	GetByCredentials(ctx context.Context, user web_helpers.LoginBodyRequest) (models.UserInfo, error)
	IsAuthUserExists(ctx context.Context, isAuth bool, userID int) (web_helpers.User, bool)
}
