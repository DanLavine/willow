package api

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/encoding"
)

//	PARAMETERS:
//	- resp - httpResponse to parse the model and relevant headers out of
//	- obj - api object that data will be parsed into
//
//	RETURNS:
//	- *errors.ServerError - any error encoutered when reading the response or validating the model
//
// DecodeAndValidateHttpResponse is the client side logic used to read and decode any http response into the provided APIObject
func DecodeAndValidateHttpResponse(resp *http.Response, obj APIObject) error {
	if obj == nil {
		panic(fmt.Errorf("client code error: API model cannot be nil when decoding api response"))
	}

	// setup the decoder for the response
	contentType := resp.Header.Get(encoding.ContentType)
	encoder, err := encoding.NewEncoder(resp.Header.Get(encoding.ContentType))
	if err != nil {
		return fmt.Errorf("unkown content type recieved from service '%s'. Unable to decode the response", contentType)
	}

	// decode the response
	if err := encoder.Decode(resp.Body, obj); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// ensure the object is valid after decoding
	if err := obj.Validate(); err != nil {
		return fmt.Errorf("failed validation for api response: %w", err)
	}

	return nil
}
