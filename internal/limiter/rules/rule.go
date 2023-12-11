package rules

import (
	"sort"

	v1limitermodels "github.com/DanLavine/willow/internal/limiter/v1_limiter_models"
	"github.com/DanLavine/willow/pkg/models/api"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=rulefakes/rule_mock.go -package=rulefakes github.com/DanLavine/willow/internal/limiter/rules Rule
type Rule interface {
	// retrieve the Name of the rules
	Name() string
	// retrieve the original group by keys
	GetGroupByKeys() []string

	// Update a particualr rule's default limit values
	Update(logger *zap.Logger, update *v1limiter.RuleUpdate)

	// set an override for a particualr group of tags
	SetOverride(logger *zap.Logger, override *v1limiter.Override) *api.Error

	// Delete an override for a particualr group of tags
	DeleteOverride(logger *zap.Logger, overrideName string) *api.Error

	// Operation that is calledfor cascading deletes because the rule is being deleted
	CascadeDeletion(logger *zap.Logger) *api.Error

	// Find the limits for a particualr group of tags including any overrides
	FindLimits(logger *zap.Logger, keyValues datatypes.KeyValues) (v1limitermodels.Limits, *api.Error)

	// generate a query based on the search tags
	//GenerateQuery(keyValues datatypes.KeyValues) datatypes.Select

	// operations for callbacks when resources are freed
	// TODO: These operations are very consuming and hard to implement. So do that after a basic limiter is
	// setup and useful
	//
	//AddClientWaiting(logger *zap.Logger, keyValues datatypes.KeyValues) <-chan struct{} // client calls when trying to "increment", but a limit is blocked, sets up a callback that triggers another "increment" call
	//TriggerClientWaiting(keyValues datatypes.KeyValues)                                 // decrement calls these whenever that happens. Can als be called on an override increase change

	// Get a rule response for Read operations
	Get(includeOverrides *v1limiter.RuleQuery) *v1limiter.RuleResponse

	// Inernal operations
	//GroupBy() []string
	//Limits()
}

func SortRulesGroupBy(rules []Rule) [][]string {
	sortedGroupBy := [][]string{}

	for _, rule := range rules {
		sortedGroupBy = append(sortedGroupBy, rule.GetGroupByKeys())
	}

	// 1. sort by lenght of rules
	// 2. sort by keys within each length
	sort.SliceStable(sortedGroupBy, func(i, j int) bool {
		if len(sortedGroupBy[i]) < len(sortedGroupBy[j]) {
			return true
		}

		if len(sortedGroupBy[i]) == len(sortedGroupBy[j]) {
			sortedKeysI := sort.StringSlice(sortedGroupBy[i])
			sortedKeysJ := sort.StringSlice(sortedGroupBy[j])

			for index, value := range sortedKeysI {
				if value < sortedKeysJ[index] {
					return true
				} else if sortedKeysJ[index] < value {
					return false
				}
			}

			// at this point, all values must be in the proper order
			return true
		}

		return false
	})

	return sortedGroupBy
}
