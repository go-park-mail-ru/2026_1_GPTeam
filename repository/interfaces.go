package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/application/models"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, userInfo models.UserInfo) (int, error)
	GetById(ctx context.Context, id int) (models.UserInfo, error)
	GetByUsername(ctx context.Context, username string) (models.UserInfo, error)
	GetByEmail(ctx context.Context, email string) (models.UserInfo, error)
}

type BudgetRepositoryInterface interface {
	Create(ctx context.Context, budget models.BudgetInfo) (int, error)
	GetById(ctx context.Context, id int) (models.BudgetInfo, error)
	GetIDsByUserId(ctx context.Context, userID int) ([]int, error)
	Delete(ctx context.Context, id int) error
	GetCurrencies() []string
}

type JWTRepositoryInterface interface {
	Create(ctx context.Context, uuid string, token models.RefreshTokenInfo) error
	DeleteByUUID(ctx context.Context, uuid string) error
	DeleteByUserID(ctx context.Context, userID int) error
	Get(ctx context.Context, uuid string) (models.RefreshTokenInfo, error)
}
