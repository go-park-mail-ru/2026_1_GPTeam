package application

import (
	"errors"
)

var (
	ErrUserNotAuthorOfBudget = errors.New("user not author of budget")
	ErrHashPassword          = errors.New("unable to hash password")
	ErrForbidden             = errors.New("forbidden")
	ErrAllFieldsEmpty        = errors.New("all fields are empty")
	ErrAccountNotFound       = errors.New("account not found")
	ErrOwnerCannotLeave      = errors.New("owner cannot leave account, delete it instead")
	ErrInviteAlreadyExists   = errors.New("invite already exists")
	ErrInviteNotFound        = errors.New("invite not found")
	ErrNotOwner              = errors.New("only owner can perform this action")
	ErrCannotRemoveOwner     = errors.New("cannot remove owner from account")
	ErrSelfInvite            = errors.New("cannot invite yourself")
	ErrAlreadyMember         = errors.New("user already member")
	ErrMemberNotFound        = errors.New("member not found")
)
