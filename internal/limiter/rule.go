package limiter

import (
	"github.com/DanLavine/willow/pkg/models/api"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

type Rule interface {
	// Update a particualr rule's default limit values
	Update(logger *zap.Logger, update *v1limiter.RuleUpdate)

	// set an override for a particualr group of tags
	SetOverride(logger *zap.Logger, override *v1limiter.Override) *api.Error

	// delete an override for a particualr group of tags
	DeleteOverride(logger *zap.Logger, query datatypes.AssociatedKeyValuesQuery) *api.Error

	// Find the limit for a particualr group of tags when checking the limits
	FindLimit(logger *zap.Logger, keyValues datatypes.KeyValues) uint64

	// check if any tags coming in match the tag group
	TagsMatch(logger *zap.Logger, keyValues datatypes.KeyValues) bool

	// generate a query based on the search tags
	//GenerateQuery(keyValues datatypes.KeyValues) datatypes.Select

	// operations for callbacks when resources are freed
	// TODO: These operations are very consuming and hard to implement. So do that after a basic limiter is
	// setup and useful
	//
	//AddClientWaiting(logger *zap.Logger, keyValues datatypes.KeyValues) <-chan struct{} // client calls when trying to "increment", but a limit is blocked, sets up a callback that triggers another "increment" call
	//TriggerClientWaiting(keyValues datatypes.KeyValues)                                 // decrement calls these whenever that happens. Can als be called on an override increase change

	// locking operation from the manager when checking limits
	Lock()
	Unlock()

	// Need to generate a query on what to search for
	GenerateQuery(keyValues datatypes.KeyValues) datatypes.AssociatedKeyValuesQuery

	// Get a rule response for Read operations
	GetRuleResponse(includeOverrides bool) *v1limiter.Rule
}
