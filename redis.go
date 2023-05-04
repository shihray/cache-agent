package cache_proxy

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"time"
)

type RedisCache[TCache any] struct {
	redis *redis.Client
}

func NewRedisCache[TCache any](client *redis.Client) *RedisCache[TCache] {
	return &RedisCache[TCache]{redis: client}
}

func (m *RedisCache[TCache]) GetValue(ctx context.Context, key string, value TCache) error {
	if ctx == nil {
		ctx = context.Background()
	}

	resp, err := m.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(resp), &value); err != nil {
		return err
	}

	return nil
}

func (m *RedisCache[TCache]) SetValue(ctx context.Context, key string, value TCache, keepTime time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}

	insertData := ""
	switch any(value).(type) {
	case string:
		insertData = any(value).(string)
	case []byte:
		insertData = string(any(value).([]byte))
	default:
		jsonVal, err := json.Marshal(value)
		if err != nil {
			log.Errorf("failed to marshal Cache:%s err:%v", key, err.Error())
			return err
		}
		insertData = string(jsonVal[:])
	}

	return m.redis.Set(ctx, key, insertData, keepTime).Err()
}
