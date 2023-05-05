package cache_proxy

import (
	"context"
	"time"
)

type BaseCacheProxy[TCache, TQry any] struct {
	Transform TransformQryOptionToCacheKey
	Cache     Cache[TCache]
	GetDB     DatabaseGetFunc[TCache, TQry]
	KeepTime  time.Duration
}

func (proxy *BaseCacheProxy[TCache, TQry]) Execute(ctx context.Context, qryOption TQry, respModel TCache) (TCache, error) {
	key := proxy.Transform(qryOption)

	// cache.get
	err := proxy.Cache.GetValue(ctx, key, respModel)
	if err == nil {
		return respModel, nil
	}

	// db.get
	respModel, err = proxy.GetDB(ctx, qryOption)
	if err != nil {
		return respModel, err
	}

	if proxy.KeepTime == 0 {
		proxy.KeepTime = DefaultTimeout
	}

	// cache.set
	err = proxy.Cache.SetValue(ctx, key, respModel, proxy.KeepTime)
	if err != nil {
		return respModel, err
	}

	return respModel, nil
}
