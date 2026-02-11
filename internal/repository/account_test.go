package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jhaprabhatt/account-transfer-project/internal/constants"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/jhaprabhatt/account-transfer-project/internal/models"
)

func setupTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *AccountRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := zap.NewNop()
	repo := NewAccountRepository(db, logger)

	return db, mock, repo
}

func TestAccountRepository_CreateAccount(t *testing.T) {
	acc := &models.Account{
		ID:      101,
		Balance: decimal.NewFromFloat(500.00),
	}

	t.Run("Success", func(t *testing.T) {
		db, mock, repo := setupTest(t)
		defer db.Close()

		mock.ExpectExec(`INSERT INTO accounts`).
			WithArgs(acc.ID, acc.Balance).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.CreateAccount(context.Background(), acc)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure: DB Error", func(t *testing.T) {
		db, mock, repo := setupTest(t)
		defer db.Close()

		mock.ExpectExec(`INSERT INTO accounts`).
			WithArgs(acc.ID, acc.Balance).
			WillReturnError(errors.New("duplicate key violation"))

		err := repo.CreateAccount(context.Background(), acc)

		assert.ErrorContains(t, err, "duplicate key violation")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAccountRepository_GetAccount(t *testing.T) {
	accountID := int64(101)
	expectedBalance := decimal.NewFromFloat(100.50)

	t.Run("Success: Account Found", func(t *testing.T) {
		db, mock, repo := setupTest(t)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"account_id", "balance"}).
			AddRow(accountID, expectedBalance)

		mock.ExpectQuery(`SELECT account_id, balance FROM accounts`).
			WithArgs(accountID).
			WillReturnRows(rows)

		acc, err := repo.GetAccount(context.Background(), accountID)

		assert.NoError(t, err)
		assert.Equal(t, accountID, acc.ID)

		assert.True(t, expectedBalance.Equal(acc.Balance))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure: Account Not Found", func(t *testing.T) {
		db, mock, repo := setupTest(t)
		defer db.Close()

		mock.ExpectQuery(`SELECT account_id, balance FROM accounts`).
			WithArgs(accountID).
			WillReturnRows(sqlmock.NewRows([]string{"account_id", "balance"}))

		acc, err := repo.GetAccount(context.Background(), accountID)

		assert.ErrorIs(t, err, constants.ErrAccountNotFound)
		assert.Nil(t, acc)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure: Generic DB Error", func(t *testing.T) {
		db, mock, repo := setupTest(t)
		defer db.Close()

		mock.ExpectQuery(`SELECT account_id, balance FROM accounts`).
			WithArgs(accountID).
			WillReturnError(errors.New("connection died"))

		acc, err := repo.GetAccount(context.Background(), accountID)

		assert.ErrorContains(t, err, "get account failed")
		assert.Nil(t, acc)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAccountRepository_GetAll(t *testing.T) {
	t.Run("Success: Returns Accounts", func(t *testing.T) {
		db, mock, repo := setupTest(t)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"account_id", "balance"}).
			AddRow(1, decimal.NewFromFloat(100.0)).
			AddRow(2, decimal.NewFromFloat(200.0))

		mock.ExpectQuery(`SELECT account_id, balance FROM accounts`).
			WillReturnRows(rows)

		accounts, err := repo.GetAll(context.Background())

		assert.NoError(t, err)
		assert.Len(t, accounts, 2)
		assert.Equal(t, int64(1), accounts[0].ID)
		assert.Equal(t, int64(2), accounts[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure: Query Error", func(t *testing.T) {
		db, mock, repo := setupTest(t)
		defer db.Close()

		mock.ExpectQuery(`SELECT account_id, balance FROM accounts`).
			WillReturnError(errors.New("syntax error"))

		accounts, err := repo.GetAll(context.Background())

		assert.ErrorContains(t, err, "syntax error")
		assert.Nil(t, accounts)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure: Rows Iteration Error", func(t *testing.T) {
		db, mock, repo := setupTest(t)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"account_id", "balance"}).
			AddRow(1, decimal.NewFromFloat(100.0)).
			RowError(0, errors.New("network packet loss"))

		mock.ExpectQuery(`SELECT account_id, balance FROM accounts`).
			WillReturnRows(rows)

		accounts, err := repo.GetAll(context.Background())

		assert.ErrorContains(t, err, "network packet loss")
		assert.Nil(t, accounts)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
