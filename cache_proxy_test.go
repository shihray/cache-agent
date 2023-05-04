package cache_proxy

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	pkgRedis "github.com/lctech-tw/face-tools-pkg/redis"
	"runtime"
	"sync"
	"testing"
)

func dependency[TCache, TQry any]() (*BaseCacheProxy[TCache, TQry], *UserDatabase) {
	dataSize := 2e3

	db := NewUserDatabase(int(dataSize))

	localCache, err := pkgRedis.NewFakeClient(0)
	if err != nil {
		panic(err)
	}

	baseProxy := &BaseCacheProxy[TCache, TQry]{
		Transform: TransformQryOptionToCacheKey(func(qryOption any) (key string) {
			return qryOption.(string)
		}),

		Cache: NewRedisCache[TCache](localCache),

		GetDB: DatabaseGetFunc[TCache, TQry](func(ctx context.Context, qryOption TQry) (result TCache, err error) {
			id := any(qryOption).(string)

			resp, err := db.QueryUserById(id)
			return any(*resp).(TCache), err
		}),
	}
	return baseProxy, db
}

func ConcurrentTester(goroutinePower uint8, fn func()) (start func(), wait func()) {
	var wg sync.WaitGroup
	ready := make(chan struct{})
	done := make(chan struct{})

	var workerCount int
	if goroutinePower == 0 {
		workerCount = 1 // sequential
	} else {
		workerCount = int(goroutinePower) * runtime.GOMAXPROCS(0)
	}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			<-ready
			fn()
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	start = func() { close(ready) }
	wait = func() { <-done }
	return start, wait
}

func CacheProxyBenchmarkConcurrentSingleKey(b *testing.B, proxy CacheProxy[gofakeit.PersonInfo, string], db *UserDatabase) {
	ids := db.GetUserIds()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id := ids[i%db.total]
		start, wait := ConcurrentTester(1, func() {
			proxy.Execute(context.Background(), id, &gofakeit.PersonInfo{})
		})
		start()
		wait()
	}

	b.Logf("single key: db qry count = %v, b.N=%v", db.qryCount, b.N)
}

func BenchmarkSyncMapProxy(b *testing.B) {

	baseProxy, db := dependency[gofakeit.PersonInfo, string]()

	CacheProxyBenchmarkConcurrentSingleKey(b, UseSyncMap[gofakeit.PersonInfo, string](baseProxy), db)
}

func BenchmarkSyncCondProxy(b *testing.B) {

	baseProxy, db := dependency[gofakeit.PersonInfo, string]()

	CacheProxyBenchmarkConcurrentSingleKey(b, UseSyncCond[gofakeit.PersonInfo, string](baseProxy), db)
}

func BenchmarkMutexProxy(b *testing.B) {

	baseProxy, db := dependency[gofakeit.PersonInfo, string]()

	CacheProxyBenchmarkConcurrentSingleKey(b, UseMutex[gofakeit.PersonInfo, string](baseProxy), db)
}
