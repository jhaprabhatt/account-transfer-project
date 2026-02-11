package repository

import (
	"context"
	"github.com/jhaprabhatt/account-transfer-project/internal/models"
)

type AccountRepo interface {
	CreateAccount(ctx context.Context, acc *models.Account) error
	GetAll(ctx context.Context) ([]models.Account, error)
	GetAccount(ctx context.Context, id int64) (*models.Account, error)
}

type Cache interface {
	Exists(ctx context.Context, id int64) (bool, error)
	SetAccount(ctx context.Context, acc *models.Account) error
}

type TransferRepo interface {
	Transfer(ctx context.Context, req *models.TransferRequest) (*models.TransferResult, error)
}
