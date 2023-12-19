package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"golang.org/x/net/http2"
)

// HTTP Client is used to make request to any of the Willow services
type HttpClient interface {
	// Perform an HTTP request that is expected to retun right away
	Do(requestData *RequestData) (*http.Response, error)

	// Perform an HTTP request that can take a while and can be canceld with the ctx
	DoWithContext(ctx context.Context, requestData *RequestData) (*http.Response, error)
}

type httpClient struct {
	client *http.Client

	contentType api.ContentType
}

// Genreate a new HTTP Client from a valid configuration
func NewHTTPClient(cfg *Config) (*httpClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client := &http.Client{}

	if cfg.CAFile != "" {
		client.Transport = &http2.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cfg.Cert()},
				RootCAs:      cfg.RootCAs(),
			},
		}
	}

	return &httpClient{
		client:      client,
		contentType: cfg.ContentType,
	}, nil
}

// Do makes a request to the remote service
func (httpClient *httpClient) Do(requestData *RequestData) (*http.Response, error) {
	return httpClient.do(context.Background(), requestData)
}

// Do makes a request to the remote service
func (httpClient *httpClient) DoWithContext(ctx context.Context, requestData *RequestData) (*http.Response, error) {
	return httpClient.do(ctx, requestData)
}

// Do makes a request to the remote service
func (httpClient *httpClient) do(ctx context.Context, requestData *RequestData) (*http.Response, error) {
	var err error
	var req *http.Request

	// always validate the requested model first
	if requestData.Model != nil {
		if err = requestData.Model.Validate(); err != nil {
			return nil, errors.ClientError(err)
		}
	}

	// make the request bassed off the content type
	switch httpClient.contentType {
	case api.ContentTypeJSON:
		// setup request with proper encoding
		if requestData.Model != nil {
			if ctx == context.Background() {
				req, err = http.NewRequest(requestData.Method, requestData.Path, bytes.NewBuffer(requestData.Model.EncodeJSON()))
			} else {
				req, err = http.NewRequestWithContext(ctx, requestData.Method, requestData.Path, bytes.NewBuffer(requestData.Model.EncodeJSON()))
			}

			// add the proper header for the content type. Responses still need to know what value to respond back as
		} else {
			if ctx == context.Background() {
				req, err = http.NewRequest(requestData.Method, requestData.Path, nil)
			} else {
				req, err = http.NewRequestWithContext(ctx, requestData.Method, requestData.Path, nil)

			}
		}

		if err != nil {
			return nil, errors.ClientError(err)
		}
	}

	// add the proper header for the content type. Any API can hit an error and still needs to know how to process correctly
	req.Header.Add("Content-Type", string(httpClient.contentType))

	// make the request
	return httpClient.client.Do(req)
}
