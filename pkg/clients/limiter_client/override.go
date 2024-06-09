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
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

//	PARAMETERS:
//	- ruleName - name of the Rule to create the Override for
//	- override - Override definition to create
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the override
//
// CreateOverride creates a new Override for a particular rule. It is important to Note that the `Override.KeyValues` must
// include the `Rule.GroubBy` Keys. At least for now, I like this enforcement to make the Overrides easier to reason about
// that they should be for a specific grouping of KeyValue items.
func (lc *LimitClient) CreateOverride(ctx context.Context, ruleName string, override *v1limiter.Override) error {
	// encode the request
	data, err := api.ObjectEncodeRequest(override)
	if err != nil {
		return err
	}

	// setup and make the request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides", lc.url, ruleName), bytes.NewBuffer(data))
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
	case http.StatusCreated:
		return nil
	case http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError:
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
//	- ruleName - name of the Rule to query the Overrides for
//	- override - Override definition
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the override
//
// Query is used to find all the rules that match the provided KeyValues
func (lc *LimitClient) QueryOverrides(ctx context.Context, ruleName string, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Overrides, error) {
	// encode the request
	data, err := api.ModelEncodeRequest(query)
	if err != nil {
		return nil, err
	}

	// setup and make the request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/query", lc.url, ruleName), bytes.NewBuffer(data))
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
		overrides := v1limiter.Overrides{}
		if err := api.ModelDecodeResponse(resp, &overrides); err != nil {
			return nil, err
		}

		return overrides, nil
	case http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError:
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
//	- ruleName - name of the Rule to query the Overrides for
//	- override - Override definition
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the override
//
// Match is used to find all the rules that match the provided KeyValues
func (lc *LimitClient) MatchOverrides(ctx context.Context, ruleName string, match *querymatchaction.MatchActionQuery) (v1limiter.Overrides, error) {
	// encode the request
	data, err := api.ModelEncodeRequest(match)
	if err != nil {
		return nil, err
	}

	// setup and make the request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/match", lc.url, ruleName), bytes.NewBuffer(data))
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
		overrides := v1limiter.Overrides{}
		if err := api.ModelDecodeResponse(resp, &overrides); err != nil {
			return nil, err
		}

		return overrides, nil
	case http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError:
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
//	- ruleName - name of the Rule that contains the override
//	- overrideName - name of the Override to obtain
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the override
//
// CreateOverride creates a new Override for a particular rule. It is important to Note that the `Override.KeyValues` must
// include the `Rule.GroubBy` Keys. At least for now, I like this enforcement to make the Overrides easier to reason about
// that they should be for a specific grouping of KeyValue items.
func (lc *LimitClient) GetOverride(ctx context.Context, ruleName string, overrideName string) (*v1limiter.Override, error) {
	// setup and make the reques
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/%s", lc.url, ruleName, overrideName), nil)
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
		override := &v1limiter.Override{}
		if err := api.ModelDecodeResponse(resp, override); err != nil {
			return nil, err
		}

		return override, nil
	case http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError:
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
//	- ruleName - name of the Rule to create the Override for
//	- override - Override definition to create
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the override
//
// CreateOverride creates a new Override for a particular rule. It is important to Note that the `Override.KeyValues` must
// include the `Rule.GroubBy` Keys. At least for now, I like this enforcement to make the Overrides easier to reason about
// that they should be for a specific grouping of KeyValue items.
func (lc *LimitClient) UpdateOverride(ctx context.Context, ruleName string, overrideName string, overrideUpdate *v1limiter.OverrideProperties) error {
	// encode the request
	data, err := api.ModelEncodeRequest(overrideUpdate)
	if err != nil {
		return err
	}

	// setup and make the request
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/%s", lc.url, ruleName, overrideName), bytes.NewBuffer(data))
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
		return nil
	case http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError:
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
//	- ruleName - name of the Rule to delete the Overide from
//	- overrideName - name of the Override to delete
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error deleting the override
//
// DeleteOverride removes a particual override for a Rule
func (lc *LimitClient) DeleteOverride(ctx context.Context, ruleName, overrideName string) error {
	// setup and make the request
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/%s", lc.url, ruleName, overrideName), nil)
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
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.ModelDecodeResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
