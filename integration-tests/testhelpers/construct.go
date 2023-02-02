package testhelpers

import (
	"bytes"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"golang.org/x/net/http2"

	. "github.com/onsi/gomega"
)

type IntegrationTestConstruct struct {
	dataDir string

	ServerPath   string
	ServerExe    *exec.Cmd
	ServerStdout *bytes.Buffer
	ServerStderr *bytes.Buffer

	ServerClient  *http.Client
	serverAddress string

	MetricsClient  *http.Client
	metricsAddress string
}

func NewIntrgrationTestConstruct(g *gomega.WithT) *IntegrationTestConstruct {
	willowPath, err := gexec.Build("github.com/DanLavine/willow/cmd/willow")
	g.Expect(err).ToNot(HaveOccurred())

	client := &http.Client{
		Transport: &http2.Transport{
			// so we don't complain about not being https
			AllowHTTP: true,
			// bypass TLS for now till we know how to set this up properly
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}

	return &IntegrationTestConstruct{
		ServerPath:     willowPath,
		ServerClient:   client,
		serverAddress:  "http://127.0.0.1:8080",
		MetricsClient:  &http.Client{},
		metricsAddress: "http://127.0.0.1:8081",
	}
}

func (itc *IntegrationTestConstruct) Start(g *gomega.WithT) {
	tmpDir, err := os.MkdirTemp("", "")
	g.Expect(err).ToNot(HaveOccurred())

	willowExe := exec.Command(itc.ServerPath, "-log-level", "debug", "-disk-storage-dir", tmpDir)
	itc.ServerStdout = new(bytes.Buffer)
	itc.ServerStderr = new(bytes.Buffer)
	willowExe.Stdout = itc.ServerStdout
	willowExe.Stderr = itc.ServerStderr

	err = willowExe.Start()
	g.Expect(err).ToNot(HaveOccurred())

	g.Eventually(func() string {
		return itc.ServerStdout.String()
	}).Should(ContainSubstring("TCP server running"))

	// record start configuration
	itc.dataDir = tmpDir
	itc.ServerExe = willowExe
}

func (itc *IntegrationTestConstruct) Shutdown(g *gomega.WithT) {
	g.Expect(os.RemoveAll(itc.dataDir)).ToNot(HaveOccurred())

	err := itc.ServerExe.Process.Signal(os.Interrupt)
	g.Expect(err).ToNot(HaveOccurred())

	g.Eventually(func() int {
		processState, _ := itc.ServerExe.Process.Wait()
		return processState.ExitCode()
	}).Should(Equal(0))
}
