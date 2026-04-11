package repository

import (
	"encoding/json"
	"errors"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

const TTLOneDay = 86400

type BucketInterface interface {
	Get(ip string) (models.BucketModel, error)
	Save(ip string, bucket models.BucketModel) error
}

type BucketRedis struct {
	db *redis.Pool
}

func NewBucketRedis(db *redis.Pool) *BucketRedis {
	return &BucketRedis{
		db: db,
	}
}

func (obj *BucketRedis) Get(ip string) (models.BucketModel, error) {
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
			return models.BucketModel{}, NoIpInSavedError
		}
		log.Error("redis error",
			zap.String("ip", ip),
			zap.Error(err))
		return models.BucketModel{}, err
	}
	bucketInfo := &models.BucketModel{}
	err = json.Unmarshal(data, bucketInfo)
	if err != nil {
		log.Error("unable to unmarshal bucket from redis",
			zap.String("ip", ip),
			zap.Error(err))
		return models.BucketModel{}, err
	}
	return *bucketInfo, nil
}

func (obj *BucketRedis) Save(ip string, bucket models.BucketModel) error {
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
	result, err := redis.String(conn.Do("SET", ip, serializedBucket, "EX", TTLOneDay))
	if err != nil {
		log.Error("error when saving bucket in redis",
			zap.String("ip", ip),
			zap.Any("bucket", bucket),
			zap.Error(err))
		return err
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
