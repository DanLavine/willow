package integrationhelpers

import (
	"github.com/DanLavine/willow/pkg/models/api/v1willow"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
)

var (
	// some general queues
	Queue1 = v1willow.Create{
		Name:                   "queue1",
		QueueMaxSize:           5,
		DeadLetterQueueMaxSize: 1,
	}

	Queue2 = v1willow.Create{
		Name:                   "queue2",
		QueueMaxSize:           5,
		DeadLetterQueueMaxSize: 1,
	}

	// some general enqueu message
	Queue1UpdateableEnqueue = v1willow.EnqueueItemRequest{
		BrokerInfo: v1willow.BrokerInfo{
			Name: "queue1",
			Tags: datatypes.KeyValues{"some": datatypes.String("tag")},
		},
		Data:       []byte(`hello world`),
		Updateable: true,
	}

	Queue1NotUpdateableEnqueue = v1willow.EnqueueItemRequest{
		BrokerInfo: v1willow.BrokerInfo{
			Name: "queue1",
			Tags: datatypes.KeyValues{"some": datatypes.String("tag")},
		},
		Data:       []byte(`hello world`),
		Updateable: false,
	}

	Queue2UpdateableEnqueue = v1willow.EnqueueItemRequest{
		BrokerInfo: v1willow.BrokerInfo{
			Name: "queue2",
			Tags: datatypes.KeyValues{"some": datatypes.String("tag")},
		},
		Data:       []byte(`hello world`),
		Updateable: true,
	}

	// some general dequeue messages
	tagValue      = datatypes.String("tag")
	Queue1Dequeue = v1willow.DequeueItemRequest{
		Name: "queue1",
		Query: query.AssociatedKeyValuesQuery{
			KeyValueSelection: &query.KeyValueSelection{
				KeyValues: map[string]query.Value{
					"some": query.Value{Value: &tagValue, ValueComparison: query.EqualsPtr()},
				},
			},
		},
	}
)
