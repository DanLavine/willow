package handlers

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/middleware"
	"github.com/DanLavine/willow/internal/willow/brokers/queues"
	"github.com/DanLavine/willow/pkg/models/api"
	"go.uber.org/zap"

	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
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
	ChannelEnqueue(w http.ResponseWriter, r *http.Request)
	ChannelDequeue(w http.ResponseWriter, r *http.Request)
	ChannelDelete(w http.ResponseWriter, r *http.Request)

	// item handlers
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

func (qh queueHandler) Create(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "Create")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the create rule request
	create := &v1willow.QueueCreate{}
	if err := api.ModelDecodeRequest(r, create); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// create the new rule
	if err := qh.queueClient.CreateQueue(ctx, create); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusCreated, nil)
}

func (qh queueHandler) List(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "List")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	queues, err := qh.queueClient.ListQueues(ctx)
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// send the response
	_, _ = api.ModelEncodeResponse(w, http.StatusOK, &queues)
}

func (qh queueHandler) Get(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "Get")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	queue, err := qh.queueClient.GetQueue(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"])
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// send the response
	if queue == nil {
		_, _ = api.ModelEncodeResponse(w, http.StatusNotFound, nil)
	} else {
		_, _ = api.ModelEncodeResponse(w, http.StatusOK, queue)
	}
}

func (qh queueHandler) Update(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "Update")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the update
	update := &v1willow.QueueUpdate{}
	if err := api.ModelDecodeRequest(r, update); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	if err := qh.queueClient.UpdateQueue(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], update); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, nil)
}

func (qh queueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "Delete")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	if err := qh.queueClient.DeleteQueue(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"]); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// send the response
	_, _ = api.ModelEncodeResponse(w, http.StatusNoContent, nil)
}
