package limiterclient

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

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
		if err := apiError.Decode(lc.contentType, resp.Body); err != nil {
			return errors.ClientError(err)
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

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
		if err := apiError.Decode(lc.contentType, resp.Body); err != nil {
			return errors.ClientError(err)
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) ListOverrides(ruleName string, query *v1common.AssociatedQuery) (*v1limiter.Overrides, error) {
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
		overrides := &v1limiter.Overrides{}
		if err := overrides.Decode(lc.contentType, resp.Body); err != nil {
			return nil, errors.ClientError(err)
		}

		return overrides, nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := apiError.Decode(lc.contentType, resp.Body); err != nil {
			return nil, errors.ClientError(err)
		}

		return nil, apiError
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
