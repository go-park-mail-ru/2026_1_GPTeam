package application

import "fmt"

type ErrorFunc func(args ...interface{}) error

var WrongSigningMethodError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("unexpected signing method: %v\n", args)
}
var InvalidTokenID ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("invalid token id %v\n", args)
}

var UserNotAuthorOfBudgetError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("user %v not author of budget %v\n", args[0], args[1])
}
