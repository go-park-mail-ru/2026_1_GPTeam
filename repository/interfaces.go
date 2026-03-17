package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, userInfo models.UserModel) (int, error)
	GetById(ctx context.Context, id int) (models.UserModel, error)
	GetByUsername(ctx context.Context, username string) (models.UserModel, error)
	GetByEmail(ctx context.Context, email string) (models.UserModel, error)
}

type BudgetRepositoryInterface interface {
	Create(ctx context.Context, budget models.BudgetModel) (int, error)
	GetById(ctx context.Context, id int) (models.BudgetModel, error)
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
	Delete(ctx context.Context, id int) error
	GetCurrencies() []string
}

type JwtRepositoryInterface interface {
	Create(ctx context.Context, token models.RefreshTokenModel) error
	DeleteByUuid(ctx context.Context, uuid string) error
	DeleteByUserId(ctx context.Context, userId int) error
	Get(ctx context.Context, uuid string) (models.RefreshTokenModel, error)
}
