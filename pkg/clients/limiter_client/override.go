package limiterclient

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

//	PARAMETERS:
//	- ruleName - name of the Rule to create the Override for
//	- override - Override definition to create
//
//	RETURNS:
//	- error - error creating the override
//
// CreateOverride creates a new Override for a particular rule. It is important to Note that the `Override.KeyValues` must
// include the `Rule.GroubBy` Keys. At least for now, I like this enforcement to make the Overrides easier to reason about
// that they should be for a specific grouping of KeyValue items.
func (lc *limiterClient) CreateOverride(ruleName string, override *v1limiter.Override) error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "POST",
		Path:   fmt.Sprintf("%s/v1/limiter/rules/%s/overrides", lc.url, ruleName),
		Model:  override,
	})

	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError:
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
//
//	RETURNS:
//	- error - error deleting the override
//
// DeleteOverride removes a particual override for a Rule
func (lc *limiterClient) DeleteOverride(ruleName, overrideName string) error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "DELETE",
		Path:   fmt.Sprintf("%s/v1/limiter/rules/%s/overrides/%s", lc.url, ruleName, overrideName),
		Model:  nil,
	})

	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError:
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
//	- query - common AssociatedQuery to find any matching overrides
//
//	RETURNS:
//	- error - unexpected errors when querying overrides
//
// QueryOverrides is used to find any Overrides that match the general Associateduery
func (lc *limiterClient) QueryOverrides(ruleName string, query *v1common.AssociatedQuery) (v1limiter.Overrides, error) {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "GET",
		Path:   fmt.Sprintf("%s/v1/limiter/rules/%s/overrides", lc.url, ruleName),
		Model:  query,
	})

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
	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return nil, err
		}

		return nil, apiError
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// 	PARAMETERS:
// 	- ruleName - name of the Rule to query the Overrides for
// 	- override - Override definition
//
// 	RETURNS:
// 	- error - error creating the override
//
// Match is used to find all the rules that match the provided KeyValues
// func (lc *limiterClient) MatchOverrides(ruleName string, query *v1limtier.RuleQuery) (v1limiter.Overrides, error) {
// 	// setup and make the request
// 	resp, err := lc.client.Do(&clients.RequestData{
// 		Method: "GET",
// 		Path:   fmt.Sprintf("%s/v1/limiter/rules/%s/overrides", lc.url, ruleName),
// 		Model:  query,
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	// parse the response
// 	switch resp.StatusCode {
// 	case http.StatusOK:
// 		overrides := v1limiter.Overrides{}
// 		if err := api.DecodeAndValidateHttpResponse(resp, &overrides); err != nil {
// 			return nil, err
// 		}

// 		return overrides, nil
// 	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError:
// 		apiError := &errors.Error{}
// 		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
// 			return nil, err
// 		}

// 		return nil, apiError
// 	default:
// 		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
// 	}
// }
