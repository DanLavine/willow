package rules

import (
	"fmt"

	"github.com/DanLavine/willow/internal/limiter/rules/memory"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

//go:generate mockgen -destination=rulefakes/constructor_mock.go -package=rulefakes github.com/DanLavine/willow/internal/limiter/rules RuleConstructor
type RuleConstructor interface {
	New(rule *v1limiter.RuleRequest) Rule
}

type memoryConstrutor struct{}

func (mc *memoryConstrutor) New(rule *v1limiter.RuleRequest) Rule {
	return memory.NewRule(rule)
}

func NewRuleConstructor(constructorType string) (RuleConstructor, error) {
	switch constructorType {
	case "memory":
		return &memoryConstrutor{}, nil
	default:
		return nil, fmt.Errorf("unkown constructor type")
	}
}
