package cache_proxy

import (
	"context"
	"sync"
)

func UseMutex[TCache, TQry any](baseProxy *BaseCacheProxy[TCache, TQry]) CacheProxy[TCache, TQry] {
	base := &SyncMapProxy[TCache, TQry]{
		BaseCacheProxy: BaseCacheProxy[TCache, TQry]{
			Transform: baseProxy.Transform,
			Cache:     baseProxy.Cache,
			GetDB:     baseProxy.GetDB,
		},
	}
	return base
}

type MutexProxy[TCache, TQry any] struct {
	BaseCacheProxy[TCache, TQry]

	mu sync.Mutex
}

func (proxy *MutexProxy[TCache, TQry]) Execute(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	return proxy.execute3(ctx, qryOption, readModelType)
}

func (proxy *MutexProxy[TCache, TQry]) execute1(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	return proxy.BaseCacheProxy.Execute(ctx, qryOption, *readModelType)
}

func (proxy *MutexProxy[TCache, TQry]) execute3(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	key := proxy.Transform(qryOption)
	err = proxy.Cache.GetValue(ctx, key, *readModelType)
	if err == nil {
		return *readModelType, nil
	}

	return proxy.execute1(ctx, qryOption, readModelType)
}
