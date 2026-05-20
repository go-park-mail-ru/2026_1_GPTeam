package repository

import (
	"errors"
)

var ErrNothingInTable = errors.New("no rows in result set")
var ErrInvalidDataInTable = errors.New("unable to scan: invalid data in table")
var ErrDuplicatedData = errors.New("duplicated data in table")
var ErrTransactionDuplicatedData = errors.New("duplicated transaction in table")
var ErrConstraint = errors.New("constraint error")
var ErrTransactionAccountForeignKey = errors.New("account id does not exist")
var ErrTooManyRows = errors.New("too many rows in result set")
var ErrUnableToReadCurrencies = errors.New("unable to read currencies from db")
var ErrUnableToReadTransactionTypes = errors.New("unable to read transaction types from db")
var ErrUnableToReadCategories = errors.New("unable to read categories from db")
var ErrAccountDuplicatedData = errors.New("duplicated account in table")
var ErrAccountForeignKey = errors.New("account id does not exist")
var ErrUnableToGetAccountUserIds = errors.New("unable to read ids from account_user")
var ErrIncorrectRowsAffected = errors.New("unexpected value of 'rows affected'")
var ErrResultNotOk = errors.New("result not ok")
var ErrNoIpInSaved = errors.New("no ip saved")
