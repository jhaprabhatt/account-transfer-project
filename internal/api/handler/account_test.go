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

func TestAccountHandler_CreateAccount(t *testing.T) {
	t.Run("Success: Account Created", func(t *testing.T) {
		mockClient := new(mocks.MockAccountServiceClient)
		logger := zap.NewNop()
		h := NewAccountHandler(mockClient, logger)
		reqBody := `{"account_id": 101, "balance": 500.00}`
		req, _ := http.NewRequest("POST", "/accounts", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		expectedGrpcReq := &pb.CreateAccountRequest{AccountId: 101, Balance: "500"}
		mockClient.On("CreateAccount", mock.Anything, expectedGrpcReq).
			Return(&pb.CreateAccountResponse{Success: true}, nil)

		h.CreateAccount(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Contains(t, rr.Body.String(), `"success":true`)

		mockClient.AssertExpectations(t)
	})

	t.Run("Failure: Invalid JSON", func(t *testing.T) {
		mockClient := new(mocks.MockAccountServiceClient)
		logger := zap.NewNop()
		h := NewAccountHandler(mockClient, logger)
		reqBody := `{"account_id": 101, "balance":`
		req, _ := http.NewRequest("POST", "/accounts", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		h.CreateAccount(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid JSON")
	})

	t.Run("Failure: gRPC Service Error", func(t *testing.T) {
		mockClient := new(mocks.MockAccountServiceClient)
		logger := zap.NewNop()
		h := NewAccountHandler(mockClient, logger)
		reqBody := `{"account_id": 101, "balance": 500.00}`
		req, _ := http.NewRequest("POST", "/accounts", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		mockClient.On("CreateAccount", mock.Anything, mock.Anything).
			Return(nil, errors.New("connection refused"))

		h.CreateAccount(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
