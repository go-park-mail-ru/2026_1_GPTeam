package repository

import (
	"context"
	"sync"

	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=enums.go -destination=mocks/mock_enums.go -package=mocks
type EnumsRepository interface {
	GetCurrencyCodesFromDB() []string
	GetTransactionTypesFromDB() []string
	GetCategoryTypesFromDB() []string
}

type EnumsPostgres struct {
	db               DB // Используем интерфейс DB вместо *pgxpool.Pool
	mu               sync.RWMutex
	currencyCodes    []string
	transactionTypes []string
	categoryTypes    []string
}

func (obj *EnumsPostgres) GetCurrencyCodesFromDB() []string {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.currencyCodes
}

func (obj *EnumsPostgres) GetTransactionTypesFromDB() []string {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.transactionTypes
}

func (obj *EnumsPostgres) GetCategoryTypesFromDB() []string {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.categoryTypes
}

// NewEnumsPostgres теперь принимает DB интерфейс. Это "лечит" ошибку в тестах.
func NewEnumsPostgres(ctx context.Context, db DB) (*EnumsPostgres, error) {
	log := logger.GetLogger()

	currencyCodes, err := getCurrenciesFromDB(ctx, db)
	if err != nil {
		log.Error("failed to get currency codes from db", zap.Error(err))
		return &EnumsPostgres{}, err
	}
	log.Info("Read currencies from db", zap.Strings("currency_codes", currencyCodes))

	transactionTypes, err := getTransactionTypesFromDB(ctx, db)
	if err != nil {
		log.Error("failed to get transaction types from db", zap.Error(err))
		return &EnumsPostgres{}, err
	}
	log.Info("Read transaction types from db", zap.Strings("transaction_types", transactionTypes))

	categoryTypes, err := getCategoriesFromDB(ctx, db)
	if err != nil {
		log.Error("failed to get categories from db", zap.Error(err))
		return &EnumsPostgres{}, err
	}
	log.Info("Read categories from db", zap.Strings("categories", categoryTypes))

	return &EnumsPostgres{
		db:               db,
		currencyCodes:    currencyCodes,
		transactionTypes: transactionTypes,
		categoryTypes:    categoryTypes,
	}, nil
}

func getCurrenciesFromDB(ctx context.Context, db DB) ([]string, error) {
	query := `select enumlabel from pg_enum where enumtypid = 'currency_code'::regtype order by enumsortorder;`
	rows, err := db.Query(ctx, query)
	if err != nil {
		return []string{}, UnableToReadCurrenciesError
	}
	defer rows.Close()
	var currencies []string
	for rows.Next() {
		var code string
		if err = rows.Scan(&code); err != nil {
			return []string{}, UnableToReadCurrenciesError
		}
		currencies = append(currencies, code)
	}
	return currencies, nil
}

func getTransactionTypesFromDB(ctx context.Context, db DB) ([]string, error) {
	query := `select enumlabel from pg_enum where enumtypid = 'transaction_type'::regtype order by enumsortorder;`
	rows, err := db.Query(ctx, query)
	if err != nil {
		return []string{}, UnableToReadTransactionTypesError
	}
	defer rows.Close()
	var transactionTypes []string
	for rows.Next() {
		var t string
		if err = rows.Scan(&t); err != nil {
			return []string{}, UnableToReadTransactionTypesError
		}
		transactionTypes = append(transactionTypes, t)
	}
	return transactionTypes, nil
}

func getCategoriesFromDB(ctx context.Context, db DB) ([]string, error) {
	query := `select enumlabel from pg_enum where enumtypid = 'category_type'::regtype order by enumsortorder;`
	rows, err := db.Query(ctx, query)
	if err != nil {
		return []string{}, UnableToReadCategoriesError
	}
	defer rows.Close()
	var categories []string
	for rows.Next() {
		var c string
		if err = rows.Scan(&c); err != nil {
			return []string{}, UnableToReadCategoriesError
		}
		categories = append(categories, c)
	}
	return categories, nil
}
