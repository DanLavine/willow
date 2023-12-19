package limiterclient

import (
	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

type LimiterClient interface {
	// rule operations
	CreateRule(rule *v1limiter.RuleRequest) error

	GetRule(ruleName string, query *v1limiter.RuleQuery) (*v1limiter.RuleResponse, error)
	ListRules(query *v1limiter.RuleQuery) (*v1limiter.Rules, error)

	UpdateRule(ruleName string, ruleUpdate *v1limiter.RuleUpdate) error

	DeleteRule(ruleName string) error

	// override operations
	ListOverrides(ruleName string, query *v1common.AssociatedQuery) (*v1limiter.Overrides, error)
	CreateOverride(ruleName string, override *v1limiter.Override) error
	DeleteOverride(ruleName string, overrideName string) error

	// counter operations
	ListCounters(query *v1common.AssociatedQuery) (*v1limiter.CountersResponse, error)
	IncrementCounter(counter *v1limiter.Counter) error
	DecrementCounter(counter *v1limiter.Counter) error
	SetCounters(counters *v1limiter.CounterSet) error
}

// client to connect with remote limiter service
type limiterClient struct {
	// url of the service to reach
	url string

	// client setup with HTTP or HTTPS certs
	client clients.HttpClient

	// type to understand request/response formats
	contentType api.ContentType
}

func NewLimiterClient(cfg *clients.Config) (LimiterClient, error) {
	httpClient, err := clients.NewHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	return &limiterClient{
		url:         cfg.URL,
		client:      httpClient,
		contentType: cfg.ContentType,
	}, nil
}
