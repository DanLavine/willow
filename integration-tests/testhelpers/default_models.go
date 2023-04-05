package testhelpers

import v1 "github.com/DanLavine/willow/pkg/models/v1"

var (
	// some general queues
	Queue1 = v1.Create{
		Name:                   v1.String("queue1"),
		QueueMaxSize:           5,
		ItemRetryAttempts:      1,
		DeadLetterQueueMaxSize: 1,
	}

	Queue2 = v1.Create{
		Name:                   v1.String("queue2"),
		QueueMaxSize:           5,
		ItemRetryAttempts:      1,
		DeadLetterQueueMaxSize: 1,
	}

	// some general enqueu message
	Queue1UpdateableEnqueue = v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name:       v1.String("queue1"),
			BrokerType: v1.Queue,
			Tags:       v1.Strings{"some tag"},
		},
		Data:       []byte(`hello world`),
		Updateable: true,
	}

	Queue2UpdateableEnqueue = v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name:       v1.String("queue2"),
			BrokerType: v1.Queue,
			Tags:       v1.Strings{"some tag"},
		},
		Data:       []byte(`hello world`),
		Updateable: true,
	}
)
