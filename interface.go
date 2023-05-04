package cache_proxy

import (
	"context"
	"time"
)

type DatabaseGetFunc[TCache, TQry any] func(ctx context.Context, qryOption TQry) (readModel TCache, err error)

type TransformQryOptionToCacheKey func(qryOption any) (key string)

// Cache infrastructure interface
type Cache[TCache any] interface {
	GetValue(ctx context.Context, key string, value TCache) (err error)
	SetValue(ctx context.Context, key string, val TCache, keepTime time.Duration) error
}

// CacheProxy Proxy interface
type CacheProxy[TCache, TQry any] interface {
	Execute(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error)
}
