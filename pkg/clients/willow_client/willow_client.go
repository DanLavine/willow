package willowclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"

	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// All Client operations for interacting with the Willow Service
//
//go:generate mockgen -destination=limiterclientfakes/limiter_client_mock.go -package=limiterclientfakes github.com/DanLavine/willow/pkg/clients/limiter_client WillowClient
type WillowServiceClient interface {
	// Ping health to know service is up and running
	Healthy() error

	// Queue Operations
	//// Create a new queue
	CreateQueue(ctx context.Context, queue *v1willow.Queue) error
	//// Get a spcific queue by name and query the possible channels
	GetQueue(ctx context.Context, queueName string, query *queryassociatedaction.AssociatedActionQuery) (*v1willow.Queue, error)
	//// List all possible queues without their channels
	ListQueues(ctx context.Context) (v1willow.Queues, error)
	//// Update a particualr queue
	UpdateQueue(ctx context.Context, queueName string, update *v1willow.QueueProperties) error
	//// Delete a particualr queue
	DeleteQueue(ctx context.Context, queueName string) error

	// channel operations
	//// enqueue a new item to a particular queue's channels
	EnqueueQueueItem(ctx context.Context, queueName string, item *v1willow.Item) error
	//// dequeue an item from a queue's channels that match the query
	DequeueQueueItem(ctx context.Context, queueName string, query *queryassociatedaction.AssociatedActionQuery) (*Item, error)
	//// delete a particu;ar channel and all enqueued items
	DeleteQueueChannel(ctx context.Context, queueName string, channelKeyValues datatypes.KeyValues) error
}

// LimiteClient to connect with remote limiter service
type WillowClient struct {
	// url of the service to reach
	url string

	// client setup with HTTP or HTTPS certs
	client *http.Client
}

//	PARAMATERS
//	- cfg - Configuration to interact with the WillowClient service
//
//	RETURNS:
//	- WillowClient - thread safe client that can be shared to any number of goroutines
//	- error - error validating the configuration or setting up the client
//
// NewWillowClient creates a new client to interact with the WillowClient service
func NewWillowClient(cfg *clients.Config) (*WillowClient, error) {
	httpClient, err := clients.NewHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	return &WillowClient{
		url:    cfg.URL,
		client: httpClient,
	}, nil
}

func (wc *WillowClient) Healthy() error {
	// setup and make the request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/health", wc.url), nil)
	if err != nil {
		return fmt.Errorf("failed to setup request to healthy api")
	}

	resp, err := wc.client.Do(req)
	if err != nil {
		return err
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("unexpected status code checking willow health: %d", resp.StatusCode)
	}
}
