package lockerclient

import (
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
)

func TestLockerClientConfig_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the URL is empty", func(t *testing.T) {
		cfg := &Config{}

		err := cfg.Vaidate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("LockerClient's Config.URL cannot be empty"))
	})

	t.Run("Context when the URL is valid", func(t *testing.T) {
		t.Run("It accepts no CA certificates", func(t *testing.T) {
			cfg := &Config{URL: "http://something.io"}

			err := cfg.Vaidate()
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("It returns an error if only one of the configuration ca certs are set", func(t *testing.T) {
			cfg := &Config{
				URL:          "http://something.io",
				LockerCAFile: "nope",
			}
			err := cfg.Vaidate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("when providing custom certs, all 3 values must be provided [LockerCAFile | LockerClientKeyFile | LockerClienCRTFile]"))

			cfg = &Config{
				URL:                 "http://something.io",
				LockerClientKeyFile: "nope",
			}
			err = cfg.Vaidate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("when providing custom certs, all 3 values must be provided [LockerCAFile | LockerClientKeyFile | LockerClienCRTFile]"))

			cfg = &Config{
				URL:                 "http://something.io",
				LockerClientCRTFile: "nope",
			}
			err = cfg.Vaidate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("when providing custom certs, all 3 values must be provided [LockerCAFile | LockerClientKeyFile | LockerClienCRTFile]"))
		})

		t.Run("It returns an error if LockerCAFile cannot be read from disk", func(t *testing.T) {
			cfg := &Config{
				URL:                 "http://something.io",
				LockerCAFile:        "nope",
				LockerClientKeyFile: "nope",
				LockerClientCRTFile: "nope",
			}
			err := cfg.Vaidate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read the LockerCAFile"))
		})

		t.Run("It returns an error if LockerCAFile is not valid", func(t *testing.T) {
			_, currentDir, _, _ := runtime.Caller(0)
			cfg := &Config{
				URL:                 "http://something.io",
				LockerCAFile:        filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
				LockerClientKeyFile: "LockerClientKeyFile",
				LockerClientCRTFile: "LockerClientCRTFile",
			}
			err := cfg.Vaidate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("error parsing LockerCAFile"))
		})

		t.Run("It returns an error if LockerClientKeyFile or LockerClientCRTFile cannot be read from disk", func(t *testing.T) {
			_, currentDir, _, _ := runtime.Caller(0)
			cfg := &Config{
				URL:                 "http://something.io",
				LockerCAFile:        filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				LockerClientKeyFile: filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
				LockerClientCRTFile: "LockerClientCRTFile",
			}
			err := cfg.Vaidate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read the LockerClientKeyFile or LockerClientCRTFile"))

			cfg = &Config{
				URL:                 "http://something.io",
				LockerCAFile:        filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				LockerClientKeyFile: "LockerClientKeyFile",
				LockerClientCRTFile: filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
			}
			err = cfg.Vaidate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read the LockerClientKeyFile or LockerClientCRTFile"))
		})

		t.Run("It returns an error if LockerClientKeyFile or LockerClientCRTFile are invalid", func(t *testing.T) {
			_, currentDir, _, _ := runtime.Caller(0)
			cfg := &Config{
				URL:                 "http://something.io",
				LockerCAFile:        filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				LockerClientKeyFile: filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				LockerClientCRTFile: "LockerClientCRTFile",
			}
			err := cfg.Vaidate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read the LockerClientKeyFile or LockerClientCRTFile"))

			cfg = &Config{
				URL:                 "http://something.io",
				LockerCAFile:        filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				LockerClientKeyFile: "LockerClientKeyFile",
				LockerClientCRTFile: filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
			}
			err = cfg.Vaidate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read the LockerClientKeyFile or LockerClientCRTFile"))
		})

		t.Run("It accepts all valid certificates ", func(t *testing.T) {
			_, currentDir, _, _ := runtime.Caller(0)
			cfg := &Config{
				URL:                 "http://something.io",
				LockerCAFile:        filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
				LockerClientKeyFile: filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
				LockerClientCRTFile: filepath.Join(currentDir, "..", "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
			}
			err := cfg.Vaidate()
			g.Expect(err).ToNot(HaveOccurred())
		})
	})
}
