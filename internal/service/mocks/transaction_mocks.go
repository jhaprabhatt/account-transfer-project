package mocks

import (
	"account-transfer-project/internal/models"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockTransactionRepo struct {
	mock.Mock
}

func (m *MockTransactionRepo) Transfer(ctx context.Context, req *models.TransferRequest) (*models.TransferResult, error) {
	args := m.Called(ctx, req)
	var res *models.TransferResult
	if args.Get(0) != nil {
		res = args.Get(0).(*models.TransferResult)
	}

	return res, args.Error(1)
}
