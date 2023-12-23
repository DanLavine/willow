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

func (lc *limiterClient) UpdateCounter(counter *v1limiter.Counter) error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "PUT",
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
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// func (lc *limiterClient) DecrementCounter(counter *v1limiter.Counter) error {
// 	// setup and make the request
// 	resp, err := lc.client.Do(&clients.RequestData{
// 		Method: "DELETE",
// 		Path:   fmt.Sprintf("%s/v1/limiter/counters", lc.url),
// 		Model:  counter,
// 	})

// 	if err != nil {
// 		return err
// 	}

// 	// parse the response
// 	switch resp.StatusCode {
// 	case http.StatusNoContent:
// 		// success case and nothing to return
// 		return nil
// 	case http.StatusBadRequest, http.StatusInternalServerError:
// 		apiError := &errors.Error{}
// 		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
// 			return err
// 		}

// 		return apiError
// 	default:
// 		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
// 	}
// }

func (lc *limiterClient) ListCounters(query *v1common.AssociatedQuery) (v1limiter.Counters, error) {
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
		counters := v1limiter.Counters{}
		if err := api.DecodeAndValidateHttpResponse(resp, &counters); err != nil {
			return nil, err
		}

		return counters, nil
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

func (lc *limiterClient) SetCounters(counters *v1limiter.Counter) error {
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
		if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
			return err
		}

		return apiError
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
