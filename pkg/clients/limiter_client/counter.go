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

func (lc *limiterClient) IncrementCounter(counter v1limiter.Counter) error {
	// always validate locally first
	if err := counter.Validate(); err != nil {
		return err
	}

	// convert the counter to bytes
	reqBody := counter.ToBytes()

	// setup and make request
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/limiter/counters", lc.url), bytes.NewBuffer(reqBody))
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
	case http.StatusOK:
		// success case and nothing to return
		return nil
	case http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError:
		apiErr, err := api.ParseError(resp.Body)
		if err != nil {
			return err
		}

		return apiErr
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) DecrementCounter(counter v1limiter.Counter) error {
	// always validate locally first
	if err := counter.Validate(); err != nil {
		return err
	}

	// convert the counter to bytes
	reqBody := counter.ToBytes()

	// setup and make request
	request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/limiter/counters", lc.url), bytes.NewBuffer(reqBody))
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
		// success case and nothing to return
		return nil
	case http.StatusBadRequest, http.StatusInternalServerError:
		apiErr, err := api.ParseError(resp.Body)
		if err != nil {
			return err
		}

		return apiErr
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) ListCounters(query v1limiter.Query) (v1limiter.CountersResponse, error) {
	var countersResponse v1limiter.CountersResponse

	// always validate locally first
	if err := query.Validate(); err != nil {
		return countersResponse, err
	}

	// convert the counter to bytes
	reqBody := query.ToBytes()

	// setup and make request
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/limiter/counters", lc.url), bytes.NewBuffer(reqBody))
	if err != nil {
		// this should never actually hit
		return countersResponse, fmt.Errorf("error setting up http request: %w", err)
	}

	resp, err := lc.client.Do(request)
	if err != nil {
		return countersResponse, fmt.Errorf("unable to make request to limiter service: %w", err)
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return countersResponse, fmt.Errorf("unable to read response from service: %w", err)
		}

		if err = json.Unmarshal(respBody, &countersResponse); err != nil {
			return countersResponse, fmt.Errorf("unable to parse response from service: %w", err)
		}

		return countersResponse, nil
	case http.StatusBadRequest, http.StatusInternalServerError:
		apiErr, err := api.ParseError(resp.Body)
		if err != nil {
			return countersResponse, err
		}

		return countersResponse, apiErr
	default:
		return countersResponse, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) SetCounters(counters v1limiter.CounterSet) error {
	// always validate locally first
	if err := counters.Validate(); err != nil {
		return err
	}

	// convert the counter to bytes
	reqBody := counters.ToBytes()

	// setup and make request
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/limiter/counters/set", lc.url), bytes.NewBuffer(reqBody))
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
	case http.StatusOK:
		// success case and nothing to return
		return nil
	case http.StatusBadRequest, http.StatusInternalServerError:
		apiErr, err := api.ParseError(resp.Body)
		if err != nil {
			return err
		}

		return apiErr
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
