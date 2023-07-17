package testclient

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"

	"golang.org/x/net/http2"

	. "github.com/onsi/gomega"
)

type Client struct {
	address        string
	metricsAddress string

	httpClient    *http.Client
	metricsClient *http.Client
}

func New(g *WithT, address, metricsAddress string) *Client {
	_, currentDir, _, _ := runtime.Caller(0)

	// parse root ca
	caPem, err := ioutil.ReadFile(filepath.Join(currentDir, "..", "..", "tls-keys", "ca.crt"))
	g.Expect(err).ToNot(HaveOccurred())

	rootCAs := x509.NewCertPool()
	g.Expect(rootCAs.AppendCertsFromPEM(caPem)).To(BeTrue())

	// parse client certs
	cert, err := tls.LoadX509KeyPair(filepath.Join(currentDir, "..", "..", "tls-keys", "client.crt"), filepath.Join(currentDir, "..", "..", "tls-keys", "client.key"))
	g.Expect(err).ToNot(HaveOccurred())

	// make a request to the server
	return &Client{
		address:        address,
		metricsAddress: metricsAddress,
		httpClient: &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
					RootCAs:      rootCAs,
				},
			},
		},
		metricsClient: &http.Client{},
	}
}

func (c *Client) Do(request *http.Request) (*http.Response, error) {
	return c.httpClient.Do(request)
}
