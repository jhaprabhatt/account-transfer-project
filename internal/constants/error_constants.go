package constants

import "errors"

var (
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	ErrInvalidAccountID     = errors.New("invalid account_id: must be positive")
	ErrSameAccount          = errors.New("source and destination account cannot be the same")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrAccountNotFound      = errors.New("account not found")
	ErrSystem               = errors.New("internal system error")
	ErrAccountAlreadyExists = errors.New("account already exists")
)
