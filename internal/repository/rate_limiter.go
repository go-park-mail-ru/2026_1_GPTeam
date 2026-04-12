package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

const TTLOneDay = 86400
const PermanentBlockedIpsKey = "permanent_blocked_ips"
const RedisOkResponse = "OK"
const RedisMethodGet = "GET"
const RedisMethodSet = "SET"

type BucketInterface interface {
	Get(ctx context.Context, ip string) (models.BucketModel, error)
	Save(ctx context.Context, ip string, bucket models.BucketModel) error
	GetPermanentBlocked(ctx context.Context) (models.PermanentBlockedIps, error)
	SetPermanentBlocked(ctx context.Context, ips models.PermanentBlockedIps) error
}

type BucketRedis struct {
	db *redis.Pool
}

func NewBucketRedis(db *redis.Pool) *BucketRedis {
	return &BucketRedis{
		db: db,
	}
}

func (obj *BucketRedis) Get(ctx context.Context, ip string) (models.BucketModel, error) {
	log := logger.GetLogger()
	conn, err := obj.db.GetContext(ctx)
	if err != nil {
		log.Error("failed to get redis connection", zap.Error(err))
		return models.BucketModel{}, err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Error("failed to close redis connection", zap.Error(err))
		}
	}()
	data, err := redis.Bytes(redis.DoContext(conn, ctx, RedisMethodGet, ip))
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

func (obj *BucketRedis) Save(ctx context.Context, ip string, bucket models.BucketModel) error {
	log := logger.GetLogger()
	serializedBucket, err := json.Marshal(bucket)
	if err != nil {
		log.Error("unable to serialize bucket",
			zap.String("ip", ip),
			zap.Any("bucket", bucket),
			zap.Error(err))
		return err
	}
	conn, err := obj.db.GetContext(ctx)
	if err != nil {
		log.Error("failed to get redis connection", zap.Error(err))
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Error("failed to close redis connection", zap.Error(err))
		}
	}()
	result, err := redis.String(redis.DoContext(conn, ctx, RedisMethodSet, ip, serializedBucket, "EX", TTLOneDay))
	if err != nil {
		log.Error("error when saving bucket in redis",
			zap.String("ip", ip),
			zap.Any("bucket", bucket),
			zap.Error(err))
		return err
	}
	if result != RedisOkResponse {
		log.Error("error when saving bucket in redis",
			zap.String("ip", ip),
			zap.Any("bucket", bucket),
			zap.String("result", result))
		return ResultNotOkError
	}
	return nil
}

func (obj *BucketRedis) GetPermanentBlocked(ctx context.Context) (models.PermanentBlockedIps, error) {
	log := logger.GetLogger()
	conn, err := obj.db.GetContext(ctx)
	if err != nil {
		log.Error("failed to get redis connection", zap.Error(err))
		return models.PermanentBlockedIps{}, err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Error("failed to close redis connection", zap.Error(err))
		}
	}()
	data, err := redis.Bytes(redis.DoContext(conn, ctx, RedisMethodGet, PermanentBlockedIpsKey))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return models.PermanentBlockedIps{}, NoIpInSavedError
		}
		log.Error("redis error",
			zap.Error(err))
		return models.PermanentBlockedIps{}, err
	}
	blockedIps := &models.PermanentBlockedIps{}
	err = json.Unmarshal(data, blockedIps)
	if err != nil {
		log.Error("unable to unmarshal bucket from redis",
			zap.Error(err))
		return models.PermanentBlockedIps{}, err
	}
	return *blockedIps, nil
}

func (obj *BucketRedis) SetPermanentBlocked(ctx context.Context, ips models.PermanentBlockedIps) error {
	log := logger.GetLogger()
	serialized, err := json.Marshal(ips)
	if err != nil {
		log.Error("unable to serialize",
			zap.Any("ips", ips),
			zap.Error(err))
		return err
	}
	conn, err := obj.db.GetContext(ctx)
	if err != nil {
		log.Error("failed to get redis connection", zap.Error(err))
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Error("failed to close redis connection", zap.Error(err))
		}
	}()
	result, err := redis.String(redis.DoContext(conn, ctx, RedisMethodSet, PermanentBlockedIpsKey, serialized))
	if err != nil {
		log.Error("error when saving bucket in redis",
			zap.Any("ips", ips),
			zap.Error(err))
		return err
	}
	if result != RedisOkResponse {
		log.Error("error when saving bucket in redis",
			zap.Any("ips", ips),
			zap.String("result", result))
		return ResultNotOkError
	}
	return nil
}
