package repository

import (
	"context"

	models2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, userInfo models2.UserModel) (int, error)
	GetById(ctx context.Context, id int) (models2.UserModel, error)
	GetByUsername(ctx context.Context, username string) (models2.UserModel, error)
	GetByEmail(ctx context.Context, email string) (models2.UserModel, error)
}

type BudgetRepositoryInterface interface {
	Create(ctx context.Context, budget models2.BudgetModel) (int, error)
	GetById(ctx context.Context, id int) (models2.BudgetModel, error)
	GetIdsByUserId(ctx context.Context, userId int) ([]int, error)
	Delete(ctx context.Context, id int) error
	GetCurrencies() []string
}

type JwtRepositoryInterface interface {
	Create(ctx context.Context, token models2.RefreshTokenModel) error
	DeleteByUuid(ctx context.Context, uuid string) error
	DeleteByUserId(ctx context.Context, userId int) error
	Get(ctx context.Context, uuid string) (models2.RefreshTokenModel, error)
}
