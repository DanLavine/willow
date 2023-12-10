package limiterclient

import (
	"bytes"
	"fmt"
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
