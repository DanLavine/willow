package btree

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestBTree_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.Find(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, OnFindTest)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFind callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.Find(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onFind is nil"))
	})

	t.Run("it does not run the callback if the item is not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) {
			called = true
		}

		g.Expect(bTree.Find(datatypes.Int64(-1), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("returns the item in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) {
			btt := item.(*BTreeTester)
			g.Expect(btt.Value).To(Equal("768"))
			called = true
		}

		bTree.Find(datatypes.Int(768), onFind)
		g.Expect(called).To(BeTrue())
	})
}

func TestBTree_FindNotEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.FindNotEqual(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, func(items any) bool { return true })
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFindSelection callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindNotEqual(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("callback cannot be nil"))
	})

	t.Run("it runs the callback on all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(items any) bool {
			foundItems = append(foundItems, items)
			return true
		}

		bTree.FindNotEqual(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(1_023)) // account for 0-1023 except for 512
		g.Expect(foundItems).ToNot(ContainElement("512"))
	})

	t.Run("it acconts for the type when comparing", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(items any) bool {
			foundItems = append(foundItems, items)
			return true
		}

		bTree.FindNotEqual(datatypes.String("512"), onFind)
		g.Expect(len(foundItems)).To(Equal(1_024)) // all ints, no strings in the setup
	})

	t.Run("it breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(items any) bool {
			foundItems = append(foundItems, items)
			return len(foundItems) < 5
		}

		bTree.FindNotEqual(datatypes.String("512"), onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindNotEqualMatchType(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.FindNotEqualMatchType(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, func(items any) bool { return true })
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFindSelection callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindNotEqualMatchType(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("callback cannot be nil"))
	})

	t.Run("it runs the callback for all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindNotEqualMatchType(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(1_023)) // account for 0-1023 except for 512
		g.Expect(foundItems).ToNot(ContainElement("512"))
	})

	t.Run("it accounts for the type when comparing", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(datatypes.Int8(1), func() any { return "1" }, OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Int8(2), func() any { return "2" }, OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Int16(1), func() any { return "int16_1" }, OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Int16(2), func() any { return "int16_2" }, OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Int16(3), func() any { return "int16_3" }, OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Int32(1), func() any { return "1" }, OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Int32(2), func() any { return "2" }, OnFindTest)).ToNot(HaveOccurred())

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		// sets up a tree with the keys like so:
		//        2
		//    1      3
		//  1   2  2   3
		//
		// want to validate that we hit those inner 2 keys

		bTree.FindNotEqualMatchType(datatypes.Int16(2), onFind)
		g.Expect(len(foundItems)).To(Equal(2))
		g.Expect(foundItems).To(ContainElements([]string{"int16_1", "int16_3"}))
	})

	t.Run("it breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindNotEqualMatchType(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindLessThan(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.FindLessThan(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, func(item any) bool { return true })
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFindSelection callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindLessThan(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("callback cannot be nil"))
	})

	t.Run("it does not run the callback if no items are not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) bool {
			called = true
			return true
		}

		g.Expect(bTree.FindLessThan(datatypes.Int64(-1), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("it runs the callback for all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindLessThan(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(512)) // account for 0-511
	})

	t.Run("it takes they type of key into account for the comparison", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindLessThan(datatypes.String("512"), onFind)
		g.Expect(len(foundItems)).To(Equal(1024)) // all ints are less than strings
	})

	t.Run("it breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindLessThan(datatypes.String("512"), onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindLessThanMatchType(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.FindLessThanMatchType(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, func(item any) bool { return true })
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFindSelection callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindLessThanMatchType(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("callback cannot be nil"))
	})

	t.Run("it does not run the callback if the item is not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(ite any) bool {
			called = true
			return true
		}

		g.Expect(bTree.FindLessThanMatchType(datatypes.Int64(-1), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("it passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindLessThanMatchType(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(512)) // account for 0-511
	})

	t.Run("it only finds values where they keys types are the same", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindLessThanMatchType(datatypes.String("512"), onFind)
		g.Expect(len(foundItems)).To(Equal(0))
	})

	t.Run("it breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindLessThanMatchType(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindLessThanOrEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.FindLessThanOrEqual(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, func(item any) bool { return true })
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFindSelection callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindLessThanOrEqual(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("callback cannot be nil"))
	})

	t.Run("it does not run the callback if the item is not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) bool {
			called = true
			return true
		}

		g.Expect(bTree.FindLessThanOrEqual(datatypes.Int64(-1), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("it passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindLessThanOrEqual(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(513)) // account for 0-512
	})

	t.Run("it takes the type into acount when comparing values", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindLessThanOrEqual(datatypes.String("512"), onFind)
		g.Expect(len(foundItems)).To(Equal(1024)) // int is less than string
	})

	t.Run("it breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindLessThanOrEqual(datatypes.String("512"), onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindLessThanOrEqualMatchType(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.FindLessThanOrEqualMatchType(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, func(item any) bool { return true })
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFindSelection callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindLessThanOrEqualMatchType(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("callback cannot be nil"))
	})

	t.Run("it does not run the callback if the item is not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) bool {
			called = true
			return true
		}

		g.Expect(bTree.FindLessThanOrEqualMatchType(datatypes.Int64(-1), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("it passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindLessThanOrEqualMatchType(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(513)) // account for 0-512
	})

	t.Run("it takes the type into acount when comparing values", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindLessThanOrEqualMatchType(datatypes.String("512"), onFind)
		g.Expect(len(foundItems)).To(Equal(0)) // int is less than string

		bTree.FindLessThanOrEqualMatchType(datatypes.Int64(512), onFind)
		g.Expect(len(foundItems)).To(Equal(0)) // int is less than string
	})

	t.Run("it breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindLessThanOrEqualMatchType(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindGreaterThan(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.FindGreaterThan(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, func(item any) bool { return true })
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFindSelection callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindGreaterThan(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("callback cannot be nil"))
	})

	t.Run("it does not run the callback if the item is not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) bool {
			called = true
			return true
		}

		g.Expect(bTree.FindGreaterThan(datatypes.Int(9_001), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("it passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindGreaterThan(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(511)) // account for 513-1023
	})

	t.Run("it tages data type int comparison", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindGreaterThan(datatypes.Int64(512), onFind)
		g.Expect(len(foundItems)).To(Equal(1024)) // ints are larger than int64
	})

	t.Run("it breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindGreaterThan(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindGreaterThanMatchType(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.FindGreaterThanMatchType(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, func(item any) bool { return true })
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFindSelection callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindGreaterThanMatchType(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("callback cannot be nil"))
	})

	t.Run("it does not run the callback if the item is not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) bool {
			called = true
			return true
		}

		g.Expect(bTree.FindGreaterThanMatchType(datatypes.Int(9_001), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("it passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindGreaterThanMatchType(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(511)) // account for 513-1023
	})

	t.Run("it only matches the exact key types we are searching for", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindGreaterThanMatchType(datatypes.Int64(512), onFind)
		g.Expect(len(foundItems)).To(Equal(0))

		bTree.FindGreaterThanMatchType(datatypes.String("512"), onFind)
		g.Expect(len(foundItems)).To(Equal(0))
	})

	t.Run("it breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindGreaterThanMatchType(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindGreaterThanOrEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.FindGreaterThanOrEqual(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, func(item any) bool { return true })
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFindSelection callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindGreaterThanOrEqual(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("callback cannot be nil"))
	})

	t.Run("it does not run the callback if the item is not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) bool {
			called = true
			return true
		}

		g.Expect(bTree.FindGreaterThanOrEqual(datatypes.Int(9_001), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("it passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindGreaterThanOrEqual(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(512)) // account for 512-1023
	})

	t.Run("it accounts for data types in the comparison as well", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindGreaterThanOrEqual(datatypes.Int64(512), onFind)
		g.Expect(len(foundItems)).To(Equal(1024)) // int is larger than int64
	})

	t.Run("it breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		findCounter := 5
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < findCounter
		}

		bTree.FindGreaterThanOrEqual(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(5))

		// this needs to be in the right sub tree, not the left since we find many larger than values,
		foundItems = []any{}
		findCounter = 0
		bTree.FindGreaterThanOrEqual(datatypes.Int(767), onFind)
		g.Expect(len(foundItems)).To(Equal(1))
	})
}

