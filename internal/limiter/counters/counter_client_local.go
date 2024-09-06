package counters

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DanLavine/channelops"
	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/internal/middleware"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
)

type counterClientLocal struct {
	// constructor for managing counters
	counterConstructor CounterConstructor

	// all possible tag groups and their counters
	counters btreeassociated.BTreeAssociated

	// client to interact with the rules and overrides
	rulesClient rules.RuleClient
}

func NewCountersClientLocal(constructor CounterConstructor, rulesClient rules.RuleClient) *counterClientLocal {
	return &counterClientLocal{
		counterConstructor: constructor,
		counters:           btreeassociated.NewThreadSafe(),
		rulesClient:        rulesClient,
	}
}

// List a all counters that match the query
func (cm *counterClientLocal) QueryCounters(ctx context.Context, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Counters, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "QueryCounters")

	countersResponse := v1limiter.Counters{}
	bTreeAssociatedOnIterate := func(item btreeassociated.AssociatedKeyValues) bool {
		countersResponse = append(countersResponse, &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: item.KeyValues(),
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf(item.Value().(Counter).Load()),
				},
			},
			State: &v1limiter.CounterState{
				Deleting: false,
			},
		})

		return true
	}

	if err := cm.counters.QueryAction(query, bTreeAssociatedOnIterate); err != nil {
		switch err {
		default:
			logger.Error("Failed to query counters", zap.Error(err))
			return countersResponse, errors.InternalServerError
		}
	}

	return countersResponse, nil
}

func (cm *counterClientLocal) IncrementCounters(ctx context.Context, lockerClient lockerclient.LockerClient, counter *v1limiter.Counter) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "IncrementCounters")

	// 1. Find the Rules /w Overrides limts that match the counter's key values
	rules, limitErrors := cm.rulesClient.FindLimits(ctx, counter.Spec.DBDefinition.KeyValues)
	if limitErrors != nil {
		return limitErrors
	}

	// 2, if there are no Rules, then just accept
	if len(rules) != 0 {
		// 3. if the last rule or override is 0, then just reject
		lastRule := rules[len(rules)-1]
		if len(lastRule.State.Overrides) == 0 {
			if *lastRule.Spec.Properties.Limit == 0 {
				return &errors.ServerError{Message: fmt.Sprintf("Limit has already been reached for rule '%s'", lastRule.State.ID), StatusCode: http.StatusConflict}
			}
		} else if *lastRule.State.Overrides[len(lastRule.State.Overrides)-1].Spec.Properties.Limit == 0 {
			return &errors.ServerError{Message: fmt.Sprintf("Limit has already been reached for rule '%s'", lastRule.State.ID), StatusCode: http.StatusConflict}
		}

		// 4. if all limits are unlimited, we can also continue without needing to check limits
		unlimited := true
		for _, rule := range rules {
			if len(rule.State.Overrides) == 0 {
				// check the rule
				if *rule.Spec.Properties.Limit != -1 {
					unlimited = false
				}
			} else {
				// enforce the overrides
				for _, override := range rule.State.Overrides {
					if *override.Spec.Properties.Limit != -1 {
						unlimited = false
						break
					}
				}

			}

			// could turn this into a goto
			if !unlimited {
				break
			}
		}

		if !unlimited {
			// 5. grab a lock for each key value to ensure that we are the only operation enforcing the rules for such values
			lockerLocks := []lockerclient.Lock{}
			defer func() {
				// release all locks when the function exits
				for _, lock := range lockerLocks {
					if err := lock.Release(ctx); err != nil {
						logger.Error("Failed to release lock", zap.Error(err))
					}
				}
			}()

			channelOps, chanReceiver := channelops.NewMergeRead[struct{}](true, ctx)
			for _, key := range counter.Spec.DBDefinition.KeyValues.SortedKeys() {
				// setup the group to lock
				lockKeyValues := &v1locker.Lock{
					Spec: &v1locker.LockSpec{
						DBDefinition: &v1locker.LockDBDefinition{
							KeyValues: datatypes.KeyValues{
								key: counter.Spec.DBDefinition.KeyValues[key],
							},
						},
						Properties: &v1locker.LockProperties{
							Timeout: helpers.PointerOf(time.Second),
						},
					},
				}

				// obtain the required lock
				lockerLock, err := lockerClient.ObtainLock(ctx, lockKeyValues, func(keyValue datatypes.KeyValues, err error) {
					logger.Error(err.Error())
				})

				if err != nil {
					logger.Error("failed to obtain a lock from the locker service", zap.Any("key values", lockKeyValues), zap.Error(err))
					return errors.InternalServerError
				}

				// setup monitor for when a lock is released
				lockerLocks = append(lockerLocks, lockerLock)
				if err := channelOps.MergeOrToOne(lockerLock.Done()); err != nil {
					// in this case, something has already been lost
					break
				}
			}

			// add a channel to manually kick. This give a chance for any lost locks to process properly
			successChan := make(chan struct{}, 1)
			successChan <- struct{}{}
			defer close(successChan)

			if err := channelOps.MergeOrToOne(successChan); err != nil {
				// lock is already lost so bail early
				logger.Error("a lock was released unexpedily")
				return errors.InternalServerError
			}

			// ensure that we didn't cancel obtaining any locks by triggering a select. there is small chance that a lock was lost,
			// but that is such a rare race condition I don't see it happening for real.
			_, ok := <-chanReceiver
			if !ok {
				// lost a lock or canceled obtaining the locks
				logger.Error("a lock was released unexpedily")
				return errors.InternalServerError
			}

			// 6. for each rule, count the possible counters that match and ensure that they are under the current limits
			for _, rule := range rules {
				// the limit is for unlimited, don't need to check this limit
				if len(rule.State.Overrides) == 0 {
					// rule enforcement
					// 1. unlimited so skip
					if *rule.Spec.Properties.Limit == -1 {
						continue
					}

					if counterErr := cm.checkCounters(ctx, rule.Spec.DBDefinition.GroupByKeyValues.Keys(), counter.Spec.DBDefinition.KeyValues, *rule.Spec.Properties.Limit); counterErr != nil {
						return counterErr
					}
				} else {
					for _, override := range rule.State.Overrides {
						// 1. unlimited so skip
						if *override.Spec.Properties.Limit == -1 {
							continue
						}

						// 2. check the limt
						if counterErr := cm.checkCounters(ctx, override.Spec.DBDefinition.GroupByKeyValues.Keys(), counter.Spec.DBDefinition.KeyValues, *override.Spec.Properties.Limit); counterErr != nil {
							return counterErr
						}
					}
				}
			}
		}
	}

	// 7. add the new counter
	createCounter := func() any {
		return cm.counterConstructor.New(counter.Spec.Properties)
	}

	incrementCounter := func(item btreeassociated.AssociatedKeyValues) {
		item.Value().(Counter).Update(counter.Spec.Properties)
	}

	if _, err := cm.counters.CreateOrFind(counter.Spec.DBDefinition.KeyValues, createCounter, incrementCounter); err != nil {
		logger.Error("Failed to find or update the counter", zap.Error(err))
		return errors.InternalServerError
	}

	return nil
}

