package rate_limiter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type RateLimiterInterface interface {
	IsIpBlocked(ip string) bool
	BlockIp(ctx context.Context, ip string)
	BlockIpPermanent(ctx context.Context, ip string)
	UnblockIp(ip string)
	Allow(ctx context.Context, ip string) bool
	IsTrustedIp(ip string) bool
	AllowN(ctx context.Context, ip string, n int) bool
}

type RateLimiter struct {
	mu               sync.RWMutex
	once             sync.Once
	permanentBlocked []string
	trustedIps       []string
	bucket           BucketInterface
}

func NewRateLimiter(bucket BucketInterface, serverIp string) (*RateLimiter, error) {
	log := logger.GetLogger()
	if net.ParseIP(serverIp) == nil {
		log.Fatal("invalid server ip address", zap.String("serverIp", serverIp))
		return &RateLimiter{}, WrongServerIpAddress
	}
	return &RateLimiter{
		permanentBlocked: []string{},
		trustedIps: []string{
			"127.0.0.1",
			"::1",
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			serverIp,
		},
		bucket: bucket,
	}, nil
}

func (obj *RateLimiter) IsIpBlocked(ip string) bool {
	log := logger.GetLogger()
	obj.mu.RLock()
	if slices.Contains(obj.permanentBlocked, ip) {
		return true
	}
	obj.mu.RUnlock()
	bucketInfo, err := obj.bucket.Get(ip)
	if err != nil {
		if errors.Is(err, NoIpInSavedError) {
			newBucketInfo := BucketModel{
				Count:          MaxCount,
				LastRefillTime: time.Now(),
				BlockedUntil:   time.Time{},
				LastSeen:       time.Now(),
			}
			err = obj.bucket.Save(ip, newBucketInfo)
			if err != nil {
				log.Error("failed to save bucket", zap.String("ip", ip), zap.Any("bucket", newBucketInfo), zap.Error(err))
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
	log := logger.GetLoggerWIthRequestId(ctx)
	bucketInfo := BucketModel{
		Count:          MaxCount,
		LastRefillTime: time.Now(),
		BlockedUntil:   time.Now().Add(BlockDuration),
		LastSeen:       time.Now(),
	}
	err := obj.bucket.Save(ip, bucketInfo)
	if err != nil {
		log.Error("failed to save bucket", zap.String("ip", ip), zap.Any("bucket", bucketInfo), zap.Error(err))
		return
	}
	log.Info("blocked ip", zap.String("ip", ip))
}

func (obj *RateLimiter) BlockIpPermanent(ctx context.Context, ip string) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.permanentBlocked = append(obj.permanentBlocked, ip)
	log := logger.GetLoggerWIthRequestId(ctx)
	log.Warn("ip blocked permanent",
		zap.String("ip", ip))
}

func (obj *RateLimiter) UnblockIp(ip string) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	for i := 0; i < len(obj.permanentBlocked); i++ {
		if obj.permanentBlocked[i] == ip {
			obj.permanentBlocked[i] = obj.permanentBlocked[len(obj.permanentBlocked)-1]
			obj.permanentBlocked = obj.permanentBlocked[:len(obj.permanentBlocked)-1]
			i--
		}
	}
	log := logger.GetLogger()
	log.Info("ip unblocked", zap.String("ip", ip))
}

func (obj *RateLimiter) Allow(ctx context.Context, ip string) bool {
	return obj.AllowN(ctx, ip, 1)
}

func (obj *RateLimiter) AllowN(ctx context.Context, ip string, n int) bool {
	log := logger.GetLoggerWIthRequestId(ctx)
	bucketInfo, err := obj.bucket.Get(ip)
	if err != nil {
		if errors.Is(err, NoIpInSavedError) {
			newBucketInfo := BucketModel{
				Count:          MaxCount,
				LastRefillTime: time.Now(),
				BlockedUntil:   time.Time{},
				LastSeen:       time.Now(),
			}
			err = obj.bucket.Save(ip, newBucketInfo)
			if err != nil {
				log.Error("failed to save bucket", zap.String("ip", ip), zap.Any("bucket", newBucketInfo), zap.Error(err))
				return false
			}
			return true
		}
		return false
	}
	fmt.Println(bucketInfo.Count)
	duration := int(time.Since(bucketInfo.LastRefillTime).Milliseconds() / 500)
	bucketInfo.Count = min(MaxCount, bucketInfo.Count+duration*RefillRate)
	bucketInfo.LastRefillTime = time.Now()
	if bucketInfo.Count >= n {
		bucketInfo.Count -= n
		err = obj.bucket.Save(ip, bucketInfo)
		if err != nil {
			log.Error("failed to save bucket", zap.String("ip", ip), zap.Any("bucket", bucketInfo), zap.Error(err))
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
		log := logger.GetLoggerWIthRequestId(r.Context())
		log.Warn("unable to get ip", zap.Error(err))
		return "", UnableToGetIp
	}
	return realIp, nil
}
