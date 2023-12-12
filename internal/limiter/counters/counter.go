package counters

import "go.uber.org/atomic"

type Counter struct {
	Count *atomic.Uint64
}

func New() *Counter {
	return &Counter{
		Count: new(atomic.Uint64),
	}
}

func (c *Counter) Increment() uint64 {
	return c.Count.Add(1)
}

func (c *Counter) Decrement() uint64 {
	return c.Count.Add(^uint64(0))
}

func (c *Counter) Set(count uint64) {
	c.Count.Swap(count)
}

func (c *Counter) Load() uint64 {
	return c.Count.Load()
}
