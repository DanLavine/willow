package handlers

import (
	"encoding/json"
	"io"

	servererrors "github.com/DanLavine/willow/internal/server_errors"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// Used for increment/decrement of a counter
func ParseCounterRequest(reader io.ReadCloser) (*v1limiter.Counter, *servererrors.ApiError) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}

	obj := &v1limiter.Counter{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, servererrors.InvalidRequestBody.With("", err.Error())
	}

	return obj, nil
}

// Used to explicitly set counters
func ParseCounterSetRequest(reader io.ReadCloser) (*v1limiter.CounterSet, *servererrors.ApiError) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}

	obj := &v1limiter.CounterSet{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, servererrors.InvalidRequestBody.With("", err.Error())
	}

	return obj, nil
}

// Server side call to parse the override request to know if it is valid
func ParseOverrideRequest(reader io.ReadCloser) (*v1limiter.Override, *servererrors.ApiError) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}

	obj := &v1limiter.Override{}
	if err := json.Unmarshal(requestBody, &obj); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, servererrors.InvalidRequestBody.With("", err.Error())
	}

	return obj, nil
}

// Server side logic to parse a Rule to know it is valid
func ParseRuleQuery(reader io.ReadCloser) (*v1limiter.RuleQuery, *servererrors.ApiError) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}

	obj := &v1limiter.RuleQuery{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, servererrors.InvalidRequestBody.With("", err.Error())
	}

	return obj, nil
}

// Server side logic to parse a Rule to know it is valid
func ParseRuleUpdateRequest(reader io.ReadCloser) (*v1limiter.RuleUpdate, *servererrors.ApiError) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}

	obj := &v1limiter.RuleUpdate{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("", err.Error())
	}

	// nothing to validate here

	return obj, nil
}

// Server side logic to parse a Rule to know it is valid
func ParseRuleRequest(reader io.ReadCloser) (*v1limiter.RuleRequest, *servererrors.ApiError) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}

	obj := &v1limiter.RuleRequest{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, servererrors.InvalidRequestBody.With("", err.Error())
	}

	return obj, nil
}
