package datastructures

import (
	"testing"

	. "github.com/onsi/gomega"
)

var (
	key0  = IntTreeKey(0)
	key1  = IntTreeKey(1)
	key2  = IntTreeKey(2)
	key3  = IntTreeKey(3)
	key4  = IntTreeKey(4)
	key5  = IntTreeKey(5)
	key6  = IntTreeKey(6)
	key7  = IntTreeKey(7)
	key7  = IntTreeKey(8)
	key7  = IntTreeKey(9)
	key10 = IntTreeKey(10)
	key20 = IntTreeKey(20)
	key30 = IntTreeKey(30)
	key40 = IntTreeKey(40)
	key50 = IntTreeKey(50)
)

func TestBTree_Delete_ShiftNode(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *BRoot {
		bTree, err := NewBTree(6)
		g.Expect(err).ToNot(HaveOccurred())

		// 3, 7, 11, 15, 19 are the root level nodes
		for i := 0; i < 24; i++ {
			_, _ = bTree.FindOrCreate(IntTreeKey(i), "OnFind", newBTreeTester("doesn't matter"))
		}

		// check the tree is 2 rows
		g.Expect(bTree.root.numberOfValues).To(Equal(5))
		g.Expect(bTree.root.children[0].numberOfChildren).To(Equal(0))

		return bTree
	}

	t.Run("shiftNodeRight", func(t *testing.T) {
		t.Run("it shifts values to the right starting from the given index", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.root.shiftNodeRight(1, 1)
			g.Expect(bTree.root.numberOfValues).To(Equal(6))
			g.Expect(bTree.root.values[0].key).To(Equal(IntTreeKey(3)))
			g.Expect(bTree.root.values[1]).To(BeNil())
			g.Expect(bTree.root.values[2].key).To(Equal(IntTreeKey(7)))
			g.Expect(bTree.root.values[3].key).To(Equal(IntTreeKey(11)))
			g.Expect(bTree.root.values[4].key).To(Equal(IntTreeKey(15)))
			g.Expect(bTree.root.values[5].key).To(Equal(IntTreeKey(19)))
		})

		t.Run("it shifts children to the right starting from the given index", func(t *testing.T) {
			bTree := setupTree(g)

			children0 := bTree.root.children[0]
			children1 := bTree.root.children[1]
			children2 := bTree.root.children[2]
			children3 := bTree.root.children[3]
			children4 := bTree.root.children[4]
			children5 := bTree.root.children[5]

			bTree.root.shiftNodeRight(1, 1)
			g.Expect(bTree.root.numberOfChildren).To(Equal(7))
			g.Expect(bTree.root.children[0]).To(Equal(children0))
			g.Expect(bTree.root.children[1]).To(BeNil())
			g.Expect(bTree.root.children[2]).To(Equal(children1))
			g.Expect(bTree.root.children[3]).To(Equal(children2))
			g.Expect(bTree.root.children[4]).To(Equal(children3))
			g.Expect(bTree.root.children[5]).To(Equal(children4))
			g.Expect(bTree.root.children[6]).To(Equal(children5))
		})
	})

	t.Run("shiftNodeLeft", func(t *testing.T) {
		t.Run("it shifts values to the left removing the given index", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.root.shiftNodeLeft(1, 1)
			g.Expect(bTree.root.numberOfValues).To(Equal(4))

			g.Expect(bTree.root.values[0].key).To(Equal(IntTreeKey(3)))
			g.Expect(bTree.root.values[1].key).To(Equal(IntTreeKey(11)))
			g.Expect(bTree.root.values[2].key).To(Equal(IntTreeKey(15)))
			g.Expect(bTree.root.values[3].key).To(Equal(IntTreeKey(19)))
		})

		t.Run("it shifts children to the left removing the given index", func(t *testing.T) {
			bTree := setupTree(g)

			children0 := bTree.root.children[0]
			children2 := bTree.root.children[2]
			children3 := bTree.root.children[3]
			children4 := bTree.root.children[4]
			children5 := bTree.root.children[5]

			bTree.root.shiftNodeLeft(1, 1)
			g.Expect(bTree.root.numberOfChildren).To(Equal(5))
			g.Expect(bTree.root.children[0]).To(Equal(children0))
			g.Expect(bTree.root.children[1]).To(Equal(children2))
			g.Expect(bTree.root.children[2]).To(Equal(children3))
			g.Expect(bTree.root.children[3]).To(Equal(children4))
			g.Expect(bTree.root.children[4]).To(Equal(children5))
		})
	})
}

