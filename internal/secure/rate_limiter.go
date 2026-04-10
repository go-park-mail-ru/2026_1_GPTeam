package secure

import (
	"errors"
	"fmt"
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

var NoIpInShardError = fmt.Errorf("no ip in shard")
var UnableToGetIp = fmt.Errorf("unable to get ip")
var WrongServerIpAddress = fmt.Errorf("wrong server ip")

const RefillRate = 1
const MaxCount = 100
const TTL = 24 * time.Hour
const BlockDuration = 15 * time.Minute
const CleanInterval = 24 * time.Hour

type BucketInterface interface {
	Allow() bool
	AllowN(n int) bool
	refill()
	GetCount() int
	Reset()
	IsBlocked() bool
	Block(duration time.Duration)
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
	log := logger.GetLogger()
	log.Info("trying to allow bucket",
		zap.Int("count", obj.count),
		zap.Int("maxCount", obj.maxCount),
		zap.Int("try", n))
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

func (obj *Bucket) GetCount() int {
	obj.refill()
	return obj.count
}

func (obj *Bucket) Reset() {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.count = obj.maxCount
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

func (obj *Bucket) Block(duration time.Duration) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.blockedUntil = time.Now().Add(duration)
	log := logger.GetLogger()
	log.Info("blocking until", zap.Time("time", obj.blockedUntil))
}

func (obj *Bucket) GetLastSeen() time.Time {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.lastSeen
}

type ShardInterface interface {
	GetBucket(ip string) (*Bucket, error)
	AddBucket(ip string)
	Clean(lastTime time.Time)
}

type Shard struct {
	mu      sync.RWMutex
	buckets map[string]*Bucket
}

func NewShard() *Shard {
	return &Shard{
		buckets: make(map[string]*Bucket),
	}
}

func (obj *Shard) GetBucket(ip string) (*Bucket, error) {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	bucket, exists := obj.buckets[ip]
	if !exists {
		return nil, NoIpInShardError
	}
	return bucket, nil
}

func (obj *Shard) AddBucket(ip string) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	newBucket := NewBucket()
	obj.buckets[ip] = newBucket
	log := logger.GetLogger()
	log.Info("add bucket for new ip",
		zap.String("ip", ip))
}

func (obj *Shard) Clean(lastTime time.Time) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	for ip, bucket := range obj.buckets {
		if !bucket.IsBlocked() {
			if bucket.GetLastSeen().Before(lastTime) {
				delete(obj.buckets, ip)
				log := logger.GetLogger()
				log.Info("delete bucket with ip",
					zap.String("ip", ip))
			}
		}
	}
}

type RateLimiterInterface interface {
	getShard(ip string) *Shard
	IsIpBlocked(ip string) bool
	BlockIp(ip string)
	BlockIpPermanent(ip string)
	UnblockIp(ip string)
	Allow(ip string) bool
	cleanupLoop()
	cleanup()
	Stop()
	GetBlacklistedIps() []string
	IsTrustedIp(ip string) bool
}

type RateLimiter struct {
	mu               sync.RWMutex
	once             sync.Once
	shards           []*Shard
	numShards        int
	permanentBlocked []string
	ttl              time.Duration
	stopCleanup      chan interface{}
	trustedIps       []string
}

func NewRateLimiter(numShards int, serverIp string) (*RateLimiter, error) {
	if net.ParseIP(serverIp) == nil {
		return &RateLimiter{}, WrongServerIpAddress
	}
	shards := make([]*Shard, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = NewShard()
	}
	log := logger.GetLogger()
	log.Info("rate limiter init", zap.Int("len shards", len(shards)))
	newRateLimiter := &RateLimiter{
		shards:           shards,
		numShards:        numShards,
		permanentBlocked: []string{},
		ttl:              TTL,
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

func (obj *RateLimiter) BlockIp(ip string) {
	shard := obj.getShard(ip)
	bucket, err := shard.GetBucket(ip)
	if err != nil {
		return
	}
	bucket.Block(BlockDuration)
}

func (obj *RateLimiter) BlockIpPermanent(ip string) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.permanentBlocked = append(obj.permanentBlocked, ip)
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
	log.Info("rate limiter cleanup")
	lastTime := time.Now().Add(-obj.ttl)
	for _, shard := range obj.shards {
		go shard.Clean(lastTime)
	}
}

func (obj *RateLimiter) Stop() {
	obj.once.Do(func() {
		close(obj.stopCleanup)
	})
}

func (obj *RateLimiter) GetBlacklistedIps() []string {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.permanentBlocked
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
		return "", UnableToGetIp
	}
	return realIp, nil
}
