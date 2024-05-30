package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/encoding"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

//	PARAMETERS:
//	- r - httpRequest to parse the model out of
//	- obj - api object that data will be parsed into
//
//	RETURNS:
//	- *errors.ServerError - any error encoutered when reading the request or validating the model
//
// DecodeAndValidateHttpRequest is the server side logic used to read and decode any http request into the provided APIObject
func DecodeRequest(r *http.Request, obj any) *errors.ServerError {
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return errors.ServerErrorDecoding(err)
	}

	return nil
}

//	PARAMETERS:
//	- r - httpRequest to parse the model and relevant headers out of
//	- obj - api object that data will be parsed into
//
//	RETURNS:
//	- *errors.ServerError - any error encoutered when reading the request or validating the model
//
// ObjectDecodeAndValidateHttpRequest is the server side logic used to read and decode any http request into the provided APIObject
func ModelDecodeRequest(r *http.Request, obj ApiModel) *errors.ServerError {
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return errors.ServerErrorDecoding(err)
	}

	if err := obj.Validate(); err != nil {
		return errors.ServerErrorModelRequestValidation(err)
	}

	return nil
}

//	PARAMETERS:
//	- r - httpRequest to parse the model and relevant headers out of
//	- obj - api object that data will be parsed into
//
//	RETURNS:
//	- *errors.ServerError - any error encoutered when reading the request or validating the model
//
// ObjectDecodeAndValidateHttpRequest is the server side logic used to read and decode any http request into the provided APIObject
func ObjectDecodeRequest(r *http.Request, obj APIObject) *errors.ServerError {
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return errors.ServerErrorDecoding(err)
	}

	if err := obj.ValidateSpecOnly(); err != nil {
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
func ModelEncodeResponse(w http.ResponseWriter, statuscode int, obj ApiModel) (int, error) {
	switch obj {
	case nil:
		// be safe and remove the content typ header just in case since there is none
		w.Header().Del(encoding.ContentTypeHeader)

		// only need to send the status code
		w.WriteHeader(statuscode)
		return 0, nil
	default:
		// validate the response is valid
		if err := obj.Validate(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"Message":"%s"}`, err.Error())))

			return 0, fmt.Errorf("failed to validate response object: %w", err)
		}

		data, err := json.Marshal(obj)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"Message":"failed to encode the response"}`))

			return 0, fmt.Errorf("failed to encode the response object: %w", err)
		}

		// write the reponse to the same encoder that we received it in
		w.WriteHeader(statuscode)

		n, writeErr := w.Write(data)
		if writeErr != nil {
			return n, fmt.Errorf("failed to write the response to the client: %w", writeErr)
		}

		return n, nil
	}
}
