package set

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestSet_New(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it can initialize a set with no entries", func(t *testing.T) {
		set := New[int]()
		g.Expect(set).ToNot(BeNil())
	})

	t.Run("it can initialize a set with optional entries", func(t *testing.T) {
		set := New[string]("one", "two")
		g.Expect(set).ToNot(BeNil())

		g.Expect(len(set.values)).To(Equal(2))
		g.Expect(set.values).To(HaveKey("one"))
		g.Expect(set.values).To(HaveKey("two"))
	})
}

func TestSet_Add(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it adds the value to the set", func(t *testing.T) {
		set := New[int]()

		set.Add(42)
		g.Expect(set.values).To(HaveKey(42))
	})

	t.Run("it ignores elements already in the set", func(t *testing.T) {
		set := New[int]()

		set.Add(42)
		set.Add(42)
		set.Add(42)
		set.Add(42)
		set.Add(42)
		g.Expect(set.values).To(HaveKey(42))
		g.Expect(len(set.values)).To(Equal(1))
	})
}

func TestSet_Remove(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it does nothing if the value is not in the set", func(t *testing.T) {
		set := New[int]()

		set.Add(42)
		set.Remove(100)
		g.Expect(len(set.values)).To(Equal(1))
		g.Expect(set.values).To(HaveKey(42))
	})

	t.Run("it removes a value from the set", func(t *testing.T) {
		set := New[int]()

		set.Add(42)
		set.Remove(42)
		g.Expect(len(set.values)).To(Equal(0))
	})
}

func TestSet_Values(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns all the values in the set", func(t *testing.T) {
		set := New[int]()

		set.Add(1)
		set.Add(2)
		set.Add(3)
		set.Add(4)
		set.Add(5)

		values := set.Values()
		g.Expect(values).To(ContainElement(1))
		g.Expect(values).To(ContainElement(2))
		g.Expect(values).To(ContainElement(3))
		g.Expect(values).To(ContainElement(4))
		g.Expect(values).To(ContainElement(5))
	})
}

func TestSet_Size(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns the number of elements in the set", func(t *testing.T) {
		set := New[int]()

		set.Add(1)
		set.Add(2)
		set.Add(3)
		set.Add(4)
		set.Add(5)

		g.Expect(set.Size()).To(Equal(5))
	})
}
