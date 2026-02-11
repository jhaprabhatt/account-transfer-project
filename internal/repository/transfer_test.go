package repository

import (
	"account-transfer-project/internal/constants"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"account-transfer-project/internal/models"
)

func setupTransferTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *TransferRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := zap.NewNop()
	repo := NewTransferRepository(db, logger)

	return db, mock, repo
}

func TestTransferRepository_Transfer(t *testing.T) {

	req := &models.TransferRequest{
		SourceID:      100,
		DestinationID: 200,
		Amount:        decimal.NewFromFloat(50.0),
	}
	correlationID := int64(987654321)

	t.Run("Success: Transaction Completed", func(t *testing.T) {
		db, mock, repo := setupTransferTest(t)
		defer db.Close()

		ctx := context.WithValue(context.Background(), CorrelationKey, correlationID)

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT balance FROM accounts WHERE account_id = \$1 FOR UPDATE`).
			WithArgs(req.SourceID).
			WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(decimal.NewFromFloat(1000.0)))

		mock.ExpectQuery(`SELECT balance FROM accounts WHERE account_id = \$1 FOR UPDATE`).
			WithArgs(req.DestinationID).
			WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(decimal.NewFromFloat(500.0)))

		mock.ExpectQuery(`INSERT INTO transfers`).
			WithArgs(
				req.SourceID, req.DestinationID, req.Amount, correlationID,
				constants.StatusPending,
				decimal.NewFromFloat(1000.0), decimal.NewFromFloat(500.0),
			).
			WillReturnRows(sqlmock.NewRows([]string{"transfer_id", "created_at"}).AddRow(1, time.Now()))

		mock.ExpectExec(`UPDATE accounts SET balance = \$1 WHERE account_id = \$2`).
			WithArgs(decimal.NewFromFloat(950.0), req.SourceID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`UPDATE accounts SET balance = \$1 WHERE account_id = \$2`).
			WithArgs(decimal.NewFromFloat(550.0), req.DestinationID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`UPDATE transfers SET status = \$1`).
			WithArgs(
				constants.StatusCompleted,
				decimal.NewFromFloat(950.0), decimal.NewFromFloat(550.0),
				int64(1),
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		result, err := repo.Transfer(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, "950", result.SourcePostBalance)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure: Source Account Not Found", func(t *testing.T) {
		db, mock, repo := setupTransferTest(t)
		defer db.Close()

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT balance FROM accounts WHERE account_id = \$1 FOR UPDATE`).
			WithArgs(req.SourceID).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectRollback()

		_, err := repo.Transfer(context.Background(), req)

		assert.Equal(t, constants.ErrAccountNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure: Insufficient Funds", func(t *testing.T) {
		db, mock, repo := setupTransferTest(t)
		defer db.Close()

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT balance FROM accounts WHERE account_id = \$1 FOR UPDATE`).
			WithArgs(req.SourceID).
			WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(decimal.NewFromFloat(40.0)))

		mock.ExpectQuery(`SELECT balance FROM accounts WHERE account_id = \$1 FOR UPDATE`).
			WithArgs(req.DestinationID).
			WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(decimal.NewFromFloat(500.0)))

		mock.ExpectRollback()

		_, err := repo.Transfer(context.Background(), req)

		assert.Equal(t, constants.ErrInsufficientFunds, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure: DB Error during Insert", func(t *testing.T) {
		db, mock, repo := setupTransferTest(t)
		defer db.Close()

		ctx := context.WithValue(context.Background(), CorrelationKey, correlationID)
		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT balance`).WithArgs(req.SourceID).WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(decimal.NewFromFloat(1000.0)))
		mock.ExpectQuery(`SELECT balance`).WithArgs(req.DestinationID).WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(decimal.NewFromFloat(500.0)))

		mock.ExpectQuery(`INSERT INTO transfers`).
			WillReturnError(errors.New("connection died"))

		mock.ExpectRollback()

		_, err := repo.Transfer(ctx, req)

		assert.Equal(t, constants.ErrSystem, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
