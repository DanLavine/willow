package integrationhelpers

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/DanLavine/willow/testhelpers/testclient"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/gomega"
)

var (
	lockerPath  string
	limiterPath string
	willowPath  string
)

// only want to build the executables once per integration run
func init() {
	var err error
	// create locker
	lockerPath, err = gexec.Build("github.com/DanLavine/willow/cmd/locker", "--race")
	if err != nil {
		panic(err)
	}

	// create limiter
	limiterPath, err = gexec.Build("github.com/DanLavine/willow/cmd/limiter", "--race")
	if err != nil {
		panic(err)
	}

	// create willow
	willowPath, err = gexec.Build("github.com/DanLavine/willow/cmd/willow", "--race")
	if err != nil {
		panic(err)
	}
}

type IntegrationTestConstruct struct {
	ServerPath   string
	ServerURL    string
	Session      *gexec.Session
	ServerStdout *bytes.Buffer
	ServerStderr *bytes.Buffer

	// generic http client that can be used for manual requests
	ServerClient *testclient.Client
}

func getFreePort(g *WithT) int {
	listener, err := net.Listen("tcp", ":0")
	g.Expect(err).ToNot(HaveOccurred())

	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

func StartLocker(g *WithT) *IntegrationTestConstruct {
	freePort := getFreePort(g)

	testConstruct := &IntegrationTestConstruct{
		ServerURL:    fmt.Sprintf("https://127.0.0.1:%d", freePort),
		ServerClient: testclient.New(g, fmt.Sprintf("https://127.0.0.1:%d", freePort)),
	}

	_, currentDir, _, _ := runtime.Caller(0)

	cmdLineFlags := []string{
		"-port", fmt.Sprintf("%d", freePort),
		"-log-level", "debug",
		"-server-ca", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		"-server-key", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "server.key"),
		"-server-crt", filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "server.crt"),
		"-lock-default-timeout", "1s",
	}

	lockerExe := exec.Command(lockerPath, cmdLineFlags...)

	testConstruct.ServerStdout = new(bytes.Buffer)
	testConstruct.ServerStderr = new(bytes.Buffer)
	session, err := gexec.Start(lockerExe, testConstruct.ServerStdout, testConstruct.ServerStderr)
	g.Expect(err).ToNot(HaveOccurred())
	g.Eventually(session.Out).Should(gbytes.Say("Locker TCP server running"))
	time.Sleep(100 * time.Millisecond)

	// record start configuration
	testConstruct.Session = session

	return testConstruct
}

func StartLimiter(g *WithT, lockerURL string) *IntegrationTestConstruct {
	freePort := getFreePort(g)

	testConstruct := &IntegrationTestConstruct{
		ServerURL:    fmt.Sprintf("https://127.0.0.1:%d", freePort),
		ServerClient: testclient.New(g, fmt.Sprintf("https://127.0.0.1:%d", freePort)),
	}

	_, currentDir, _, _ := runtime.Caller(0)

	cmdLineFlags := []string{
		"-port", fmt.Sprintf("%d", freePort),
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

	limiterExe := exec.Command(limiterPath, cmdLineFlags...)

	testConstruct.ServerStdout = new(bytes.Buffer)
	testConstruct.ServerStderr = new(bytes.Buffer)
	session, err := gexec.Start(limiterExe, testConstruct.ServerStdout, testConstruct.ServerStderr)
	g.Expect(err).ToNot(HaveOccurred())
	g.Eventually(session.Out).Should(gbytes.Say("Limiter TCP server running"))
	time.Sleep(100 * time.Millisecond)

	// record start configuration
	testConstruct.Session = session

	return testConstruct
}

func StartWillow(g *WithT, limiterURL string) *IntegrationTestConstruct {
	freePort := getFreePort(g)

	testConstruct := &IntegrationTestConstruct{
		ServerURL:    fmt.Sprintf("https://127.0.0.1:%d", freePort),
		ServerClient: testclient.New(g, fmt.Sprintf("https://127.0.0.1:%d", freePort)),
	}

	_, currentDir, _, _ := runtime.Caller(0)

	cmdLineFlags := []string{
		"-port", fmt.Sprintf("%d", freePort),
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

	willowExe := exec.Command(willowPath, cmdLineFlags...)

	testConstruct.ServerStdout = new(bytes.Buffer)
	testConstruct.ServerStderr = new(bytes.Buffer)
	session, err := gexec.Start(willowExe, testConstruct.ServerStdout, testConstruct.ServerStderr)
	g.Expect(err).ToNot(HaveOccurred())

	g.Eventually(session.Out).Should(gbytes.Say("Willow TCP server running"))
	time.Sleep(100 * time.Millisecond)

	// record start configuration
	testConstruct.Session = session

	return testConstruct
}

func (itc *IntegrationTestConstruct) Shutdown(g *WithT) {
	session := itc.Session.Interrupt()
	time.Sleep(time.Second)

	fmt.Println(string(itc.ServerStderr.String()))
	fmt.Println(string(itc.ServerStdout.String()))

	g.Eventually(session, "2s").Should(gexec.Exit(0), func() string { return itc.ServerStdout.String() + itc.ServerStderr.String() })
}

func (itc *IntegrationTestConstruct) Cleanup(g *WithT) {
	gexec.CleanupBuildArtifacts()
}
