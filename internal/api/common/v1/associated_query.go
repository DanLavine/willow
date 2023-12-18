package v1

import (
	"encoding/json"
	"io"

	servererrors "github.com/DanLavine/willow/internal/server_errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
)

// Parse the AssoiciatedQuery that is used to find many server resources
func ParseAssociatedQuery(reader io.ReadCloser) (*v1common.AssociatedQuery, *servererrors.ApiError) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}

	obj := &v1common.AssociatedQuery{}
	if err := json.Unmarshal(requestBody, &obj); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, servererrors.InvalidRequestBody.With("", validateErr.Error())
	}

	return obj, nil
}
