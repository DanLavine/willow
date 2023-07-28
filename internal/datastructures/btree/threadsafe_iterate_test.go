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
		g.Expect(err.Error()).To(Equal("callback cannot be nil"))
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

func TestBTree_IterateMatchType(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.IterateMatchType(datatypes.T_float32, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("callback cannot be nil"))
	})

	t.Run("it does not run the iterative function if there are no values", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		iterate := func(_ any) {
			panic("should not call")
		}

		g.Expect(func() { bTree.IterateMatchType(datatypes.T_any, iterate) }).ToNot(Panic())
	})

	t.Run("it calls the iterative function on each tree item with a value where the datatypes match", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(datatypes.Int(1), NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Int8(1), NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Int16(1), NewBTreeTester("3"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Int32(1), NewBTreeTester("4"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Int64(1), NewBTreeTester("5"), OnFindTest)).ToNot(HaveOccurred())

		count := 0

		bTree.IterateMatchType(datatypes.T_int, func(item any) {
			bTreeTester := item.(*BTreeTester)

			g.Expect(bTreeTester.Value).To(Equal("1"))
			count++
		})
		bTree.IterateMatchType(datatypes.T_int8, func(item any) {
			bTreeTester := item.(*BTreeTester)

			g.Expect(bTreeTester.Value).To(Equal("2"))
			count++
		})
		bTree.IterateMatchType(datatypes.T_int16, func(item any) {
			bTreeTester := item.(*BTreeTester)

			g.Expect(bTreeTester.Value).To(Equal("3"))
			count++
		})
		bTree.IterateMatchType(datatypes.T_int32, func(item any) {
			bTreeTester := item.(*BTreeTester)

			g.Expect(bTreeTester.Value).To(Equal("4"))
			count++
		})
		bTree.IterateMatchType(datatypes.T_int64, func(item any) {
			bTreeTester := item.(*BTreeTester)

			g.Expect(bTreeTester.Value).To(Equal("5"))
			count++
		})

		g.Expect(count).To(Equal(5))
	})
}
