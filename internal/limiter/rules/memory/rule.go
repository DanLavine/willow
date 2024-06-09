package memory

import (
	"sync"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

type rule struct {
	lock       *sync.RWMutex
	properties *v1limiter.RuleProperties
}

func New(ruleProperties *v1limiter.RuleProperties) *rule {
	return &rule{
		lock:       new(sync.RWMutex),
		properties: ruleProperties,
	}
}

func (r *rule) Limit() int64 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return *r.properties.Limit
}

func (r *rule) Update(ruleProperties *v1limiter.RuleProperties) *errors.ServerError {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.properties = ruleProperties
	return nil
}

func (r *rule) Delete() *errors.ServerError {
	// nothing to do for local deletion
	return nil
}
