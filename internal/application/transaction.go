package application

import (
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"

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
	BulkCreate(ctx context.Context, transactions []models.TransactionModel, accounts []models.AccountModel) error
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
	account, err := obj.accountRepo.GetByAccountId(ctx, transaction.AccountId)
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
	oldAccount, err := obj.accountRepo.GetByAccountId(ctx, oldTransaction.AccountId)
	if err != nil {
		return err
	}
	newAccount, err := obj.accountRepo.GetByAccountId(ctx, transaction.AccountId)
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
	account, err := obj.accountRepo.GetByAccountId(ctx, transaction.AccountId)
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

func (obj *Transaction) BulkCreate(ctx context.Context, transactions []models.TransactionModel, accounts []models.AccountModel) error {
	return obj.repository.BulkCreate(ctx, transactions, accounts)
}

type CsvTransactionsReaderStrategy interface {
	ReadTransactions(ctx context.Context) ([]models.TransactionModel, []models.AccountModel, error)
}

type GpteamReaderStrategy struct {
	csvReader  *csv.Reader
	userId     int
	accountApp AccountUseCase
}

func NewGpteamReaderStrategy(reader *csv.Reader, userId int, accApp AccountUseCase) *GpteamReaderStrategy {
	return &GpteamReaderStrategy{
		csvReader:  reader,
		userId:     userId,
		accountApp: accApp,
	}
}

func (obj *GpteamReaderStrategy) ReadTransactions(ctx context.Context) ([]models.TransactionModel, []models.AccountModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	var accounts []models.AccountModel
	var transactions []models.TransactionModel
	for {
		record, err := obj.csvReader.Read()
		if err == io.EOF {
			return transactions, accounts, nil
		}
		if err != nil {
			log.Warn("failed to read record", zap.Error(err))
			continue
		}
		accountId, err := strconv.Atoi(record[2])
		if err != nil {
			log.Warn("failed to convert account id to int", zap.Error(err))
			continue
		}
		account, err := obj.accountApp.GetById(ctx, obj.userId, accountId)
		if err != nil {
			continue
		}
		value, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Warn("failed to convert value to int", zap.Error(err))
			continue
		}
		transactionDate, err := time.Parse(time.RFC3339, strings.TrimSpace(record[5]))
		if err != nil {
			log.Warn("failed to convert date to time", zap.Error(err))
			continue
		}
		transaction := models.TransactionModel{
			UserId:          obj.userId,
			AccountId:       accountId,
			Value:           value,
			Type:            record[3],
			Category:        record[4],
			Title:           record[0],
			Description:     record[6],
			TransactionDate: transactionDate,
		}
		transactions = append(transactions, transaction)
		accounts = append(accounts, account)
	}
}

type SberReaderStrategy struct {
	csvReader        *csv.Reader
	userId           int
	defaultAccountId int
	accountApp       AccountUseCase
}

func NewSberReaderStrategy(reader *csv.Reader, userId int, defaultAccountId int, accApp AccountUseCase) *SberReaderStrategy {
	return &SberReaderStrategy{
		csvReader:        reader,
		userId:           userId,
		defaultAccountId: defaultAccountId,
		accountApp:       accApp,
	}
}

func (obj *SberReaderStrategy) ReadTransactions(ctx context.Context) ([]models.TransactionModel, []models.AccountModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	defaultAccount, err := obj.accountApp.GetById(ctx, obj.userId, obj.defaultAccountId)
	if err != nil {
		return []models.TransactionModel{}, []models.AccountModel{}, err
	}
	var accounts []models.AccountModel
	var transactions []models.TransactionModel
	for {
		record, err := obj.csvReader.Read()
		if err == io.EOF {
			return transactions, accounts, nil
		}
		if err != nil {
			log.Warn("failed to read record", zap.Error(err))
			continue
		}
		transactionDate, err := time.Parse("02.01.2006 15:04", strings.TrimSpace(record[0]))
		if err != nil {
			log.Warn("failed to convert date to time", zap.Error(err))
			continue
		}
		value, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Warn("failed to convert value to int", zap.Error(err))
			continue
		}
		var transactionType string
		if record[1][0] == '+' {
			transactionType = "INCOME"
		} else {
			transactionType = "EXPENSE"
		}
		transaction := models.TransactionModel{
			UserId:          obj.userId,
			AccountId:       obj.defaultAccountId,
			Value:           value,
			Type:            transactionType,
			Category:        record[1],
			Title:           record[1],
			Description:     record[1],
			TransactionDate: transactionDate,
		}
		transactions = append(transactions, transaction)
		accounts = append(accounts, defaultAccount)
	}
}
