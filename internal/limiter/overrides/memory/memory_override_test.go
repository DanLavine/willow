package memory

import (
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	. "github.com/onsi/gomega"
)

func Test_overrideMemory_Limit(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It reutrns the limit set to the override", func(t *testing.T) {
		overrideReq := &v1.OverrideProperties{
			Limit: helpers.PointerOf[int64](18),
		}
		g.Expect(overrideReq.Validate()).ToNot(HaveOccurred())

		overrideMemory := New(overrideReq)
		g.Expect(overrideMemory.Limit()).To(Equal(int64(18)))
	})
}

func Test_overrideMemory_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can update the limit in the override", func(t *testing.T) {
		// create the request
		overrideReq := &v1.OverrideProperties{
			Limit: helpers.PointerOf[int64](18),
		}
		g.Expect(overrideReq.Validate()).ToNot(HaveOccurred())

		overrideMemory := New(overrideReq)

		// update the request
		updateReq := &v1.OverrideProperties{
			Limit: helpers.PointerOf[int64](99),
		}
		g.Expect(updateReq.Validate()).ToNot(HaveOccurred())

		overrideMemory.Update(updateReq)
		g.Expect(overrideMemory.Limit()).To(Equal(int64(99)))
	})
}

func Test_overrideMemory_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It performs a no-op", func(t *testing.T) {
		overrideReq := &v1.OverrideProperties{
			Limit: helpers.PointerOf[int64](18),
		}
		g.Expect(overrideReq.Validate()).ToNot(HaveOccurred())

		overrideMemory := New(overrideReq)
		g.Expect(overrideMemory.Delete()).ToNot(HaveOccurred())
	})
}