func (cm *counterClientLocal) checkCounters(ctx context.Context, ruleKeys []string, counteKeyValyes datatypes.KeyValues, limit int64) *errors.ServerError {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "checkCounters")

	// construct the query for all possible rules that need to be found
	query := &queryassociatedaction.AssociatedActionQuery{
		Selection: &queryassociatedaction.Selection{
			KeyValues: queryassociatedaction.SelectionKeyValues{},
		},
	}

	for _, ruleKey := range ruleKeys {
		query.Selection.KeyValues[ruleKey] = queryassociatedaction.ValueQuery{
			Value:      counteKeyValyes[ruleKey],
			Comparison: v1common.Equals,
			TypeRestrictions: v1common.TypeRestrictions{
				MinDataType: counteKeyValyes[ruleKey].Type,
				MaxDataType: counteKeyValyes[ruleKey].Type,
			},
		}
	}

	counter := int64(0)
	bTreeAssociatedOnIterate := func(item btreeassociated.AssociatedKeyValues) bool {
		counter += item.Value().(Counter).Load()
		return counter < limit // check to exit query early if this fails
	}

	if err := cm.counters.QueryAction(query, bTreeAssociatedOnIterate); err != nil {
		logger.Error("Failed to query the current counters", zap.Error(err))
		return errors.InternalServerError
	}

	// final check of the counters
	if counter >= limit {
		logger.Info("Limit already reached")
		return &errors.ServerError{Message: fmt.Sprintf("Limit has already been reached for rule"), StatusCode: http.StatusConflict}
	}

	return nil
}

// Decrement removes a single instance from the key values group. If the total count would become 0, then the
// key values are removed entierly
//
// Decrement is muuch easier than increment because we don't need to ensure any rules validation. So no locks are required
// and we can just decrement the key values directly
func (cm *counterClientLocal) DecrementCounters(ctx context.Context, counter *v1limiter.Counter) *errors.ServerError {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "DecrementCounters")

	bTreeAssociatedCanDelete := func(item btreeassociated.AssociatedKeyValues) bool {
		return item.Value().(Counter).Update(counter.Spec.Properties) <= 0
	}

	if err := cm.counters.Delete(counter.Spec.DBDefinition.KeyValues, bTreeAssociatedCanDelete); err != nil {
		logger.Error("Failed to find or update the counter", zap.Error(err))
		return errors.InternalServerError
	}

	return nil
}

func (cm *counterClientLocal) SetCounter(ctx context.Context, counter *v1limiter.Counter) *errors.ServerError {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "SetCounter")

	if *counter.Spec.Properties.Counters <= 0 {
		// need to remove the key values
		bTreeAssociatedCanDelete := func(item btreeassociated.AssociatedKeyValues) bool {
			return true
		}

		if err := cm.counters.Delete(counter.Spec.DBDefinition.KeyValues, bTreeAssociatedCanDelete); err != nil {
			logger.Error("Failed to delete the set counters", zap.Error(err))
			return errors.InternalServerError
		}

		return nil
	} else {
		// need to create or set the key values
		bTreeAssociatedOnCreate := func() any {
			return cm.counterConstructor.New(counter.Spec.Properties)
		}

		bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) {
			item.Value().(Counter).Set(counter.Spec.Properties)
		}

		if _, err := cm.counters.CreateOrFind(counter.Spec.DBDefinition.KeyValues, bTreeAssociatedOnCreate, bTreeAssociatedOnFind); err != nil {
			logger.Error("Failed to find or update the set counter", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return nil
}
