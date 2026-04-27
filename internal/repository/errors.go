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
var AccountDuplicatedDataError = errors.New("duplicated account in table")
var AccountForeignKeyError = errors.New("account id does not exist")
var UnableToGetAccountUserIdsError = errors.New("unable to read ids from account_user")
var IncorrectRowsAffectedError = errors.New("unexpected value of 'rows affected'")
var ResultNotOkError = errors.New("result not ok")
var NoIpInSavedError = errors.New("no ip saved")

var ErrAccountNotFound = errors.New("account not found")
