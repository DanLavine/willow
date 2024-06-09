package limiterclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// All Client operations for interacting with the Limiter Service
//
//go:generate mockgen -destination=limiterclientfakes/limiter_client_mock.go -package=limiterclientfakes github.com/DanLavine/willow/pkg/clients/limiter_client LimiterClient
type LimiterClient interface {
	// Ping health to know service is up and running
	Healthy() error

	// Rule operations
	// Create a new Rule
	CreateRule(ctx context.Context, rule *v1limiter.Rule) error
	// Get a spcific rule by name and query the possible overrides
	GetRule(ctx context.Context, ruleName string) (*v1limiter.Rule, error)
	// Query Rules for specific key values
	QueryRules(ctx context.Context, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Rules, error)
	// Match any Rules for the provided key values
	MatchRules(ctx context.Context, match *querymatchaction.MatchActionQuery) (v1limiter.Rules, error)
	// Update a Rule by name
	UpdateRule(ctx context.Context, ruleName string, ruleUpdate *v1limiter.RuleProperties) error
	// Delete a Rule by name
	DeleteRule(ctx context.Context, ruleName string) error

	// Override operations
	// Create an Override for a particular Rule
	CreateOverride(ctx context.Context, ruleName string, override *v1limiter.Override) error
	// Get an Override for a particular Rule
	GetOverride(ctx context.Context, ruleName string, overrideName string) (*v1limiter.Override, error)
	// Query Overrides
	QueryOverrides(ctx context.Context, ruleName string, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Overrides, error)
	// Match Overrides
	MatchOverrides(ctx context.Context, ruleName string, match *querymatchaction.MatchActionQuery) (v1limiter.Overrides, error)
	// Update a particular Override
	UpdateOverride(ctx context.Context, ruleName string, overrideName string, overrideUpdate *v1limiter.OverrideProperties) error
	// Delete an Override for a particual Rule
	DeleteOverride(ctx context.Context, ruleName string, overrideName string) error

	// Counter operations
	// Increment Or Decrement a particual Counter
	UpdateCounter(ctx context.Context, counter *v1limiter.Counter) error
	// Query Counters
	QueryCounters(ctx context.Context, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Counters, error)
	// Forcefully set the Counter without enforcing any rules
	SetCounters(ctx context.Context, counters *v1limiter.Counter) error
}

// LimiteClient to connect with remote limiter service
type LimitClient struct {
	// url of the service to reach
	url string

	// client setup with HTTP or HTTPS certs
	client *http.Client
}

//	PARAMATERS
//	- cfg - Configuration to interact with the Limiter service
//
//	RETURNS:
//	- LimiterClient - thread safe client that can be shared to any number of goroutines
//	- error - error validating the configuration or setting up the client
//
// NewLimiterClient creates a new client to interact with the Limiter service
func NewLimiterClient(cfg *clients.Config) (*LimitClient, error) {
	httpClient, err := clients.NewHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	return &LimitClient{
		url:    cfg.URL,
		client: httpClient,
	}, nil
}

func (lc *LimitClient) Healthy() error {
	// setup and make the request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/health", lc.url), nil)
	if err != nil {
		return fmt.Errorf("failed to setup request to healthy api")
	}

	resp, err := lc.client.Do(req)
	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("unexpected status code checking health: %d", resp.StatusCode)
	}
}
