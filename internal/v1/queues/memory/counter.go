package memory

import (
	"sync"
)

type Counter struct {
	lock  *sync.Mutex
	max   uint64
	total uint64
}

func NewCounter(max uint64) *Counter {
	return &Counter{
		lock:  new(sync.Mutex),
		max:   max,
		total: 0,
	}
}

func (c *Counter) Add() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.total < c.max {
		c.total++
		return true
	}

	return false
}

func (c *Counter) Decrement() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.total--
}

func (c *Counter) Total() uint64 {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.total
}
