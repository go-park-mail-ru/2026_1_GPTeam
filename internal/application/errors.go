package application

import (
	"errors"
)

var WrongSigningMethodError = errors.New("unexpected signing method")
var InvalidTokenID = errors.New("invalid token id")
var UserNotAuthorOfBudgetError = errors.New("user not author of budget")
var HashPasswordError = errors.New("unable to hash password")
var ForbiddenError = errors.New("forbidden")
var AllFieldsEmptyError = errors.New("all fields are empty")
