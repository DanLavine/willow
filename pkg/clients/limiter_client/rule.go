package limiterclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

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

func (lc *limiterClient) GetRule(ruleName string, includeOverrides bool) (*v1limiter.Rule, error) {
	// setup and make request
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/limiter/rules/%s", lc.url, ruleName), nil)
	if err != nil {
		// this should never actually hit
		return nil, fmt.Errorf("internal error setting up http request: %w", err)
	}

	// add the query parameters
	query := request.URL.Query()
	if includeOverrides {
		query.Add("includeOverrides", "true")
	} else {
		query.Add("includeOverrides", "false")
	}
	request.URL.RawQuery = query.Encode()

	resp, err := lc.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to make request to limiter service: %w", err)
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		rule := &v1limiter.Rule{}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body from service: %w", err)
		}

		if err := json.Unmarshal(data, rule); err != nil {
			return nil, fmt.Errorf("failed to parse response body fom service: %w", err)
		}

		return rule, nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		apiErr, err := api.ParseError(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, apiErr
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
