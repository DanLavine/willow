package memory

import (
	"testing"

	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func Test_overrideMemory_Limit(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It reutrns the limit set to the override", func(t *testing.T) {
		overrideReq := &v1.Override{
			Name: "override",
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.Float32(1.0),
			},
			Limit: 18,
		}
		g.Expect(overrideReq.Validate()).ToNot(HaveOccurred())

		overrideMemory := New(overrideReq)
		g.Expect(overrideMemory.Limit()).To(Equal(int64(18)))
	})
}

func Test_overrideMemory_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can update the limit in the override", func(t *testing.T) {
		overrideReq := &v1.Override{
			Name: "override",
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.Float32(1.0),
			},
			Limit: 18,
		}
		g.Expect(overrideReq.Validate()).ToNot(HaveOccurred())

		overrideMemory := New(overrideReq)

		updateReq := &v1.OverrideUpdate{
			Limit: 99,
		}
		g.Expect(updateReq.Validate()).ToNot(HaveOccurred())

		overrideMemory.Update(updateReq)
		g.Expect(overrideMemory.Limit()).To(Equal(int64(99)))
	})
}

func Test_overrideMemory_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It performs a no-op", func(t *testing.T) {
		overrideReq := &v1.Override{
			Name: "override",
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.Float32(1.0),
			},
			Limit: 18,
		}
		g.Expect(overrideReq.Validate()).ToNot(HaveOccurred())

		overrideMemory := New(overrideReq)
		g.Expect(overrideMemory.Delete()).ToNot(HaveOccurred())
	})
}
