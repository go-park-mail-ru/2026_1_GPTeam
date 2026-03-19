package repository

import (
	"errors"
)

var NothingInTableError = errors.New("no rows in result set")
var InvalidDataInTableError = errors.New("unable to scan: invalid data in table")
var DuplicatedDataError = errors.New("duplicated data in table")
var TransactionDuplicatedDataError = errors.New("duplicated transaction in table")
var ConstraintError = errors.New("constraint error")
var TransactionAccountForeignKeyError = errors.New("account id does not exist")
var TooManyRowsError = errors.New("too many rows in result set")
var UnableToReadCurrenciesError = errors.New("unable to read currencies from db")
var UnableToReadTransactionTypesError = errors.New("unable to read transaction types from db")
var UnableToReadCategoriesError = errors.New("unable to read categories from db")
