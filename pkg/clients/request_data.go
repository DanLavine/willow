package clients

import (
	"github.com/DanLavine/willow/pkg/models/api"
)

type RequestData struct {
	// String value for API method [GET, POST, PUT, DELETE]
	Method string

	// URL path to make the request against
	Path string

	// API object model to encode
	Model api.APIObject
}
