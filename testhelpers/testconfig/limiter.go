package testconfig

import (
	"path/filepath"
	"runtime"

	"github.com/DanLavine/willow/pkg/config"

	. "github.com/onsi/gomega"
)

func Limiter(g *WithT) *config.LimiterConfig {
	_, currentDir, _, _ := runtime.Caller(0)

	args := []string{
		"-log-level", "debug",
		"-limiter-port", "8080",
		"-limiter-ca", filepath.Join(currentDir, "..", "..", "tls-keys", "ca.crt"),
		"-limiter-server-key", filepath.Join(currentDir, "..", "..", "tls-keys", "server.key"),
		"-limiter-server-crt", filepath.Join(currentDir, "..", "..", "tls-keys", "server.crt"),
	}

	config, err := config.Limiter(args)
	g.Expect(err).ToNot(HaveOccurred())

	return config
}
