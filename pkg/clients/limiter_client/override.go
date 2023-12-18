package limiterclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

func (lc *limiterClient) CreateOverride(ruleName string, override v1limiter.Override) error {
	// always validate locally first
	if err := override.Validate(); err != nil {
		return err
	}

	// convert the override to bytes
	reqBody := override.ToBytes()

	// setup and make request
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides", lc.url, ruleName), bytes.NewBuffer(reqBody))
	if err != nil {
		// this should never actually hit
		return fmt.Errorf("error setting up http request: %w", err)
	}

	resp, err := lc.client.Do(request)
	if err != nil {
		return fmt.Errorf("unable to make request to limiter service: %w", err)
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		return errors.ParseError(resp.Body)
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) DeleteOverride(ruleName, overrideName string) error {
	// setup and make request
	request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/%s", lc.url, ruleName, overrideName), nil)
	if err != nil {
		// this should never actually hit
		return fmt.Errorf("error setting up http request: %w", err)
	}

	resp, err := lc.client.Do(request)
	if err != nil {
		return fmt.Errorf("unable to make request to limiter service: %w", err)
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		return errors.ParseError(resp.Body)
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) ListOverrides(ruleName string, query v1common.AssociatedQuery) (v1limiter.Overrides, error) {
	var overrides v1limiter.Overrides

	// always validate locally first
	if err := query.Validate(); err != nil {
		return overrides, err
	}

	// convert the query to bytes
	reqBody := query.ToBytes()

	// setup and make request
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides", lc.url, ruleName), bytes.NewBuffer(reqBody))
	if err != nil {
		// this should never actually hit
		return overrides, fmt.Errorf("error setting up http request: %w", err)
	}

	resp, err := lc.client.Do(request)
	if err != nil {
		return overrides, fmt.Errorf("unable to make request to limiter service: %w", err)
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:

		respData, err := io.ReadAll(resp.Body)
		if err != nil {
			return overrides, fmt.Errorf("failed to read response from limiter service: %w", err)
		}

		if err := json.Unmarshal(respData, &overrides); err != nil {
			return overrides, fmt.Errorf("failed to parse response from limiter service: %w", err)
		}

		return overrides, nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		return overrides, errors.ParseError(resp.Body)
	default:
		return overrides, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
