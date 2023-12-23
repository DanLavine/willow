package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api"
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

	contentType string
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
		contentType: cfg.ContentEncoding,
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

	// make the request bassed off the content type
	switch httpClient.contentType {
	case api.ContentTypeJSON:

		// setup request with proper encoding
		if requestData.Model != nil {
			// validate the model
			if err := requestData.Model.Validate(); err != nil {
				return nil, fmt.Errorf("failed to validate model localy, not sending request: %w", err)
			}

			// encode the mode
			data, err := requestData.Model.EncodeJSON()
			if err != nil {
				return nil, fmt.Errorf("failed to encode the model local, not sending request: %w", err)
			}

			req, err = http.NewRequestWithContext(ctx, requestData.Method, requestData.Path, bytes.NewBuffer(data))
			if err != nil {
				return nil, fmt.Errorf("unexpected error setting up http request: %w", err)
			}
		} else {
			req, err = http.NewRequestWithContext(ctx, requestData.Method, requestData.Path, nil)
			if err != nil {
				return nil, fmt.Errorf("unexpected error setting up http request: %w", err)
			}
		}
	}

	// add the proper header for the content type. Any API can hit an error and still needs to know how to process correctly
	req.Header.Add("Content-Type", string(httpClient.contentType))

	// make the request
	return httpClient.client.Do(req)
}
