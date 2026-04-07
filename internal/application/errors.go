package application

import (
	"errors"
)

var (
	WrongSigningMethodError    = errors.New("unexpected signing method")
	InvalidTokenID             = errors.New("invalid token id")
	UserNotAuthorOfBudgetError = errors.New("user not author of budget")
	HashPasswordError          = errors.New("unable to hash password")
	ForbiddenError             = errors.New("forbidden")
	AllFieldsEmptyError        = errors.New("all fields are empty")
	ErrAccountNotFound         = errors.New("account not found")
	InternalParserError        = errors.New("internal parser error")
	InternalTranscriptionError = errors.New("internal transcription error")
)
