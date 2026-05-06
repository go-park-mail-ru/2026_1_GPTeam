package rate_limiter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	mockrepo "github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newRL(t *testing.T) (*RateLimiter, *mockrepo.MockBucketInterface, func()) {
	t.Helper()
	ctrl := gomock.NewController(t)
	bucket := mockrepo.NewMockBucketInterface(ctrl)
	rl, err := NewRateLimiter(bucket, "8.8.8.8")
	require.NoError(t, err)
	return rl, bucket, ctrl.Finish
}

func TestNewRateLimiter(t *testing.T) {
	t.Run("ok with valid IP", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		bucket := mockrepo.NewMockBucketInterface(ctrl)

		rl, err := NewRateLimiter(bucket, "1.2.3.4")

		require.NoError(t, err)
		require.NotNil(t, rl)
		require.Contains(t, rl.trustedIps, "1.2.3.4")
		require.Contains(t, rl.trustedIps, "127.0.0.1")
	})
}

func TestRateLimiter_IsTrustedIp(t *testing.T) {
	rl, _, done := newRL(t)
	defer done()

	require.True(t, rl.IsTrustedIp("127.0.0.1"))
	require.True(t, rl.IsTrustedIp("8.8.8.8"))
	require.False(t, rl.IsTrustedIp("4.4.4.4"))
}

func TestRateLimiter_IsIpBlocked(t *testing.T) {
	ctx := context.Background()

	t.Run("permanent blocked", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{Ips: []string{"1.1.1.1"}}, nil)

		require.True(t, rl.IsIpBlocked(ctx, "1.1.1.1"))
	})

	t.Run("permanent fetch error then bucket missing -> save ok -> not blocked", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{}, errors.New("oops"))
		bucket.EXPECT().Get(gomock.Any(), "9.9.9.9").
			Return(models.BucketModel{}, repository.NoIpInSavedError)
		bucket.EXPECT().Save(gomock.Any(), "9.9.9.9", gomock.Any()).Return(nil)

		require.False(t, rl.IsIpBlocked(ctx, "9.9.9.9"))
	})

	t.Run("bucket missing and save fails -> blocked", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{}, nil)
		bucket.EXPECT().Get(gomock.Any(), "9.9.9.9").
			Return(models.BucketModel{}, repository.NoIpInSavedError)
		bucket.EXPECT().Save(gomock.Any(), "9.9.9.9", gomock.Any()).Return(errors.New("save err"))

		require.True(t, rl.IsIpBlocked(ctx, "9.9.9.9"))
	})

	t.Run("bucket Get returns generic error -> blocked", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{}, nil)
		bucket.EXPECT().Get(gomock.Any(), "9.9.9.9").
			Return(models.BucketModel{}, errors.New("boom"))

		require.True(t, rl.IsIpBlocked(ctx, "9.9.9.9"))
	})

	t.Run("bucket exists, BlockedUntil zero -> not blocked", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{}, nil)
		bucket.EXPECT().Get(gomock.Any(), "9.9.9.9").
			Return(models.BucketModel{Count: 5}, nil)

		require.False(t, rl.IsIpBlocked(ctx, "9.9.9.9"))
	})

	t.Run("bucket exists, BlockedUntil in the future -> blocked", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{}, nil)
		bucket.EXPECT().Get(gomock.Any(), "9.9.9.9").
			Return(models.BucketModel{BlockedUntil: time.Now().Add(time.Minute)}, nil)

		require.True(t, rl.IsIpBlocked(ctx, "9.9.9.9"))
	})

	t.Run("bucket exists, BlockedUntil in the past -> not blocked", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{}, nil)
		bucket.EXPECT().Get(gomock.Any(), "9.9.9.9").
			Return(models.BucketModel{BlockedUntil: time.Now().Add(-time.Minute)}, nil)

		require.False(t, rl.IsIpBlocked(ctx, "9.9.9.9"))
	})
}

func TestRateLimiter_BlockIp(t *testing.T) {
	ctx := context.Background()

	t.Run("save ok", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().Save(gomock.Any(), "9.9.9.9", gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, b models.BucketModel) error {
				require.Equal(t, MaxCount, b.Count)
				require.False(t, b.BlockedUntil.IsZero())
				return nil
			})

		rl.BlockIp(ctx, "9.9.9.9")
	})

	t.Run("save error", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().Save(gomock.Any(), "9.9.9.9", gomock.Any()).
			Return(errors.New("err"))

		rl.BlockIp(ctx, "9.9.9.9")
	})
}

func TestRateLimiter_BlockIpPermanent(t *testing.T) {
	ctx := context.Background()

	t.Run("get permanent error: nothing happens", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{}, errors.New("err"))

		rl.BlockIpPermanent(ctx, "9.9.9.9")
	})

	t.Run("set permanent error: still attempts", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{Ips: []string{"a"}}, nil)
		bucket.EXPECT().SetPermanentBlocked(gomock.Any(), gomock.Any()).
			Return(errors.New("err"))

		rl.BlockIpPermanent(ctx, "b")
	})

	t.Run("ok: ip appended", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{Ips: []string{"a"}}, nil)
		bucket.EXPECT().SetPermanentBlocked(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, ips models.PermanentBlockedIps) error {
				require.Contains(t, ips.Ips, "b")
				require.Contains(t, ips.Ips, "a")
				return nil
			})

		rl.BlockIpPermanent(ctx, "b")
	})
}

