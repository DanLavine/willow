package willowclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

//	PARAMETERS:
//	- queueName - name of the queue to add the item
//	- item - item to be stored in the queue. Including all the detaisl about update and retry operations
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the queue
//
// EnqueueQueueItem enqueus an item to the proper channel for clients to dequeue and process
func (wc *WillowClient) EnqueueQueueItem(ctx context.Context, queueName string, item *v1willow.Item) error {
	// encode the request
	data, err := api.ObjectEncodeRequest(item)
	if err != nil {
		return err
	}

	// setup and make the request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/queues/%s/channels/items", wc.url, queueName), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	clients.AddHeadersFromContext(req, ctx)

	resp, err := wc.client.Do(req)
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
//	- cancelContext - context to cancel the dequeue operation if nothing has been received
//	- queueName - name of the queue to dequeue from
//	- query - query to be applied to any channels on the queue for items to process
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- item - item that will automatically be setup to heartbeat as long as the client is processing
//	- error - error creating the queue
//
// DequeueQueueItem retrieves a particular item that matches the dequeue query
func (wc *WillowClient) DequeueQueueItem(ctx context.Context, queueName string, query *queryassociatedaction.AssociatedActionQuery) (*Item, error) {
	// encode the request
	data, err := api.ModelEncodeRequest(query)
	if err != nil {
		return nil, err
	}

	// setup and make the request
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/queues/%s/channels/items", wc.url, queueName), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	clients.AddHeadersFromContext(req, ctx)

	resp, err := wc.client.Do(req)
	if err != nil {
		return nil, err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		dequeueItem := &v1willow.Item{}
		if err := api.ModelDecodeResponse(resp, dequeueItem); err != nil {
			return nil, err
		}

		return newItem(wc.url, wc.client, queueName, dequeueItem), nil
	case http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError:
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
//	- queueName - name of the queue to delete
//	- channelKeyValues - key value group that defines the channel to be deleted
//	- headers (optional) - any headers to apply to the request
//
//	RETURNS:
//	- error - error creating the queue
//
// DeleteQueueChannel removes a specific channel and any items enqueued
func (wc *WillowClient) DeleteQueueChannel(ctx context.Context, queueName string, channelDelete datatypes.KeyValues) error {
	// encode the request
	if err := channelDelete.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return err
	}

	data, err := json.Marshal(channelDelete)
	if err != nil {
		return err
	}

	// setup and make the request
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/queues/%s/channels", wc.url, queueName), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	clients.AddHeadersFromContext(req, ctx)

	resp, err := wc.client.Do(req)
	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusNoContent:
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
