package application

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

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
	log        *zap.Logger
}

func NewTransaction(repo repository.TransactionRepository) *Transaction {
	return &Transaction{
		repository: repo,
		log:        logger.GetLogger(),
	}
}

func (obj *Transaction) Create(ctx context.Context, transaction models.TransactionModel) (int, error) {
	obj.log.Info("creating transaction",
		zap.String("title", transaction.Title),
		zap.String("request_id", ctx.Value("request_id").(string)))
	id, err := obj.repository.Create(ctx, transaction)
	return id, err
}

func (obj *Transaction) GetTransactionIdsOfUser(ctx context.Context, user models.UserModel) ([]int, error) {
	obj.log.Info("getting transaction ids of user",
		zap.Int("user_id", user.Id),
		zap.String("request_id", ctx.Value("request_id").(string)))
	ids, err := obj.repository.GetIdsByUserId(ctx, user.Id)
	return ids, err
}

func (obj *Transaction) Update(ctx context.Context, transaction models.TransactionModel) error {
	obj.log.Info("updating transaction",
		zap.Int("transaction_id", transaction.Id),
		zap.Int("user_id", transaction.UserId),
		zap.String("request_id", ctx.Value("request_id").(string)))
	err := obj.repository.Update(ctx, transaction)
	return err
}

func (obj *Transaction) Delete(ctx context.Context, transactionId int, userId int) (int, error) {
	obj.log.Info("deleting transaction",
		zap.Int("transaction_id", transactionId),
		zap.Int("user_id", userId),
		zap.String("request_id", ctx.Value("request_id").(string)))
	transaction, err := obj.repository.Detail(ctx, transactionId)
	if err != nil {
		return 0, err
	}
	if transaction.UserId != userId {
		obj.log.Warn("user is not author of transaction",
			zap.Int("user_id", userId),
			zap.Int("transaction_id", transactionId),
			zap.String("request_id", ctx.Value("request_id").(string)))
		return 0, ForbiddenError
	}
	id, err := obj.repository.Delete(ctx, transactionId)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (obj *Transaction) Detail(ctx context.Context, transactionId int, userId int) (models.TransactionModel, error) {
	obj.log.Info("getting transaction detail info",
		zap.Int("transaction_id", transactionId),
		zap.Int("user_id", userId),
		zap.String("request_id", ctx.Value("request_id").(string)))
	transaction, err := obj.repository.Detail(ctx, transactionId)
	if err != nil {
		return models.TransactionModel{}, err
	}
	if transaction.UserId != userId {
		obj.log.Warn("user is not author of transaction",
			zap.Int("user_id", userId),
			zap.Int("transaction_id", transactionId),
			zap.String("request_id", ctx.Value("request_id").(string)))
		return models.TransactionModel{}, ForbiddenError
	}
	return transaction, nil
}

func (obj *Transaction) IsUserAuthorOfTransaction(user models.UserModel, transaction models.TransactionModel) bool {
	obj.log.Info("checking if user author of transaction",
		zap.Int("user_id", user.Id),
		zap.Int("transaction_id", transaction.Id),
		zap.Bool("is_author", user.Id == transaction.UserId))
	return transaction.UserId == user.Id
}
