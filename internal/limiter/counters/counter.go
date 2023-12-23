package counters

import "go.uber.org/atomic"

type Counter struct {
	Count *atomic.Int64
}

func New() *Counter {
	return &Counter{
		Count: new(atomic.Int64),
	}
}

func (c *Counter) Update(count int64) int64 {
	return c.Count.Add(count)
}

func (c *Counter) Set(count int64) {
	c.Count.Swap(count)
}

func (c *Counter) Load() int64 {
	return c.Count.Load()
}
