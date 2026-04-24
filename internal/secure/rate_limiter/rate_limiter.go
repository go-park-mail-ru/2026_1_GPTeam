package rate_limiter

import (
	"context"
	"errors"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type RateLimiterInterface interface {
	IsIpBlocked(ctx context.Context, ip string) bool
	BlockIp(ctx context.Context, ip string)
	BlockIpPermanent(ctx context.Context, ip string)
	UnblockIp(ctx context.Context, ip string)
	Allow(ctx context.Context, ip string) bool
	IsTrustedIp(ip string) bool
	AllowN(ctx context.Context, ip string, n int) bool
}

type RateLimiter struct {
	mu         sync.RWMutex
	trustedIps []string
	bucket     repository.BucketInterface
}

func NewRateLimiter(bucket repository.BucketInterface, serverIp string) (*RateLimiter, error) {
	log := logger.GetLogger()
	if net.ParseIP(serverIp) == nil {
		log.Fatal("invalid server ip address", zap.String("serverIp", serverIp))
		return &RateLimiter{}, WrongServerIpAddress
	}
	return &RateLimiter{
		trustedIps: []string{
			"127.0.0.1",
			"::1",
			"10.0.0.0",
			"172.16.0.0",
			"192.168.0.0",
			serverIp,
		},
		bucket: bucket,
	}, nil
}

func (obj *RateLimiter) IsIpBlocked(ctx context.Context, ip string) bool {
	log := logger.GetLogger()
	permanentBlockedIps, err := obj.bucket.GetPermanentBlocked(ctx)
	if err != nil {
		log.Error("skip checking permanent blocked ips", zap.Error(err))
	} else {
		obj.mu.RLock()
		if slices.Contains(permanentBlockedIps.Ips, ip) {
			obj.mu.RUnlock()
			return true
		}
		obj.mu.RUnlock()
	}
	bucketInfo, err := obj.bucket.Get(ctx, ip)
	if err != nil {
		if errors.Is(err, repository.NoIpInSavedError) {
			newBucketInfo := models.BucketModel{
				Count:          MaxCount,
				LastRefillTime: time.Now(),
				BlockedUntil:   time.Time{},
				LastSeen:       time.Now(),
			}
			err = obj.bucket.Save(ctx, ip, newBucketInfo)
			if err != nil {
				log.Error("failed to save bucket",
					zap.String("ip", ip),
					zap.Any("bucket", newBucketInfo),
					zap.Error(err))
				return true
			}
			return false
		}
		return true
	}
	if bucketInfo.BlockedUntil.IsZero() {
		return false
	}
	now := time.Now()
	return bucketInfo.BlockedUntil.After(now)
}

func (obj *RateLimiter) BlockIp(ctx context.Context, ip string) {
	log := logger.GetLoggerWithRequestId(ctx)
	bucketInfo := models.BucketModel{
		Count:          MaxCount,
		LastRefillTime: time.Now(),
		BlockedUntil:   time.Now().Add(BlockDuration),
		LastSeen:       time.Now(),
	}
	err := obj.bucket.Save(ctx, ip, bucketInfo)
	if err != nil {
		log.Error("failed to save bucket",
			zap.String("ip", ip),
			zap.Any("bucket", bucketInfo),
			zap.Error(err))
		return
	}
	log.Info("blocked ip", zap.String("ip", ip))
}

func (obj *RateLimiter) BlockIpPermanent(ctx context.Context, ip string) {
	log := logger.GetLoggerWithRequestId(ctx)
	obj.mu.Lock()
	defer obj.mu.Unlock()
	permanentBlockedIps, err := obj.bucket.GetPermanentBlocked(ctx)
	if err != nil {
		return
	}
	permanentBlockedIps.Ips = append(permanentBlockedIps.Ips, ip)
	err = obj.bucket.SetPermanentBlocked(ctx, permanentBlockedIps)
	if err == nil {
		log.Warn("ip blocked permanent",
			zap.String("ip", ip))
	}
}

func (obj *RateLimiter) UnblockIp(ctx context.Context, ip string) {
	log := logger.GetLogger()
	obj.mu.Lock()
	defer obj.mu.Unlock()
	permanentBlockedIps, err := obj.bucket.GetPermanentBlocked(ctx)
	if err != nil {
		return
	}
	for i := 0; i < len(permanentBlockedIps.Ips); i++ {
		if permanentBlockedIps.Ips[i] == ip {
			permanentBlockedIps.Ips[i] = permanentBlockedIps.Ips[len(permanentBlockedIps.Ips)-1]
			permanentBlockedIps.Ips = permanentBlockedIps.Ips[:len(permanentBlockedIps.Ips)-1]
			i--
		}
	}
	err = obj.bucket.SetPermanentBlocked(ctx, permanentBlockedIps)
	if err == nil {
		log.Info("ip unblocked", zap.String("ip", ip))
	}
}

func (obj *RateLimiter) Allow(ctx context.Context, ip string) bool {
	return obj.AllowN(ctx, ip, 1)
}

func (obj *RateLimiter) AllowN(ctx context.Context, ip string, n int) bool {
	log := logger.GetLoggerWithRequestId(ctx)
	bucketInfo, err := obj.bucket.Get(ctx, ip)
	if err != nil {
		if errors.Is(err, repository.NoIpInSavedError) {
			newBucketInfo := models.BucketModel{
				Count:          MaxCount,
				LastRefillTime: time.Now(),
				BlockedUntil:   time.Time{},
				LastSeen:       time.Now(),
			}
			err = obj.bucket.Save(ctx, ip, newBucketInfo)
			if err != nil {
				log.Error("failed to save bucket",
					zap.String("ip", ip),
					zap.Any("bucket", newBucketInfo),
					zap.Error(err))
				return false
			}
			return true
		}
		return false
	}
	duration := int(time.Since(bucketInfo.LastRefillTime).Milliseconds() / 500)
	bucketInfo.Count = min(MaxCount, bucketInfo.Count+duration*RefillRateInHalfSecond)
	bucketInfo.LastRefillTime = time.Now()
	if bucketInfo.Count >= n {
		bucketInfo.Count -= n
		err = obj.bucket.Save(ctx, ip, bucketInfo)
		if err != nil {
			log.Error("failed to save bucket",
				zap.String("ip", ip),
				zap.Any("bucket", bucketInfo),
				zap.Error(err))
			return false
		}
		return true
	}
	return false
}

func (obj *RateLimiter) IsTrustedIp(ip string) bool {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return slices.Contains(obj.trustedIps, ip)
}

func GetRealIp(r *http.Request) (string, error) {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		clientIp := strings.TrimSpace(ips[0])
		if net.ParseIP(clientIp) != nil {
			return clientIp, nil
		}
	}
	realIp := r.Header.Get("X-Real-IP")
	if realIp != "" {
		if net.ParseIP(realIp) != nil {
			return realIp, nil
		}
	}
	realIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log := logger.GetLoggerWithRequestId(r.Context())
		log.Warn("unable to get ip", zap.Error(err))
		return "", UnableToGetIp
	}
	return realIp, nil
}
