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
	testBucket := models.BucketModel{
		Count:          10,
		LastRefillTime: time.Now(),
		BlockedUntil:   time.Time{},
		LastSeen:       time.Now(),
	}
	jsonData, err := json.Marshal(testBucket)
	require.NoError(t, err)
	existingBucket := string(jsonData)

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		repository, mock := newRateLimiterRedis(t)
		mock.Command("GET", ip).Expect(existingBucket)
		bucket, err := repository.Get(context.Background(), ip)
		require.NoError(t, err)
		require.Equal(t, bucket.Count, testBucket.Count)
		require.True(t, bucket.LastRefillTime.Equal(testBucket.LastRefillTime))
		require.True(t, bucket.BlockedUntil.Equal(testBucket.BlockedUntil))
		require.True(t, bucket.LastSeen.Equal(testBucket.LastSeen))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no ip", func(t *testing.T) {
		t.Parallel()
		repository, mock := newRateLimiterRedis(t)

		mock.Command("GET", ip).ExpectError(redis.ErrNil)

		_, err = repository.Get(context.Background(), ip)
		require.ErrorIs(t, err, NoIpInSavedError)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBucketRedis_Save(t *testing.T) {
	ip := "125.20.150.1"
	testBucket := models.BucketModel{
		Count:          10,
		LastRefillTime: time.Now(),
		BlockedUntil:   time.Time{},
		LastSeen:       time.Now(),
	}
	serializedBucket, err := json.Marshal(testBucket)
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		repository, mock := newRateLimiterRedis(t)
		mock.Command("SET", ip, serializedBucket, "EX", TTLOneDay).Expect("OK")
		err = repository.Save(context.Background(), ip, testBucket)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBucketRedis_GetPermanentBlocked(t *testing.T) {
	test := models.PermanentBlockedIps{
		Ips: []string{
			"125.20.150.1",
			"125.20.150.2",
		},
	}
	jsonData, err := json.Marshal(test)
	require.NoError(t, err)
	existing := string(jsonData)

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		repository, mock := newRateLimiterRedis(t)
		mock.Command("GET", PermanentBlockedIpsKey).Expect(existing)
		permanentBlocked, err := repository.GetPermanentBlocked(context.Background())
		require.NoError(t, err)
		require.Equal(t, permanentBlocked, test)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no key", func(t *testing.T) {
		t.Parallel()
		repository, mock := newRateLimiterRedis(t)
		mock.Command("GET", PermanentBlockedIpsKey).ExpectError(redis.ErrNil)
		permanentBlocked, err := repository.GetPermanentBlocked(context.Background())
		require.ErrorIs(t, err, NoIpInSavedError)
		require.Equal(t, permanentBlocked, models.PermanentBlockedIps{})
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBucketRedis_SetPermanentBlocked(t *testing.T) {
	test := models.PermanentBlockedIps{
		Ips: []string{
			"125.20.150.1",
			"125.20.150.2",
		},
	}
	serialized, err := json.Marshal(test)
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		repository, mock := newRateLimiterRedis(t)
		mock.Command("SET", PermanentBlockedIpsKey, serialized).Expect("OK")
		err = repository.SetPermanentBlocked(context.Background(), test)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
