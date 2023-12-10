package counters

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Counter_Increment(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It adds 1 to the counter", func(t *testing.T) {
		counter := New()

		g.Expect(counter.Increment()).To(Equal(uint64(1)))
		g.Expect(counter.Increment()).To(Equal(uint64(2)))
		g.Expect(counter.Increment()).To(Equal(uint64(3)))
	})
}

func Test_Counter_Decrement(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It removed 1 from the counter", func(t *testing.T) {
		counter := New()

		g.Expect(counter.Increment()).To(Equal(uint64(1)))
		g.Expect(counter.Increment()).To(Equal(uint64(2)))
		g.Expect(counter.Increment()).To(Equal(uint64(3)))

		g.Expect(counter.Decrement()).To(Equal(uint64(2)))
		g.Expect(counter.Decrement()).To(Equal(uint64(1)))
		g.Expect(counter.Decrement()).To(Equal(uint64(0)))
	})
}

func Test_Counter_Load(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It loads the current value for the counter", func(t *testing.T) {
		counter := New()

		g.Expect(counter.Increment()).To(Equal(uint64(1)))
		g.Expect(counter.Increment()).To(Equal(uint64(2)))
		g.Expect(counter.Increment()).To(Equal(uint64(3)))

		g.Expect(counter.Load()).To(Equal(uint64(3)))
	})
}