func TestBTree_FindGreaterThanOrEqualMatchType(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.FindGreaterThanOrEqualMatchType(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, func(item any) bool { return true })
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if the onFindSelection callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindGreaterThanOrEqualMatchType(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("callback cannot be nil"))
	})

	t.Run("it does not run the callback if the item is not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item any) bool {
			called = true
			return true
		}

		g.Expect(bTree.FindGreaterThanOrEqualMatchType(datatypes.Int(9_001), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("it passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindGreaterThanOrEqualMatchType(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(512)) // account for 512-1023
	})

	t.Run("it oonly counts values who's key type matches", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		bTree.FindGreaterThanOrEqualMatchType(datatypes.Int64(512), onFind)
		g.Expect(len(foundItems)).To(Equal(0))

		bTree.FindGreaterThanOrEqualMatchType(datatypes.String("512"), onFind)
		g.Expect(len(foundItems)).To(Equal(0))
	})

	t.Run("it breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		findCounter := 5
		onFind := func(item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < findCounter
		}

		bTree.FindGreaterThanOrEqualMatchType(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(5))

		// this needs to be in the right sub tree, not the left since we find many larger than values,
		foundItems = []any{}
		findCounter = 0
		bTree.FindGreaterThanOrEqualMatchType(datatypes.Int(767), onFind)
		g.Expect(len(foundItems)).To(Equal(1))
	})
}
