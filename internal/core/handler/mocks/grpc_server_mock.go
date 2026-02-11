package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"account-transfer-project/internal/models"
)

type MockTransferService struct {
	mock.Mock
}

func (m *MockTransferService) MakeTransfer(ctx context.Context, req *models.TransferRequest) (*models.TransferResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransferResult), args.Error(1)
}

func (m *MockTransferService) ValidateTransfer(ctx context.Context, sourceID, destID int64) error {
	return m.Called(ctx, sourceID, destID).Error(0)
}

type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount(ctx context.Context, acc *models.Account) error {
	args := m.Called(ctx, acc)
	return args.Error(0)
}
