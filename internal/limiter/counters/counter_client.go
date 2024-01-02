package counters

import (
	"context"

	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"go.uber.org/zap"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

type CounterClient interface {
	// counter operations
	//// create or update a particualr counter, checking that no rules are violated
	IncrementCounters(logger *zap.Logger, requestContext context.Context, lockerClient lockerclient.LockerClient, counters *v1limiter.Counter) *errors.ServerError

	//// Decrement a prticular counter
	DecrementCounters(logger *zap.Logger, counters *v1limiter.Counter) *errors.ServerError

	//// Query all possible counters
	QueryCounters(logger *zap.Logger, query *v1common.AssociatedQuery) (v1limiter.Counters, *errors.ServerError)

	//// set a counter to a particular value, without ensuring any rules
	SetCounter(logger *zap.Logger, counters *v1limiter.Counter) *errors.ServerError
}
