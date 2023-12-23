package limiterclient

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

func (lc *limiterClient) CreateRule(rule *v1limiter.RuleCreateRequest) error {
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
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) GetRule(ruleName string, query *v1limiter.RuleQuery) (*v1limiter.Rule, error) {
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
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return nil, err
		}

		return nil, apiError
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) ListRules(query *v1limiter.RuleQuery) (v1limiter.Rules, error) {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "GET",
		Path:   fmt.Sprintf("%s/v1/limiter/rules", lc.url),
		Model:  query,
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
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return nil, err
		}

		return nil, apiError
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) UpdateRule(ruleName string, ruleUpdate *v1limiter.RuleUpdateRquest) error {
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
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		apiError := &errors.Error{}
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) DeleteRule(ruleName string) error {
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
