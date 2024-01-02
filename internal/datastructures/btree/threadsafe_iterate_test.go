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

		iterate := func(_ datatypes.EncapsulatedValue, _ any) bool {
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

		keys := map[datatypes.EncapsulatedValue]struct{}{}
		seenValues := map[string]struct{}{}
		count := 0
		iterate := func(key datatypes.EncapsulatedValue, val any) bool {
			BTreeTester := val.(*BTreeTester)

			// check that each value is unique
			g.Expect(keys).ToNot(HaveKey(key))
			g.Expect(seenValues).ToNot(HaveKey(BTreeTester.Value))
			keys[key] = struct{}{}
			seenValues[BTreeTester.Value] = struct{}{}

			count++
			return true
		}

		bTree.Iterate(iterate)
		g.Expect(count).To(Equal(1_024))
	})

	t.Run("it breaks the iteration when the callback returns false", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		keys := map[datatypes.EncapsulatedValue]struct{}{}
		seenValues := map[string]struct{}{}
		iterate := func(key datatypes.EncapsulatedValue, val any) bool {
			BTreeTester := val.(*BTreeTester)

			// check that each value is unique
			g.Expect(keys).ToNot(HaveKey(key))
			g.Expect(seenValues).ToNot(HaveKey(BTreeTester.Value))
			keys[key] = struct{}{}
			seenValues[BTreeTester.Value] = struct{}{}

			return len(seenValues) < 5
		}

		bTree.Iterate(iterate)
		g.Expect(len(seenValues)).To(Equal(5))
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

		iterate := func(_ datatypes.EncapsulatedValue, _ any) bool {
			panic("should not call")
		}

		g.Expect(func() { bTree.IterateMatchType(datatypes.T_uint, iterate) }).ToNot(Panic())
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

		bTree.IterateMatchType(datatypes.T_int, func(key datatypes.EncapsulatedValue, item any) bool {
			bTreeTester := item.(*BTreeTester)

			g.Expect(key).To(Equal(datatypes.Int(1)))
			g.Expect(bTreeTester.Value).To(Equal("1"))
			count++
			return true
		})
		bTree.IterateMatchType(datatypes.T_int8, func(key datatypes.EncapsulatedValue, item any) bool {
			bTreeTester := item.(*BTreeTester)

			g.Expect(key).To(Equal(datatypes.Int8(1)))
			g.Expect(bTreeTester.Value).To(Equal("2"))
			count++
			return true
		})
		bTree.IterateMatchType(datatypes.T_int16, func(key datatypes.EncapsulatedValue, item any) bool {
			bTreeTester := item.(*BTreeTester)

			g.Expect(key).To(Equal(datatypes.Int16(1)))
			g.Expect(bTreeTester.Value).To(Equal("3"))
			count++
			return true
		})
		bTree.IterateMatchType(datatypes.T_int32, func(key datatypes.EncapsulatedValue, item any) bool {
			bTreeTester := item.(*BTreeTester)

			g.Expect(key).To(Equal(datatypes.Int32(1)))
			g.Expect(bTreeTester.Value).To(Equal("4"))
			count++
			return true
		})
		bTree.IterateMatchType(datatypes.T_int64, func(key datatypes.EncapsulatedValue, item any) bool {
			bTreeTester := item.(*BTreeTester)

			g.Expect(key).To(Equal(datatypes.Int64(1)))
			g.Expect(bTreeTester.Value).To(Equal("5"))
			count++
			return true
		})

		g.Expect(count).To(Equal(5))
	})

	t.Run("it breaks the iteration when the callback returns false", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		keys := map[datatypes.EncapsulatedValue]struct{}{}
		seenValues := map[string]struct{}{}
		iterate := func(key datatypes.EncapsulatedValue, val any) bool {
			BTreeTester := val.(*BTreeTester)

			// check that each value is unique
			g.Expect(keys).ToNot(HaveKey(key))
			g.Expect(seenValues).ToNot(HaveKey(BTreeTester.Value))
			keys[key] = struct{}{}
			seenValues[BTreeTester.Value] = struct{}{}

			return len(seenValues) < 5
		}

		bTree.IterateMatchType(datatypes.T_int, iterate)
		g.Expect(len(seenValues)).To(Equal(5))
	})
}
