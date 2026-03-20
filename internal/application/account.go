package application

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
)

type AccountUseCase interface {
	Create(ctx context.Context, account models.AccountModel) (int, error)
	LinkAccountAndUser(ctx context.Context, accountId int, userId int) error
	IsUserAuthorOfAccount(ctx context.Context, userId int, accountId int) bool
}

type Account struct {
	repository repository.AccountRepository
}

func NewAccount(repo repository.AccountRepository) *Account {
	return &Account{repository: repo}
}

func (obj *Account) Create(ctx context.Context, account models.AccountModel) (int, error) {
	id, err := obj.repository.Create(ctx, account)
	return id, err
}

func (obj *Account) LinkAccountAndUser(ctx context.Context, accountId int, userId int) error {
	_, err := obj.repository.LinkAccountAndUser(ctx, accountId, userId)
	return err
}

func (obj *Account) IsUserAuthorOfAccount(ctx context.Context, userId int, accountId int) bool {
	ids, err := obj.repository.GetIdsByUserAndAccount(ctx, userId, accountId)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return len(ids) > 0
}
