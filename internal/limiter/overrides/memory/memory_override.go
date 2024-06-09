package memory

import (
	"sync"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// this is the minimum amoun of data needed for an override
type overrideMemory struct {
	lock       *sync.RWMutex
	properties *v1limiter.OverrideProperties
}

func New(overrideProperties *v1limiter.OverrideProperties) *overrideMemory {
	return &overrideMemory{
		lock:       new(sync.RWMutex),
		properties: overrideProperties,
	}
}

func (override *overrideMemory) Limit() int64 {
	override.lock.RLock()
	defer override.lock.RUnlock()

	return *override.properties.Limit
}

func (override *overrideMemory) Update(overrideProperties *v1limiter.OverrideProperties) {
	override.lock.Lock()
	defer override.lock.Unlock()

	override.properties = overrideProperties
}

func (overrideMemory *overrideMemory) Delete() *errors.ServerError {
	return nil
}
