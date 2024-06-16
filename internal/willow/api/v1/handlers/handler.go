package handlers

import (
	"net/http"

	"github.com/DanLavine/willow/internal/willow/brokers/queues"
)

type V1QueueHandler interface {
	// all queue operations
	Create(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)

	// queue specific operations
	Get(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)

	// channel handlers
	ChannelQuery(w http.ResponseWriter, r *http.Request)
	ChannelDelete(w http.ResponseWriter, r *http.Request)

	// item handlers
	ChannelEnqueue(w http.ResponseWriter, r *http.Request)
	ChannelDequeue(w http.ResponseWriter, r *http.Request)
	ItemACK(w http.ResponseWriter, r *http.Request)
	ItemHeartbeat(w http.ResponseWriter, r *http.Request)
}

type queueHandler struct {
	queueClient queues.QueuesClient
}

func NewV1QueueHandler(queueClient queues.QueuesClient) *queueHandler {
	return &queueHandler{
		queueClient: queueClient,
	}
}
