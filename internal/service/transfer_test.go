package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/jhaprabhatt/account-transfer-project/internal/models"
	"github.com/jhaprabhatt/account-transfer-project/internal/service"
	"github.com/jhaprabhatt/account-transfer-project/internal/service/mocks"
)

func TestTransferService_MakeTransfer(t *testing.T) {
	req := &models.TransferRequest{
		SourceID:      1,
		DestinationID: 2,
		Amount:        decimal.NewFromFloat(100.0),
	}

	tests := []struct {
		name          string
		mockBehavior  func(repo *mocks.MockTransactionRepo, cache *mocks.MockCache)
		expectedError string
	}{
		{
			name: "Success",
			mockBehavior: func(repo *mocks.MockTransactionRepo, cache *mocks.MockCache) {
				cache.On("Exists", mock.Anything, int64(1)).Return(true, nil)
				cache.On("Exists", mock.Anything, int64(2)).Return(true, nil)
				res := &models.TransferResult{Status: "SUCCESS", CorrelationID: 12345}
				repo.On("Transfer", mock.Anything, req).Return(res, nil)
			},
			expectedError: "",
		},
		{
			name: "Failure: DB Error",
			mockBehavior: func(repo *mocks.MockTransactionRepo, cache *mocks.MockCache) {
				cache.On("Exists", mock.Anything, int64(1)).Return(true, nil)
				cache.On("Exists", mock.Anything, int64(2)).Return(true, nil)
				repo.On("Transfer", mock.Anything, req).Return(nil, errors.New("db error"))
			},
			expectedError: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, mockCache, svc := newTestSetup(t)
			tt.mockBehavior(mockRepo, mockCache)
			_, err := svc.MakeTransfer(context.Background(), req)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}

	t.Run("Failure: Repo Transfer Fails", func(t *testing.T) {
		repo, cache, svc := newTestSetup(t)

		cache.On("Exists", mock.Anything, int64(1)).Return(true, nil)
		cache.On("Exists", mock.Anything, int64(2)).Return(true, nil)

		repo.On("Transfer", mock.Anything, req).Return(nil, errors.New("db connection lost"))

		_, err := svc.MakeTransfer(context.Background(), req)

		assert.ErrorContains(t, err, "db connection lost")
	})
}

func newTestSetup(t *testing.T) (*mocks.MockTransactionRepo, *mocks.MockCache, *service.TransferService) {
	mockRepo := new(mocks.MockTransactionRepo)
	mockCache := new(mocks.MockCache)
	logger := zap.NewNop()
	svc := service.NewTransferService(mockRepo, mockCache, logger)
	return mockRepo, mockCache, svc
}
