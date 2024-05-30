package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ObjectEncodeRequest(obj APIObject) ([]byte, error) {
	if err := obj.ValidateSpecOnly(); err != nil {
		return nil, err
	}

	return json.Marshal(obj)
}

func ModelEncodeRequest(obj ApiModel) ([]byte, error) {
	if err := obj.Validate(); err != nil {
		return nil, err
	}

	return json.Marshal(obj)
}

//	PARAMETERS:
//	- resp - httpResponse to parse the model and relevant headers out of
//	- obj - api object that data will be parsed into
//
//	RETURNS:
//	- *errors.ServerError - any error encoutered when reading the response or validating the model
//
// DecodeAndValidateHttpResponse is the client side logic used to read and decode any http response into the provided APIObject
func ModelDecodeResponse(resp *http.Response, obj ApiModel) error {
	// decode the response
	if err := json.NewDecoder(resp.Body).Decode(obj); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// ensure the object is valid after decoding
	if err := obj.Validate(); err != nil {
		return err
	}

	return nil
}
