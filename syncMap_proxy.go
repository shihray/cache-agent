package cache_proxy

import (
	"context"
	"sync"
)

func UseSyncMap[TCache, TQry any](baseProxy *BaseCacheProxy[TCache, TQry]) CacheProxy[TCache, TQry] {

	base := &SyncMapProxy[TCache, TQry]{
		BaseCacheProxy: BaseCacheProxy[TCache, TQry]{
			Transform: baseProxy.Transform,
			Cache:     baseProxy.Cache,
			GetDB:     baseProxy.GetDB,
		},
	}
	return base
}

type SyncMapProxy[TCache, TQry any] struct {
	BaseCacheProxy[TCache, TQry]
	// Map.LoadOrStore(key, value any) (actual any, loaded bool)
	shards sync.Map
}

func (proxy *SyncMapProxy[TCache, TQry]) Execute(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	return proxy.execute(ctx, qryOption, readModelType)
}

func (proxy *SyncMapProxy[TCache, TQry]) execute(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	key := proxy.Transform(qryOption)

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

	return proxy.BaseCacheProxy.Execute(ctx, qryOption, *readModelType)
}
