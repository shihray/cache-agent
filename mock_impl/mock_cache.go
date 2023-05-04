package mock_impl

import (
	"context"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func NewFakeClient(db int) (cache *redis.Client, err error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, err
	}
	cache = redis.NewClient(
		&redis.Options{
			Addr: mr.Addr(),
			DB:   db,
		},
	)
	err = cache.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}
	return cache, nil
}
