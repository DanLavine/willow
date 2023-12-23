package limiterclient

import (
	"github.com/DanLavine/willow/pkg/clients"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

type LimiterClient interface {
	// rule operations
	CreateRule(rule *v1limiter.RuleCreateRequest) error

	GetRule(ruleName string, query *v1limiter.RuleQuery) (*v1limiter.Rule, error)
	ListRules(query *v1limiter.RuleQuery) (v1limiter.Rules, error)

	UpdateRule(ruleName string, ruleUpdate *v1limiter.RuleUpdateRquest) error

	DeleteRule(ruleName string) error

	// override operations
	ListOverrides(ruleName string, query *v1common.AssociatedQuery) (v1limiter.Overrides, error)
	CreateOverride(ruleName string, override *v1limiter.Override) error
	DeleteOverride(ruleName string, overrideName string) error

	// counter operations
	ListCounters(query *v1common.AssociatedQuery) (v1limiter.Counters, error)
	UpdateCounter(counter *v1limiter.Counter) error
	SetCounters(counters *v1limiter.Counter) error
}

// client to connect with remote limiter service
type limiterClient struct {
	// url of the service to reach
	url string

	// client setup with HTTP or HTTPS certs
	client clients.HttpClient

	// type to understand request/response formats
	contentType string
}

func NewLimiterClient(cfg *clients.Config) (LimiterClient, error) {
	httpClient, err := clients.NewHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	return &limiterClient{
		url:         cfg.URL,
		client:      httpClient,
		contentType: cfg.ContentEncoding,
	}, nil
}
