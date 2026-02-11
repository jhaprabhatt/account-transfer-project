package repository

import (
	"account-transfer-project/internal/constants"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"

	"account-transfer-project/internal/models"
)

type AccountRepository struct {
	db  *sql.DB
	log *zap.Logger
}

func NewAccountRepository(db *sql.DB, log *zap.Logger) *AccountRepository {
	return &AccountRepository{
		db:  db,
		log: log,
	}
}

func (r *AccountRepository) GetAccount(ctx context.Context, id int64) (*models.Account, error) {
	query := `SELECT account_id, balance FROM accounts WHERE account_id = $1`

	row := r.db.QueryRowContext(ctx, query, id)

	var acc models.Account
	err := row.Scan(&acc.ID, &acc.Balance)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constants.ErrAccountNotFound
		}
		r.log.Error("Failed to get account", zap.Int64("id", id), zap.Error(err))
		return nil, fmt.Errorf("get account failed: %w", err)
	}

	return &acc, nil
}

func (r *AccountRepository) CreateAccount(ctx context.Context, acc *models.Account) error {
	query := `INSERT INTO accounts (account_id, balance) VALUES ($1, $2)`

	_, err := r.db.ExecContext(ctx, query, acc.ID, acc.Balance)
	if err != nil {
		r.log.Error("Failed to create account",
			zap.Int64("account_id", acc.ID),
			zap.Error(err))
		return err
	}
	return nil
}

func (r *AccountRepository) GetAll(ctx context.Context) ([]models.Account, error) {
	query := `SELECT account_id, balance FROM accounts`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.log.Error("Failed to query all accounts", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var acc models.Account
		if err := rows.Scan(&acc.ID, &acc.Balance); err != nil {
			r.log.Error("Row scan failed", zap.Error(err))
			continue
		}
		accounts = append(accounts, acc)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Row iteration error", zap.Error(err))
		return nil, err
	}

	return accounts, nil
}
