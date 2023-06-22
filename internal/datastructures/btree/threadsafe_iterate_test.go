package btree

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestBTree_Iterate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.Iterate(nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Iterate callback is nil"))
	})

	t.Run("it does not run the iterative function if there are no values", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		iterate := func(_ any) {
			panic("should not call")
		}

		g.Expect(func() { bTree.Iterate(iterate) }).ToNot(Panic())
	})

	t.Run("it calls the iterative function on each tree item with a value", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		seenValues := map[string]struct{}{}
		count := 0
		iterate := func(val any) {
			BTreeTester := val.(*BTreeTester)

			// check that each value is unique
			g.Expect(seenValues).ToNot(HaveKey(BTreeTester.Value))
			seenValues[BTreeTester.Value] = struct{}{}

			count++
		}

		bTree.Iterate(iterate)
		g.Expect(count).To(Equal(1_024))
	})
}
