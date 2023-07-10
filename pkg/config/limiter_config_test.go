package config

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestLimiterConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	// global ca certificate
	caCrt, err := ioutil.TempFile("", "")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(caCrt.Close()).ToNot(HaveOccurred())
	defer os.RemoveAll(caCrt.Name())

	// global test key
	serverKey, err := ioutil.TempFile("", "")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(serverKey.Close()).ToNot(HaveOccurred())
	defer os.RemoveAll(serverKey.Name())

	// global test cert
	serverCRT, err := ioutil.TempFile("", "")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(serverCRT.Close()).ToNot(HaveOccurred())
	defer os.RemoveAll(serverCRT.Name())

	baseArgs := []string{"-limiter-server-key", serverKey.Name(), "-limiter-server-crt", serverCRT.Name()}

	t.Run("It returns an error if the 'limiter-server-key' is not set", func(t *testing.T) {
		cfg, err := Limiter(nil)
		g.Expect(cfg).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("param 'limiter-server-key' is not set"))
	})

	t.Run("It returns an error if the 'limiter-server-crt' is not set", func(t *testing.T) {
		cfg, err := Limiter([]string{"-limiter-server-key", serverKey.Name()})
		g.Expect(cfg).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("param 'limiter-server-crt' is not set"))
	})

	t.Run("limter-server keys can be set via env vars", func(t *testing.T) {
		os.Setenv("LIMITER_SERVER_KEY", serverKey.Name())
		os.Setenv("LIMITER_SERVER_CRT", serverCRT.Name())
		defer func() {
			defer os.Unsetenv("LIMITER_SERVER_KEY")
			defer os.Unsetenv("LIMITER_SERVER_CRT")
		}()

		cfg, err := Limiter(nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(*cfg.LimiterServerKey).To(Equal(serverKey.Name()))
		g.Expect(*cfg.LimiterServerCRT).To(Equal(serverCRT.Name()))
	})

	t.Run("Describe CA certificate", func(t *testing.T) {
		t.Run("It can be set via command line params", func(t *testing.T) {
			cfg, err := Limiter(append(baseArgs, "-limiter-ca", caCrt.Name()))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(*cfg.LimiterCA).To(Equal(caCrt.Name()))
		})

		t.Run("It can be set via an env var", func(t *testing.T) {
			os.Setenv("LIMITER_CA", caCrt.Name())
			defer func() {
				defer os.Unsetenv("LOG_LEVEL")
			}()

			cfg, err := Limiter(baseArgs)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(*cfg.LimiterCA).To(Equal(caCrt.Name()))
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
				g.Expect(err.Error()).To(Equal("param 'log-level' is invalid: 'bad'. Must be set to [debug | info]"))
			})
		})
	})

	t.Run("Describe server validation", func(t *testing.T) {
		t.Run("Context limiter-port", func(t *testing.T) {
			t.Run("It can be set via command line params", func(t *testing.T) {
				cfg, err := Limiter(append(baseArgs, "-limiter-port", "9001"))
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
				cfg, err := Limiter(append(baseArgs, "-limiter-port", "bad"))
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("error parsing 'limiter-port'"))
			})

			t.Run("It returns an error if the value if the port is bad", func(t *testing.T) {
				cfg, err := Limiter(append(baseArgs, "-limiter-port", "100000"))
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(Equal("param 'limiter-port' is invalid: '100000'. Must be set to a proper port below 65536"))
			})
		})
	})
}
