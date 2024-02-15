package config

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestWillowConfig(t *testing.T) {
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

	baseArgs := []string{"-server-ca", caCrt.Name(), "-server-key", serverKey.Name(), "-server-crt", serverCRT.Name()}

	t.Run("Describe willow server", func(t *testing.T) {
		t.Run("Context insecure-http", func(t *testing.T) {
			t.Run("It can be set via command line", func(t *testing.T) {
				cfg, err := Willow([]string{"-insecure-http"})
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.InsecureHttp).To(BeTrue())
			})

			t.Run("It can be set via env vars", func(t *testing.T) {
				os.Setenv("WILLOW_INSECURE_HTTP", "true")
				defer os.Unsetenv("WILLOW_INSECURE_HTTP")

				cfg, err := Willow([]string{})
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.InsecureHttp).To(BeTrue())
			})

			t.Run("It errors if the ca cert is provided", func(t *testing.T) {
				cfg, err := Willow([]string{"-insecure-http", "-server-ca", caCrt.Name()})
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(Equal("parameter 'server-ca' is set, but also configured to run in plain http"))
			})

			t.Run("It errors if the ca crt is provided", func(t *testing.T) {
				cfg, err := Willow([]string{"-insecure-http", "-server-crt", serverCRT.Name()})
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(Equal("parameter 'server-crt' is set, but also configured to run in plain http"))
			})

			t.Run("It errors if the ca key is provided", func(t *testing.T) {
				cfg, err := Willow([]string{"-insecure-http", "-server-key", serverCRT.Name()})
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(Equal("parameter 'server-key' is set, but also configured to run in plain http"))
			})
		})

		t.Run("Context when using http2", func(t *testing.T) {
			t.Run("Context server-key", func(t *testing.T) {
				t.Run("It returns an error if there is no server-key", func(t *testing.T) {
					cfg, err := Willow(nil)
					g.Expect(cfg).To(BeNil())
					g.Expect(err).To(HaveOccurred())
					g.Expect(err.Error()).To(ContainSubstring("parameter 'server-key' is not set"))
				})

				t.Run("It can be set via env vars", func(t *testing.T) {
					os.Setenv("WILLOW_SERVER_KEY", serverKey.Name())
					defer os.Unsetenv("WILLOW_SERVER_KEY")

					cfg, err := Willow([]string{"-server-ca", caCrt.Name(), "-server-crt", serverCRT.Name()})
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(*cfg.ServerKey).To(Equal(serverKey.Name()))
				})
			})

			t.Run("Context server-crt", func(t *testing.T) {
				t.Run("It returns an error if there is no server-crt", func(t *testing.T) {
					cfg, err := Willow([]string{"-server-key", serverKey.Name()})
					g.Expect(cfg).To(BeNil())
					g.Expect(err).To(HaveOccurred())
					g.Expect(err.Error()).To(ContainSubstring("parameter 'server-crt' is not set"))
				})

				t.Run("It can be set via env vars", func(t *testing.T) {
					os.Setenv("WILLOW_SERVER_CRT", serverCRT.Name())
					defer os.Unsetenv("WILLOW_SERVER_CRT")

					cfg, err := Willow([]string{"-server-ca", caCrt.Name(), "-server-key", serverKey.Name()})
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(*cfg.ServerCRT).To(Equal(serverCRT.Name()))
				})
			})
		})

		t.Run("Context port", func(t *testing.T) {
			t.Run("It can be set via command line", func(t *testing.T) {
				cfg, err := Willow(append(baseArgs, "-port", "8765"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.Port).To(Equal("8765"))
			})

			t.Run("It can be set via env vars", func(t *testing.T) {
				os.Setenv("WILLOW_PORT", "8888")
				defer os.Unsetenv("WILLOW_PORT")

				cfg, err := Willow(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.Port).To(Equal("8888"))
			})
		})
	})

	t.Run("Describe storage validation", func(t *testing.T) {
		t.Run("Context memory", func(t *testing.T) {
			t.Run("It can be set via command line", func(t *testing.T) {
				cfg, err := Willow(append(baseArgs, "-storage-type", "memory"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.StorageConfig.Type).To(Equal("memory"))
			})

			t.Run("can be set by env var", func(t *testing.T) {
				os.Setenv("STORAGE_TYPE", "memory")
				defer os.Unsetenv("STORAGE_TYPE")

				cfg, err := Willow(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.StorageConfig.Type).To(Equal("memory"))
			})
		})

		t.Run("It returns an error for an unknown type", func(t *testing.T) {
			cfg, err := Willow(append(baseArgs, "-storage-type", "foo"))
			g.Expect(cfg).To(BeNil())
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("invalid storage type selected 'foo'"))
		})
	})

	t.Run("Describe logging", func(t *testing.T) {
		t.Run("Context log-level", func(t *testing.T) {
			t.Run("It can be set via command line", func(t *testing.T) {
				cfg, err := Willow(append(baseArgs, "-log-level", "debug"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(cfg.LogLevel()).To(Equal("debug"))
			})

			t.Run("It can be set via env vars", func(t *testing.T) {
				os.Setenv("WILLOW_LOG_LEVEL", "info")
				defer os.Unsetenv("WILLOW_LOG_LEVEL")

				cfg, err := Willow(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(cfg.LogLevel()).To(Equal("info"))
			})

			t.Run("It returns an error on unkown log level", func(t *testing.T) {
				cfg, err := Willow(append(baseArgs, "-log-level", "no idea"))
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("expected config 'LogLevel' to be [debug | info]. Received: 'no idea'"))
			})
		})
	})

	t.Run("Describe limiter client", func(t *testing.T) {
		t.Run("Context limiter-url", func(t *testing.T) {
			t.Run("It can be set via command line", func(t *testing.T) {
				cfg, err := Willow(append(baseArgs, "-limiter-url", "http://127.0.0.1:8082/"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LimiterURL).To(Equal("http://127.0.0.1:8082/"))
			})

			t.Run("It can be set via env vars", func(t *testing.T) {
				os.Setenv("WILLOW_LIMITER_URL", "http://127.0.0.1:8082/")
				defer os.Unsetenv("WILLOW_LIMITER_URL")

				cfg, err := Willow(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LimiterURL).To(Equal("http://127.0.0.1:8082/"))
			})
		})

		t.Run("Context limiter-client-ca", func(t *testing.T) {
			t.Run("It can be set via command line", func(t *testing.T) {
				cfg, err := Willow(append(baseArgs, "-limiter-client-ca", caCrt.Name()))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LimiterClientCA).To(Equal(caCrt.Name()))
			})

			t.Run("It can be set via env vars", func(t *testing.T) {
				os.Setenv("WILLOW_LIMITER_CA", caCrt.Name())
				defer os.Unsetenv("WILLOW_LIMITER_CLIENT_CA")

				cfg, err := Willow(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LimiterClientCA).To(Equal(caCrt.Name()))
			})
		})

		t.Run("Context limiter-client-crt", func(t *testing.T) {
			t.Run("It can be set via command line", func(t *testing.T) {
				cfg, err := Willow(append(baseArgs, "-limiter-client-crt", serverCRT.Name()))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LimiterClientCRT).To(Equal(serverCRT.Name()))
			})

			t.Run("It can be set via env vars", func(t *testing.T) {
				os.Setenv("WILLOW_LIMITER_CLIENT_CRT", serverCRT.Name())
				defer os.Unsetenv("WILLOW_LIMITER_CLIENT_CRT")

				cfg, err := Willow(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LimiterClientCRT).To(Equal(serverCRT.Name()))
			})
		})

		t.Run("Context limiter-client-key", func(t *testing.T) {
			t.Run("It can be set via command line", func(t *testing.T) {
				cfg, err := Willow(append(baseArgs, "-limiter-client-key", serverKey.Name()))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LimiterClientKey).To(Equal(serverKey.Name()))
			})

			t.Run("It can be set via env vars", func(t *testing.T) {
				os.Setenv("WILLOW_LIMITER_CLIENT_KEY", serverCRT.Name())
				defer os.Unsetenv("WILLOW_LIMITER_CLIENT_KEY")

				cfg, err := Willow(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.LimiterClientKey).To(Equal(serverCRT.Name()))
			})
		})

	})
}
