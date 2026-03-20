package application

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
)

type TransactionUseCase interface {
	Create(ctx context.Context, transaction models.TransactionModel) (int, error)
	GetTransactionIdsOfUser(ctx context.Context, user models.UserModel) ([]int, error)
	Delete(ctx context.Context, transactionId int) (int, error)
	Detail(ctx context.Context, transactionId int) (models.TransactionModel, error)
	IsUserAuthorOfTransaction(transaction models.TransactionModel, user models.UserModel) (bool, error)
}

type Transaction struct {
	repository repository.TransactionRepository
}

func NewTransaction(repo repository.TransactionRepository) *Transaction {
	return &Transaction{repository: repo}
}

func (obj *Transaction) Create(ctx context.Context, transaction models.TransactionModel) (int, error) {
	id, err := obj.repository.Create(ctx, transaction)
	return id, err
}

func (obj *Transaction) GetTransactionIdsOfUser(ctx context.Context, user models.UserModel) ([]int, error) {
	ids, err := obj.repository.GetIdsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Transaction) Delete(ctx context.Context, transactionId int) (int, error) {
	id, err := obj.repository.Delete(ctx, transactionId)
	if err != nil {
		return 0, err
	}
	return id, err
}

func (obj *Transaction) Detail(ctx context.Context, transactionId int) (models.TransactionModel, error) {
	transaction, err := obj.repository.Detail(ctx, transactionId)
	if err != nil {
		return models.TransactionModel{}, err
	}
	return transaction, nil
}

func (obj *Transaction) IsUserAuthorOfTransaction(transaction models.TransactionModel, user models.UserModel) (bool, error) {
	return transaction.UserId == user.Id, nil
}
