package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jhaprabhatt/account-transfer-project/internal/models"
	"github.com/jhaprabhatt/account-transfer-project/internal/repository"

	"go.uber.org/zap"
)

type AccountService struct {
	accRepo repository.AccountRepo
	cache   repository.Cache
	log     *zap.Logger
}

func NewAccountService(
	accRepo repository.AccountRepo,
	cache repository.Cache,
	log *zap.Logger,
) *AccountService {
	return &AccountService{
		accRepo: accRepo,
		cache:   cache,
		log:     log,
	}
}

func (s *AccountService) LoadAllAccountsToCache(ctx context.Context) error {
	accounts, err := s.accRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch accounts for warmup: %w", err)
	}

	count := 0
	for _, acc := range accounts {
		if err := s.cache.SetAccount(ctx, &acc); err != nil {
			s.log.Error("Failed to cache account during warmup",
				zap.Int64("id", acc.ID), zap.Error(err))
		} else {
			count++
		}
	}
	s.log.Info("Cache Warm-up Complete", zap.Int("cached_count", count))
	return nil
}

func (s *AccountService) CreateAccount(ctx context.Context, acc *models.Account) error {
	if exists, err := s.cache.Exists(ctx, acc.ID); err == nil && exists {
		return errors.New("account already exists")
	}

	if err := s.accRepo.CreateAccount(ctx, acc); err != nil {
		s.log.Error("Failed to create account in DB", zap.Error(err))
		return err
	}

	if err := s.cache.SetAccount(ctx, acc); err != nil {
		s.log.Error("Cache write-through failed (data is safe in DB)",
			zap.Int64("account_id", acc.ID),
			zap.Error(err),
		)
	}

	s.log.Info("Account created successfully", zap.Int64("account_id", acc.ID))

	return nil
}

func (s *AccountService) GetAccount(ctx context.Context, id int64) (*models.Account, error) {
	acc, err := s.accRepo.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = s.cache.SetAccount(ctx, acc)

	return acc, nil
}
