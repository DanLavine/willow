package limiterclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

//	PARAMETERS:
//	- counter - Counter object to increment or decrement counters from. This value is checked against all Rules
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - Error if there was a limit reached, or encoding issue
//
// UpdateCounter can be used to increment or decrement a particual counter.
// 1. If the `Counter.Counters` is positive, then the counter will either be created or incremented server side.
// 2. If the `Counter.Counters` is negative, then the counter will be decremented server side. If this value reaches 0, then the counter is automatically deleted
// An error will be returned if there is a Rule that limits the total number of Counters associatted with the KeyValues
func (lc *LimitClient) UpdateCounter(ctx context.Context, counter *v1limiter.Counter) error {
	// encode the request
	data, err := api.ObjectEncodeRequest(counter)
	if err != nil {
		return err
	}

	// setup and make the request
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/v1/limiter/counters", lc.url), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	clients.AddHeadersFromContext(req, ctx)

	resp, err := lc.client.Do(req)
	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		// success case and nothing to return
		return nil
	case http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.ModelDecodeResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

//	PARAMETERS:
//	- query - Common query operation to find any number of saved counters
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - Error if there was an unexpcted or encoding issue
//
// LisCounters can be used to query any Counters
func (lc *LimitClient) QueryCounters(ctx context.Context, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Counters, error) {
	// encode the request
	data, err := api.ModelEncodeRequest(query)
	if err != nil {
		return nil, err
	}

	// setup and make the request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/limiter/counters/query", lc.url), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	clients.AddHeadersFromContext(req, ctx)

	resp, err := lc.client.Do(req)
	if err != nil {
		return nil, err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		counters := v1limiter.Counters{}
		if err := api.ModelDecodeResponse(resp, &counters); err != nil {
			return nil, err
		}

		return counters, nil
	case http.StatusBadRequest, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.ModelDecodeResponse(resp, apiError); err != nil {
			return nil, err
		}

		return nil, apiError
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

//	PARAMETERS:
//	- counter - Counter object to set specific counters for
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - Error if there was a limit reached, or encoding issue
//
// SetCounters is used to forcefully set the number of counters for a particual KeyValues without enforcing any rules
func (lc *LimitClient) SetCounters(ctx context.Context, counters *v1limiter.Counter) error {
	// encode the request
	data, err := api.ObjectEncodeRequest(counters)
	if err != nil {
		return err
	}

	// setup and make the request
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/v1/limiter/counters/set", lc.url), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	clients.AddHeadersFromContext(req, ctx)

	resp, err := lc.client.Do(req)
	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		// success case and nothing to return
		return nil
	case http.StatusBadRequest, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.ModelDecodeResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
