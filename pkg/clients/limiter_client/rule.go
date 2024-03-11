package limiterclient

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

//	PARAMETERS:
//	- rule - Rule definition to create
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the rule
//
// CreateRule creates a new Rule to enforce Counters against
func (lc *LimitClient) CreateRule(rule *v1limiter.RuleCreateRequest, headers http.Header) error {
	// setup and make the request
	req, err := lc.client.EncodedRequest("POST", fmt.Sprintf("%s/v1/limiter/rules", lc.url), rule)
	if err != nil {
		return err
	}

	clients.AppendHeaders(req, headers)

	resp, err := lc.client.Do(req)
	if err != nil {
		return nil
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError:
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
//	- ruleName - name of the Rule to get
//	- query - overrides to include for the found rule
//	- headers (optional) - any headers to apply to the request
//	RETURNS:
//	- error - error findinig the Rule or Overrides
//
// GetRule is used to find a specific Rule and any optional Overrides that match the query
func (lc *LimitClient) GetRule(ruleName string, query *v1limiter.RuleGet, headers http.Header) (*v1limiter.Rule, error) {
	// setup and make the request
	req, err := lc.client.EncodedRequest("GET", fmt.Sprintf("%s/v1/limiter/rules/%s", lc.url, ruleName), query)
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
		rule := &v1limiter.Rule{}
		if err := api.DecodeAndValidateHttpResponse(resp, rule); err != nil {
			return nil, err
		}

		return rule, nil
	case http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError:
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
//	- matchQuery - query that can be used to match KeyValues to Rules
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - unexpected errors when querying Rule or Overrides
//
// MatchRules is used to find any Rules and optional Overrides for the matchQuery
func (lc *LimitClient) MatchRules(matchQuery *v1limiter.RuleMatch, headers http.Header) (v1limiter.Rules, error) {
	// setup and make the request
	req, err := lc.client.EncodedRequest("GET", fmt.Sprintf("%s/v1/limiter/rules", lc.url), matchQuery)
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
		rules := v1limiter.Rules{}
		if err := api.DecodeAndValidateHttpResponse(resp, &rules); err != nil {
			return nil, err
		}

		return rules, nil
	case http.StatusBadRequest, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return nil, err
		}

		return nil, apiError
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

//	PARAMATERS
//	- ruleName - name of the Rule to update
//	- ruleUpdate - update definition for a particular Rule
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - unexpected errors when updating the Rule
//
// UpdateRule can update the default limits for a particular Rule
func (lc *LimitClient) UpdateRule(ruleName string, ruleUpdate *v1limiter.RuleUpdateRquest, headers http.Header) error {
	// setup and make the request
	req, err := lc.client.EncodedRequest("PUT", fmt.Sprintf("%s/v1/limiter/rules/%s", lc.url, ruleName), ruleUpdate)
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

//	PARAMATERS
//	- ruleName - name of the Rule to delete
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - unexpected errors when deleting the Rule
//
// DeleteRule deletes a Rule and all Overrides for the associated Rule
func (lc *LimitClient) DeleteRule(ruleName string, headers http.Header) error {
	// setup and make the request
	req, err := lc.client.SetupRequest("DELETE", fmt.Sprintf("%s/v1/limiter/rules/%s", lc.url, ruleName))
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
	case http.StatusConflict, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
