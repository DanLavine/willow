package memory

import (
	"sync"

	"github.com/DanLavine/willow/internal/helpers"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

type Counter struct {
	lock       *sync.RWMutex
	properties *v1limiter.CounteProperties
}

func New(counterProperties *v1limiter.CounteProperties) *Counter {
	return &Counter{
		lock: new(sync.RWMutex),
		properties: &v1limiter.CounteProperties{
			// NOTE: we want to perform a deep copy of the values.
			Counters: helpers.PointerOf[int64](*counterProperties.Counters),
		},
	}
}

func (c *Counter) Update(counterProperties *v1limiter.CounteProperties) int64 {
	c.lock.Lock()
	defer c.lock.Unlock()

	*c.properties.Counters += *counterProperties.Counters
	return *c.properties.Counters
}

func (c *Counter) Set(counterProperties *v1limiter.CounteProperties) {
	c.lock.Lock()
	defer c.lock.Unlock()

	*c.properties.Counters = *counterProperties.Counters
}

func (c *Counter) Load() int64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return *c.properties.Counters
}
