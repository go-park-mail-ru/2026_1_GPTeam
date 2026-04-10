package rate_limiter

import (
	"context"
	"errors"
	"hash/crc32"
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
	Allow(ip string) bool
	IsTrustedIp(ip string) bool
}

type RateLimiter struct {
	mu               sync.RWMutex
	once             sync.Once
	shards           []*Shard
	numShards        int
	permanentBlocked []string
	stopCleanup      chan interface{}
	trustedIps       []string
}

func NewRateLimiter(numShards int, serverIp string) (*RateLimiter, error) {
	log := logger.GetLogger()
	if net.ParseIP(serverIp) == nil {
		log.Fatal("invalid server ip address", zap.String("serverIp", serverIp))
		return &RateLimiter{}, WrongServerIpAddress
	}
	shards := make([]*Shard, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = NewShard()
	}
	newRateLimiter := &RateLimiter{
		shards:           shards,
		numShards:        numShards,
		permanentBlocked: []string{},
		stopCleanup:      make(chan interface{}),
		trustedIps: []string{
			"127.0.0.1",
			"::1",
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			serverIp,
		},
	}
	go newRateLimiter.cleanupLoop()
	return newRateLimiter, nil
}

func (obj *RateLimiter) getShard(ip string) *Shard {
	key := int(crc32.ChecksumIEEE([]byte(ip)))
	index := key % obj.numShards
	return obj.shards[index]
}

func (obj *RateLimiter) IsIpBlocked(ip string) bool {
	obj.mu.RLock()
	if slices.Contains(obj.permanentBlocked, ip) {
		return true
	}
	obj.mu.RUnlock()
	shard := obj.getShard(ip)
	bucket, err := shard.GetBucket(ip)
	if err != nil {
		if errors.Is(err, NoIpInShardError) {
			shard.AddBucket(ip)
			return false
		}
		return true
	}
	return bucket.IsBlocked()
}

func (obj *RateLimiter) BlockIp(ctx context.Context, ip string) {
	shard := obj.getShard(ip)
	bucket, err := shard.GetBucket(ip)
	if err != nil {
		return
	}
	bucket.Block(ctx, BlockDuration)
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

func (obj *RateLimiter) Allow(ip string) bool {
	shard := obj.getShard(ip)
	bucket, err := shard.GetBucket(ip)
	if err != nil {
		if errors.Is(err, NoIpInShardError) {
			shard.AddBucket(ip)
			return true
		}
		return false
	}
	return bucket.Allow()
}

func (obj *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(CleanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			obj.cleanup()
		case <-obj.stopCleanup:
			return
		}
	}
}

func (obj *RateLimiter) cleanup() {
	log := logger.GetLogger()
	log.Info("rate limiter cleanup started")
	lastTime := time.Now().Add(-TTL)
	for _, shard := range obj.shards {
		go shard.Clean(lastTime)
	}
}

func (obj *RateLimiter) Stop() {
	obj.once.Do(func() {
		close(obj.stopCleanup)
	})
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
