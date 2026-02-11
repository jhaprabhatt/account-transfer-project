package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"account-transfer-project/internal/constants"
	"account-transfer-project/internal/core/handler/mocks"
	"account-transfer-project/internal/core/interceptors"
	"account-transfer-project/internal/models"
	pb "account-transfer-project/internal/proto"
)

func TestGrpcHandler_MakeTransfer(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Success: Transfer Executed", func(t *testing.T) {
		mockSvc := new(mocks.MockTransferService)

		h := NewGrpcHandler(nil, mockSvc, logger)

		ctx := context.WithValue(context.Background(), interceptors.CorrelationKey, int64(12345))

		req := &pb.TransferRequest{
			SourceId:      100,
			DestinationId: 200,
			Amount:        "50.00",
		}

		expectedResult := &models.TransferResult{
			AuditID:           999,
			CorrelationID:     12345,
			SourcePostBalance: "950.00",
		}

		mockSvc.On("MakeTransfer", ctx, mock.MatchedBy(func(r *models.TransferRequest) bool {
			return r.SourceID == 100 && r.DestinationID == 200 && r.Amount.Equal(decimal.NewFromInt(50))
		})).Return(expectedResult, nil)

		resp, err := h.MakeTransfer(ctx, req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, int64(999), resp.AuditId)
		assert.Equal(t, int64(12345), resp.TransactionId)
	})

	t.Run("Failure: Invalid Amount Format", func(t *testing.T) {
		mockSvc := new(mocks.MockTransferService)
		h := NewGrpcHandler(nil, mockSvc, logger)

		req := &pb.TransferRequest{Amount: "not-a-number"}

		resp, err := h.MakeTransfer(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "invalid amount format")
	})

	t.Run("Failure: Account Not Found (Translation Check)", func(t *testing.T) {
		mockSvc := new(mocks.MockTransferService)
		h := NewGrpcHandler(nil, mockSvc, logger)

		ctx := context.Background()
		req := &pb.TransferRequest{Amount: "50.00"}

		mockSvc.On("MakeTransfer", mock.Anything, mock.Anything).
			Return(nil, constants.ErrAccountNotFound)

		_, err := h.MakeTransfer(ctx, req)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Equal(t, "account not found", st.Message())
	})

	t.Run("Failure: Insufficient Funds (Translation Check)", func(t *testing.T) {
		mockSvc := new(mocks.MockTransferService)
		h := NewGrpcHandler(nil, mockSvc, logger)

		req := &pb.TransferRequest{Amount: "50.00"}

		mockSvc.On("MakeTransfer", mock.Anything, mock.Anything).
			Return(nil, constants.ErrInsufficientFunds)

		_, err := h.MakeTransfer(context.Background(), req)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
	})

	t.Run("Failure: System Error (Default Fallback)", func(t *testing.T) {
		mockSvc := new(mocks.MockTransferService)
		h := NewGrpcHandler(nil, mockSvc, logger)

		req := &pb.TransferRequest{Amount: "50.00"}

		mockSvc.On("MakeTransfer", mock.Anything, mock.Anything).
			Return(nil, errors.New("db connection died"))

		_, err := h.MakeTransfer(context.Background(), req)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestGrpcHandler_CreateAccount(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Success: Account Created", func(t *testing.T) {
		mockAccSvc := new(mocks.MockAccountService)

		h := NewGrpcHandler(mockAccSvc, nil, logger)

		req := &pb.CreateAccountRequest{
			AccountId: 101,
			Balance:   "500.00",
		}

		mockAccSvc.On("CreateAccount", mock.Anything, mock.MatchedBy(func(a *models.Account) bool {
			return a.ID == 101 && a.Balance.String() == "500"
		})).Return(nil)

		resp, err := h.CreateAccount(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("Failure: Invalid Balance", func(t *testing.T) {
		mockAccSvc := new(mocks.MockAccountService)
		h := NewGrpcHandler(mockAccSvc, nil, logger)

		req := &pb.CreateAccountRequest{
			AccountId: 101,
			Balance:   "invalid-decimal",
		}

		resp, err := h.CreateAccount(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("Failure: Account Already Exists", func(t *testing.T) {
		mockAccSvc := new(mocks.MockAccountService)
		h := NewGrpcHandler(mockAccSvc, nil, logger)

		req := &pb.CreateAccountRequest{AccountId: 101, Balance: "500.00"}

		mockAccSvc.On("CreateAccount", mock.Anything, mock.Anything).
			Return(constants.ErrAccountAlreadyExists)

		_, err := h.CreateAccount(context.Background(), req)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})
}
