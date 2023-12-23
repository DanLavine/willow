package counters

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Counter_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can add to the counter", func(t *testing.T) {
		counter := New()

		g.Expect(counter.Update(1)).To(Equal(int64(1)))
		g.Expect(counter.Update(1)).To(Equal(int64(2)))
		g.Expect(counter.Update(2)).To(Equal(int64(4)))
	})

	t.Run("It cam decrement the counter", func(t *testing.T) {
		counter := New()

		g.Expect(counter.Update(1)).To(Equal(int64(1)))
		g.Expect(counter.Update(1)).To(Equal(int64(2)))
		g.Expect(counter.Update(2)).To(Equal(int64(4)))
		g.Expect(counter.Update(-1)).To(Equal(int64(3)))
		g.Expect(counter.Update(-1)).To(Equal(int64(2)))
		g.Expect(counter.Update(-2)).To(Equal(int64(0)))
	})
}

func Test_Counter_Load(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It loads the current value for the counter", func(t *testing.T) {
		counter := New()

		g.Expect(counter.Update(1)).To(Equal(int64(1)))
		g.Expect(counter.Update(1)).To(Equal(int64(2)))
		g.Expect(counter.Update(1)).To(Equal(int64(3)))

		g.Expect(counter.Load()).To(Equal(int64(3)))
	})
}
