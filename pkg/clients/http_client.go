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
	Do(request *http.Request) (*http.Response, error)

	// SsetupRequest with any headers setup for the clients configuration
	SetupRequest(method, url string) (*http.Request, error)

	// EncodeModel
	EncodedRequest(method, url string, Model api.APIObject) (*http.Request, error)

	// EncodeModelWithCancel
	EncodedRequestWithCancel(ctx context.Context, method, url string, Model api.APIObject) (*http.Request, error)
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

// SetupRequest to the client's configured settings. This includes any headers for encoding
// the client is expecting to recieve on an error or successful response
func (httpClient *httpClient) SetupRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unexpected error setting up http request: %w", err)
	}

	req.Header.Add("Content-Type", string(httpClient.contentType))

	return req, nil
}

// EncodeRequest to the client's configured settings. The returnd HTTP request has a requried header set
// 'Content-Type' for services to recognize, but additional headers can be added such as the 'x_request_id'
// to add a trace log id for each service to track a single request
func (httpClient *httpClient) EncodedRequest(method, url string, model api.APIObject) (*http.Request, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	var err error
	var req *http.Request

	switch httpClient.contentType {
	case api.ContentTypeJSON:
		// validate the model
		if err := model.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate model localy, not sending request: %w", err)
		}

		// encode the mode
		data, err := model.EncodeJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to encode the model local, not sending request: %w", err)
		}

		req, err = http.NewRequestWithContext(context.Background(), method, url, bytes.NewBuffer(data))
		if err != nil {
			return nil, fmt.Errorf("unexpected error setting up http request: %w", err)
		}

		req.Header.Add("Content-Type", string(httpClient.contentType))
	}

	return req, err
}

// EncodeRequest to the client's configured settings. The returnd HTTP request has a requried header set
// 'Content-Type' for services to recognize, but additional headers can be added such as the 'x_request_id'
// to add a trace log id for each service to track a single request
func (httpClient *httpClient) EncodedRequestWithCancel(ctx context.Context, method, url string, model api.APIObject) (*http.Request, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	var err error
	var req *http.Request

	switch httpClient.contentType {
	case api.ContentTypeJSON:
		// validate the model
		if err := model.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate model localy, not sending request: %w", err)
		}

		// encode the mode
		data, err := model.EncodeJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to encode the model local, not sending request: %w", err)
		}

		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
		if err != nil {
			return nil, fmt.Errorf("unexpected error setting up http request: %w", err)
		}

		req.Header.Add("Content-Type", string(httpClient.contentType))
	}

	return req, err
}

// Do makes a request to the remote service
func (httpClient *httpClient) Do(request *http.Request) (*http.Response, error) {
	return httpClient.client.Do(request)
}
