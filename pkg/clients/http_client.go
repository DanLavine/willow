package clients

import (
	"crypto/tls"
	"net/http"

	"golang.org/x/net/http2"
)

// Genreate a new HTTP Client from a valid configuration
func NewHTTPClient(cfg *Config) (*http.Client, error) {
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

	return client, nil
}

// // NOW that I have a proper encoder, this seems a bit over kill and not needed anymore. Just need to
// // setup the possible headers properly and setup the clien'ts certs properly.

// // HTTP Client is used to make request to any of the Willow services
// type HttpClient interface {
// 	// Perform an HTTP request that is expected to retun right away
// 	Do(request *http.Request) (*http.Response, error)

// 	// SetupRequest with any headers setup for the clients configuration
// 	SetupRequest(method, url string) (*http.Request, error)

// 	EncodedRequestWithoutValidation(method, url string, Model any) (*http.Request, error)

// 	// EncodeModel into a http request
// 	EncodedRequest(method, url string, Model api.APIObject) (*http.Request, error)

// 	// EncodeModelWithCancel into a http request
// 	EncodedRequestWithCancel(ctx context.Context, method, url string, Model api.APIObject) (*http.Request, error)
// }

// type httpClient struct {
// 	client *http.Client

// 	encoder encoding.Encoder
// }

// // Genreate a new HTTP Client from a valid configuration
// func NewHTTPClient(cfg *Config) (*httpClient, error) {
// 	if err := cfg.Validate(); err != nil {
// 		return nil, err
// 	}

// 	client := &http.Client{}

// 	if cfg.CAFile != "" {
// 		client.Transport = &http2.Transport{
// 			TLSClientConfig: &tls.Config{
// 				Certificates: []tls.Certificate{cfg.Cert()},
// 				RootCAs:      cfg.RootCAs(),
// 			},
// 		}
// 	}

// 	encoder, err := encoding.NewEncoder(cfg.ContentEncoding)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &httpClient{
// 		client:  client,
// 		encoder: encoder,
// 	}, nil
// }

// // SetupRequest to the client's configured settings. This includes any headers for encoding
// // the client is expecting to recieve on an error or successful response
// func (httpClient *httpClient) SetupRequest(method, url string) (*http.Request, error) {
// 	req, err := http.NewRequestWithContext(context.Background(), method, url, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("unexpected error setting up http request: %w", err)
// 	}

// 	req.Header.Add("Content-Type", httpClient.encoder.ContentType())
// 	return req, nil
// }

// // EncodeRequest to the client's configured settings. The returnd HTTP request has a requried header set
// // 'Content-Type' for services to recognize, but additional headers can be added such as the 'x_request_id'
// // to add a trace log id for each service to track a single request
// func (httpClient *httpClient) EncodedRequestWithoutValidation(method, url string, model any) (*http.Request, error) {
// 	if model == nil {
// 		return nil, fmt.Errorf("model cannot be nil")
// 	}

// 	data, encodeErr := httpClient.encoder.Encode(model)
// 	if encodeErr != nil {
// 		return nil, encodeErr
// 	}

// 	req, err := http.NewRequestWithContext(context.Background(), method, url, bytes.NewBuffer(data))
// 	if err != nil {
// 		return nil, fmt.Errorf("unexpected error setting up http request: %w", err)
// 	}

// 	req.Header.Add("Content-Type", httpClient.encoder.ContentType())
// 	return req, nil
// }

// // EncodeRequest to the client's configured settings. The returnd HTTP request has a requried header set
// // 'Content-Type' for services to recognize, but additional headers can be added such as the 'x_request_id'
// // to add a trace log id for each service to track a single request
// func (httpClient *httpClient) EncodedRequest(method, url string, model encoding.APIModel) (*http.Request, error) {
// 	if model == nil {
// 		return nil, fmt.Errorf("model cannot be nil")
// 	}

// 	if err := model.Validate(); err != nil {
// 		return nil, err
// 	}

// 	data, encodeErr := httpClient.encoder.Encode(model)
// 	if encodeErr != nil {
// 		return nil, encodeErr
// 	}

// 	req, err := http.NewRequestWithContext(context.Background(), method, url, bytes.NewBuffer(data))
// 	if err != nil {
// 		return nil, fmt.Errorf("unexpected error setting up http request: %w", err)
// 	}

// 	req.Header.Add("Content-Type", httpClient.encoder.ContentType())
// 	return req, nil
// }

// // EncodeRequest to the client's configured settings. The returnd HTTP request has a requried header set
// // 'Content-Type' for services to recognize, but additional headers can be added such as the 'x_request_id'
// // to add a trace log id for each service to track a single request
// func (httpClient *httpClient) EncodedRequestWithCancel(ctx context.Context, method, url string, model encoding.APIModel) (*http.Request, error) {
// 	if model == nil {
// 		return nil, fmt.Errorf("model cannot be nil")
// 	}

// 	if err := model.Validate(); err != nil {
// 		return nil, err
// 	}

// 	data, err := httpClient.encoder.Encode(model)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
// 	if err != nil {
// 		return nil, fmt.Errorf("unexpected error setting up http request: %w", err)
// 	}

// 	req.Header.Add("Content-Type", httpClient.encoder.ContentType())
// 	return req, err
// }

// // Do makes a request to the remote service
// func (httpClient *httpClient) Do(request *http.Request) (*http.Response, error) {
// 	return httpClient.client.Do(request)
// }
