package application

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
)

type TransactionUseCase interface {
	GetTransactionsOfUser(ctx context.Context, user models.UserModel) ([]int, error)
}

type Transaction struct {
	repository repository.TransactionRepository
}

func NewTransaction(repo repository.TransactionRepository) *Transaction {
	return &Transaction{repository: repo}
}

func (obj *Transaction) GetTransactionsOfUser(ctx context.Context, user models.UserModel) ([]int, error) {
	ids, err := obj.repository.GetIdsByUserId(ctx, user.Id)
	return ids, err
}
