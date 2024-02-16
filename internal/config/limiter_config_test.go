package config

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestLimiterConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	// global ca certificate
	caCrt, err := os.CreateTemp("", "")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(caCrt.Close()).ToNot(HaveOccurred())
	defer os.RemoveAll(caCrt.Name())

	// global test key
	serverKey, err := os.CreateTemp("", "")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(serverKey.Close()).ToNot(HaveOccurred())
	defer os.RemoveAll(serverKey.Name())

	// global test cert
	serverCRT, err := os.CreateTemp("", "")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(serverCRT.Close()).ToNot(HaveOccurred())
	defer os.RemoveAll(serverCRT.Name())

	baseArgs := []string{"-server-key", serverKey.Name(), "-server-crt", serverCRT.Name()}

	t.Run("Describe Limiter server", func(t *testing.T) {
		t.Run("Context insecure-http", func(t *testing.T) {
			t.Run("It can be set via command line", func(t *testing.T) {
				cfg, err := Limiter([]string{"-insecure-http"})
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.InsecureHttp).To(BeTrue())
			})

			t.Run("It can be set via env vars", func(t *testing.T) {
				os.Setenv("LIMITER_INSECURE_HTTP", "true")
				defer os.Unsetenv("LIMITER_INSECURE_HTTP")

				cfg, err := Limiter([]string{})
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.InsecureHttp).To(BeTrue())
			})

			t.Run("It reurns an err if server-ca is set", func(t *testing.T) {
				os.Setenv("LIMITER_INSECURE_HTTP", "true")
				defer os.Unsetenv("LIMITER_INSECURE_HTTP")

				cfg, err := Limiter([]string{"-server-ca", caCrt.Name()})
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("flag 'server-ca' cannot be set with 'insecure-http'"))
			})

			t.Run("It reurns an err if server-key is set", func(t *testing.T) {
				os.Setenv("LIMITER_INSECURE_HTTP", "true")
				defer os.Unsetenv("LIMITER_INSECURE_HTTP")

				cfg, err := Limiter([]string{"-server-key", serverKey.Name()})
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("flag 'server-key' cannot be set with 'insecure-http'"))
			})

			t.Run("It reurns an err if server-crt is set", func(t *testing.T) {
				os.Setenv("LIMITER_INSECURE_HTTP", "true")
				defer os.Unsetenv("LIMITER_INSECURE_HTTP")

				cfg, err := Limiter([]string{"-server-crt", serverCRT.Name()})
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("flag 'server-crt' cannot be set with 'insecure-http'"))
			})
		})

		t.Run("Context when using https", func(t *testing.T) {
			t.Run("Context server-crt", func(t *testing.T) {
				t.Run("It returns an error when not set", func(t *testing.T) {
					cfg, err := Limiter([]string{"-server-key", serverKey.Name()})
					g.Expect(cfg).To(BeNil())

					g.Expect(err).To(HaveOccurred())
					g.Expect(err.Error()).To(Equal("flag 'server-crt' is not set"))
				})

				t.Run("It can be set via env vars", func(t *testing.T) {
					os.Setenv("LIMITER_SERVER_CRT", serverCRT.Name())
					defer os.Unsetenv("LIMITER_SERVER_CRT")

					cfg, err := Limiter([]string{"-server-key", serverKey.Name()})
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(*cfg.ServerKey).To(Equal(serverKey.Name()))
				})
			})

			t.Run("Context server-key", func(t *testing.T) {
				t.Run("It returns an error if the 'locker-server-key' is not set", func(t *testing.T) {
					cfg, err := Limiter([]string{"-server-crt", serverCRT.Name()})
					g.Expect(cfg).To(BeNil())

					g.Expect(err).To(HaveOccurred())
					g.Expect(err.Error()).To(Equal("flag 'server-key' is not set"))
				})

				t.Run("It can be set via env vars", func(t *testing.T) {
					os.Setenv("LIMITER_SERVER_KEY", serverKey.Name())
					defer os.Unsetenv("LIMITER_SERVER_KEY")

					cfg, err := Limiter([]string{"-server-crt", serverCRT.Name()})
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(*cfg.ServerKey).To(Equal(serverKey.Name()))
				})
			})

			t.Run("Describe server-ca", func(t *testing.T) {
				t.Run("It can be set via command line params", func(t *testing.T) {
					cfg, err := Limiter(append(baseArgs, "-server-ca", caCrt.Name()))
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(*cfg.ServerCA).To(Equal(caCrt.Name()))
				})

				t.Run("It can be set via an env var", func(t *testing.T) {
					os.Setenv("LIMITER_SERVER_CA", caCrt.Name())
					defer os.Unsetenv("LIMITER_SERVER_CA")

					cfg, err := Limiter(baseArgs)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(*cfg.ServerCA).To(Equal(caCrt.Name()))
				})
			})
		})

		t.Run("Context port", func(t *testing.T) {
			t.Run("It can be set via command line params", func(t *testing.T) {
				cfg, err := Limiter(append(baseArgs, "-port", "9001"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LimiterPort).To(Equal("9001"))
			})

			t.Run("It can be set via an env var", func(t *testing.T) {
				os.Setenv("LIMITER_PORT", "9001")
				defer func() {
					os.Unsetenv("LIMITER_PORT")
				}()

				cfg, err := Limiter(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LimiterPort).To(Equal("9001"))
			})

			t.Run("It returns an error if the value if the port is not an int", func(t *testing.T) {
				cfg, err := Limiter(append(baseArgs, "-port", "bad"))
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("error parsing 'limiter-port'"))
			})

			t.Run("It returns an error if the value if the port is bad", func(t *testing.T) {
				cfg, err := Limiter(append(baseArgs, "-port", "100000"))
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(Equal("flag 'limiter-port' is invalid: '100000'. Must be set to a proper port below 65536"))
			})
		})
	})

	t.Run("Describe log validation", func(t *testing.T) {
		t.Run("Context log-level", func(t *testing.T) {
			t.Run("It can be set via command line params", func(t *testing.T) {
				cfg, err := Limiter(append(baseArgs, "-log-level", "debug"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.logLevel).To(Equal("debug"))
			})

			t.Run("It can be set via an env var", func(t *testing.T) {
				os.Setenv("LOG_LEVEL", "debug")
				defer func() {
					defer os.Unsetenv("LOG_LEVEL")
				}()

				cfg, err := Limiter(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.logLevel).To(Equal("debug"))
			})

			t.Run("It returns an error if the value is invalid", func(t *testing.T) {
				cfg, err := Limiter(append(baseArgs, "-log-level", "bad"))
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(Equal("flag 'log-level' is invalid: 'bad'. Must be set to [debug | info]"))
			})
		})
	})
}
