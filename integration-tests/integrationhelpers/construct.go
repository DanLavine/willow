package integrationhelpers

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/DanLavine/willow/testhelpers/testclient"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/gomega"
)

type IntegrationTestConstruct struct {
	dataDir string

	ServerPath   string
	Session      *gexec.Session
	ServerStdout *bytes.Buffer
	ServerStderr *bytes.Buffer

	ServerClient *testclient.Client
}

func NewIntrgrationTestConstruct(g *WithT) *IntegrationTestConstruct {
	willowPath, err := gexec.Build("github.com/DanLavine/willow/cmd/willow")
	g.Expect(err).ToNot(HaveOccurred())

	return &IntegrationTestConstruct{
		ServerPath:   willowPath,
		ServerClient: testclient.New(g, "https://127.0.0.1:8080", "http://127.0.0.1:8081"),
	}
}

func (itc *IntegrationTestConstruct) Start(g *WithT) {
	tmpDir, err := os.MkdirTemp("", "")
	g.Expect(err).ToNot(HaveOccurred())
	itc.dataDir = tmpDir

	_, currentDir, _, _ := runtime.Caller(0)

	cmdLineFlags := []string{
		"-log-level", "debug",
		"-willow-ca", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		"-willow-server-key", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "server.key"),
		"-willow-server-crt", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "server.crt"),
	}

	willowExe := exec.Command(itc.ServerPath, cmdLineFlags...)

	itc.ServerStdout = new(bytes.Buffer)
	itc.ServerStderr = new(bytes.Buffer)
	session, err := gexec.Start(willowExe, itc.ServerStdout, itc.ServerStderr)
	g.Expect(err).ToNot(HaveOccurred())
	g.Eventually(session.Out).Should(gbytes.Say("Willow TCP server running"))

	// record start configuration
	itc.Session = session
}

func (itc *IntegrationTestConstruct) Shutdown(g *WithT) {
	session := itc.Session.Interrupt()
	g.Eventually(session, 2*time.Second).Should(gexec.Exit(0), string(itc.Session.Out.Contents()))

	g.Expect(os.RemoveAll(itc.dataDir)).ToNot(HaveOccurred())
}

func (itc *IntegrationTestConstruct) Cleanup(g *WithT) {
	gexec.CleanupBuildArtifacts()
}
