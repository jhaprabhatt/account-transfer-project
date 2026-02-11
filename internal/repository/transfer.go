package repository

import (
	"account-transfer-project/internal/constants"
	"account-transfer-project/internal/models"
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type contextKey string

const CorrelationKey contextKey = "correlation_id"

type TransferRepository struct {
	db  *sql.DB
	log *zap.Logger
}

func NewTransferRepository(db *sql.DB, log *zap.Logger) *TransferRepository {
	return &TransferRepository{db: db, log: log}
}

func (r *TransferRepository) Transfer(ctx context.Context, req *models.TransferRequest) (*models.TransferResult, error) {

	correlationID, _ := ctx.Value(CorrelationKey).(int64)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		r.log.Error("failed to begin tx", zap.Error(err))
		return nil, constants.ErrSystem
	}
	defer tx.Rollback()

	firstID, secondID := req.SourceID, req.DestinationID
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}

	var srcPre, destPre decimal.Decimal

	err = tx.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE", req.SourceID).Scan(&srcPre)
	if err != nil {
		return nil, constants.ErrAccountNotFound
	}

	err = tx.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE", req.DestinationID).Scan(&destPre)
	if err != nil {
		return nil, constants.ErrAccountNotFound
	}

	if srcPre.LessThan(req.Amount) {
		return nil, constants.ErrInsufficientFunds
	}

	var transferID int64
	var createdAt time.Time
	err = tx.QueryRowContext(ctx, `
        INSERT INTO transfers (
            source_account_id, destination_account_id, amount, 
            correlation_id, status, source_prev_balance, destination_prev_balance
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7) 
        RETURNING transfer_id, created_at`,
		req.SourceID, req.DestinationID, req.Amount, correlationID,
		constants.StatusPending, srcPre, destPre,
	).Scan(&transferID, &createdAt)

	if err != nil {
		r.log.Error("failed to create audit log", zap.Error(err))
		return nil, constants.ErrSystem
	}

	var srcPost, destPost decimal.Decimal
	srcPost = srcPre.Sub(req.Amount)
	destPost = destPre.Add(req.Amount)

	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = $1 WHERE account_id = $2", srcPost, req.SourceID)
	if err != nil {
		return nil, constants.ErrSystem
	}

	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = $1 WHERE account_id = $2", destPost, req.DestinationID)
	if err != nil {
		return nil, constants.ErrSystem
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE transfers SET 
            status = $1,
            source_post_balance = $2, 
            destination_post_balance = $3
        WHERE transfer_id = $4`,
		constants.StatusCompleted, srcPost, destPost, transferID,
	)
	if err != nil {
		r.log.Error("failed to finalize audit", zap.Error(err))
		return nil, constants.ErrSystem
	}

	if err := tx.Commit(); err != nil {
		return nil, constants.ErrSystem
	}

	return &models.TransferResult{
		AuditID:           transferID,
		CorrelationID:     correlationID,
		Status:            "SUCCESS",
		SourcePostBalance: srcPost.String(),
		CreatedAt:         createdAt,
	}, nil
}
