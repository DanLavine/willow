package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	willowerr "github.com/DanLavine/willow/pkg/models/api/common/errors"
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

	// can always attempt to read the http response since we want to decode something.
	// if the data is empty, it still means we want to use whatever the decode object returns
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read http response body: %w", err)
	}

	contentType := ContentTypeHeader(resp.Header)
	switch contentType {
	case ContentTypeJSON:
		if err := obj.DecodeJSON(data); err != nil {
			// special case for an error that occured due to parsing an non 2xx, 3xx status code
			var apiErr *willowerr.Error
			if errors.As(err, &apiErr) {
				return err
			}

			return fmt.Errorf("failed to decode response: %w", err)
		}
	default:
		return fmt.Errorf("unkown content type recieved from service: %s", contentType)
	}

	if err := obj.Validate(); err != nil {
		return fmt.Errorf("failed validation for api response: %w", err)
	}

	return nil
}
