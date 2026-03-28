package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EnumsRepository interface {
	GetCurrencyCodesFromDB() []string
	GetTransactionTypesFromDB() []string
	GetCategoryTypesFromDB() []string
}

type EnumsPostgres struct {
	db               DB
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

func NewEnumsPostgres(ctx context.Context, db *pgxpool.Pool) (*EnumsPostgres, error) {
	currencyCodes, err := getCurrenciesFromDB(ctx, db)
	if err != nil {
		return &EnumsPostgres{}, err
	}
	fmt.Printf("Read currencies from db: %v\n", currencyCodes)
	transactionTypes, err := getTransactionTypesFromDB(ctx, db)
	if err != nil {
		return &EnumsPostgres{}, err
	}
	fmt.Printf("Read transaction types from db: %v\n", transactionTypes)
	categoryTypes, err := getCategoriesFromDB(ctx, db)
	if err != nil {
		return &EnumsPostgres{}, err
	}
	fmt.Printf("Read categories from db: %v\n", categoryTypes)
	return &EnumsPostgres{
		db:               db,
		currencyCodes:    currencyCodes,
		transactionTypes: transactionTypes,
		categoryTypes:    categoryTypes,
	}, nil
}

func getCurrenciesFromDB(ctx context.Context, db *pgxpool.Pool) ([]string, error) {
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

func getTransactionTypesFromDB(ctx context.Context, db *pgxpool.Pool) ([]string, error) {
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

func getCategoriesFromDB(ctx context.Context, db *pgxpool.Pool) ([]string, error) {
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
