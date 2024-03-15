package willowclient

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"

	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

//	PARAMETERS:
//	- queue - Queue definition to create
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the queue
//
// CreateQueue creates a new Queue. This will return an error if the queue name already exists
func (wc *WillowClient) CreateQueue(queue *v1willow.QueueCreate, headers http.Header) error {
	// setup and make the request
	req, err := wc.client.EncodedRequest("POST", fmt.Sprintf("%s/v1/queues", wc.url), queue)
	if err != nil {
		return err
	}

	clients.AppendHeaders(req, headers)

	resp, err := wc.client.Do(req)
	if err != nil {
		return err
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
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the queue
//
// List all the queues without any of the details about the queue's channels
func (wc *WillowClient) ListQueues(headers http.Header) (v1willow.Queues, error) {
	// setup and make the request
	req, err := wc.client.SetupRequest("GET", fmt.Sprintf("%s/v1/queues", wc.url))
	if err != nil {
		return nil, err
	}

	clients.AppendHeaders(req, headers)

	resp, err := wc.client.Do(req)
	if err != nil {
		return nil, err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		queues := v1willow.Queues{}
		if err := api.DecodeAndValidateHttpResponse(resp, &queues); err != nil {
			return nil, err
		}

		return queues, nil
	case http.StatusInternalServerError:
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
//	- queueName - name of the queue to obtain
//	- channelsQuery - query for specific channels to find
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the queue
//
// GetQueue retrieves a queue and any channels that match the provided query
func (wc *WillowClient) GetQueue(queueName string, channelsQuery *queryassociatedaction.AssociatedActionQuery, headers http.Header) (*v1willow.Queue, error) {
	// setup and make the request
	req, err := wc.client.EncodedRequest("GET", fmt.Sprintf("%s/v1/queues/%s", wc.url, queueName), channelsQuery)
	if err != nil {
		return nil, err
	}

	clients.AppendHeaders(req, headers)

	resp, err := wc.client.Do(req)
	if err != nil {
		return nil, err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		queue := &v1willow.Queue{}

		if err := api.DecodeAndValidateHttpResponse(resp, queue); err != nil {
			return nil, err
		}

		return queue, nil
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
//	- queueName - name of the queue to update
//	- queueUpdate - full configuration to apply to the existing queue
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the queue
//
// UpdateQueue is used to upadte the queue's specific properties
func (wc *WillowClient) UpdateQueue(queueName string, queueUpdate *v1willow.QueueUpdate, headers http.Header) error {
	// setup and make the request
	req, err := wc.client.EncodedRequest("PUT", fmt.Sprintf("%s/v1/queues/%s", wc.url, queueName), queueUpdate)
	if err != nil {
		return err
	}

	clients.AppendHeaders(req, headers)

	resp, err := wc.client.Do(req)
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
//	- queueName - name of the queue to update
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the queue
//
// Delete a queue and any channels associated with the Queue
func (wc *WillowClient) DeleteQueue(queueName string, headers http.Header) error {
	// setup and make the request
	req, err := wc.client.SetupRequest("DELETE", fmt.Sprintf("%s/v1/queues/%s", wc.url, queueName))
	if err != nil {
		return err
	}

	clients.AppendHeaders(req, headers)

	resp, err := wc.client.Do(req)
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
