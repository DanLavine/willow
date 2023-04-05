package testhelpers

import v1 "github.com/DanLavine/willow/pkg/models/v1"

func DefaultCreate() v1.Create {
	return v1.Create{
		Name:                   "test queue",
		QueueMaxSize:           5,
		ItemRetryAttempts:      0,
		DeadLetterQueueMaxSize: 0,
	}
}

func DefaultEnqueueItemRequestNotUpdateable() v1.EnqueueItemRequest {
	return v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name:       "test queue",
			BrokerType: v1.Queue,
			Tags:       v1.Strings{"some tag"},
		},
		Data:       []byte(`hello world`),
		Updateable: false,
	}
}

func DefaultEnqueueItemRequestUpdateable() v1.EnqueueItemRequest {
	return v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name:       "test queue",
			BrokerType: v1.Queue,
			Tags:       v1.Strings{"some tag"},
		},
		Data:       []byte(`hello world`),
		Updateable: true,
	}
}
