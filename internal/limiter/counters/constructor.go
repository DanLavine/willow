package counters

import (
	"fmt"

	"github.com/DanLavine/willow/internal/limiter/counters/memory"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

//go:generate mockgen -destination=rulefakes/counter_constructor_mock.go -package=rulefakes github.com/DanLavine/willow/internal/limiter/counters CounterConstructor
type Counter interface {
	Update(count int64) int64

	Set(count int64)

	Load() int64
}

type CounterConstructor interface {
	New(createRequest *v1limiter.Counter) Counter
}

func NewCountersConstructor(constructorType string) (CounterConstructor, error) {
	switch constructorType {
	case "memory":
		return &memoryConstrutor{}, nil
	default:
		return nil, fmt.Errorf("unkown constructor type")
	}
}

type memoryConstrutor struct{}

func (mc *memoryConstrutor) New(counter *v1limiter.Counter) Counter {
	return memory.New(counter)
}
