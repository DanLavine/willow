package handlers

import (
	"net/http"

	"github.com/DanLavine/contextops"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"go.uber.org/zap"

	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// Increment the Counters if they do not conflict with any rules
func (grh *groupRuleHandler) UpsertCounters(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("UpsertCounters"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the counter increment
	counter := &v1limiter.Counter{}
	if err := api.DecodeAndValidateHttpRequest(r, counter); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if counter.Counters > 0 {
		// this is an increment request

		// need to always setup a new lock client for each request. This is because each update to the counters
		// are independent and multiple request to the same counter need to happen serialy
		lockerClient, lockerErr := lockerclient.NewLockClient(grh.lockerClientConfig)
		if lockerErr != nil {
			logger.Error("failed to create locker client on increment counter request", zap.Error(lockerErr))
			err := errors.InternalServerError
			_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
			return
		}

		if err := grh.counterClient.IncrementCounters(logger, contextops.MergeForDone(r.Context(), grh.shutdownContext), lockerClient, counter); err != nil {
			_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
			return

		}

		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
	} else {
		// this is a decrement request
		if err := grh.counterClient.DecrementCounters(logger, counter); err != nil {
			_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
			return
		}

		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
	}
}

// Query the counters to see what is already provided
func (grh *groupRuleHandler) QueryCounters(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("QueryCounters"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the query from the counters
	query := &v1common.AssociatedQuery{}
	if err := api.DecodeAndValidateHttpRequest(r, query); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	countersResp, err := grh.counterClient.QueryCounters(logger, query)
	if err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, &countersResp)
}

func (grh *groupRuleHandler) SetCounters(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("SetCounters"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the countrs set
	counter := &v1limiter.Counter{}
	if err := api.DecodeAndValidateHttpRequest(r, counter); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// forcefully set the counters
	if err := grh.counterClient.SetCounter(logger, counter); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
}
