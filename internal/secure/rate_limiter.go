package secure

import (
	"fmt"
	"sync"
	"time"
)

var LimitExceededError = fmt.Errorf("limit exceeded")
var WrongMaxCountError = fmt.Errorf("wrong max count")

type BucketInterface interface {
	Allow() bool
	AllowN() bool
	refill()
	GetTokens()
	Reset()
}

type Bucket struct {
	count          int
	maxCount       int
	refillRate     int
	lastRefillTime time.Time
	mu             sync.RWMutex
}

func NewBucket(refillRate int, maxCount int) (*Bucket, error) {
	if maxCount <= 0 {
		return &Bucket{}, WrongMaxCountError
	}
	return &Bucket{
		count:          0,
		maxCount:       maxCount,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}, nil
}

func (obj *Bucket) Allow() bool {
	//TODO implement me
	panic("implement me")
}

func (obj *Bucket) AllowN() bool {
	//TODO implement me
	panic("implement me")
}

func (obj *Bucket) refill() {
	//TODO implement me
	panic("implement me")
}

func (obj *Bucket) GetTokens() {
	//TODO implement me
	panic("implement me")
}

func (obj *Bucket) Reset() {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.count = obj.maxCount
	obj.lastRefillTime = time.Now()
}
