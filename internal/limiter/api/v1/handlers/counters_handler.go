package handlers

import (
	"net/http"

	"github.com/DanLavine/contextops"
	"github.com/DanLavine/willow/internal/middleware"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"go.uber.org/zap"

	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// Increment the Counters if they do not conflict with any rules
func (grh *groupRuleHandler) UpsertCounters(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "UpsertCounters")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the counter increment
	counter := &v1limiter.Counter{}
	if err := api.ObjectDecodeRequest(r, counter); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	if *counter.Spec.Properties.Counters > 0 {
		// this is an increment request

		// need to always setup a new lock client for each request. This is because each update to the counters
		// are independent and multiple request to the same counter need to happen serialy
		lockerClient, lockerErr := lockerclient.NewLockClient(grh.lockerClientConfig)
		if lockerErr != nil {
			logger.Error("failed to create locker client on increment counter request", zap.Error(lockerErr))
			err := errors.InternalServerError
			_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
			return
		}

		ctx, cancel := contextops.MergeDone(ctx, grh.shutdownContext)
		defer cancel()

		if err := grh.counterClient.IncrementCounters(ctx, lockerClient, counter); err != nil {
			_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
			return
		}

		_, _ = api.ModelEncodeResponse(w, http.StatusOK, nil)
	} else {
		// this is a decrement request
		if err := grh.counterClient.DecrementCounters(ctx, counter); err != nil {
			_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
			return
		}

		_, _ = api.ModelEncodeResponse(w, http.StatusOK, nil)
	}
}

// Query the counters to see what is already provided
func (grh *groupRuleHandler) QueryCounters(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "QueryCounters")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the query from the counters
	query := &queryassociatedaction.AssociatedActionQuery{}
	if err := api.ModelDecodeRequest(r, query); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	countersResp, err := grh.counterClient.QueryCounters(ctx, query)
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, &countersResp)
}

func (grh *groupRuleHandler) SetCounters(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "SetCounters")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the countrs set
	counter := &v1limiter.Counter{}
	if err := api.ObjectDecodeRequest(r, counter); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// forcefully set the counters
	if err := grh.counterClient.SetCounter(ctx, counter); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, nil)
}
