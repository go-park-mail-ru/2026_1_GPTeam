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
	log        *zap.Logger
}

func NewAccount(repo repository.AccountRepository) *Account {
	return &Account{
		repository: repo,
		log:        logger.GetLogger(),
	}
}

func (obj *Account) Create(ctx context.Context, account models.AccountModel) (int, error) {
	obj.log.Info("creating account", zap.String("name", account.Name))
	id, err := obj.repository.Create(ctx, account)
	return id, err
}

func (obj *Account) GetAccountIdByUserId(ctx context.Context, userId int) (int, error) {
	obj.log.Info("getting account by user id", zap.Int("user_id", userId))
	return obj.repository.GetAccountIdByUserId(ctx, userId)
}

func (obj *Account) LinkAccountAndUser(ctx context.Context, accountId int, userId int) error {
	obj.log.Info("linking account and user", zap.Int("account_id", accountId), zap.Int("user_id", userId))
	_, err := obj.repository.LinkAccountAndUser(ctx, accountId, userId)
	return err
}

func (obj *Account) IsUserAuthorOfAccount(ctx context.Context, userId int, accountId int) bool {
	obj.log.Info("checking is user author of account", zap.Int("user_id", userId), zap.Int("account_id", accountId))
	ids, err := obj.repository.GetIdsByUserAndAccount(ctx, userId, accountId)
	if err != nil {
		obj.log.Warn("user is not author of account", zap.Int("user_id", userId), zap.Int("account_id", accountId))
		return false
	}
	return len(ids) > 0
}