func TestBTree_Delete_ParamChecks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it panics if the tree key is nil", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(func() { bTree.Delete(nil) }).To(Panic())
	})
}

func TestBTree_Delete_LeafOnly(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns when the tree has no items", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(func() { bTree.Delete(IntTreeKey(1)) }).ToNot(Panic())
	})

	t.Run("removes a tree with a single item sets the root tree to nil", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		_, err = bTree.FindOrCreate(IntTreeKey(1), "", newBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())

		bTree.Delete(IntTreeKey(1))
		g.Expect(bTree.root).To(BeNil())
	})

	t.Run("when there are 2 or more indexes in the root node", func(t *testing.T) {
		// Each test will run with a tree like
		// [1,2]
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			_, _ = bTree.FindOrCreate(key1, "OnFind", newBTreeTester("1"))
			_, _ = bTree.FindOrCreate(key2, "OnFind", newBTreeTester("2"))
			return bTree
		}

		t.Run("it shifts the elements after removing the first index", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(key1)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1]).To(BeNil())
		})

		t.Run("it can remove the last element", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(key2)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key1))
			g.Expect(bTree.root.values[1]).To(BeNil())
		})
	})
}

func TestBTree_Delete_HeightTwo(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("deleting children without any action to tak", func(t *testing.T) {
		// Each test will start with a tree like
		//     2
		//   /   \
		//  0,1  3,4
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			_, _ = bTree.FindOrCreate(key1, "OnFind", newBTreeTester("1"))
			_, _ = bTree.FindOrCreate(key2, "OnFind", newBTreeTester("2"))
			_, _ = bTree.FindOrCreate(key3, "OnFind", newBTreeTester("3"))
			_, _ = bTree.FindOrCreate(key0, "OnFind", newBTreeTester("0"))
			_, _ = bTree.FindOrCreate(key4, "OnFind", newBTreeTester("4"))

			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 := bTree.root.children[0]
			child2 := bTree.root.children[1]
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))
			g.Expect(child2.values[0].key).To(Equal(key3))
			g.Expect(child2.values[1].key).To(Equal(key4))

			return bTree
		}

		t.Run("it can remove a left child value", func(t *testing.T) {
			bTree := setupTree(g)

			// final tree
			//     2
			//   /   \
			//  1   3,4
			bTree.Delete(key0)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))
			g.Expect(child2.values[1].key).To(Equal(key4))

			// final tree
			//     2
			//   /   \
			//  0   3,4
			_, _ = bTree.FindOrCreate(key0, "OnFind", newBTreeTester("0"))
			bTree.Delete(key1)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 = bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))

			child2 = bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))
			g.Expect(child2.values[1].key).To(Equal(key4))
		})

		t.Run("it can remove a right child value", func(t *testing.T) {
			bTree := setupTree(g)

			// final tree
			//     2
			//   /   \
			//  0,1   4
			bTree.Delete(key3)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key4))

			// final tree
			//     2
			//   /   \
			//  0,1   3
			_, _ = bTree.FindOrCreate(key3, "OnFind", newBTreeTester("3"))
			bTree.Delete(key4)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 = bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 = bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))
		})
	})

	t.Run("swapping top layer keys", func(t *testing.T) {
		// Each test will start with a tree like
		//       2,5
		//   /    |   \
		//  0,1  3,4  6,7
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			_, _ = bTree.FindOrCreate(key1, "OnFind", newBTreeTester("1"))
			_, _ = bTree.FindOrCreate(key2, "OnFind", newBTreeTester("2"))
			_, _ = bTree.FindOrCreate(key3, "OnFind", newBTreeTester("3"))
			_, _ = bTree.FindOrCreate(key0, "OnFind", newBTreeTester("0"))
			_, _ = bTree.FindOrCreate(key6, "OnFind", newBTreeTester("6"))
			_, _ = bTree.FindOrCreate(key5, "OnFind", newBTreeTester("5"))
			_, _ = bTree.FindOrCreate(key4, "OnFind", newBTreeTester("4"))
			_, _ = bTree.FindOrCreate(key7, "OnFind", newBTreeTester("7"))

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key5))

			g.Expect(bTree.root.numberOfChildren).To(Equal(3))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.values[0].key).To(Equal(key3))
			g.Expect(child2.values[1].key).To(Equal(key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.values[0].key).To(Equal(key6))
			g.Expect(child3.values[1].key).To(Equal(key7))

			return bTree
		}

		t.Run("removing left keys swaps with the left child", func(t *testing.T) {
			// final tree
			//       1,5
			//     /  |   \
			//    0  3,4  6,7
			bTree := setupTree(g)

			bTree.Delete(key2)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key1))
			g.Expect(bTree.root.values[1].key).To(Equal(key5))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))
			g.Expect(child2.values[1].key).To(Equal(key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key6))
			g.Expect(child3.values[1].key).To(Equal(key7))
		})

		t.Run("removing a right key swaps with the left key first", func(t *testing.T) {
			// final tree
			//       2,4
			//     /  |   \
			//    0,1 3  6,7
			bTree := setupTree(g)

			bTree.Delete(key5)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key4))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key6))
			g.Expect(child3.values[1].key).To(Equal(key7))
		})

		t.Run("removing a key swaps with the right as a last resort", func(t *testing.T) {
			// final tree
			//       2,6
			//     /  |  \
			//    0,1 3   7
			bTree := setupTree(g)

			bTree.Delete(key5)
			bTree.Delete(key4)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key6))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key7))
		})
	})

	t.Run("rotating single layer keys", func(t *testing.T) {
		// Each test will start with a tree like
		//       2,5
		//   /    |   \
		//  0,1  3,4  6,7
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			_, _ = bTree.FindOrCreate(key1, "OnFind", newBTreeTester("1"))
			_, _ = bTree.FindOrCreate(key2, "OnFind", newBTreeTester("2"))
			_, _ = bTree.FindOrCreate(key3, "OnFind", newBTreeTester("3"))
			_, _ = bTree.FindOrCreate(key0, "OnFind", newBTreeTester("0"))
			_, _ = bTree.FindOrCreate(key6, "OnFind", newBTreeTester("6"))
			_, _ = bTree.FindOrCreate(key5, "OnFind", newBTreeTester("5"))
			_, _ = bTree.FindOrCreate(key4, "OnFind", newBTreeTester("4"))
			_, _ = bTree.FindOrCreate(key7, "OnFind", newBTreeTester("7"))

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key5))

			g.Expect(bTree.root.numberOfChildren).To(Equal(3))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.values[0].key).To(Equal(key3))
			g.Expect(child2.values[1].key).To(Equal(key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.values[0].key).To(Equal(key6))
			g.Expect(child3.values[1].key).To(Equal(key7))

			return bTree
		}

		t.Run("removing left keys rotates a middle child left", func(t *testing.T) {
			// final tree
			//       3,5
			//     /  |  \
			//    2   4  6,7
			bTree := setupTree(g)

			bTree.Delete(key0)
			bTree.Delete(key1)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key3))
			g.Expect(bTree.root.values[1].key).To(Equal(key5))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key2))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key6))
			g.Expect(child3.values[1].key).To(Equal(key7))
		})

		t.Run("removing a middle key rotates a right child to the left", func(t *testing.T) {
			// final tree
			//       2,6
			//     /  |  \
			//   0,1  5   7
			bTree := setupTree(g)

			bTree.Delete(key3)
			bTree.Delete(key4)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key6))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key5))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key7))
		})

		t.Run("removing a right key rotates a left child to the right iff its the only option", func(t *testing.T) {
			// final tree
			//       2,4
			//     /  |  \
			//   0,1  3   5
			bTree := setupTree(g)

			bTree.Delete(key6)
			bTree.Delete(key7)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key4))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key5))
		})
	})

	t.Run("squashing single layer trees", func(t *testing.T) {
		// Each test will start with a tree like
		//     2
		//   /   \
		//  1     3
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			_, _ = bTree.FindOrCreate(key1, "OnFind", newBTreeTester("1"))
			_, _ = bTree.FindOrCreate(key2, "OnFind", newBTreeTester("2"))
			_, _ = bTree.FindOrCreate(key3, "OnFind", newBTreeTester("3"))

			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 := bTree.root.children[0]
			child2 := bTree.root.children[1]
			g.Expect(child1.values[0].key).To(Equal(key1))
			g.Expect(child2.values[0].key).To(Equal(key3))

			return bTree
		}

		t.Run("it setups a one node tree when the left child is deleted", func(t *testing.T) {
			// final tree
			// [2,3]
			bTree := setupTree(g)

			bTree.Delete(key1)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key3))
		})

		t.Run("it setups a one node tree when the right child is deleted", func(t *testing.T) {
			// final tree
			// [1,2]
			bTree := setupTree(g)

			bTree.Delete(key3)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key1))
			g.Expect(bTree.root.values[1].key).To(Equal(key2))
		})

		t.Run("it sets up a one node tree when the root value is deleted", func(t *testing.T) {
			// final tree
			// [1,3]
			bTree := setupTree(g)

			bTree.Delete(key2)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key1))
			g.Expect(bTree.root.values[1].key).To(Equal(key3))
		})
	})

	t.Run("merging keys", func(t *testing.T) {
		// Each test will start with a tree like
		//     2,4
		//   /  |  \
		//  1   3   5
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			_, _ = bTree.FindOrCreate(key1, "OnFind", newBTreeTester("1"))
			_, _ = bTree.FindOrCreate(key2, "OnFind", newBTreeTester("2"))
			_, _ = bTree.FindOrCreate(key3, "OnFind", newBTreeTester("3"))
			_, _ = bTree.FindOrCreate(keyFour, "OnFind", newBTreeTester("4"))
			_, _ = bTree.FindOrCreate(keyFive, "OnFind", newBTreeTester("5"))
			return bTree
		}

		t.Run("deleteing the left child merges properly", func(t *testing.T) {
			// final tree
			//      4
			//    /   \
			//  2,3    5
			bTree := setupTree(g)

			bTree.Delete(key1)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(keyFour))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.values[0].key).To(Equal(key2))
			g.Expect(child0.values[1].key).To(Equal(key3))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(keyFive))
		})

		t.Run("deleteing a middle child merges properly", func(t *testing.T) {
			// final tree
			//      4
			//    /   \
			//  1,2    5
			bTree := setupTree(g)

			bTree.Delete(key3)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key4))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.values[0].key).To(Equal(key1))
			g.Expect(child0.values[1].key).To(Equal(key2))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key5))
		})

		t.Run("deleteing a right child merges properly", func(t *testing.T) {
			// final tree
			//      2
			//    /   \
			//  1     3,4
			bTree := setupTree(g)

			bTree.Delete(keyFive)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(1))
			g.Expect(child0.values[0].key).To(Equal(key1))

			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.values[0].key).To(Equal(key3))
			g.Expect(child1.values[1].key).To(Equal(keyFour))
		})

		t.Run("deleteing a left node value merges properly", func(t *testing.T) {
			// final tree
			//      4
			//    /   \
			//  1,3		 5
			bTree := setupTree(g)

			bTree.Delete(key2)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key4))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.values[0].key).To(Equal(key1))
			g.Expect(child0.values[1].key).To(Equal(key3))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key5))
		})

		t.Run("deleteing a right node value merges properly", func(t *testing.T) {
			// note, we always swap on the left and merge to left
			// final tree
			//      3
			//    /   \
			//  1,2		 5
			bTree := setupTree(g)

			bTree.Delete(keyFour)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key3))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.values[0].key).To(Equal(key1))
			g.Expect(child0.values[1].key).To(Equal(key2))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key5))
		})
	})
}

