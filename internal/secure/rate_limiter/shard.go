package rate_limiter

import (
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

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
