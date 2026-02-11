package models

import (
	"github.com/jhaprabhatt/account-transfer-project/internal/constants"

	"github.com/shopspring/decimal"
)

type TransferRequest struct {
	SourceID      int64           `json:"source_account_id"`
	DestinationID int64           `json:"destination_account_id"`
	Amount        decimal.Decimal `json:"amount"`
}

func (r *TransferRequest) Validate() error {
	if r.Amount.LessThanOrEqual(decimal.Zero) {
		return constants.ErrAmountMustBePositive
	}

	if r.SourceID <= 0 || r.DestinationID <= 0 {
		return constants.ErrInvalidAccountID
	}

	if r.SourceID == r.DestinationID {
		return constants.ErrSameAccount
	}

	return nil
}
