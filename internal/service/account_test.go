package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"account-transfer-project/internal/models"
	"account-transfer-project/internal/service"
	"account-transfer-project/internal/service/mocks"
)

func TestAccountService_CreateAccount(t *testing.T) {
	acc := &models.Account{
		ID:      101,
		Balance: decimal.NewFromFloat(500.00),
	}

	tests := []struct {
		name          string
		mockBehaviour func(repo *mocks.MockAccountRepo, cache *mocks.MockCache)
		expectedError string
	}{
		{
			name: "Success: Account Created",
			mockBehaviour: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				cache.On("Exists", mock.Anything, acc.ID).Return(false, nil)
				repo.On("CreateAccount", mock.Anything, acc).Return(nil)
				cache.On("SetAccount", mock.Anything, acc).Return(nil)

			},
			expectedError: "",
		},
		{
			name: "Failure: Account Already Exists",
			mockBehaviour: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				cache.On("Exists", mock.Anything, acc.ID).Return(true, nil)
			},
			expectedError: "account already exists",
		},
		{
			name: "Failure: Cache Error (Exists Check)",
			mockBehaviour: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				cache.On("Exists", mock.Anything, acc.ID).Return(false, errors.New("redis connection refused"))
				repo.On("CreateAccount", mock.Anything, acc).Return(nil)
				cache.On("SetAccount", mock.Anything, acc).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "Failure: Database Error",
			mockBehaviour: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				cache.On("Exists", mock.Anything, acc.ID).Return(false, nil)
				repo.On("CreateAccount", mock.Anything, acc).Return(errors.New("db connection failed"))
			},
			expectedError: "db connection failed",
		},
		{
			name: "Success: Cache Write Failure (Should not fail request)",
			mockBehaviour: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				cache.On("Exists", mock.Anything, acc.ID).Return(false, nil)
				repo.On("CreateAccount", mock.Anything, acc).Return(nil)
				cache.On("SetAccount", mock.Anything, acc).Return(errors.New("redis timeout"))
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockAccountRepo)
			mockCache := new(mocks.MockCache)
			logger := zap.NewNop()

			tt.mockBehaviour(mockRepo, mockCache)

			svc := service.NewAccountService(mockRepo, mockCache, logger)

			err := svc.CreateAccount(context.Background(), acc)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})

	}
}

func TestAccountService_GetAccount(t *testing.T) {
	accountID := int64(101)
	expectedAcc := &models.Account{
		ID:      accountID,
		Balance: decimal.NewFromFloat(1000.00),
	}

	tests := []struct {
		name          string
		mockBehaviour func(repo *mocks.MockAccountRepo, cache *mocks.MockCache)
		expectedAcc   *models.Account
		expectedError string
	}{
		{
			name: "Success: Account Found (Trigger Read-Repair)",
			mockBehaviour: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				repo.On("GetAccount", mock.Anything, accountID).Return(expectedAcc, nil)
				cache.On("SetAccount", mock.Anything, expectedAcc).Return(nil)
			},
			expectedAcc:   expectedAcc,
			expectedError: "",
		},
		{
			name: "Failure: Account Not Found",
			mockBehaviour: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				repo.On("GetAccount", mock.Anything, accountID).Return(nil, errors.New("account not found"))
			},
			expectedAcc:   nil,
			expectedError: "account not found",
		},
		{
			name: "Failure: Database Error",
			mockBehaviour: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				repo.On("GetAccount", mock.Anything, accountID).Return(nil, errors.New("db connection error"))
			},
			expectedAcc:   nil,
			expectedError: "db connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockAccountRepo)
			mockCache := new(mocks.MockCache)
			logger := zap.NewNop()

			tt.mockBehaviour(mockRepo, mockCache)

			svc := service.NewAccountService(mockRepo, mockCache, logger)

			acc, err := svc.GetAccount(context.Background(), accountID)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, acc)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAcc, acc)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestAccountService_LoadAllAccountsToCache(t *testing.T) {
	acc1 := models.Account{ID: 1, Balance: decimal.NewFromFloat(100.0)}
	acc2 := models.Account{ID: 2, Balance: decimal.NewFromFloat(200.0)}
	accounts := []models.Account{acc1, acc2}

	tests := []struct {
		name          string
		mockBehavior  func(repo *mocks.MockAccountRepo, cache *mocks.MockCache)
		expectedError string
	}{
		{
			name: "Success: All Accounts Cached",
			mockBehavior: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				repo.On("GetAll", mock.Anything).Return(accounts, nil)
				cache.On("SetAccount", mock.Anything, &acc1).Return(nil)
				cache.On("SetAccount", mock.Anything, &acc2).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "Failure: Repository Error",
			mockBehavior: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				repo.On("GetAll", mock.Anything).Return(nil, errors.New("db disconnect"))
			},
			expectedError: "failed to fetch accounts for warmup",
		},
		{
			name: "Success: Partial Cache Failure (Resilience)",
			mockBehavior: func(repo *mocks.MockAccountRepo, cache *mocks.MockCache) {
				repo.On("GetAll", mock.Anything).Return(accounts, nil)
				cache.On("SetAccount", mock.Anything, &acc1).Return(errors.New("redis timeout"))
				cache.On("SetAccount", mock.Anything, &acc2).Return(nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockAccountRepo)
			mockCache := new(mocks.MockCache)
			logger := zap.NewNop()

			tt.mockBehavior(mockRepo, mockCache)

			svc := service.NewAccountService(mockRepo, mockCache, logger)

			err := svc.LoadAllAccountsToCache(context.Background())

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}
