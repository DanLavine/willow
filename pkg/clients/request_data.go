package clients

import (
	"net/http"
)

func AppendHeaders(req *http.Request, headers http.Header) {
	for headerKey, headerValues := range headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerKey, headerValue)
		}
	}
}
