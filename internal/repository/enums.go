package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
)

type EnumsRepository interface {
	GetCurrencyCodesFromDB() []string
	GetTransactionTypesFromDB() []string
	GetCategoryTypesFromDB() []string
}

type EnumsPostgres struct {
	db               *pgx.Conn
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

func NewEnumsPostgres(db *pgx.Conn) (*EnumsPostgres, error) {
	currencyCodes, err := getCurrenciesFromDB(db)
	if err != nil {
		return &EnumsPostgres{}, err
	}
	fmt.Printf("Read currencies from db: %v\n", currencyCodes)
	transactionTypes, err := getTransactionTypesFromDB(db)
	if err != nil {
		return &EnumsPostgres{}, err
	}
	fmt.Printf("Read transaction types from db: %v\n", transactionTypes)
	categoryTypes, err := getCategoriesFromDB(db)
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

func getCurrenciesFromDB(db *pgx.Conn) ([]string, error) {
	query := `select enumlabel from pg_enum where enumtypid = 'currency_code'::regtype order by enumsortorder;`
	row, err := db.Query(context.Background(), query)
	if err != nil {
		return []string{}, UnableToReadCurrenciesError
	}
	var currencies []string
	for row.Next() {
		var code string
		err = row.Scan(&code)
		if err != nil {
			return []string{}, UnableToReadCurrenciesError
		}
		currencies = append(currencies, code)
	}
	return currencies, nil
}

func getTransactionTypesFromDB(db *pgx.Conn) ([]string, error) {
	query := `select enumlabel from pg_enum where enumtypid = 'transaction_type'::regtype order by enumsortorder;`
	row, err := db.Query(context.Background(), query)
	if err != nil {
		return []string{}, UnableToReadTransactionTypesError
	}
	var transactionTypes []string
	for row.Next() {
		var transactionType string
		err = row.Scan(&transactionType)
		if err != nil {
			return []string{}, UnableToReadTransactionTypesError
		}
		transactionTypes = append(transactionTypes, transactionType)
	}
	return transactionTypes, nil
}

func getCategoriesFromDB(db *pgx.Conn) ([]string, error) {
	query := `select enumlabel from pg_enum where enumtypid = 'category_type'::regtype order by enumsortorder;`
	row, err := db.Query(context.Background(), query)
	if err != nil {
		return []string{}, UnableToReadCategoriesError
	}
	var categories []string
	for row.Next() {
		var category string
		err = row.Scan(&category)
		if err != nil {
			return []string{}, UnableToReadCategoriesError
		}
		categories = append(categories, category)
	}
	return categories, nil
}
