package config

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestWillowConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Describe storage validation", func(t *testing.T) {
		t.Run("Context memory", func(t *testing.T) {
			t.Run("can be set by flags", func(t *testing.T) {
				originalArgs := os.Args
				os.Args = []string{"test", "-storage-type", "memory"}
				defer func() {
					os.Args = originalArgs
				}()

				cfg, err := Willow()
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.StorageConfig.Type).To(Equal("memory"))
			})

			t.Run("can be set by env var", func(t *testing.T) {
				os.Setenv("STORAGE_TYPE", "memory")
				originalArgs := os.Args
				os.Args = []string{"test"}
				defer func() {
					os.Args = originalArgs
					os.Unsetenv("STORAGE_TYPE")
				}()

				cfg, err := Willow()
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(*cfg.StorageConfig.Type).To(Equal("memory"))
			})
		})

		t.Run("Context disk", func(t *testing.T) {
			t.Run("It requires a storage dir when set via flags", func(t *testing.T) {
				originalArgs := os.Args
				os.Args = []string{"test", "-storage-type", "disk"}
				defer func() {
					os.Args = originalArgs
				}()

				cfg, err := Willow()
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("'disk-storage-dir' is required when storage type is 'disk'"))
			})

			t.Run("It requires a storage dir when set via env vars", func(t *testing.T) {
				os.Setenv("STORAGE_TYPE", "disk")
				originalArgs := os.Args
				os.Args = []string{"test"}
				defer func() {
					os.Args = originalArgs
					os.Unsetenv("STORAGE_TYPE")
				}()

				cfg, err := Willow()
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("'disk-storage-dir' is required when storage type is 'disk'"))
			})

			t.Run("Context when setting a storage dir", func(t *testing.T) {
				t.Run("can be set by flags", func(t *testing.T) {
					originalArgs := os.Args
					tmpdDir := os.TempDir()

					os.Args = []string{"test", "-storage-type", "disk", "-storage-dir", tmpdDir}
					defer func() {
						os.Args = originalArgs
						os.Remove(tmpdDir)
					}()

					cfg, err := Willow()
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(*cfg.StorageConfig.Type).To(Equal("disk"))
					g.Expect(*cfg.StorageConfig.Disk.StorageDir).To(Equal(tmpdDir))
				})

				t.Run("can be set by env var", func(t *testing.T) {
					tmpdDir := os.TempDir()

					os.Setenv("STORAGE_TYPE", "disk")
					os.Setenv("DISK_STORAGE_DIR", tmpdDir)
					originalArgs := os.Args
					os.Args = []string{"test"}
					defer func() {
						os.Args = originalArgs
						os.Unsetenv("STORAGE_TYPE")
						os.Unsetenv("DISK_STORAGE_DIR")
						os.Remove(tmpdDir)
					}()

					cfg, err := Willow()
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(*cfg.StorageConfig.Type).To(Equal("disk"))
					g.Expect(*cfg.StorageConfig.Disk.StorageDir).To(Equal(tmpdDir))
				})
			})
		})

		t.Run("Context an unkown type", func(t *testing.T) {
			t.Run("It returns an error if set by flags", func(t *testing.T) {
				originalArgs := os.Args
				os.Args = []string{"test", "-storage-type", "foo"}
				defer func() {
					os.Args = originalArgs
				}()

				cfg, err := Willow()
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("invalid storage type selected 'foo'"))
			})

			t.Run("It returns an error if set via env vars", func(t *testing.T) {
				os.Setenv("STORAGE_TYPE", "foo")
				originalArgs := os.Args
				os.Args = []string{"test"}
				defer func() {
					os.Args = originalArgs
					os.Unsetenv("STORAGE_TYPE")
				}()

				cfg, err := Willow()
				g.Expect(cfg).To(BeNil())
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("invalid storage type selected 'foo'"))
			})
		})
	})
}
