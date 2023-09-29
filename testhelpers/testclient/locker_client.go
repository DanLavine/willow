package testclient

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/net/http2"

	. "github.com/onsi/gomega"
)

type LockerClient struct {
	address string

	httpClient *http.Client
}

func NewLockerClient(g *WithT, address string) *LockerClient {
	_, currentDir, _, _ := runtime.Caller(0)

	// parse root ca
	caPem, err := os.ReadFile(filepath.Join(currentDir, "..", "..", "tls-keys", "ca.crt"))
	g.Expect(err).ToNot(HaveOccurred())

	rootCAs := x509.NewCertPool()
	g.Expect(rootCAs.AppendCertsFromPEM(caPem)).To(BeTrue())

	// parse client certs
	cert, err := tls.LoadX509KeyPair(filepath.Join(currentDir, "..", "..", "tls-keys", "client.crt"), filepath.Join(currentDir, "..", "..", "tls-keys", "client.key"))
	g.Expect(err).ToNot(HaveOccurred())

	// make a request to the server
	return &LockerClient{
		address: address,
		httpClient: &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
					RootCAs:      rootCAs,
				},
			},
		},
	}
}

func (lc *LockerClient) Do(request *http.Request) (*http.Response, error) {
	return lc.httpClient.Do(request)
}

func (lc *LockerClient) Address() string {
	return lc.address
}

func (lc *LockerClient) Transport() *tls.Config {
	return lc.httpClient.Transport.(*http2.Transport).TLSClientConfig
}

func (lc *LockerClient) Disconnect() {
	lc.httpClient.CloseIdleConnections()
}
