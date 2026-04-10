package rate_limiter

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

type BucketModel struct {
	Count          int
	LastRefillTime time.Time
	BlockedUntil   time.Time
	LastSeen       time.Time
}

type BucketInterface interface {
	Get(ip string) (BucketModel, error)
	Save(ip string, bucket BucketModel) error
}

type BucketRedis struct {
	db *redis.Pool
}

func NewBucketRedis(db *redis.Pool) *BucketRedis {
	return &BucketRedis{
		db: db,
	}
}

func (obj *BucketRedis) Get(ip string) (BucketModel, error) {
	log := logger.GetLogger()
	conn := obj.db.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Error("failed to close redis connection", zap.Error(err))
		}
	}()
	data, err := redis.Bytes(conn.Do("GET", ip))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return BucketModel{}, NoIpInSavedError
		}
		log.Error("redis error",
			zap.String("ip", ip),
			zap.Error(err))
		return BucketModel{}, err
	}
	bucketInfo := &BucketModel{}
	err = json.Unmarshal(data, bucketInfo)
	if err != nil {
		log.Error("unable to unmarshal bucket from redis",
			zap.String("ip", ip),
			zap.Error(err))
		return BucketModel{}, err
	}
	return *bucketInfo, nil
}

func (obj *BucketRedis) Save(ip string, bucket BucketModel) error {
	log := logger.GetLogger()
	serializedBucket, err := json.Marshal(bucket)
	if err != nil {
		log.Error("unable to serialize bucket",
			zap.String("ip", ip),
			zap.Any("bucket", bucket),
			zap.Error(err))
		return err
	}
	conn := obj.db.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Error("failed to close redis connection", zap.Error(err))
		}
	}()
	result, err := redis.String(conn.Do("SET", ip, serializedBucket, "EX", TTL))
	if err != nil {
		log.Error("error when saving bucket in redis",
			zap.String("ip", ip),
			zap.Any("bucket", bucket),
			zap.Error(err))
	}
	if result != "OK" {
		log.Error("error when saving bucket in redis",
			zap.String("ip", ip),
			zap.Any("bucket", bucket),
			zap.String("result", result))
		return ResultNotOkError
	}
	return nil
}
