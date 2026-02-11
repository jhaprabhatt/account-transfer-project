package handler

import (
	"context"
	"errors"

	"github.com/jhaprabhatt/account-transfer-project/internal/constants"
	"github.com/jhaprabhatt/account-transfer-project/internal/core/interceptors"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jhaprabhatt/account-transfer-project/internal/models"
	pb "github.com/jhaprabhatt/account-transfer-project/internal/proto"
)

type TransferUseCase interface {
	MakeTransfer(ctx context.Context, req *models.TransferRequest) (*models.TransferResult, error)
}

type AccountUseCase interface {
	CreateAccount(ctx context.Context, acc *models.Account) error
}

type GrpcHandler struct {
	pb.UnimplementedAccountServiceServer
	pb.UnimplementedTransferServiceServer

	accountService  AccountUseCase
	transferService TransferUseCase

	log *zap.Logger
}

func NewGrpcHandler(accSvc AccountUseCase, txSvc TransferUseCase, log *zap.Logger) *GrpcHandler {
	return &GrpcHandler{
		accountService:  accSvc,
		transferService: txSvc,
		log:             log,
	}
}

func (h *GrpcHandler) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	balance, err := decimal.NewFromString(req.Balance)
	if err != nil {
		h.log.Error("Invalid balance format", zap.String("balance", req.Balance), zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "invalid balance format")
	}

	acc := &models.Account{
		ID:      req.AccountId,
		Balance: balance,
	}

	if err := h.accountService.CreateAccount(ctx, acc); err != nil {
		h.log.Error("Failed to create account", zap.Error(err))

		if errors.Is(err, constants.ErrAccountAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "account already exists")
		}

		if errors.Is(err, constants.ErrAmountMustNotBeNegative) {
			return nil, status.Error(codes.InvalidArgument, "")
		}

		return nil, status.Error(codes.Internal, "internal system error")
	}

	return &pb.CreateAccountResponse{Success: true}, nil
}

func (h *GrpcHandler) MakeTransfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	correlationID, _ := ctx.Value(interceptors.CorrelationKey).(int64)

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		h.log.Error("Invalid amount format",
			zap.String("amount", req.Amount),
			zap.Int64("correlation_id", correlationID))
		return nil, status.Error(codes.InvalidArgument, "invalid amount format")
	}

	modelReq := &models.TransferRequest{
		SourceID:      req.SourceId,
		DestinationID: req.DestinationId,
		Amount:        amount,
	}

	result, err := h.transferService.MakeTransfer(ctx, modelReq)
	if err != nil {
		h.log.Error("Transfer execution failed",
			zap.Error(err),
			zap.Int64("correlation_id", correlationID))
		switch {
		case errors.Is(err, constants.ErrAccountNotFound):
			return nil, status.Error(codes.NotFound, "account not found")

		case errors.Is(err, constants.ErrInsufficientFunds):
			return nil, status.Error(codes.FailedPrecondition, "insufficient funds")

		case errors.Is(err, constants.ErrSameAccount):
			return nil, status.Error(codes.InvalidArgument, "source and destination cannot be same")

		default:
			return nil, status.Error(codes.Internal, "internal system error")
		}
	}

	if result != nil {
		return &pb.TransferResponse{
			Success:          true,
			TransactionId:    result.CorrelationID,
			AuditId:          result.AuditID,
			NewSourceBalance: result.SourcePostBalance,
		}, nil
	}

	return nil, err
}
