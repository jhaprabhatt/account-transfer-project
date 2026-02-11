package service

import (
	"account-transfer-project/internal/constants"
	"account-transfer-project/internal/models"
	"account-transfer-project/internal/repository"
	"context"
	"fmt"

	"go.uber.org/zap"
)

type TransferService struct {
	transferRepo repository.TransferRepo
	cache        repository.Cache
	log          *zap.Logger
}

func NewTransferService(
	transferRepo repository.TransferRepo,
	cache repository.Cache,
	log *zap.Logger,
) *TransferService {
	return &TransferService{
		transferRepo: transferRepo,
		cache:        cache,
		log:          log,
	}
}

func (s *TransferService) MakeTransfer(ctx context.Context, req *models.TransferRequest) (*models.TransferResult, error) {

	if err := s.ValidateTransfer(ctx, req.SourceID, req.DestinationID); err != nil {
		return nil, err
	}

	s.log.Info("Transfer Validated via Redis",
		zap.Int64("from", req.SourceID),
		zap.Int64("to", req.DestinationID))

	result, err := s.transferRepo.Transfer(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *TransferService) ValidateTransfer(ctx context.Context, sourceID, destID int64) error {
	if sourceID == destID {
		return constants.ErrSameAccount
	}
	srcExists, err := s.cache.Exists(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("failed to check source account cache: %w", err)
	}
	if !srcExists {
		return constants.ErrAccountNotFound
	}

	destExists, err := s.cache.Exists(ctx, destID)
	if err != nil {
		return fmt.Errorf("failed to check destination account cache: %w", err)
	}
	if !destExists {
		return constants.ErrAccountNotFound
	}

	return nil
}
