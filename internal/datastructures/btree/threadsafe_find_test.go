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

		err := bTree.Find(datatypes.EncapsulatedData{DataType: -1, Value: "bad"}, OnFindTest)
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

		err := bTree.FindNotEqual(datatypes.EncapsulatedData{DataType: -1, Value: "bad"}, func(items []any) {})
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

	t.Run("passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(items []any) {
			foundItems = items
		}

		//bTree.root.print("")
		bTree.FindNotEqual(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(1_023)) // account for 0-1023 except for 512
		g.Expect(foundItems).ToNot(ContainElement("512"))
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

		err := bTree.FindLessThan(datatypes.EncapsulatedData{DataType: -1, Value: "bad"}, func(items []any) {})
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

	t.Run("it does not run the callback if the item is not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onFind := func(item []any) {
			called = true
		}

		g.Expect(bTree.FindLessThan(datatypes.Int64(-1), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(items []any) {
			foundItems = items
		}

		//bTree.root.print("")
		bTree.FindLessThan(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(512)) // account for 0-511
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

		err := bTree.FindLessThanOrEqual(datatypes.EncapsulatedData{DataType: -1, Value: "bad"}, func(items []any) {})
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
		onFind := func(item []any) {
			called = true
		}

		g.Expect(bTree.FindLessThanOrEqual(datatypes.Int64(-1), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(items []any) {
			foundItems = items
		}

		//bTree.root.print("")
		bTree.FindLessThanOrEqual(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(513)) // account for 0-512
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

		err := bTree.FindGreaterThan(datatypes.EncapsulatedData{DataType: -1, Value: "bad"}, func(items []any) {})
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
		onFind := func(item []any) {
			called = true
		}

		g.Expect(bTree.FindGreaterThan(datatypes.Int64(9_001), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(items []any) {
			foundItems = items
		}

		//bTree.root.print("")
		bTree.FindGreaterThan(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(511)) // account for 513-1023
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

		err := bTree.FindGreaterThanOrEqual(datatypes.EncapsulatedData{DataType: -1, Value: "bad"}, func(items []any) {})
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
		onFind := func(item []any) {
			called = true
		}

		g.Expect(bTree.FindGreaterThanOrEqual(datatypes.Int64(9_001), onFind)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("passes a slice of all items found less than the key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(items []any) {
			foundItems = items
		}

		//bTree.root.print("")
		bTree.FindGreaterThanOrEqual(datatypes.Int(512), onFind)
		g.Expect(len(foundItems)).To(Equal(512)) // account for 512-1023
	})
}
