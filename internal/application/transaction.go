package application

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=transaction.go -destination=mocks/mock_transaction.go -package=mocks
type TransactionUseCase interface {
	Create(ctx context.Context, transaction models.TransactionModel) (int, error)
	GetTransactionIdsOfUser(ctx context.Context, user models.UserModel) ([]int, error)
	Update(ctx context.Context, transaction models.TransactionModel) error
	Delete(ctx context.Context, transactionId int, userId int) (int, error)
	Detail(ctx context.Context, transactionId int, userId int) (models.TransactionModel, error)
	IsUserAuthorOfTransaction(user models.UserModel, transaction models.TransactionModel) bool
	Search(ctx context.Context, userId int, filters repository.TransactionFilters) ([]models.TransactionModel, error)
}

type Transaction struct {
	repository  repository.TransactionRepository
	accountRepo repository.AccountRepository
}

func NewTransaction(repo repository.TransactionRepository, accRepo repository.AccountRepository) *Transaction {
	return &Transaction{
		repository:  repo,
		accountRepo: accRepo,
	}
}

func (obj *Transaction) Create(ctx context.Context, transaction models.TransactionModel) (int, error) {
	account, err := obj.accountRepo.GetById(ctx, transaction.AccountId)
	if err != nil {
		return 0, err
	}
	id, err := obj.repository.Create(ctx, transaction, account)
	return id, err
}

func (obj *Transaction) GetTransactionIdsOfUser(ctx context.Context, user models.UserModel) ([]int, error) {
	ids, err := obj.repository.GetIdsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Transaction) Update(ctx context.Context, transaction models.TransactionModel) error {
	oldTransaction, err := obj.repository.Detail(ctx, transaction.Id)
	if err != nil {
		return err
	}
	oldAccount, err := obj.accountRepo.GetById(ctx, oldTransaction.AccountId)
	if err != nil {
		return err
	}
	newAccount, err := obj.accountRepo.GetById(ctx, transaction.AccountId)
	if err != nil {
		return err
	}
	err = obj.repository.Update(ctx, transaction, oldTransaction, newAccount, oldAccount)
	return err
}

func (obj *Transaction) Delete(ctx context.Context, transactionId int, userId int) (int, error) {
	log := logger.GetLoggerWithRequestId(ctx)
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
	account, err := obj.accountRepo.GetById(ctx, transaction.AccountId)
	if err != nil {
		return 0, err
	}
	id, err := obj.repository.Delete(ctx, transactionId, account)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (obj *Transaction) Detail(ctx context.Context, transactionId int, userId int) (models.TransactionModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
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

func (obj *Transaction) Search(ctx context.Context, userId int, filters repository.TransactionFilters) ([]models.TransactionModel, error) {
	transactions, err := obj.repository.Search(ctx, userId, filters)
	return transactions, err
}
