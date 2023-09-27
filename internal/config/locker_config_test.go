package config

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestLockerConfig(t *testing.T) {
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

	baseArgs := []string{"-Locker-server-key", serverKey.Name(), "-Locker-server-crt", serverCRT.Name()}

	t.Run("It returns an error if the 'Locker-server-key' is not set", func(t *testing.T) {
		cfg, err := Locker(nil)
		g.Expect(cfg).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("param 'Locker-server-key' is not set"))
	})

	t.Run("It returns an error if the 'Locker-server-crt' is not set", func(t *testing.T) {
		cfg, err := Locker([]string{"-Locker-server-key", serverKey.Name()})
		g.Expect(cfg).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("param 'Locker-server-crt' is not set"))
	})

	t.Run("limter-server keys can be set via env vars", func(t *testing.T) {
		os.Setenv("Locker_SERVER_KEY", serverKey.Name())
		os.Setenv("Locker_SERVER_CRT", serverCRT.Name())
		defer func() {
			defer os.Unsetenv("Locker_SERVER_KEY")
			defer os.Unsetenv("Locker_SERVER_CRT")
		}()

		cfg, err := Locker(nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(*cfg.LockerServerKey).To(Equal(serverKey.Name()))
		g.Expect(*cfg.LockerServerCRT).To(Equal(serverCRT.Name()))
	})

	t.Run("Describe CA certificate", func(t *testing.T) {
		t.Run("It can be set via command line params", func(t *testing.T) {
			cfg, err := Locker(append(baseArgs, "-Locker-ca", caCrt.Name()))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(*cfg.LockerCA).To(Equal(caCrt.Name()))
		})

		t.Run("It can be set via an env var", func(t *testing.T) {
			os.Setenv("Locker_CA", caCrt.Name())
			defer func() {
				defer os.Unsetenv("LOG_LEVEL")
			}()

			cfg, err := Locker(baseArgs)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(*cfg.LockerCA).To(Equal(caCrt.Name()))
		})
	})

	t.Run("Describe log validation", func(t *testing.T) {
		t.Run("Context log-level", func(t *testing.T) {
			t.Run("It can be set via command line params", func(t *testing.T) {
				cfg, err := Locker(append(baseArgs, "-log-level", "debug"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.logLevel).To(Equal("debug"))
			})

			t.Run("It can be set via an env var", func(t *testing.T) {
				os.Setenv("LOG_LEVEL", "debug")
				defer func() {
					defer os.Unsetenv("LOG_LEVEL")
				}()

				cfg, err := Locker(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.logLevel).To(Equal("debug"))
			})

			t.Run("It returns an error if the value is invalid", func(t *testing.T) {
				cfg, err := Locker(append(baseArgs, "-log-level", "bad"))
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(Equal("param 'log-level' is invalid: 'bad'. Must be set to [debug | info]"))
			})
		})
	})

	t.Run("Describe server validation", func(t *testing.T) {
		t.Run("Context Locker-port", func(t *testing.T) {
			t.Run("It can be set via command line params", func(t *testing.T) {
				cfg, err := Locker(append(baseArgs, "-Locker-port", "9001"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LockerPort).To(Equal("9001"))
			})

			t.Run("It can be set via an env var", func(t *testing.T) {
				os.Setenv("Locker_PORT", "9001")
				defer func() {
					os.Unsetenv("Locker_PORT")
				}()

				cfg, err := Locker(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LockerPort).To(Equal("9001"))
			})

			t.Run("It returns an error if the value if the port is not an int", func(t *testing.T) {
				cfg, err := Locker(append(baseArgs, "-Locker-port", "bad"))
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("error parsing 'Locker-port'"))
			})

			t.Run("It returns an error if the value if the port is bad", func(t *testing.T) {
				cfg, err := Locker(append(baseArgs, "-Locker-port", "100000"))
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(Equal("param 'Locker-port' is invalid: '100000'. Must be set to a proper port below 65536"))
			})
		})
	})
}
