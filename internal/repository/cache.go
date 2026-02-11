package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jhaprabhatt/account-transfer-project/internal/config"
	"github.com/jhaprabhatt/account-transfer-project/internal/models"

	"github.com/redis/go-redis/v9"
)

type AccountCache struct {
	client *redis.Client
}

func NewAccountCache() *AccountCache {
	addr := config.GetEnv("REDIS_ADDR", "localhost:6379")
	password := config.GetEnv("REDIS_PASSWORD", "")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	return &AccountCache{client: rdb}
}

func (c *AccountCache) SetAccount(ctx context.Context, acc *models.Account) error {
	data, err := json.Marshal(acc)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	key := fmt.Sprintf("account:%d", acc.ID)
	if err := c.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

func (c *AccountCache) Exists(ctx context.Context, accountID int64) (bool, error) {
	key := fmt.Sprintf("account:%d", accountID)
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return count > 0, nil
}
