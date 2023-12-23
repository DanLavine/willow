package clients

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DanLavine/willow/pkg/models/api"
	. "github.com/onsi/gomega"
)

func TestClientConfig_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the URL is empty", func(t *testing.T) {
		cfg := &Config{}

		err := cfg.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("client's Config.URL cannot be empty"))
	})

	t.Run("Context when the URL is valid", func(t *testing.T) {
		t.Run("It sets the default conttent encoding to json", func(t *testing.T) {
			cfg := &Config{URL: "http://something.io"}

			err := cfg.Validate()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cfg.ContentEncoding).To(Equal(api.ContentTypeJSON))
		})

		t.Run("It returns an error for unkown content types", func(t *testing.T) {
			cfg := &Config{URL: "http://something.io", ContentEncoding: "bad"}

			err := cfg.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("unknown content type: bad"))
		})

		t.Run("It accepts json content types", func(t *testing.T) {
			cfg := &Config{URL: "http://something.io", ContentEncoding: "application/json"}

			err := cfg.Validate()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(cfg.ContentEncoding).To(Equal(api.ContentTypeJSON))
		})

		t.Run("It accepts no CA certificates", func(t *testing.T) {
			cfg := &Config{URL: "http://something.io"}

			err := cfg.Validate()
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("It returns an error if only one of the configuration ca certs are set", func(t *testing.T) {
			cfg := &Config{
				URL:    "http://something.io",
				CAFile: "nope",
			}
			err := cfg.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("when providing custom certs, all 3 values must be provided [CAFile | ClientKeyFile | ClienCRTFile]"))

			cfg = &Config{
				URL:           "http://something.io",
				ClientKeyFile: "nope",
			}
			err = cfg.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("when providing custom certs, all 3 values must be provided [CAFile | ClientKeyFile | ClienCRTFile]"))

			cfg = &Config{
				URL:           "http://something.io",
				ClientCRTFile: "nope",
			}
			err = cfg.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("when providing custom certs, all 3 values must be provided [CAFile | ClientKeyFile | ClienCRTFile]"))
		})

		t.Run("It returns an error if CAFile cannot be read from disk", func(t *testing.T) {
			cfg := &Config{
				URL:           "http://something.io",
				CAFile:        "nope",
				ClientKeyFile: "nope",
				ClientCRTFile: "nope",
			}
			err := cfg.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read the CAFile"))
		})

		t.Run("It returns an error if CAFile is not valid", func(t *testing.T) {
			_, currentDir, _, _ := runtime.Caller(0)
			cfg := &Config{
				URL:           "http://something.io",
				CAFile:        filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
				ClientKeyFile: "ClientKeyFile",
				ClientCRTFile: "ClientCRTFile",
			}
			err := cfg.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("error parsing CAFile"))
		})

		t.Run("It returns an error if ClientKeyFile or ClientCRTFile cannot be read from disk", func(t *testing.T) {
			_, currentDir, _, _ := runtime.Caller(0)
			cfg := &Config{
				URL:           "http://something.io",
				CAFile:        filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				ClientKeyFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
				ClientCRTFile: "ClientCRTFile",
			}
			err := cfg.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read the ClientKeyFile or ClientCRTFile"))

			cfg = &Config{
				URL:           "http://something.io",
				CAFile:        filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				ClientKeyFile: "ClientKeyFile",
				ClientCRTFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
			}
			err = cfg.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read the ClientKeyFile or ClientCRTFile"))
		})

		t.Run("It returns an error if ClientKeyFile or ClientCRTFile are invalid", func(t *testing.T) {
			_, currentDir, _, _ := runtime.Caller(0)
			cfg := &Config{
				URL:           "http://something.io",
				CAFile:        filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				ClientKeyFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				ClientCRTFile: "ClientCRTFile",
			}
			err := cfg.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read the ClientKeyFile or ClientCRTFile"))

			cfg = &Config{
				URL:           "http://something.io",
				CAFile:        filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				ClientKeyFile: "ClientKeyFile",
				ClientCRTFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
			}
			err = cfg.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read the ClientKeyFile or ClientCRTFile"))
		})

		t.Run("It accepts all valid certificates ", func(t *testing.T) {
			_, currentDir, _, _ := runtime.Caller(0)
			cfg := &Config{
				URL:           "http://something.io",
				CAFile:        filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				ClientKeyFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
				ClientCRTFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
			}
			err := cfg.Validate()
			g.Expect(err).ToNot(HaveOccurred())
		})
	})
}