// TODO this is the start tomorrow
func TestBTree_Delete_HeightThreeAndAbove(t *testing.T) {
	g := NewGomegaWithT(t)

	// Each test will start with a tree like
	//       2,5
	//   /    |   \
	//  0,1  3,4  6,7
	setupTree := func(g *GomegaWithT) *BRoot {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		_, _ = bTree.FindOrCreate(key0, "OnFind", newBTreeTester("0"))
		_, _ = bTree.FindOrCreate(key1, "OnFind", newBTreeTester("1"))
		_, _ = bTree.FindOrCreate(key2, "OnFind", newBTreeTester("2"))
		_, _ = bTree.FindOrCreate(key3, "OnFind", newBTreeTester("3"))
		_, _ = bTree.FindOrCreate(key4, "OnFind", newBTreeTester("4"))
		_, _ = bTree.FindOrCreate(key5, "OnFind", newBTreeTester("5"))
		_, _ = bTree.FindOrCreate(key6, "OnFind", newBTreeTester("6"))
		_, _ = bTree.FindOrCreate(key7, "OnFind", newBTreeTester("7"))
		_, _ = bTree.FindOrCreate(key8, "OnFind", newBTreeTester("8"))
		_, _ = bTree.FindOrCreate(key9, "OnFind", newBTreeTester("9"))
		_, _ = bTree.FindOrCreate(key10, "OnFind", newBTreeTester("10"))

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(key2))
		g.Expect(bTree.root.values[1].key).To(Equal(key5))

		g.Expect(bTree.root.numberOfChildren).To(Equal(3))

		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfChildren).To(Equal(0))
		g.Expect(child1.numberOfValues).To(Equal(2))
		g.Expect(child1.values[0].key).To(Equal(key0))
		g.Expect(child1.values[1].key).To(Equal(key1))

		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfChildren).To(Equal(0))
		g.Expect(child2.numberOfValues).To(Equal(2))
		g.Expect(child2.values[0].key).To(Equal(key3))
		g.Expect(child2.values[1].key).To(Equal(key4))

		child3 := bTree.root.children[2]
		g.Expect(child3.numberOfChildren).To(Equal(0))
		g.Expect(child3.numberOfValues).To(Equal(2))
		g.Expect(child3.values[0].key).To(Equal(key6))
		g.Expect(child3.values[1].key).To(Equal(key7))

		return bTree
	}

	t.Run("swapping top layer keys", func(t *testing.T) {
		t.Run("removing left keys swaps with the left child", func(t *testing.T) {
			// final tree
			//       1,5
			//     /  |   \
			//    0  3,4  6,7
			bTree := setupTree(g)

			bTree.Delete(key2)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key1))
			g.Expect(bTree.root.values[1].key).To(Equal(key5))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))
			g.Expect(child2.values[1].key).To(Equal(key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key6))
			g.Expect(child3.values[1].key).To(Equal(key7))
		})

		t.Run("removing a right key swaps with the left key first", func(t *testing.T) {
			// final tree
			//       2,4
			//     /  |   \
			//    0,1 3  6,7
			bTree := setupTree(g)

			bTree.Delete(key5)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key4))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key6))
			g.Expect(child3.values[1].key).To(Equal(key7))
		})

		t.Run("removing a key swaps with the right as a last resort", func(t *testing.T) {
			// final tree
			//       2,6
			//     /  |  \
			//    0,1 3   7
			bTree := setupTree(g)

			bTree.Delete(key5)
			bTree.Delete(key4)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key6))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key7))
		})
	})

	t.Run("rotating single layer keys", func(t *testing.T) {
		// Each test will start with a tree like
		//       2,5
		//   /    |   \
		//  0,1  3,4  6,7
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			_, _ = bTree.FindOrCreate(key1, "OnFind", newBTreeTester("1"))
			_, _ = bTree.FindOrCreate(key2, "OnFind", newBTreeTester("2"))
			_, _ = bTree.FindOrCreate(key3, "OnFind", newBTreeTester("3"))
			_, _ = bTree.FindOrCreate(key0, "OnFind", newBTreeTester("0"))
			_, _ = bTree.FindOrCreate(key6, "OnFind", newBTreeTester("6"))
			_, _ = bTree.FindOrCreate(key5, "OnFind", newBTreeTester("5"))
			_, _ = bTree.FindOrCreate(key4, "OnFind", newBTreeTester("4"))
			_, _ = bTree.FindOrCreate(key7, "OnFind", newBTreeTester("7"))

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key5))

			g.Expect(bTree.root.numberOfChildren).To(Equal(3))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.values[0].key).To(Equal(key3))
			g.Expect(child2.values[1].key).To(Equal(key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.values[0].key).To(Equal(key6))
			g.Expect(child3.values[1].key).To(Equal(key7))

			return bTree
		}

		t.Run("removing left keys rotates a middle child left", func(t *testing.T) {
			// final tree
			//       3,5
			//     /  |  \
			//    2   4  6,7
			bTree := setupTree(g)

			bTree.Delete(key0)
			bTree.Delete(key1)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key3))
			g.Expect(bTree.root.values[1].key).To(Equal(key5))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key2))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key6))
			g.Expect(child3.values[1].key).To(Equal(key7))
		})

		t.Run("removing a middle key rotates a right child to the left", func(t *testing.T) {
			// final tree
			//       2,6
			//     /  |  \
			//   0,1  5   7
			bTree := setupTree(g)

			bTree.Delete(key3)
			bTree.Delete(key4)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key6))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key5))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key7))
		})

		t.Run("removing a right key rotates a left child to the right iff its the only option", func(t *testing.T) {
			// final tree
			//       2,4
			//     /  |  \
			//   0,1  3   5
			bTree := setupTree(g)

			bTree.Delete(key6)
			bTree.Delete(key7)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key4))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.values[0].key).To(Equal(key0))
			g.Expect(child1.values[1].key).To(Equal(key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.values[0].key).To(Equal(key3))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.values[0].key).To(Equal(key5))
		})
	})

	t.Run("squashing single layer trees", func(t *testing.T) {
		t.Run("it setups a one node tree when the left child is deleted", func(t *testing.T) {
			// final tree
			// [2,3]
			bTree := setupTree(g)

			bTree.Delete(key1)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))
			g.Expect(bTree.root.values[1].key).To(Equal(key3))
		})

		t.Run("it setups a one node tree when the right child is deleted", func(t *testing.T) {
			// final tree
			// [1,2]
			bTree := setupTree(g)

			bTree.Delete(key3)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key1))
			g.Expect(bTree.root.values[1].key).To(Equal(key2))
		})

		t.Run("it sets up a one node tree when the root value is deleted", func(t *testing.T) {
			// final tree
			// [1,3]
			bTree := setupTree(g)

			bTree.Delete(key2)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key1))
			g.Expect(bTree.root.values[1].key).To(Equal(key3))
		})
	})

	t.Run("merging keys", func(t *testing.T) {
		t.Run("deleteing the left child merges properly", func(t *testing.T) {
			// final tree
			//      4
			//    /   \
			//  2,3    5
			bTree := setupTree(g)

			bTree.Delete(key1)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(keyFour))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.values[0].key).To(Equal(key2))
			g.Expect(child0.values[1].key).To(Equal(key3))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(keyFive))
		})

		t.Run("deleteing a middle child merges properly", func(t *testing.T) {
			// final tree
			//      4
			//    /   \
			//  1,2    5
			bTree := setupTree(g)

			bTree.Delete(key3)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key4))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.values[0].key).To(Equal(key1))
			g.Expect(child0.values[1].key).To(Equal(key2))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key5))
		})

		t.Run("deleteing a right child merges properly", func(t *testing.T) {
			// final tree
			//      2
			//    /   \
			//  1     3,4
			bTree := setupTree(g)

			bTree.Delete(keyFive)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key2))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(1))
			g.Expect(child0.values[0].key).To(Equal(key1))

			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.values[0].key).To(Equal(key3))
			g.Expect(child1.values[1].key).To(Equal(keyFour))
		})

		t.Run("deleteing a left node value merges properly", func(t *testing.T) {
			// final tree
			//      4
			//    /   \
			//  1,3		 5
			bTree := setupTree(g)

			bTree.Delete(key2)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key4))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.values[0].key).To(Equal(key1))
			g.Expect(child0.values[1].key).To(Equal(key3))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key5))
		})

		t.Run("deleteing a right node value merges properly", func(t *testing.T) {
			// note, we always swap on the left and merge to left
			// final tree
			//      3
			//    /   \
			//  1,2		 5
			bTree := setupTree(g)

			bTree.Delete(keyFour)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.values[0].key).To(Equal(key3))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.values[0].key).To(Equal(key1))
			g.Expect(child0.values[1].key).To(Equal(key2))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key5))
		})
	})
}
