package testhelpers

import (
	"bytes"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"golang.org/x/net/http2"

	. "github.com/onsi/gomega"
)

type IntegrationTestConstruct struct {
	dataDir string

	ServerPath   string
	Session      *gexec.Session
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
	itc.dataDir = tmpDir

	//willowExe := exec.Command(itc.ServerPath, "-log-level", "debug", "-disk-storage-dir", tmpDir)
	willowExe := exec.Command(itc.ServerPath, "-log-level", "debug")

	itc.ServerStdout = new(bytes.Buffer)
	itc.ServerStderr = new(bytes.Buffer)
	session, err := gexec.Start(willowExe, itc.ServerStdout, itc.ServerStderr)
	g.Expect(err).ToNot(HaveOccurred())
	g.Eventually(session.Out).Should(gbytes.Say("TCP server running"))

	// record start configuration
	itc.Session = session
}

func (itc *IntegrationTestConstruct) Shutdown(g *gomega.WithT) {
	session := itc.Session.Interrupt()
	g.Eventually(session).Should(gexec.Exit(0))

	g.Expect(os.RemoveAll(itc.dataDir)).ToNot(HaveOccurred())
}

func (itc *IntegrationTestConstruct) ServerAddress() string {
	return itc.serverAddress
}

func (itc *IntegrationTestConstruct) Cleanup(g *gomega.WithT) {
	gexec.CleanupBuildArtifacts()
}
