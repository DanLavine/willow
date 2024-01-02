package memory

import (
	"sync/atomic"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// this is the minimum amoun of data needed for an override
type overrideMemory struct {
	limit *atomic.Int64
}

func New(req *v1limiter.Override) *overrideMemory {
	limit := &atomic.Int64{}
	limit.Add(req.Limit)

	return &overrideMemory{
		limit: limit,
	}
}

func (override *overrideMemory) Limit() int64 {
	return override.limit.Load()
}

func (override *overrideMemory) Update(overrideUpdate *v1limiter.OverrideUpdate) {
	override.limit.Store(overrideUpdate.Limit)
}

func (overrideMemory *overrideMemory) Delete() *errors.ServerError {
	return nil
}
