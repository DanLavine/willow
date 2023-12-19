package limiterclient

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

func (lc *limiterClient) IncrementCounter(counter *v1limiter.Counter) error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "POST",
		Path:   fmt.Sprintf("%s/v1/limiter/counters", lc.url),
		Model:  counter,
	})

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
		if err := apiError.Decode(lc.contentType, resp.Body); err != nil {
			return errors.ClientError(err)
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) DecrementCounter(counter *v1limiter.Counter) error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "DELETE",
		Path:   fmt.Sprintf("%s/v1/limiter/counters", lc.url),
		Model:  counter,
	})

	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusNoContent:
		// success case and nothing to return
		return nil
	case http.StatusBadRequest, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := apiError.Decode(lc.contentType, resp.Body); err != nil {
			return errors.ClientError(err)
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) ListCounters(query *v1common.AssociatedQuery) (*v1limiter.CountersResponse, error) {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "GET",
		Path:   fmt.Sprintf("%s/v1/limiter/counters", lc.url),
		Model:  query,
	})

	if err != nil {
		return nil, err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		countersResponse := &v1limiter.CountersResponse{}
		if err := countersResponse.Decode(lc.contentType, resp.Body); err != nil {
			return nil, errors.ClientError(err)
		}

		return countersResponse, nil
	case http.StatusBadRequest, http.StatusInternalServerError:
		apiError := &errors.Error{}
		if err := apiError.Decode(lc.contentType, resp.Body); err != nil {
			return nil, errors.ClientError(err)
		}

		return nil, apiError
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (lc *limiterClient) SetCounters(counters *v1limiter.CounterSet) error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "POST",
		Path:   fmt.Sprintf("%s/v1/limiter/counters/set", lc.url),
		Model:  counters,
	})

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
		if err := apiError.Decode(lc.contentType, resp.Body); err != nil {
			return errors.ClientError(err)
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
