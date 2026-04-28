package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock/v3"
	"github.com/stretchr/testify/require"
)

func newRateLimiterRedis(t *testing.T) (*BucketRedis, *redigomock.Conn) {
	t.Helper()
	mock := redigomock.NewConn()
	pool := &redis.Pool{
		MaxIdle:   3,
		MaxActive: 10,
		Dial: func() (redis.Conn, error) {
			return mock, nil
		},
	}
	t.Cleanup(func() {
		_ = pool.Close()
		_ = mock.Close()
	})
	return NewBucketRedis(pool), mock
}

func TestBucketRedis_Get(t *testing.T) {
	ip := "125.20.150.1"
	bucket := models.BucketModel{
		Count:          10,
		LastRefillTime: time.Now(),
		BlockedUntil:   time.Time{},
		LastSeen:       time.Now(),
	}
	jsonData, err := json.Marshal(bucket)
	require.NoError(t, err)
	existingBucket := string(jsonData)

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		repository, mock := newRateLimiterRedis(t)
		mock.Command("GET", ip).Expect(existingBucket)
		_, err = repository.Get(context.Background(), ip)
		require.NoError(t, err)
	})

	t.Run("no ip", func(t *testing.T) {
		t.Parallel()
		repository, mock := newRateLimiterRedis(t)

		mock.Command("GET", ip).ExpectError(redis.ErrNil)

		_, err := repository.Get(context.Background(), ip)
		require.ErrorIs(t, err, NoIpInSavedError)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBucketRedis_Save(t *testing.T) {
	t.Parallel()
}

func TestBucketRedis_GetPermanentBlocked(t *testing.T) {
	t.Parallel()
}

func TestBucketRedis_SetPermanentBlocked(t *testing.T) {
	t.Parallel()
}
