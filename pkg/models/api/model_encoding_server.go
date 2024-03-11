package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

//	PARAMETERS:
//	- r - httpRequest to parse the model and relevant headers out of
//	- obj - api object that data will be parsed into
//
//	RETURNS:
//	- *errors.ServerError - any error encoutered when reading the request or validating the model
//
// DecodeAndValidateHttpRequest is the server side logic used to read and decode any http request into the provided APIObject
func DecodeAndValidateHttpRequest(r *http.Request, obj APIObject) *errors.ServerError {
	if obj == nil {
		return errors.ServerErrorNoAPIModel()
	}

	// can always attempt to read the http response since we want to decode something.
	// if the data is empty, it still means we want to use whatever the decode object returns
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.ServerErrorReadingRequestBody(err)
	}

	contentType := ContentTypeHeader(r.Header)
	switch contentType {
	case ContentTypeJSON:
		if data != nil {
			if err := obj.DecodeJSON(data); err != nil {
				return errors.ServerErrorDecodeingJson(err)
			}
		}
	default:
		return errors.ServerUnknownContentType(contentType)
	}

	if err := obj.Validate(); err != nil {
		return errors.ServerErrorModelRequestValidation(err)
	}

	return nil
}

//	PARAMETERS:
//	- headers - Headers from the original http request. Can pull content type and tracking IDs
//	- w - http.ResponseWriter that will be sent the encoded reponse. Or a `StatusInternalServerError` if the APIObject's encoding and validation fails
//	- statusCode - http status code to send on a successful response
//	- obj - api object to be encoded and validated before responding. If this is nil, then only the statusCode will be sent
//
//	RETURNS:
//	- error - any error encoutered when reading the response or Validating the model
//
// HttpResponse can be used to encode any APIObject and send thr response to the http.ResponseWriter.
// If there us an error encoding or validatiing the data, the http.ResponseWriter will be sent a `http.InternalServerError`
// and an error will be returned that can be logged server side to fix the unexpected issue
func EncodeAndSendHttpResponse(headers http.Header, w http.ResponseWriter, statuscode int, obj APIObject) (int, error) {
	switch obj {
	case nil:
		// only need to send the status code
		w.WriteHeader(statuscode)
		return 0, nil
	default:
		// validate the response is valid
		if err := obj.Validate(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return 0, fmt.Errorf("failed to validate response object: %w", err)
		}

		contentType := ContentTypeHeader(headers)
		switch contentType {
		case ContentTypeJSON:
			// encode the object
			data, err := obj.EncodeJSON()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return 0, fmt.Errorf("failed to encode the response object: %w", err)
			}

			// every request needs to set this header so the client knows how to process the response
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(statuscode)
			n, err := w.Write(data)
			if err != nil {
				return 0, fmt.Errorf("failed to write the respoonse to the client. Closed unexpectedly: %w", err)
			}

			return n, nil
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return 0, fmt.Errorf("unknown content type to send back to the client: %s", contentType)
		}
	}
}