func TestRateLimiter_UnblockIp(t *testing.T) {
	ctx := context.Background()

	t.Run("get permanent error: nothing happens", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{}, errors.New("err"))

		rl.UnblockIp(ctx, "x")
	})

	t.Run("ip removed when present", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{Ips: []string{"a", "b", "c"}}, nil)
		bucket.EXPECT().SetPermanentBlocked(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, ips models.PermanentBlockedIps) error {
				require.NotContains(t, ips.Ips, "b")
				require.Contains(t, ips.Ips, "a")
				require.Contains(t, ips.Ips, "c")
				return nil
			})

		rl.UnblockIp(ctx, "b")
	})

	t.Run("ip absent: list unchanged, set still called", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().GetPermanentBlocked(gomock.Any()).
			Return(models.PermanentBlockedIps{Ips: []string{"a", "c"}}, nil)
		bucket.EXPECT().SetPermanentBlocked(gomock.Any(), gomock.Any()).Return(nil)

		rl.UnblockIp(ctx, "b")
	})
}

func TestRateLimiter_AllowN(t *testing.T) {
	ctx := context.Background()

	t.Run("new ip, save ok -> allowed", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().Get(gomock.Any(), "ip").
			Return(models.BucketModel{}, repository.NoIpInSavedError)
		bucket.EXPECT().Save(gomock.Any(), "ip", gomock.Any()).Return(nil)

		require.True(t, rl.AllowN(ctx, "ip", 1))
	})

	t.Run("new ip, save fails -> denied", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().Get(gomock.Any(), "ip").
			Return(models.BucketModel{}, repository.NoIpInSavedError)
		bucket.EXPECT().Save(gomock.Any(), "ip", gomock.Any()).
			Return(errors.New("err"))

		require.False(t, rl.AllowN(ctx, "ip", 1))
	})

	t.Run("get returns generic error -> denied", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().Get(gomock.Any(), "ip").
			Return(models.BucketModel{}, errors.New("boom"))

		require.False(t, rl.AllowN(ctx, "ip", 1))
	})

	t.Run("existing bucket with enough tokens -> allowed", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().Get(gomock.Any(), "ip").
			Return(models.BucketModel{Count: 10, LastRefillTime: time.Now()}, nil)
		bucket.EXPECT().Save(gomock.Any(), "ip", gomock.Any()).Return(nil)

		require.True(t, rl.AllowN(ctx, "ip", 5))
	})

	t.Run("existing bucket, save fails -> denied", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().Get(gomock.Any(), "ip").
			Return(models.BucketModel{Count: 10, LastRefillTime: time.Now()}, nil)
		bucket.EXPECT().Save(gomock.Any(), "ip", gomock.Any()).Return(errors.New("err"))

		require.False(t, rl.AllowN(ctx, "ip", 1))
	})

	t.Run("existing bucket without enough tokens -> denied", func(t *testing.T) {
		rl, bucket, done := newRL(t)
		defer done()

		bucket.EXPECT().Get(gomock.Any(), "ip").
			Return(models.BucketModel{Count: 0, LastRefillTime: time.Now()}, nil)

		require.False(t, rl.AllowN(ctx, "ip", 1))
	})
}

func TestRateLimiter_Allow(t *testing.T) {
	rl, bucket, done := newRL(t)
	defer done()

	bucket.EXPECT().Get(gomock.Any(), "ip").
		Return(models.BucketModel{Count: 10, LastRefillTime: time.Now()}, nil)
	bucket.EXPECT().Save(gomock.Any(), "ip", gomock.Any()).Return(nil)

	require.True(t, rl.Allow(context.Background(), "ip"))
}

func TestGetRealIp(t *testing.T) {
	t.Run("X-Forwarded-For valid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")

		ip, err := GetRealIp(req)
		require.NoError(t, err)
		require.Equal(t, "1.2.3.4", ip)
	})

	t.Run("X-Forwarded-For invalid -> falls back", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "not-an-ip")
		req.Header.Set("X-Real-IP", "9.9.9.9")

		ip, err := GetRealIp(req)
		require.NoError(t, err)
		require.Equal(t, "9.9.9.9", ip)
	})

	t.Run("X-Real-IP valid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "9.9.9.9")

		ip, err := GetRealIp(req)
		require.NoError(t, err)
		require.Equal(t, "9.9.9.9", ip)
	})

	t.Run("X-Real-IP invalid -> falls back to RemoteAddr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "garbage")
		req.RemoteAddr = "10.0.0.5:12345"

		ip, err := GetRealIp(req)
		require.NoError(t, err)
		require.Equal(t, "10.0.0.5", ip)
	})

	t.Run("no headers -> uses RemoteAddr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.5:12345"

		ip, err := GetRealIp(req)
		require.NoError(t, err)
		require.Equal(t, "10.0.0.5", ip)
	})

	t.Run("malformed RemoteAddr -> error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "no-port-here"

		ip, err := GetRealIp(req)
		require.ErrorIs(t, err, UnableToGetIp)
		require.Empty(t, ip)
	})
}
