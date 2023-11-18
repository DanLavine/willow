package limiterclient

import (
	"crypto/tls"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"golang.org/x/net/http2"
)

type LimiterClient interface {
	// rule operations
	CreateRule(rule v1limiter.RuleRequest) error

	GetRule(ruleName string, query v1limiter.RuleQuery) (*v1limiter.RuleResponse, error)
	ListRules(query v1limiter.RuleQuery) (v1limiter.Rules, error)

	UpdateRule(ruleName string, ruleUpdate v1limiter.RuleUpdate) error

	DeleteRule(ruleName string) error

	// override operations
	CreateOverride(ruleName string, override v1limiter.Override) error
	DeleteOverride(ruleName string, overrideName string) error
}

type limiterClient struct {

	// client to connect with remote limiter service
	url    string
	client *http.Client
}

func NewLimiterClient(cfg *clients.Config) (LimiterClient, error) {
	if err := cfg.Vaidate(); err != nil {
		return nil, err
	}

	httpClient := &http.Client{}

	if cfg.CAFile != "" {
		httpClient.Transport = &http2.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cfg.Cert()},
				RootCAs:      cfg.RootCAs(),
			},
		}
	}

	limiterClient := &limiterClient{
		url:    cfg.URL,
		client: httpClient,
	}

	return limiterClient, nil
}
