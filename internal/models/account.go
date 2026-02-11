package models

import "github.com/shopspring/decimal"

type Account struct {
	ID      int64           `json:"account_id"`
	Balance decimal.Decimal `json:"balance"`
}

func (a *Account) CanWithdraw(amount decimal.Decimal) bool {
	return a.Balance.GreaterThanOrEqual(amount)
}

type CreateAccountRequest struct {
	ID      int64           `json:"account_id"`
	Balance decimal.Decimal `json:"balance"`
}
