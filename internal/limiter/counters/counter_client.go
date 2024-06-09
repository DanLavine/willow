package counters

import (
	"context"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"

	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

type CounterClient interface {
	// counter operations
	//// create or update a particualr counter, checking that no rules are violated
	IncrementCounters(ctx context.Context, lockerClient lockerclient.LockerClient, counters *v1limiter.Counter) *errors.ServerError

	//// Decrement a prticular counter
	DecrementCounters(ctx context.Context, counters *v1limiter.Counter) *errors.ServerError

	//// Query all possible counters
	QueryCounters(ctx context.Context, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Counters, *errors.ServerError)

	//// set a counter to a particular value, without ensuring any rules
	SetCounter(ctx context.Context, counters *v1limiter.Counter) *errors.ServerError
}
