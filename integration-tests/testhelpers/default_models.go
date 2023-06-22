package testhelpers

import (
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

var (
	// some general queues
	Queue1 = v1.Create{
		Name:                   "queue1",
		QueueMaxSize:           5,
		DeadLetterQueueMaxSize: 1,
	}

	Queue2 = v1.Create{
		Name:                   "queue2",
		QueueMaxSize:           5,
		DeadLetterQueueMaxSize: 1,
	}

	// some general enqueu message
	Queue1UpdateableEnqueue = v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name: "queue1",
			Tags: datatypes.StringMap{"some": datatypes.String("tag")},
		},
		Data:       []byte(`hello world`),
		Updateable: true,
	}

	Queue1NotUpdateableEnqueue = v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name: "queue1",
			Tags: datatypes.StringMap{"some": datatypes.String("tag")},
		},
		Data:       []byte(`hello world`),
		Updateable: false,
	}

	Queue2UpdateableEnqueue = v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name: "queue2",
			Tags: datatypes.StringMap{"some": datatypes.String("tag")},
		},
		Data:       []byte(`hello world`),
		Updateable: true,
	}

	// some general dequeue messages
	Queue1Dequeue = v1.ReaderSelect{
		BrokerName: "queue1",
		Queries: []v1.ReaderQuery{
			{
				Type: v1.ReaderExactly,
				Tags: datatypes.StringMap{"some": datatypes.String("tag")},
			},
		},
	}
)
