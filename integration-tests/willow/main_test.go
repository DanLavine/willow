package willow_integration_tests

import (
	"os"
	"testing"

	"github.com/onsi/gomega/gexec"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()

	// run cleanup
	// NOTE: this cannot be in a defer function as `os.Exit()` does not run the defer functions
	gexec.CleanupBuildArtifacts()

	// exit with test exit code
	os.Exit(exitCode)
}
