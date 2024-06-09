package memory

import (
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"

	. "github.com/onsi/gomega"
)

func Test_Counter_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can add to the counter", func(t *testing.T) {
		counterResquest := &v1limiter.CounteProperties{
			Counters: helpers.PointerOf[int64](0),
		}
		g.Expect(counterResquest.Validate()).ToNot(HaveOccurred())

		counter := New(counterResquest)

		g.Expect(counter.Update(&v1limiter.CounteProperties{Counters: helpers.PointerOf[int64](1)})).To(Equal(int64(1)))
		g.Expect(counter.Update(&v1limiter.CounteProperties{Counters: helpers.PointerOf[int64](1)})).To(Equal(int64(2)))
		g.Expect(counter.Update(&v1limiter.CounteProperties{Counters: helpers.PointerOf[int64](2)})).To(Equal(int64(4)))
	})

	t.Run("It can decrement the counter", func(t *testing.T) {
		counterResquest := &v1limiter.CounteProperties{
			Counters: helpers.PointerOf[int64](15),
		}
		g.Expect(counterResquest.Validate()).ToNot(HaveOccurred())

		counter := New(counterResquest)

		g.Expect(counter.Update(&v1limiter.CounteProperties{Counters: helpers.PointerOf[int64](-1)})).To(Equal(int64(14)))
		g.Expect(counter.Update(&v1limiter.CounteProperties{Counters: helpers.PointerOf[int64](-1)})).To(Equal(int64(13)))
		g.Expect(counter.Update(&v1limiter.CounteProperties{Counters: helpers.PointerOf[int64](-2)})).To(Equal(int64(11)))
	})
}

func Test_Counter_Load(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It loads the current value for the counter", func(t *testing.T) {
		counterResquest := &v1limiter.CounteProperties{
			Counters: helpers.PointerOf[int64](15),
		}
		g.Expect(counterResquest.Validate()).ToNot(HaveOccurred())

		counter := New(counterResquest)
		g.Expect(counter.Load()).To(Equal(int64(15)))
	})
}
