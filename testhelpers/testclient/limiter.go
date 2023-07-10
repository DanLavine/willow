package testclient

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"

	. "github.com/onsi/gomega"
	"golang.org/x/net/http2"
)

func Limiter(g *WithT) *http.Client {
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
	return &http.Client{
		Transport: &http2.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      rootCAs,
			},
		},
	}
}
