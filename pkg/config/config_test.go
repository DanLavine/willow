package config_test

import (
	"os"
	"testing"

	"github.com/DanLavine/willow/pkg/config"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("storage validation", func(t *testing.T) {
		t.Run("by flags returns an error on an unkown storage type", func(t *testing.T) {
			originalArgs := os.Args
			os.Args = []string{"test", "-storage-type", "foo"}
			defer func() {
				os.Args = originalArgs
			}()

			cfg := config.Default()
			err := cfg.Parse()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("invalid storage type selected: foo"))
		})

		t.Run("by env var requires an error on an unkown storage type", func(t *testing.T) {
			os.Setenv("STORAGE_TYPE", "foo")
			defer os.Unsetenv("STORAGE_TYPE")

			cfg := config.Default()
			err := cfg.Parse()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("invalid storage type selected: foo"))
		})

		t.Run("memory", func(t *testing.T) {
			t.Run("can be set by flags", func(t *testing.T) {
				originalArgs := os.Args
				os.Args = []string{"test", "-storage-type", "memory"}
				defer func() {
					os.Args = originalArgs
				}()

				cfg := config.Default()
				err := cfg.Parse()
				g.Expect(err).ToNot(HaveOccurred())
			})

			t.Run("can be set by env var", func(t *testing.T) {
				os.Setenv("STORAGE_TYPE", "memory")
				defer os.Unsetenv("STORAGE_TYPE")

				cfg := config.Default()
				err := cfg.Parse()
				g.Expect(err).ToNot(HaveOccurred())
			})
		})

		t.Run("disk", func(t *testing.T) {
			t.Run("by flags requires a storage dir", func(t *testing.T) {
				originalArgs := os.Args
				os.Args = []string{"test", "-storage-type", "disk"}
				defer func() {
					os.Args = originalArgs
				}()

				cfg := config.Default()
				err := cfg.Parse()
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("'disk-storage-dir' is required when storage type is 'disk'"))
			})

			t.Run("by env var requires a storage dir", func(t *testing.T) {
				os.Setenv("STORAGE_TYPE", "disk")
				defer os.Unsetenv("STORAGE_TYPE")

				cfg := config.Default()
				err := cfg.Parse()
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("'disk-storage-dir' is required when storage type is 'disk'"))
			})

			t.Run("by flags accepts a storage dir", func(t *testing.T) {
				originalArgs := os.Args
				os.Args = []string{"test", "-storage-type", "disk", "-disk-storage-dir", os.TempDir()}
				defer func() {
					os.Args = originalArgs
				}()

				cfg := config.Default()
				err := cfg.Parse()
				g.Expect(err).ToNot(HaveOccurred())
			})

			t.Run("by env accepts a storage dir", func(t *testing.T) {
				os.Setenv("STORAGE_TYPE", "disk")
				os.Setenv("DISK_STORAGE_DIR", os.TempDir())
				defer os.Unsetenv("STORAGE_TYPE")
				defer os.Unsetenv("DISK_STORAGE_DIR")

				cfg := config.Default()
				err := cfg.Parse()
				g.Expect(err).ToNot(HaveOccurred())
			})
		})
	})
}
