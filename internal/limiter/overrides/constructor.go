package overrides

import (
	"fmt"

	"github.com/DanLavine/willow/internal/limiter/overrides/memory"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// NOTE: this is not yet used beause there is no manager, but it will be. The current though to associate relationships
// between data models is to provide a `_[key]` name. The keys that start with an `_` will need to be treated specially
// to know that those are "service made". I had tried it with custom types at one point, but that was aweful for the encodeing
// and decoding between the various clients

type Override interface {
	// get the limit for the override
	Limit() int64

	// update the saved override
	Update(OverrideUpdate *v1limiter.OverrideUpdate)

	// Delete the override
	Delete() *errors.ServerError
}

//go:generate mockgen -destination=rulefakes/constructor_mock.go -package=rulefakes github.com/DanLavine/willow/internal/limiter/rules RuleConstructor
type OverrideConstructor interface {
	New(override *v1limiter.Override) Override
}

func NewOverrideConstructor(constructorType string) (OverrideConstructor, error) {
	switch constructorType {
	case "memory":
		return &memoryConstrutor{}, nil
	default:
		return nil, fmt.Errorf("unkown constructor type")
	}
}

type memoryConstrutor struct{}

func (mc *memoryConstrutor) New(override *v1limiter.Override) Override {
	return memory.New(override)
}
