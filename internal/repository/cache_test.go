package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/jhaprabhatt/account-transfer-project/internal/models"
)

func TestAccountCache_SetAccount(t *testing.T) {
	acc := &models.Account{
		ID:      101,
		Balance: decimal.NewFromFloat(500.00),
	}
	expectedKey := fmt.Sprintf("account:%d", acc.ID)

	expectedJSON, _ := json.Marshal(acc)

	t.Run("Success: Account Cached", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		cache := &AccountCache{client: db}

		mock.ExpectSet(expectedKey, expectedJSON, 0).
			SetVal("OK")

		err := cache.SetAccount(context.Background(), acc)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure: Redis Connection Error", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		cache := &AccountCache{client: db}

		mock.ExpectSet(expectedKey, expectedJSON, 0).
			SetErr(errors.New("connection refused"))

		err := cache.SetAccount(context.Background(), acc)

		assert.ErrorContains(t, err, "failed to set key")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAccountCache_Exists(t *testing.T) {
	accountID := int64(101)
	expectedKey := fmt.Sprintf("account:%d", accountID)

	t.Run("Success: Account Exists (Returns True)", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		cache := &AccountCache{client: db}

		mock.ExpectExists(expectedKey).SetVal(1)

		exists, err := cache.Exists(context.Background(), accountID)

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success: Account Does Not Exist (Returns False)", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		cache := &AccountCache{client: db}

		mock.ExpectExists(expectedKey).SetVal(0)

		exists, err := cache.Exists(context.Background(), accountID)

		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure: Redis Error", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		cache := &AccountCache{client: db}

		mock.ExpectExists(expectedKey).
			SetErr(errors.New("redis timeout"))

		exists, err := cache.Exists(context.Background(), accountID)

		assert.ErrorContains(t, err, "failed to check existence")
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNewAccountCache(t *testing.T) {
	t.Setenv("REDIS_ADDR", "localhost:9999")
	t.Setenv("REDIS_PASSWORD", "secret_pass")
	cache := NewAccountCache()
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.client)
	opt := cache.client.Options()

	assert.Equal(t, "localhost:9999", opt.Addr)
	assert.Equal(t, "secret_pass", opt.Password)
	assert.Equal(t, 0, opt.DB)
}

func TestNewAccountCache_Defaults(t *testing.T) {
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("REDIS_PASSWORD", "")

	cache := NewAccountCache()
	opt := cache.client.Options()
	assert.Equal(t, "localhost:6379", opt.Addr)
	assert.Empty(t, opt.Password)
}
