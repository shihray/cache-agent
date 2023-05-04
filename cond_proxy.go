package cache_proxy

import (
	"context"
	"sync"
)

func UseSyncCond[TCache, TQry any](baseProxy *BaseCacheProxy[TCache, TQry]) CacheProxy[TCache, TQry] {

	proxy := &SyncCondProxy[TCache, TQry]{
		transform: baseProxy.Transform,
		cache:     baseProxy.Cache,

		baseProxy: baseProxy,
	}
	proxy.idle.L = &proxy.mu
	return proxy
}

type SyncCondProxy[TCache, TQry any] struct {
	transform TransformQryOptionToCacheKey
	cache     Cache[TCache]
	baseProxy *BaseCacheProxy[TCache, TQry]

	mu   sync.Mutex
	idle sync.Cond
	busy bool
}

func (proxy *SyncCondProxy[TCache, TQry]) Execute(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	return proxy.execute(ctx, qryOption, readModelType)
}

func (proxy *SyncCondProxy[TCache, TQry]) execute(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {

	for proxy.busy {
		return proxy.AwaitIdle(ctx, qryOption, readModelType)
	}

	proxy.SetBusy(true)

	readModel, err = proxy.baseProxy.Execute(ctx, qryOption, *readModelType)

	proxy.SetBusy(false)
	return
}

func (proxy *SyncCondProxy[TCache, TQry]) SetBusy(b bool) {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	wasBusy := proxy.busy
	proxy.busy = b
	if wasBusy && !proxy.busy {
		proxy.idle.Broadcast()
	}
}

func (proxy *SyncCondProxy[TCache, TQry]) AwaitIdle(ctx context.Context, qryOption TQry, readModelType *TCache) (readModel TCache, err error) {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	for proxy.busy {
		proxy.idle.Wait()
	}
	return proxy.baseProxy.Execute(ctx, qryOption, *readModelType)
}
