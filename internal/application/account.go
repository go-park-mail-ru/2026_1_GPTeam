package application

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type AccountUseCase interface {
	Create(ctx context.Context, account models.AccountModel) (int, error)
	LinkAccountAndUser(ctx context.Context, accountId int, userId int) error
	IsUserAuthorOfAccount(ctx context.Context, userId int, accountId int) bool
	GetAccountIdByUserId(ctx context.Context, userId int) (int, error)
}

type Account struct {
	repository repository.AccountRepository
}

func NewAccount(repo repository.AccountRepository) *Account {
	return &Account{
		repository: repo,
	}
}

func (obj *Account) Create(ctx context.Context, account models.AccountModel) (int, error) {
	id, err := obj.repository.Create(ctx, account)
	return id, err
}

func (obj *Account) GetAccountIdByUserId(ctx context.Context, userId int) (int, error) {
	return obj.repository.GetAccountIdByUserId(ctx, userId)
}

func (obj *Account) LinkAccountAndUser(ctx context.Context, accountId int, userId int) error {
	_, err := obj.repository.LinkAccountAndUser(ctx, accountId, userId)
	return err
}

func (obj *Account) IsUserAuthorOfAccount(ctx context.Context, userId int, accountId int) bool {
	log := logger.GetLoggerWIthRequestId(ctx)
	ids, err := obj.repository.GetIdsByUserAndAccount(ctx, userId, accountId)
	if err != nil {
		return false
	}
	if len(ids) == 0 {
		log.Warn("user is not author of account",
			zap.Int("userId", userId),
			zap.Int("accountId", accountId))
		return false
	}
	return true
}
