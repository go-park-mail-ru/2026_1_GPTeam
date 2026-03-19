package application

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
)

type TransactionUseCase interface {
	Create(ctx context.Context, transaction models.TransactionModel) (int, error)
	GetTransactionIdsOfUser(ctx context.Context, user models.UserModel) ([]int, error)
}

type Transaction struct {
	repository repository.TransactionRepository
}

func NewTransaction(repo repository.TransactionRepository) *Transaction {
	return &Transaction{repository: repo}
}

func (obj *Transaction) Create(ctx context.Context, transaction models.TransactionModel) (int, error) {
	transaction.AccountId = 1 // ToDo: это заглушка, так как счета надо будет делать позже -> надо самостоятельно сделать себе счёт с id=1
	id, err := obj.repository.Create(ctx, transaction)
	return id, err
}

func (obj *Transaction) GetTransactionIdsOfUser(ctx context.Context, user models.UserModel) ([]int, error) {
	ids, err := obj.repository.GetIdsByUserId(ctx, user.Id)
	return ids, err
}
