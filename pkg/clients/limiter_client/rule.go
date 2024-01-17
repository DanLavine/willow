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
//
//	RETURNS:
//	- error - error creating the rule
//
// CreateRule creates a new Rule to enforce Counters against
func (lc *LimitClient) CreateRule(rule *v1limiter.RuleCreateRequest) error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "POST",
		Path:   fmt.Sprintf("%s/v1/limiter/rules", lc.url),
		Model:  rule,
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
//	- ruleName - name of the Rule to get
//	- query - overrides to include for the found rule
//
//	RETURNS:
//	- error - error findinig the Rule or Overrides
//
// GetRule is used to find a specific Rule and any optional Overrides that match the query
func (lc *LimitClient) GetRule(ruleName string, query *v1limiter.RuleGet) (*v1limiter.Rule, error) {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "GET",
		Path:   fmt.Sprintf("%s/v1/limiter/rules/%s", lc.url, ruleName),
		Model:  query,
	})

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
	case http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError:
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
//
//	RETURNS:
//	- error - unexpected errors when querying Rule or Overrides
//
// MatchRules is used to find any Rules and optional Overrides for the matchQuery
func (lc *LimitClient) MatchRules(matchQuery *v1limiter.RuleMatch) (v1limiter.Rules, error) {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "GET",
		Path:   fmt.Sprintf("%s/v1/limiter/rules", lc.url),
		Model:  matchQuery,
	})

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
	case http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError:
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
//
//	RETURNS:
//	- error - unexpected errors when updating the Rule
//
// UpdateRule can update the default limits for a particular Rule
func (lc *LimitClient) UpdateRule(ruleName string, ruleUpdate *v1limiter.RuleUpdateRquest) error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "PUT",
		Path:   fmt.Sprintf("%s/v1/limiter/rules/%s", lc.url, ruleName),
		Model:  ruleUpdate,
	})

	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError:
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
//
//	RETURNS:
//	- error - unexpected errors when deleting the Rule
//
// DeleteRule deletes a Rule and all Overrides for the associated Rule
func (lc *LimitClient) DeleteRule(ruleName string) error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "DELETE",
		Path:   fmt.Sprintf("%s/v1/limiter/rules/%s", lc.url, ruleName),
		Model:  nil,
	})

	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
