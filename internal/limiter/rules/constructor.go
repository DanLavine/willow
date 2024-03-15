package rules

import (
	"fmt"

	"github.com/DanLavine/willow/internal/limiter/rules/memory"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"go.uber.org/zap"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

//go:generate mockgen -destination=rulesfakes/rule_cloner_mock.go -package=rulesfakes github.com/DanLavine/willow/internal/limiter/rules Rule
type Rule interface {
	// Limit retrieves the saved limit
	Limit() int64

	// Update the stored Rule
	Update(logger *zap.Logger, updateRequest *v1limiter.RuleUpdateRquest) *errors.ServerError

	// Delete the rule
	Delete() *errors.ServerError
}

//go:generate mockgen -destination=rulesfakes/rule_constructor_mock.go -package=rulesfakes github.com/DanLavine/willow/internal/limiter/rules RuleConstructor
type RuleConstructor interface {
	New(rule *v1limiter.Rule) Rule
}

func NewRuleConstructor(constructorType string) (RuleConstructor, error) {
	switch constructorType {
	case "memory":
		return &memoryConstrutor{}, nil
	default:
		return nil, fmt.Errorf("unkown constructor type")
	}
}

type memoryConstrutor struct{}

func (mc *memoryConstrutor) New(rule *v1limiter.Rule) Rule {
	return memory.New(rule)
}
