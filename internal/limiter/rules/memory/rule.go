package memory

import (
	"sync"
	"sync/atomic"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"go.uber.org/zap"
)

type rule struct {
	lock  *sync.RWMutex
	limit *atomic.Int64
}

func New(createRequest *v1limiter.Rule) *rule {
	limit := new(atomic.Int64)
	limit.Add(createRequest.Limit)

	return &rule{
		lock:  new(sync.RWMutex),
		limit: limit,
	}
}

func (r *rule) Limit() int64 {
	return r.limit.Load()
}

func (r *rule) Update(logger *zap.Logger, updateRequest *v1limiter.RuleUpdateRquest) *errors.ServerError {
	r.limit.Store(updateRequest.Limit)
	return nil
}

func (r *rule) Delete() *errors.ServerError {
	// nothing to do for local deletion
	return nil
}
