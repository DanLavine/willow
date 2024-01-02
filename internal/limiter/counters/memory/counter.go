package memory

import (
	"sync/atomic"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

type Counter struct {
	counter *atomic.Int64
}

func New(createRequest *v1limiter.Counter) *Counter {
	counter := new(atomic.Int64)
	counter.Store(createRequest.Counters)

	return &Counter{
		counter: counter,
	}
}

func (c *Counter) Update(count int64) int64 {
	return c.counter.Add(count)
}

func (c *Counter) Set(count int64) {
	c.counter.Swap(count)
}

func (c *Counter) Load() int64 {
	return c.counter.Load()
}
