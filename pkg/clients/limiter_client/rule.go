package limiterclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

func (lc *limiterClient) CreateRule(rule v1limiter.RuleRequest) error {
	// always validate locally first
	if err := rule.Validate(); err != nil {
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
		return errors.ParseError(resp.Body)
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) GetRule(ruleName string, query v1limiter.RuleQuery) (*v1limiter.RuleResponse, error) {
	// always validate locally first
	if err := query.Validate(); err != nil {
		return nil, err
	}

	// convert the rule query to bytes
	reqBody := query.ToBytes()

	// setup and make request
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/limiter/rules/%s", lc.url, ruleName), bytes.NewBuffer(reqBody))
	if err != nil {
		// this should never actually hit
		return nil, fmt.Errorf("internal error setting up http request: %w", err)
	}

	resp, err := lc.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to make request to limiter service: %w", err)
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		rule := &v1limiter.RuleResponse{}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body from service: %w", err)
		}

		if err := json.Unmarshal(data, rule); err != nil {
			return nil, fmt.Errorf("failed to parse response body fom service: %w", err)
		}

		return rule, nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return nil, errors.ParseError(resp.Body)
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) ListRules(query v1limiter.RuleQuery) (v1limiter.Rules, error) {
	// always validate locally first
	if err := query.Validate(); err != nil {
		return nil, err
	}

	// convert the rule query to bytes
	reqBody := query.ToBytes()

	// setup and make request
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/limiter/rules", lc.url), bytes.NewBuffer(reqBody))
	if err != nil {
		// this should never actually hit
		return nil, fmt.Errorf("internal error setting up http request: %w", err)
	}

	resp, err := lc.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to make request to limiter service: %w", err)
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		rules, err := v1limiter.ParseRulesResponse(resp.Body)
		if err != nil {
			return nil, err
		}

		return rules, nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return nil, errors.ParseError(resp.Body)
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) UpdateRule(ruleName string, ruleUpdate v1limiter.RuleUpdate) error {
	reqBody := ruleUpdate.ToBytes()

	// setup and make request
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/v1/limiter/rules/%s", lc.url, ruleName), bytes.NewBuffer(reqBody))
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
	case http.StatusOK:
		return nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return errors.ParseError(resp.Body)
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) DeleteRule(ruleName string) error {
	// setup and make request
	request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/limiter/rules/%s", lc.url, ruleName), nil)
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
	case http.StatusNoContent:
		return nil
	case http.StatusInternalServerError:
		return errors.ParseError(resp.Body)
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
