package mocks

import (
	"account-transfer-project/internal/models"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockAccountRepo struct {
	mock.Mock
}

func (m *MockAccountRepo) CreateAccount(ctx context.Context, acc *models.Account) error {
	args := m.Called(ctx, acc)
	return args.Error(0)
}

func (m *MockAccountRepo) GetAll(ctx context.Context) ([]models.Account, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Account), args.Error(1)
}

func (m *MockAccountRepo) GetAccount(ctx context.Context, id int64) (*models.Account, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Account), args.Error(1)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Exists(ctx context.Context, id int64) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) SetAccount(ctx context.Context, acc *models.Account) error {
	args := m.Called(ctx, acc)
	return args.Error(0)
}
