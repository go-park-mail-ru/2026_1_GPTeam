package application

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

//go:generate mockgen -source=account.go -destination=mocks/account.go -package=mocks
type AccountUseCase interface {
	Create(ctx context.Context, account models.AccountModel) (int, error)
	CreateForUser(ctx context.Context, userId int, account models.AccountCreateModel) (models.AccountModel, error)
	LinkAccountAndUser(ctx context.Context, accountId int, userId int) error
	IsUserAuthorOfAccount(ctx context.Context, userId int, accountId int) bool
	GetAccountIdByUserId(ctx context.Context, userId int) (int, error)

	GetById(ctx context.Context, userId int, accountId int) (models.AccountModel, error)
	GetByUserId(ctx context.Context, userId int) ([]models.AccountModel, error)
	GetAllAccountsByUserIdWithBalance(ctx context.Context, userId int) ([]models.AccountModel, []float64, []float64, error)
	GetAllAccountsByUserId(ctx context.Context, userId int) ([]models.AccountModel, error)

	Update(ctx context.Context, userId int, accountId int, account models.AccountUpdateModel) (models.AccountModel, error)
	Delete(ctx context.Context, userId int, accountId int) error
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

func (obj *Account) CreateForUser(ctx context.Context, userId int, account models.AccountCreateModel) (models.AccountModel, error) {
	now := time.Now()
	accountId, err := obj.repository.Create(ctx, models.AccountModel{
		Name:      account.Name,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return models.AccountModel{}, err
	}
	if _, err = obj.repository.LinkAccountAndUser(ctx, accountId, userId); err != nil {
		return models.AccountModel{}, err
	}
	return obj.repository.GetById(ctx, userId, accountId)
}

func (obj *Account) GetAccountIdByUserId(ctx context.Context, userId int) (int, error) {
	return obj.repository.GetAccountIdByUserId(ctx, userId)
}

func (obj *Account) GetById(ctx context.Context, userId int, accountId int) (models.AccountModel, error) {
	account, err := obj.repository.GetById(ctx, userId, accountId)
	if err != nil {
		if err == repository.ErrAccountNotFound {
			return models.AccountModel{}, ErrAccountNotFound
		}
		return models.AccountModel{}, err
	}
	return account, nil
}

func (obj *Account) GetByUserId(ctx context.Context, userId int) ([]models.AccountModel, error) {
	return obj.repository.GetByUserId(ctx, userId)
}

func (obj *Account) Update(ctx context.Context, userId int, accountId int, account models.AccountUpdateModel) (models.AccountModel, error) {
	if account.Name == nil && account.Balance == nil && account.Currency == nil {
		return models.AccountModel{}, AllFieldsEmptyError
	}
	updated, err := obj.repository.Update(ctx, userId, accountId, account)
	if err != nil {
		if err == repository.ErrAccountNotFound {
			return models.AccountModel{}, ErrAccountNotFound
		}
		return models.AccountModel{}, err
	}
	return updated, nil
}

func (obj *Account) Delete(ctx context.Context, userId int, accountId int) error {
	err := obj.repository.Delete(ctx, userId, accountId)
	if err == repository.ErrAccountNotFound {
		return ErrAccountNotFound
	}
	return err
}

func (obj *Account) LinkAccountAndUser(ctx context.Context, accountId int, userId int) error {
	_, err := obj.repository.LinkAccountAndUser(ctx, accountId, userId)
	return err
}

func (obj *Account) IsUserAuthorOfAccount(ctx context.Context, userId int, accountId int) bool {
	log := logger.GetLoggerWithRequestId(ctx)
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

func (obj *Account) GetAllAccountsByUserIdWithBalance(ctx context.Context, userId int) ([]models.AccountModel, []float64, []float64, error) {
	accounts, income, expenses, err := obj.repository.GetAllAccountsByUserIdWithBalance(ctx, userId)
	return accounts, income, expenses, err
}

func (obj *Account) GetAllAccountsByUserId(ctx context.Context, userId int) ([]models.AccountModel, error) {
	accounts, err := obj.repository.GetAllAccountsByUserId(ctx, userId)
	return accounts, err
}
