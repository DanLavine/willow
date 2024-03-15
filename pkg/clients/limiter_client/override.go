package limiterclient

import (
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
func (lc *LimitClient) CreateOverride(ruleName string, override *v1limiter.Override, headers http.Header) error {
	// setup and make the request
	req, err := lc.client.EncodedRequest("POST", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides", lc.url, ruleName), override)
	if err != nil {
		return err
	}

	clients.AppendHeaders(req, headers)

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
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
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
func (lc *LimitClient) QueryOverrides(ruleName string, query *queryassociatedaction.AssociatedActionQuery, headers http.Header) (v1limiter.Overrides, error) {
	// setup and make the request
	req, err := lc.client.EncodedRequest("GET", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/query", lc.url, ruleName), query)
	if err != nil {
		return nil, err
	}

	clients.AppendHeaders(req, headers)

	resp, err := lc.client.Do(req)
	if err != nil {
		return nil, err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		overrides := v1limiter.Overrides{}
		if err := api.DecodeAndValidateHttpResponse(resp, &overrides); err != nil {
			return nil, err
		}

		return overrides, nil
	case http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
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
func (lc *LimitClient) MatchOverrides(ruleName string, match *querymatchaction.MatchActionQuery, headers http.Header) (v1limiter.Overrides, error) {
	// setup and make the request
	req, err := lc.client.EncodedRequest("GET", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/match", lc.url, ruleName), match)
	if err != nil {
		return nil, err
	}

	clients.AppendHeaders(req, headers)

	resp, err := lc.client.Do(req)
	if err != nil {
		return nil, err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		overrides := v1limiter.Overrides{}
		if err := api.DecodeAndValidateHttpResponse(resp, &overrides); err != nil {
			return nil, err
		}

		return overrides, nil
	case http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
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
func (lc *LimitClient) GetOverride(ruleName string, overrideName string, headers http.Header) (*v1limiter.Override, error) {
	// setup and make the reques
	req, err := lc.client.SetupRequest("GET", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/%s", lc.url, ruleName, overrideName))
	if err != nil {
		return nil, err
	}

	clients.AppendHeaders(req, headers)

	resp, err := lc.client.Do(req)
	if err != nil {
		return nil, err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		override := &v1limiter.Override{}
		if err := api.DecodeAndValidateHttpResponse(resp, override); err != nil {
			return nil, err
		}

		return override, nil
	case http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
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
func (lc *LimitClient) UpdateOverride(ruleName string, overrideName string, overrideUpdate *v1limiter.OverrideUpdate, headers http.Header) error {
	// setup and make the request
	req, err := lc.client.EncodedRequest("PUT", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/%s", lc.url, ruleName, overrideName), overrideUpdate)
	if err != nil {
		return err
	}

	clients.AppendHeaders(req, headers)

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
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
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
func (lc *LimitClient) DeleteOverride(ruleName, overrideName string, headers http.Header) error {
	// setup and make the request
	req, err := lc.client.SetupRequest("DELETE", fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/%s", lc.url, ruleName, overrideName))
	if err != nil {
		return err
	}

	clients.AppendHeaders(req, headers)

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
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
