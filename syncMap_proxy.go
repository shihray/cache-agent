package cache_proxy

import (
	"context"
	"sync"
)

func UseSyncMap[TCache, TQry any](baseProxy *BaseCacheProxy[TCache, TQry]) CacheProxy[TCache, TQry] {
	return &SyncMapProxy[TCache, TQry]{
		transform: baseProxy.Transform,
		cache:     baseProxy.Cache,
		baseProxy: baseProxy,
	}
}

type SyncMapProxy[TCache, TQry any] struct {
	transform TransformQryOptionToCacheKey
	cache     Cache[TCache]

	// Map.LoadOrStore(key, value any) (actual any, loaded bool)
	shards sync.Map

	baseProxy *BaseCacheProxy[TCache, TQry]
}

func (proxy *SyncMapProxy[TCache, TQry]) Execute(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	return proxy.execute(ctx, qryOption, readModelType)
}

func (proxy *SyncMapProxy[TCache, TQry]) execute(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	key := proxy.transform(qryOption)

	var wg sync.WaitGroup
	wg.Add(1)
	getReadModelFunc := func() (TCache, error) {
		wg.Wait()
		return readModel, err
	}

	fn, isSecond := proxy.shards.LoadOrStore(key, getReadModelFunc)
	if isSecond {
		// 其他 goroutine 拿到的是, first goroutine 的 閉包 func
		return fn.(func() (TCache, error))()
	}

	defer func() {
		wg.Done()
		proxy.shards.Delete(key)
	}()

	return proxy.baseProxy.Execute(ctx, qryOption, *readModelType)
}
