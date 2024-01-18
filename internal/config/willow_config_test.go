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

	baseArgs := []string{"-server-key", serverKey.Name(), "-server-crt", serverCRT.Name()}

	t.Run("Describe willow server", func(t *testing.T) {
		t.Run("It returns an error if there is no willow-server-key", func(t *testing.T) {
			cfg, err := Willow(nil)
			g.Expect(cfg).To(BeNil())
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("parameter 'willow-server-key' is not set"))
		})

		t.Run("It returns an error if there is no willow-server-crt", func(t *testing.T) {
			cfg, err := Willow([]string{"-server-key", serverKey.Name()})
			g.Expect(cfg).To(BeNil())
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("parameter 'willow-server-crt' is not set"))
		})

		t.Run("Context willow-port", func(t *testing.T) {
			t.Run("It can be set via command line", func(t *testing.T) {
				cfg, err := Willow(append(baseArgs, "-port", "8765"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.WillowPort).To(Equal("8765"))
			})

			t.Run("It can be set via env vars", func(t *testing.T) {
				os.Setenv("WILLOW_PORT", "8888")
				defer os.Unsetenv("WILLOW_PORT")

				cfg, err := Willow(baseArgs)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.WillowPort).To(Equal("8888"))
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

		t.Run("Context disk", func(t *testing.T) {
			t.Run("It requires a storage dir when set via cmd line", func(t *testing.T) {
				cfg, err := Willow(append(baseArgs, "-storage-type", "disk"))
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("'disk-storage-dir' is required when storage type is 'disk'"))
			})

			t.Run("It requires a storage dir when set via env vars", func(t *testing.T) {
				os.Setenv("STORAGE_TYPE", "disk")
				defer os.Unsetenv("STORAGE_TYPE")

				cfg, err := Willow(baseArgs)
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("'disk-storage-dir' is required when storage type is 'disk'"))
			})

			t.Run("Context when setting a storage dir", func(t *testing.T) {
				tmpDir := os.TempDir()
				defer os.Remove(tmpDir)

				t.Run("can be set by flags", func(t *testing.T) {
					cfg, err := Willow(append(baseArgs, "-storage-type", "disk", "-storage-dir", tmpDir))
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(*cfg.StorageConfig.Type).To(Equal("disk"))
					g.Expect(*cfg.StorageConfig.Disk.StorageDir).To(Equal(tmpDir))
				})

				t.Run("can be set by env var", func(t *testing.T) {
					os.Setenv("STORAGE_TYPE", "disk")
					os.Setenv("DISK_STORAGE_DIR", tmpDir)
					defer func() {
						os.Unsetenv("STORAGE_TYPE")
						os.Unsetenv("DISK_STORAGE_DIR")
					}()

					cfg, err := Willow(baseArgs)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(*cfg.StorageConfig.Type).To(Equal("disk"))
					g.Expect(*cfg.StorageConfig.Disk.StorageDir).To(Equal(tmpDir))
				})
			})
		})

		t.Run("It returns an error for an unknown type", func(t *testing.T) {
			cfg, err := Willow(append(baseArgs, "-storage-type", "foo"))
			g.Expect(cfg).To(BeNil())
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("invalid storage type selected 'foo'"))
		})
	})
}
