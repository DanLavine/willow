package willowclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

func (wc *WillowClient) EnqueueQueueItem(queueName string, item *v1willow.EnqueueQueueItem) error {
	// setup and make the request
	resp, err := wc.client.Do(&clients.RequestData{
		Method: "POST",
		Path:   fmt.Sprintf("%s/v1/brokers/queues/%s", wc.url, queueName),
		Model:  item,
	})

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

func (wc *WillowClient) DequeueQueueItem(cancelContext context.Context, queueName string, query *datatypes.AssociatedKeyValuesQuery) (*Item, error) {
	// setup and make the request
	resp, err := wc.client.DoWithContext(cancelContext, &clients.RequestData{
		Method: "GET",
		Path:   fmt.Sprintf("%s/v1/brokers/queues/%s/channel", wc.url, queueName),
		Model:  query,
	})

	if err != nil {
		return nil, err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		dequeueItem := v1willow.DequeueQueueItem{}
		if err := api.DecodeAndValidateHttpResponse(resp, &dequeueItem); err != nil {
			return nil, err
		}

		return newItem(wc.url, wc.client, queueName, &dequeueItem), nil
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

func (wc *WillowClient) DeleteQueueChannel(queueName string, channelKeyValues *datatypes.KeyValues) error {
	// setup and make the request
	resp, err := wc.client.Do(&clients.RequestData{
		Method: "DELETE",
		Path:   fmt.Sprintf("%s/v1/brokers/queues/%s/channel", wc.url, queueName),
		Model:  channelKeyValues,
	})

	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusNoContent:
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
