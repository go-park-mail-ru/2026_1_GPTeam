package rate_limiter

import (
	"context"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type BucketInterface interface {
	Allow() bool
	AllowN(n int) bool
	refill()
	IsBlocked() bool
	Block(ctx context.Context, duration time.Duration)
	GetLastSeen() time.Time
}

type Bucket struct {
	count          int
	maxCount       int
	refillRate     int
	lastRefillTime time.Time
	mu             sync.RWMutex
	blockedUntil   time.Time
	lastSeen       time.Time
}

func NewBucket() *Bucket {
	return &Bucket{
		count:          MaxCount,
		maxCount:       MaxCount,
		refillRate:     RefillRate,
		lastRefillTime: time.Now(),
		blockedUntil:   time.Time{},
		lastSeen:       time.Now(),
	}
}

func (obj *Bucket) Allow() bool {
	return obj.AllowN(1)
}

func (obj *Bucket) AllowN(n int) bool {
	obj.refill()
	obj.mu.Lock()
	defer obj.mu.Unlock()
	if obj.count >= n {
		obj.count -= n
		return true
	}
	return false
}

func (obj *Bucket) refill() {
	duration := int(time.Since(obj.lastRefillTime).Milliseconds() / 500)
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.count = min(obj.maxCount, obj.count+obj.refillRate*duration)
	obj.lastRefillTime = time.Now()
}

func (obj *Bucket) IsBlocked() bool {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	if obj.blockedUntil.IsZero() {
		return false
	}
	now := time.Now()
	return obj.blockedUntil.After(now)
}

func (obj *Bucket) Block(ctx context.Context, duration time.Duration) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.blockedUntil = time.Now().Add(duration)
	log := logger.GetLoggerWIthRequestId(ctx)
	log.Info("blocking until", zap.Time("time", obj.blockedUntil))
}

func (obj *Bucket) GetLastSeen() time.Time {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.lastSeen
}
