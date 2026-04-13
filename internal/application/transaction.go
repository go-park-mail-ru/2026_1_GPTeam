package application

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

//go:generate mockgen -source=transaction.go -destination=mocks/transaction.go -package=mocks
type TransactionUseCase interface {
	Create(ctx context.Context, transaction models.TransactionModel) (int, error)
	GetTransactionIdsOfUser(ctx context.Context, user models.UserModel) ([]int, error)
	Update(ctx context.Context, transaction models.TransactionModel) error
	Delete(ctx context.Context, transactionId int, userId int) (int, error)
	Detail(ctx context.Context, transactionId int, userId int) (models.TransactionModel, error)
	IsUserAuthorOfTransaction(user models.UserModel, transaction models.TransactionModel) bool
}

type Transaction struct {
	repository repository.TransactionRepository
}

func NewTransaction(repo repository.TransactionRepository) *Transaction {
	return &Transaction{
		repository: repo,
	}
}

func (obj *Transaction) Create(ctx context.Context, transaction models.TransactionModel) (int, error) {
	id, err := obj.repository.Create(ctx, transaction)
	return id, err
}

func (obj *Transaction) GetTransactionIdsOfUser(ctx context.Context, user models.UserModel) ([]int, error) {
	ids, err := obj.repository.GetIdsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Transaction) Update(ctx context.Context, transaction models.TransactionModel) error {
	err := obj.repository.Update(ctx, transaction)
	return err
}

func (obj *Transaction) Delete(ctx context.Context, transactionId int, userId int) (int, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	transaction, err := obj.repository.Detail(ctx, transactionId)
	if err != nil {
		return 0, err
	}
	if transaction.UserId != userId {
		log.Warn("user is not author of transaction",
			zap.Int("user_id", userId),
			zap.Int("transaction_id", transactionId))
		return 0, ForbiddenError
	}
	id, err := obj.repository.Delete(ctx, transactionId)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (obj *Transaction) Detail(ctx context.Context, transactionId int, userId int) (models.TransactionModel, error) {
	log := logger.GetLoggerWIthRequestId(ctx)
	transaction, err := obj.repository.Detail(ctx, transactionId)
	if err != nil {
		return models.TransactionModel{}, err
	}
	if transaction.UserId != userId {
		log.Warn("user is not author of transaction",
			zap.Int("user_id", userId),
			zap.Int("transaction_id", transactionId))
		return models.TransactionModel{}, ForbiddenError
	}
	return transaction, nil
}

func (obj *Transaction) IsUserAuthorOfTransaction(user models.UserModel, transaction models.TransactionModel) bool {
	return transaction.UserId == user.Id
}
