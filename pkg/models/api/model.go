package api

import "github.com/DanLavine/willow/pkg/models/api/common/errors"

// APIObjects are all representtations that can be saved in the DB
type APIObject interface {
	ApiModel

	ValidateSpecOnly() *errors.ModelError
}

// APIAction are models that have information for an action to perform
type ApiModel interface {
	Validate() *errors.ModelError
}
