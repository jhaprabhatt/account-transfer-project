package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"account-transfer-project/internal/api/handler/mocks"
	pb "account-transfer-project/internal/proto"
)

func TestTransactionHandler_MakeTransfer(t *testing.T) {

	t.Run("Success: Transfer Completed", func(t *testing.T) {
		mockClient := new(mocks.MockTransferServiceClient)
		logger := zap.NewNop()
		h := NewTransactionHandler(mockClient, logger)

		reqBody := `{"source_account_id": 100, "destination_account_id": 200, "amount": 50.00}`
		req, _ := http.NewRequest("POST", "/transfer", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		mockResponse := &pb.TransferResponse{
			Success:       true,
			TransactionId: 12345,
		}

		mockClient.On("MakeTransfer", mock.Anything, mock.MatchedBy(func(req *pb.TransferRequest) bool {
			return req.SourceId == 100 && req.DestinationId == 200 && req.Amount == "50"
		})).Return(mockResponse, nil)

		h.MakeTransfer(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "12345")
		mockClient.AssertExpectations(t)
	})

	t.Run("Failure: Invalid JSON", func(t *testing.T) {
		mockClient := new(mocks.MockTransferServiceClient)
		logger := zap.NewNop()
		h := NewTransactionHandler(mockClient, logger)

		reqBody := `{"source_account_id": 100, "amount":`
		req, _ := http.NewRequest("POST", "/transfer", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		h.MakeTransfer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid JSON")
	})

	t.Run("Failure: gRPC Service Error", func(t *testing.T) {
		mockClient := new(mocks.MockTransferServiceClient)
		logger := zap.NewNop()
		h := NewTransactionHandler(mockClient, logger)

		reqBody := `{"source_account_id": 100, "destination_account_id": 200, "amount": 50.00}`
		req, _ := http.NewRequest("POST", "/transfer", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		mockClient.On("MakeTransfer", mock.Anything, mock.Anything).
			Return(nil, errors.New("network timeout"))

		h.MakeTransfer(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal Server Error")

		mockClient.AssertExpectations(t)
	})

	t.Run("Failure: Validation Error (Negative Amount)", func(t *testing.T) {
		mockClient := new(mocks.MockTransferServiceClient)
		logger := zap.NewNop()
		h := NewTransactionHandler(mockClient, logger)

		reqBody := `{"source_account_id": 100, "destination_account_id": 200, "amount": -50.00}`
		req, _ := http.NewRequest("POST", "/transfer", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		h.MakeTransfer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "amount must be greater than zero")

		mockClient.AssertNotCalled(t, "MakeTransfer")
	})

	t.Run("Failure: Validation Error (Same Accounts)", func(t *testing.T) {
		mockClient := new(mocks.MockTransferServiceClient)
		logger := zap.NewNop()
		h := NewTransactionHandler(mockClient, logger)

		reqBody := `{"source_account_id": 100, "destination_account_id": 100, "amount": 50.00}`
		req, _ := http.NewRequest("POST", "/transfer", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		h.MakeTransfer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "source and destination account cannot be the same")

		mockClient.AssertNotCalled(t, "MakeTransfer")
	})
}
