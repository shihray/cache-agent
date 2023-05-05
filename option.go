package cache_proxy

import (
	"math/rand"
	"sort"
	"time"
)

type Option func(*Options)

func KeepTime(t int64) Option {
	return func(o *Options) {
		o.keepTime = t
	}
}

func KeepTimeRange(t1, t2 int64) Option {
	return func(o *Options) {
		o.keepTimeRange = []int64{t1, t2}
	}
}

// Options setting
type Options struct {
	keepTime      int64
	keepTimeRange []int64
}

func (o *Options) GetKeepTime() time.Duration {

	if o.keepTime != 0 {
		return time.Duration(o.keepTime) * time.Second
	}

	sort.Slice(o.keepTimeRange, func(i, j int) bool {
		return i > j
	})

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Int63n(o.keepTimeRange[1] - o.keepTimeRange[0])

	return time.Duration(o.keepTimeRange[0]+i) * time.Second
}
