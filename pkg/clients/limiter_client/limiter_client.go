package limiterclient

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"golang.org/x/net/http2"
)

type LimiterClient interface {
	CreateRule(rule *v1limiter.Rule) error
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

func (lc *limiterClient) CreateRule(rule *v1limiter.Rule) error {
	// always validate locally first
	if err := rule.ValidateRequest(); err != nil {
		return err
	}

	// convert the rule to bytes
	reqBody := rule.ToBytes()

	// setup and make request
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/limiter/rules", lc.url), bytes.NewBuffer(reqBody))
	if err != nil {
		// this should never actually hit
		return fmt.Errorf("internal error setting up http request: %w", err)
	}

	resp, err := lc.client.Do(request)
	if err != nil {
		return fmt.Errorf("unable to make request to limiter service: %w", err)
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		apiErr, err := api.ParseError(resp.Body)
		if err != nil {
			return err
		}

		return apiErr
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
