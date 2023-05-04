package cache_proxy

import (
	"context"
	"sync"
)

func UseMutex[TCache, TQry any](baseProxy *BaseCacheProxy[TCache, TQry]) CacheProxy[TCache, TQry] {
	return &MutexProxy[TCache, TQry]{
		transform: baseProxy.Transform,
		cache:     baseProxy.Cache,
		baseProxy: baseProxy,
	}
}

type MutexProxy[TCache, TQry any] struct {
	transform TransformQryOptionToCacheKey
	cache     Cache[TCache]

	mu sync.Mutex

	baseProxy *BaseCacheProxy[TCache, TQry]
}

func (proxy *MutexProxy[TCache, TQry]) Execute(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	return proxy.execute3(ctx, qryOption, readModelType)
}

func (proxy *MutexProxy[TCache, TQry]) execute1(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	return proxy.baseProxy.Execute(ctx, qryOption, *readModelType)
}

func (proxy *MutexProxy[TCache, TQry]) execute3(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	key := proxy.transform(qryOption)
	err = proxy.cache.GetValue(ctx, key, *readModelType)
	if err == nil {
		return *readModelType, nil
	}

	return proxy.execute1(ctx, qryOption, readModelType)
}
