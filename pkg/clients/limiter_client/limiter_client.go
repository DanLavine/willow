package limiterclient

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
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
	CreateRule(rule *v1limiter.RuleCreateRequest) error
	// Get a spcific rule by name and query the possible overrides
	GetRule(ruleName string, query *v1limiter.RuleGet) (*v1limiter.Rule, error)
	// MAtch Rules and possible Overrides for a specific KeyValues group
	MatchRules(query *v1limiter.RuleMatch) (v1limiter.Rules, error)
	// Update a Rule by name
	UpdateRule(ruleName string, ruleUpdate *v1limiter.RuleUpdateRquest) error
	// Delete a Rule by name
	DeleteRule(ruleName string) error

	// Override operations
	// Create an Override for a particular Rule
	CreateOverride(ruleName string, override *v1limiter.Override) error
	// Get an Override for a particular Rule
	GetOverride(ruleName string, overrideName string) (*v1limiter.Override, error)
	// Match Overrides
	MatchOverrides(ruleName string, query *v1common.MatchQuery) (v1limiter.Overrides, error)
	// Update a particular Override
	UpdateOverride(ruleName string, overrideName string, overrideUpdate *v1limiter.OverrideUpdate) error
	// Delete an Override for a particual Rule
	DeleteOverride(ruleName string, overrideName string) error

	// Counter operations
	// Increment Or Decrement a particual Counter
	UpdateCounter(counter *v1limiter.Counter) error
	// Query Counters
	QueryCounters(query *v1common.AssociatedQuery) (v1limiter.Counters, error)
	// Forcefully set the Counter without enforcing any rules
	SetCounters(counters *v1limiter.Counter) error
}

// LimiteClient to connect with remote limiter service
type LimitClient struct {
	// url of the service to reach
	url string

	// client setup with HTTP or HTTPS certs
	client clients.HttpClient

	// type to understand request/response formats
	contentType string
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
		url:         cfg.URL,
		client:      httpClient,
		contentType: cfg.ContentEncoding,
	}, nil
}

func (lc *LimitClient) Healthy() error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "GET",
		Path:   fmt.Sprintf("%s/health", lc.url),
		Model:  nil,
	})

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
