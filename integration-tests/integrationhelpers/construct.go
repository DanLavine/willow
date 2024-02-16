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
	ServerURL    string
	Session      *gexec.Session
	ServerStdout *bytes.Buffer
	ServerStderr *bytes.Buffer

	// generic http client that can be used for manual requests
	ServerClient *testclient.Client
}

func NewIntrgrationWillowTestConstruct(g *WithT) *IntegrationTestConstruct {
	willowPath, err := gexec.Build("github.com/DanLavine/willow/cmd/willow", "--race")
	g.Expect(err).ToNot(HaveOccurred())

	return &IntegrationTestConstruct{
		ServerPath:   willowPath,
		ServerURL:    "https://127.0.0.1:8080",
		ServerClient: testclient.New(g, "https://127.0.0.1:8080"),
	}
}

func NewIntrgrationLimiterTestConstruct(g *WithT) *IntegrationTestConstruct {
	limiterPath, err := gexec.Build("github.com/DanLavine/willow/cmd/limiter", "--race")
	g.Expect(err).ToNot(HaveOccurred())

	return &IntegrationTestConstruct{
		ServerPath:   limiterPath,
		ServerURL:    "https://127.0.0.1:8082",
		ServerClient: testclient.New(g, "https://127.0.0.1:8082"),
	}
}

func NewIntrgrationLockerTestConstruct(g *WithT) *IntegrationTestConstruct {
	lockerPath, err := gexec.Build("github.com/DanLavine/willow/cmd/locker", "--race")
	g.Expect(err).ToNot(HaveOccurred())

	return &IntegrationTestConstruct{
		ServerPath:   lockerPath,
		ServerURL:    "https://127.0.0.1:8083",
		ServerClient: testclient.New(g, "https://127.0.0.1:8083"),
	}
}

func (itc *IntegrationTestConstruct) StartLocker(g *WithT) {
	tmpDir, err := os.MkdirTemp("", "")
	g.Expect(err).ToNot(HaveOccurred())
	itc.dataDir = tmpDir

	_, currentDir, _, _ := runtime.Caller(0)

	cmdLineFlags := []string{
		"-log-level", "debug",
		"-server-ca", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		"-server-key", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "server.key"),
		"-server-crt", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "server.crt"),
		"-lock-default-timeout", "1s",
	}

	lockerExe := exec.Command(itc.ServerPath, cmdLineFlags...)

	itc.ServerStdout = new(bytes.Buffer)
	itc.ServerStderr = new(bytes.Buffer)
	session, err := gexec.Start(lockerExe, itc.ServerStdout, itc.ServerStderr)
	g.Expect(err).ToNot(HaveOccurred())
	g.Eventually(session.Out).Should(gbytes.Say("Locker TCP server running"))
	time.Sleep(100 * time.Millisecond)

	// record start configuration
	itc.Session = session
}

func (itc *IntegrationTestConstruct) StartLimiter(g *WithT, lockerURL string) {
	tmpDir, err := os.MkdirTemp("", "")
	g.Expect(err).ToNot(HaveOccurred())
	itc.dataDir = tmpDir

	_, currentDir, _, _ := runtime.Caller(0)

	cmdLineFlags := []string{
		"-log-level", "debug",
		"-server-ca", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		"-server-key", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "server.key"),
		"-server-crt", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "server.crt"),
		// configuration to point to the locker service
		"-locker-url", lockerURL,
		"-locker-client-ca", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		"-locker-client-key", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
		"-locker-client-crt", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
	}

	lockerExe := exec.Command(itc.ServerPath, cmdLineFlags...)

	itc.ServerStdout = new(bytes.Buffer)
	itc.ServerStderr = new(bytes.Buffer)
	session, err := gexec.Start(lockerExe, itc.ServerStdout, itc.ServerStderr)
	g.Expect(err).ToNot(HaveOccurred())
	g.Eventually(session.Out).Should(gbytes.Say("Limiter TCP server running"))
	time.Sleep(100 * time.Millisecond)

	// record start configuration
	itc.Session = session
}

func (itc *IntegrationTestConstruct) StartWillow(g *WithT, limiterURL string) {
	tmpDir, err := os.MkdirTemp("", "")
	g.Expect(err).ToNot(HaveOccurred())
	itc.dataDir = tmpDir

	_, currentDir, _, _ := runtime.Caller(0)

	cmdLineFlags := []string{
		"-log-level", "debug",
		"-server-ca", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		"-server-key", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "server.key"),
		"-server-crt", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "server.crt"),
		// configuration to point to limiter service
		"-limiter-url", limiterURL,
		"-limiter-client-ca", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		"-limiter-client-key", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
		"-limiter-client-crt", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
	}

	willowExe := exec.Command(itc.ServerPath, cmdLineFlags...)

	itc.ServerStdout = new(bytes.Buffer)
	itc.ServerStderr = new(bytes.Buffer)
	session, err := gexec.Start(willowExe, itc.ServerStdout, itc.ServerStderr)
	g.Expect(err).ToNot(HaveOccurred())

	g.Eventually(session.Out).Should(gbytes.Say("Willow TCP server running"))
	time.Sleep(100 * time.Millisecond)

	// record start configuration
	itc.Session = session
}

func (itc *IntegrationTestConstruct) Shutdown(g *WithT) {
	session := itc.Session.Interrupt()
	time.Sleep(time.Second)

	g.Eventually(session, "2s").Should(gexec.Exit(0), func() string { return itc.ServerStdout.String() + itc.ServerStderr.String() })

	g.Expect(os.RemoveAll(itc.dataDir)).ToNot(HaveOccurred())
}

func (itc *IntegrationTestConstruct) Cleanup(g *WithT) {
	gexec.CleanupBuildArtifacts()
}
