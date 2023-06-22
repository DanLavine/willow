package btree

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestBTree_Delete_ParamChecks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.Delete(datatypes.EncapsulatedData{DataType: -1, Value: "bad"}, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})
}

func TestBTree_Delete_ShiftNode(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(6)
		g.Expect(err).ToNot(HaveOccurred())

		// 3, 7, 11, 15, 19, nil are the root level nodes
		for i := 0; i < 24; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester("doesn't matter"), OnFindTest)).ToNot(HaveOccurred())
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
			g.Expect(bTree.root.keyValues[0].key).To(Equal(datatypes.Int(3)))
			g.Expect(bTree.root.keyValues[1]).To(BeNil())
			g.Expect(bTree.root.keyValues[2].key).To(Equal(datatypes.Int(7)))
			g.Expect(bTree.root.keyValues[3].key).To(Equal(datatypes.Int(11)))
			g.Expect(bTree.root.keyValues[4].key).To(Equal(datatypes.Int(15)))
			g.Expect(bTree.root.keyValues[5].key).To(Equal(datatypes.Int(19)))
		})

		t.Run("it shifts children to the right starting from the given index", func(t *testing.T) {
			bTree := setupTree(g)

			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]
			child2 := bTree.root.children[2]
			child3 := bTree.root.children[3]
			child4 := bTree.root.children[4]
			child5 := bTree.root.children[5]

			bTree.root.shiftNodeRight(1, 1)
			g.Expect(bTree.root.numberOfChildren).To(Equal(7))
			g.Expect(bTree.root.children[0]).To(Equal(child0))
			g.Expect(bTree.root.children[1]).To(BeNil())
			g.Expect(bTree.root.children[2]).To(Equal(child1))
			g.Expect(bTree.root.children[3]).To(Equal(child2))
			g.Expect(bTree.root.children[4]).To(Equal(child3))
			g.Expect(bTree.root.children[5]).To(Equal(child4))
			g.Expect(bTree.root.children[6]).To(Equal(child5))
		})
	})

	t.Run("shiftNodeLeft", func(t *testing.T) {
		t.Run("it shifts values to the left removing the given index", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.root.shiftNodeLeft(1, 1)
			g.Expect(bTree.root.numberOfValues).To(Equal(4))

			g.Expect(bTree.root.keyValues[0].key).To(Equal(datatypes.Int(3)))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(datatypes.Int(11)))
			g.Expect(bTree.root.keyValues[2].key).To(Equal(datatypes.Int(15)))
			g.Expect(bTree.root.keyValues[3].key).To(Equal(datatypes.Int(19)))
		})

		t.Run("it shifts children to the left removing the given index", func(t *testing.T) {
			bTree := setupTree(g)

			child0 := bTree.root.children[0]
			child2 := bTree.root.children[2]
			child3 := bTree.root.children[3]
			child4 := bTree.root.children[4]
			child5 := bTree.root.children[5]

			bTree.root.shiftNodeLeft(1, 1)
			g.Expect(bTree.root.numberOfChildren).To(Equal(5))
			g.Expect(bTree.root.children[0]).To(Equal(child0))
			g.Expect(bTree.root.children[1]).To(Equal(child2))
			g.Expect(bTree.root.children[2]).To(Equal(child3))
			g.Expect(bTree.root.children[3]).To(Equal(child4))
			g.Expect(bTree.root.children[4]).To(Equal(child5))
		})
	})
}

func TestBTree_Delete_CanDeleteChecks(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 100; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("always deletes the item if the CanDelete callback check is nil", func(t *testing.T) {
		bTree := setupTree(g)
		bTree.Delete(Key1, nil)

		found := false
		onFind := func(item any) {
			found = true
		}

		bTree.Find(Key1, onFind)
		g.Expect(found).To(BeFalse())
	})

	t.Run("always deletes the item if canDelete returns true", func(t *testing.T) {
		canDelete := func(item any) bool {
			return true
		}

		t.Run("on leaf keys", func(t *testing.T) {
			bTree := setupTree(g)
			bTree.Delete(Key0, canDelete)

			found := false
			onFind := func(item any) {
				found = true
			}

			bTree.Find(Key0, onFind)
			g.Expect(found).To(BeFalse())
		})

		t.Run("on internal node keys", func(t *testing.T) {
			bTree := setupTree(g)
			key := bTree.root.keyValues[0].key
			bTree.Delete(key, canDelete)

			found := false
			onFind := func(item any) {
				found = true
			}

			bTree.Find(key, onFind)
			g.Expect(found).To(BeFalse())
		})
	})

	t.Run("does not deletes the item if canDelete returns false", func(t *testing.T) {
		canDelete := func(item any) bool {
			return false
		}

		t.Run("on leaf keys", func(t *testing.T) {
			bTree := setupTree(g)
			bTree.Delete(Key0, canDelete)

			found := false
			onFind := func(item any) {
				found = true
			}

			bTree.Find(Key0, onFind)
			g.Expect(found).To(BeTrue())
		})

		t.Run("on internal node keys", func(t *testing.T) {
			bTree := setupTree(g)
			key := bTree.root.keyValues[0].key
			bTree.Delete(key, canDelete)

			found := false
			onFind := func(item any) {
				found = true
			}

			bTree.Find(Key0, onFind)
			g.Expect(found).To(BeTrue())
		})
	})
}

func TestBTree_Delete_SingleNodeHeight(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns when the tree has no items", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Delete(Key1, nil)).ToNot(HaveOccurred())
	})

	t.Run("removes a tree with a single item sets the root tree to nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())

		bTree.Delete(Key1, nil)
		g.Expect(bTree.root).To(BeNil())
	})

	t.Run("when there are 2 or more indexes in the root node", func(t *testing.T) {
		// Each test will run with a tree like
		// [1,2]
		setupTree := func(g *GomegaWithT) *threadSafeBTree {
			bTree, err := NewThreadSafe(2)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())

			return bTree
		}

		t.Run("it shifts the elements after removing the first index", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key1, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.keyValues[1]).To(BeNil())
		})

		t.Run("it can remove the last element", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key2, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key1))
			g.Expect(bTree.root.keyValues[1]).To(BeNil())
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
		setupTree := func(g *GomegaWithT) *threadSafeBTree {
			bTree, err := NewThreadSafe(2)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key3, NewBTreeTester("3"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key0, NewBTreeTester("0"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key4, NewBTreeTester("4"), OnFindTest)).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 := bTree.root.children[0]
			child2 := bTree.root.children[1]
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))
			g.Expect(child1.keyValues[1].key).To(Equal(Key1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))
			g.Expect(child2.keyValues[1].key).To(Equal(Key4))

			return bTree
		}

		// final tree
		//     2
		//   /   \
		//  1   3,4
		t.Run("it can remove a left[0] child value", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key0, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))
			g.Expect(child2.keyValues[1].key).To(Equal(Key4))
		})

		// final tree
		//     2
		//   /   \
		//  0   3,4
		t.Run("it can remove a left[1] child value", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key1, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))
			g.Expect(child2.keyValues[1].key).To(Equal(Key4))
		})

		// final tree
		//     2
		//   /   \
		//  0,1   4
		t.Run("it can remove a right[0] child value", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key3, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))
			g.Expect(child1.keyValues[1].key).To(Equal(Key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key4))
		})

		// final tree
		//     2
		//   /   \
		//  0,1   3
		t.Run("it can remove a right[1] child value", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key4, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))
			g.Expect(child1.keyValues[1].key).To(Equal(Key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))
		})
	})

	t.Run("swapping top layer keys", func(t *testing.T) {
		// Each test will start with a tree like
		//       2,5
		//   /    |   \
		//  0,1  3,4  6,7
		setupTree := func(g *GomegaWithT) *threadSafeBTree {
			bTree, err := NewThreadSafe(2)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key3, NewBTreeTester("3"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key0, NewBTreeTester("0"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key6, NewBTreeTester("6"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key5, NewBTreeTester("5"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key4, NewBTreeTester("4"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key7, NewBTreeTester("7"), OnFindTest)).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key5))

			g.Expect(bTree.root.numberOfChildren).To(Equal(3))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))
			g.Expect(child1.keyValues[1].key).To(Equal(Key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))
			g.Expect(child2.keyValues[1].key).To(Equal(Key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.keyValues[0].key).To(Equal(Key6))
			g.Expect(child3.keyValues[1].key).To(Equal(Key7))

			return bTree
		}

		// final tree
		//       1,5
		//     /  |   \
		//    0  3,4  6,7
		t.Run("removing key[0] swaps with the left child", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key2, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key1))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key5))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))
			g.Expect(child2.keyValues[1].key).To(Equal(Key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.keyValues[0].key).To(Equal(Key6))
			g.Expect(child3.keyValues[1].key).To(Equal(Key7))
		})

		// final tree
		//       2,4
		//     /  |   \
		//    0,1 3  6,7
		t.Run("removing a key[1] swaps with the left child key first", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key5, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key4))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))
			g.Expect(child1.keyValues[1].key).To(Equal(Key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.keyValues[0].key).To(Equal(Key6))
			g.Expect(child3.keyValues[1].key).To(Equal(Key7))
		})

		// final tree
		//       2,6
		//     /  |  \
		//    0,1 3   7
		t.Run("removing key[1] swaps with the right child as a last resort", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key5, nil)
			bTree.Delete(Key4, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key6))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))
			g.Expect(child1.keyValues[1].key).To(Equal(Key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.keyValues[0].key).To(Equal(Key7))
		})
	})

	t.Run("rotating single layer keys", func(t *testing.T) {
		// Each test will start with a tree like
		//       2,5
		//   /    |   \
		//  0,1  3,4  6,7
		setupTree := func(g *GomegaWithT) *threadSafeBTree {
			bTree, err := NewThreadSafe(2)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key3, NewBTreeTester("3"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key0, NewBTreeTester("0"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key6, NewBTreeTester("6"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key5, NewBTreeTester("5"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key4, NewBTreeTester("4"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key7, NewBTreeTester("7"), OnFindTest)).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key5))

			g.Expect(bTree.root.numberOfChildren).To(Equal(3))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))
			g.Expect(child1.keyValues[1].key).To(Equal(Key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))
			g.Expect(child2.keyValues[1].key).To(Equal(Key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.keyValues[0].key).To(Equal(Key6))
			g.Expect(child3.keyValues[1].key).To(Equal(Key7))

			return bTree
		}

		// final tree
		//       3,5
		//     /  |  \
		//    2   4  6,7
		t.Run("removing all child[0] keys rotates a child[1] key to the left", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key0, nil)
			bTree.Delete(Key1, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key3))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key5))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key2))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key4))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(2))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.keyValues[0].key).To(Equal(Key6))
			g.Expect(child3.keyValues[1].key).To(Equal(Key7))
		})

		// final tree
		//       2,6
		//     /  |  \
		//   0,1  5   7
		t.Run("removing all child[1] keys rotates a child[2] key to the left", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key3, nil)
			bTree.Delete(Key4, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key6))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))
			g.Expect(child1.keyValues[1].key).To(Equal(Key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key5))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.keyValues[0].key).To(Equal(Key7))
		})

		// final tree
		//       1,5
		//     /  |  \
		//   0    2   7
		t.Run("removing all child[1] keys rotates a child[0] key to the right iff child[2] has not available keys", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key6, nil)
			bTree.Delete(Key3, nil)
			bTree.Delete(Key4, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key1))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key5))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key2))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.keyValues[0].key).To(Equal(Key7))
		})

		// final tree
		//       2,4
		//     /  |  \
		//   0,1  3   5
		t.Run("removing all child[2] keys rotates a child[1] to the right iff its the only option", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key6, nil)
			bTree.Delete(Key7, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key4))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.numberOfChildren).To(Equal(0))
			g.Expect(child1.keyValues[0].key).To(Equal(Key0))
			g.Expect(child1.keyValues[1].key).To(Equal(Key1))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.numberOfChildren).To(Equal(0))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.numberOfChildren).To(Equal(0))
			g.Expect(child3.keyValues[0].key).To(Equal(Key5))
		})
	})

	t.Run("squashing single layer trees", func(t *testing.T) {
		// Each test will start with a tree like
		//     2
		//   /   \
		//  1     3
		setupTree := func(g *GomegaWithT) *threadSafeBTree {
			bTree, err := NewThreadSafe(2)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key3, NewBTreeTester("3"), OnFindTest)).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))

			child1 := bTree.root.children[0]
			child2 := bTree.root.children[1]
			g.Expect(child1.keyValues[0].key).To(Equal(Key1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key3))

			return bTree
		}

		// final tree
		// [2,3]
		t.Run("it setups a one node tree when child[0] is deleted", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key1, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key3))
		})

		// final tree
		// [1,2]
		t.Run("it setups a one node tree when child[1] is deleted", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key3, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key1))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key2))
		})

		// final tree
		// [1,3]
		t.Run("it sets up a one node tree when the root value is deleted", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key2, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key1))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key3))
		})
	})

	t.Run("merging keys", func(t *testing.T) {
		// Each test will start with a tree like
		//     2,4
		//   /  |  \
		//  1   3   5
		setupTree := func(g *GomegaWithT) *threadSafeBTree {
			bTree, err := NewThreadSafe(2)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key3, NewBTreeTester("3"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key4, NewBTreeTester("4"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key5, NewBTreeTester("5"), OnFindTest)).ToNot(HaveOccurred())

			return bTree
		}

		// final tree
		//      4
		//    /   \
		//  2,3    5
		t.Run("deleteing the child[0] merges properly", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key1, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key4))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.keyValues[0].key).To(Equal(Key2))
			g.Expect(child0.keyValues[1].key).To(Equal(Key3))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key5))
		})

		// final tree
		//      4
		//    /   \
		//  1,2    5
		t.Run("deleteing a child[1] merges properly", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key3, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key4))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.keyValues[0].key).To(Equal(Key1))
			g.Expect(child0.keyValues[1].key).To(Equal(Key2))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key5))
		})

		// final tree
		//      2
		//    /   \
		//  1     3,4
		t.Run("deleteing a child[2] merges properly", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key5, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(1))
			g.Expect(child0.keyValues[0].key).To(Equal(Key1))

			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.keyValues[0].key).To(Equal(Key3))
			g.Expect(child1.keyValues[1].key).To(Equal(Key4))
		})

		// final tree
		//      4
		//    /   \
		//  1,3		 5
		t.Run("deleteing root[0] value merges properly", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key2, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key4))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.keyValues[0].key).To(Equal(Key1))
			g.Expect(child0.keyValues[1].key).To(Equal(Key3))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key5))
		})

		// note, we always swap on the left and merge to left
		// final tree
		//      3
		//    /   \
		//  1,2		 5
		t.Run("deleteing root[1] value merges properly", func(t *testing.T) {
			bTree := setupTree(g)

			bTree.Delete(Key4, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key3))

			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			child0 := bTree.root.children[0]
			child1 := bTree.root.children[1]

			g.Expect(child0.numberOfValues).To(Equal(2))
			g.Expect(child0.keyValues[0].key).To(Equal(Key1))
			g.Expect(child0.keyValues[1].key).To(Equal(Key2))

			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key5))
		})
	})
}

func TestBTree_Delete_HeightThreeAndAbove(t *testing.T) {
	g := NewGomegaWithT(t)

	// Each test will start with a tree like
	//
	//                40,80
	//       /          |         \
	//    20,35        60,75      100,120
	//   /  |   \     /  |  \    /   |   \
	//  10, 30 ,38   50  70 78  90  110  130
	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(Key10, NewBTreeTester("10"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key20, NewBTreeTester("20"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key30, NewBTreeTester("30"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key40, NewBTreeTester("40"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key50, NewBTreeTester("50"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key60, NewBTreeTester("60"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key70, NewBTreeTester("70"), OnFindTest)).ToNot(HaveOccurred())

		//  fill in left tree
		g.Expect(bTree.CreateOrFind(Key35, NewBTreeTester("35"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key38, NewBTreeTester("38"), OnFindTest)).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(Key80, NewBTreeTester("80"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key90, NewBTreeTester("90"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key100, NewBTreeTester("100"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key110, NewBTreeTester("110"), OnFindTest)).ToNot(HaveOccurred())

		// fill in middle tree
		g.Expect(bTree.CreateOrFind(Key75, NewBTreeTester("75"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key78, NewBTreeTester("78"), OnFindTest)).ToNot(HaveOccurred())

		// fill in right tree
		g.Expect(bTree.CreateOrFind(Key120, NewBTreeTester("120"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key130, NewBTreeTester("130"), OnFindTest)).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.numberOfChildren).To(Equal(3))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key40))
		g.Expect(bTree.root.keyValues[1].key).To(Equal(Key80))

		// left tree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfChildren).To(Equal(3))
		g.Expect(child1.numberOfValues).To(Equal(2))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))
		g.Expect(child1.keyValues[1].key).To(Equal(Key35))

		gchild1_1 := child1.children[0]
		g.Expect(gchild1_1.numberOfChildren).To(Equal(0))
		g.Expect(gchild1_1.numberOfValues).To(Equal(1))
		g.Expect(gchild1_1.keyValues[0].key).To(Equal(Key10))

		gchild1_2 := child1.children[1]
		g.Expect(gchild1_2.numberOfChildren).To(Equal(0))
		g.Expect(gchild1_2.numberOfValues).To(Equal(1))
		g.Expect(gchild1_2.keyValues[0].key).To(Equal(Key30))

		gchild1_3 := child1.children[2]
		g.Expect(gchild1_3.numberOfChildren).To(Equal(0))
		g.Expect(gchild1_3.numberOfValues).To(Equal(1))
		g.Expect(gchild1_3.keyValues[0].key).To(Equal(Key38))

		// middle tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfChildren).To(Equal(3))
		g.Expect(child2.numberOfValues).To(Equal(2))
		g.Expect(child2.keyValues[0].key).To(Equal(Key60))
		g.Expect(child2.keyValues[1].key).To(Equal(Key75))

		gchild2_1 := child2.children[0]
		g.Expect(gchild2_1.numberOfChildren).To(Equal(0))
		g.Expect(gchild2_1.numberOfValues).To(Equal(1))
		g.Expect(gchild2_1.keyValues[0].key).To(Equal(Key50))

		gchild2_2 := child2.children[1]
		g.Expect(gchild2_2.numberOfChildren).To(Equal(0))
		g.Expect(gchild2_2.numberOfValues).To(Equal(1))
		g.Expect(gchild2_2.keyValues[0].key).To(Equal(Key70))

		gchild2_3 := child2.children[2]
		g.Expect(gchild2_3.numberOfChildren).To(Equal(0))
		g.Expect(gchild2_3.numberOfValues).To(Equal(1))
		g.Expect(gchild2_3.keyValues[0].key).To(Equal(Key78))

		// right tree
		child3 := bTree.root.children[2]
		g.Expect(child3.numberOfChildren).To(Equal(3))
		g.Expect(child3.numberOfValues).To(Equal(2))
		g.Expect(child3.keyValues[0].key).To(Equal(Key100))
		g.Expect(child3.keyValues[1].key).To(Equal(Key120))

		gchild3_1 := child3.children[0]
		g.Expect(gchild3_1.numberOfChildren).To(Equal(0))
		g.Expect(gchild3_1.numberOfValues).To(Equal(1))
		g.Expect(gchild3_1.keyValues[0].key).To(Equal(Key90))

		gchild3_2 := child3.children[1]
		g.Expect(gchild3_2.numberOfChildren).To(Equal(0))
		g.Expect(gchild3_2.numberOfValues).To(Equal(1))
		g.Expect(gchild3_2.keyValues[0].key).To(Equal(Key110))

		gchild3_3 := child3.children[2]
		g.Expect(gchild3_3.numberOfChildren).To(Equal(0))
		g.Expect(gchild3_3.numberOfValues).To(Equal(1))
		g.Expect(gchild3_3.keyValues[0].key).To(Equal(Key130))

		return bTree
	}

	t.Run("swapping internal node keys", func(t *testing.T) {
		t.Run("it choses the left child's greates value when the child nodes have the same number of values", func(t *testing.T) {
			// final tree
			//
			//              38,80
			//       /        |         \
			//     20       60,75      100,120
			//   /   \     /  |  \    /   |   \
			//  10, 30,35  50  70 78  90  110  130
			bTree := setupTree(g)

			bTree.Delete(Key40, nil)
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key38))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key80))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfChildren).To(Equal(2))
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key20))

			gchild1_1 := child1.children[0]
			g.Expect(gchild1_1.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_1.numberOfValues).To(Equal(1))
			g.Expect(gchild1_1.keyValues[0].key).To(Equal(Key10))

			gchild1_2 := child1.children[1]
			g.Expect(gchild1_2.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_2.numberOfValues).To(Equal(2))
			g.Expect(gchild1_2.keyValues[0].key).To(Equal(Key30))
			g.Expect(gchild1_2.keyValues[1].key).To(Equal(Key35))

			// validate the rest of tree
			validateThreadSafeTree(g, bTree.root)
		})

		t.Run("the right child's smalles value is chosen iff the left child has less values", func(t *testing.T) {
			// final tree
			//
			//              50,80
			//       /        |         \
			//     20        75      100,120
			//   /   \      /  \    /   |   \
			//  10, 30,35 60,70 78  90  110  130
			bTree := setupTree(g)

			bTree.Delete(Key40, nil)
			bTree.Delete(Key38, nil)
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key50))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key80))

			child1 := bTree.root.children[1]
			g.Expect(child1.numberOfChildren).To(Equal(2))
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key75))

			gchild1_1 := child1.children[0]
			g.Expect(gchild1_1.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_1.numberOfValues).To(Equal(2))
			g.Expect(gchild1_1.keyValues[0].key).To(Equal(Key60))
			g.Expect(gchild1_1.keyValues[1].key).To(Equal(Key70))

			gchild1_2 := child1.children[1]
			g.Expect(gchild1_2.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_2.numberOfValues).To(Equal(1))
			g.Expect(gchild1_2.keyValues[0].key).To(Equal(Key78))

			// validate the rest of tree
			validateThreadSafeTree(g, bTree.root)
		})
	})

	t.Run("when internal nodes are below minimum required values", func(t *testing.T) {
		t.Run("it can rotate a right value left with children", func(t *testing.T) {
			bTree := setupTree(g)
			// setup tree
			//              40,80
			//       /        |         \
			//     35        60,75      100,120
			//   /   \     /  |  \    /   |   \
			//  30   38   50  70 78  90  110  130
			bTree.Delete(Key10, nil)
			bTree.Delete(Key20, nil)
			validateThreadSafeTree(g, bTree.root)

			// deleting anything on left tree (in this case 35) to merge down to the left and look like so:
			//            40,80
			//       /      |         \
			//[  empty]   60,75      100,120
			//   /       /  |  \    /   |   \
			//  30,38  38   50  70 78  90  110  130
			//
			//
			// with a final rotation seting up the final tree:
			//              60,80
			//       /        |         \
			//     40        75       100,120
			//    /   \     /  \    /   |   \
			// 30,38   50  70  78  90  110  130
			bTree.Delete(Key35, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.numberOfChildren).To(Equal(3))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key60))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key80))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfChildren).To(Equal(2))
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key40))

			gchild1_1 := child1.children[0]
			g.Expect(gchild1_1.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_1.numberOfValues).To(Equal(2))
			g.Expect(gchild1_1.keyValues[0].key).To(Equal(Key30))
			g.Expect(gchild1_1.keyValues[1].key).To(Equal(Key38))

			gchild1_2 := child1.children[1]
			g.Expect(gchild1_2.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_2.numberOfValues).To(Equal(1))
			g.Expect(gchild1_2.keyValues[0].key).To(Equal(Key50))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfChildren).To(Equal(2))
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key75))

			gchild2_1 := child2.children[0]
			g.Expect(gchild2_1.numberOfChildren).To(Equal(0))
			g.Expect(gchild2_1.numberOfValues).To(Equal(1))
			g.Expect(gchild2_1.keyValues[0].key).To(Equal(Key70))

			gchild2_2 := child2.children[1]
			g.Expect(gchild2_2.numberOfChildren).To(Equal(0))
			g.Expect(gchild2_2.numberOfValues).To(Equal(1))
			g.Expect(gchild2_2.keyValues[0].key).To(Equal(Key78))

			// validate the rest of tree
			validateThreadSafeTree(g, bTree.root)
		})

		t.Run("it can rotate a left value right with children", func(t *testing.T) {
			bTree := setupTree(g)
			// setup tree
			//               40,80
			//        /        |      \
			//    20,35      60,75     100
			//   /  |  \    /  |  \   /  \
			//  10 30 38  50  70 78  90   110
			bTree.Delete(Key120, nil)
			bTree.Delete(Key130, nil)
			validateThreadSafeTree(g, bTree.root)

			// deleting anything on right tree (in this case 110) to merge down to the left and look like so:
			//               40,80
			//        /        |         \
			//    20,35      60,75      [empty]
			//   /  |  \    /  |  \     /
			//  10 30 38  50  70 78  90,100
			//
			//
			// with a final rotation seting up the final tree:
			//              40,75
			//        /       |       \
			//    20,35      60 			80
			//   /  |  \    /  \    /   \
			//  10 30 38   50  70  78  90,100
			bTree.Delete(Key110, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.numberOfChildren).To(Equal(3))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key40))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key75))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfChildren).To(Equal(2))
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key60))

			gchild2_1 := child2.children[0]
			g.Expect(gchild2_1.numberOfChildren).To(Equal(0))
			g.Expect(gchild2_1.numberOfValues).To(Equal(1))
			g.Expect(gchild2_1.keyValues[0].key).To(Equal(Key50))

			gchild2_2 := child2.children[1]
			g.Expect(gchild2_2.numberOfChildren).To(Equal(0))
			g.Expect(gchild2_2.numberOfValues).To(Equal(1))
			g.Expect(gchild2_2.keyValues[0].key).To(Equal(Key70))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfChildren).To(Equal(2))
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key80))

			gchild3_1 := child3.children[0]
			g.Expect(gchild3_1.numberOfChildren).To(Equal(0))
			g.Expect(gchild3_1.numberOfValues).To(Equal(1))
			g.Expect(gchild3_1.keyValues[0].key).To(Equal(Key78))

			gchild3_2 := child3.children[1]
			g.Expect(gchild3_2.numberOfChildren).To(Equal(0))
			g.Expect(gchild3_2.numberOfValues).To(Equal(2))
			g.Expect(gchild3_2.keyValues[0].key).To(Equal(Key90))
			g.Expect(gchild3_2.keyValues[1].key).To(Equal(Key100))

			// validate the rest of tree
			validateThreadSafeTree(g, bTree.root)
		})

		t.Run("it can merge a parent value down when the left is under the required minimum values", func(t *testing.T) {
			bTree := setupTree(g)
			// setup tree
			//               40,80
			//       /          |         \
			//      35         75      100,120
			//   /     \     /    \    /   |   \
			//  30     38   70    78  90  110  130
			bTree.Delete(Key10, nil)
			bTree.Delete(Key20, nil)
			bTree.Delete(Key50, nil)
			bTree.Delete(Key60, nil)
			validateThreadSafeTree(g, bTree.root)

			// deleting anything on left or middle tree (in this case 30) to merge down to the left and look like so:
			//               40,80
			//       /          |         \
			//    [empty]      75      100,120
			//   /           /    \    /   |   \
			//  35,38      70    78  90  110  130
			//
			//
			// with a final merge seting up the final tree:
			//              80
			//         /          \
			//       40,75      100,120
			//     /   |  \    /   |   \
			//  35,38 70  78  90  110  130
			bTree.Delete(Key30, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key80))

			// left child
			child1 := bTree.root.children[00]
			g.Expect(child1.numberOfChildren).To(Equal(3))
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.keyValues[0].key).To(Equal(Key40))
			g.Expect(child1.keyValues[1].key).To(Equal(Key75))

			gchild1_1 := child1.children[0]
			g.Expect(gchild1_1.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_1.numberOfValues).To(Equal(2))
			g.Expect(gchild1_1.keyValues[0].key).To(Equal(Key35))
			g.Expect(gchild1_1.keyValues[1].key).To(Equal(Key38))

			gchild1_2 := child1.children[1]
			g.Expect(gchild1_2.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_2.numberOfValues).To(Equal(1))
			g.Expect(gchild1_2.keyValues[0].key).To(Equal(Key70))

			gchild1_3 := child1.children[2]
			g.Expect(gchild1_3.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_3.numberOfValues).To(Equal(1))
			g.Expect(gchild1_3.keyValues[0].key).To(Equal(Key78))

			// right child
			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfChildren).To(Equal(3))
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.keyValues[0].key).To(Equal(Key100))
			g.Expect(child2.keyValues[1].key).To(Equal(Key120))

			gchild2_1 := child2.children[0]
			g.Expect(gchild2_1.numberOfChildren).To(Equal(0))
			g.Expect(gchild2_1.numberOfValues).To(Equal(1))
			g.Expect(gchild2_1.keyValues[0].key).To(Equal(Key90))

			gchild2_2 := child2.children[1]
			g.Expect(gchild2_2.numberOfChildren).To(Equal(0))
			g.Expect(gchild2_2.numberOfValues).To(Equal(1))
			g.Expect(gchild2_2.keyValues[0].key).To(Equal(Key110))

			gchild2_3 := child2.children[2]
			g.Expect(gchild2_3.numberOfChildren).To(Equal(0))
			g.Expect(gchild2_3.numberOfValues).To(Equal(1))
			g.Expect(gchild2_3.keyValues[0].key).To(Equal(Key130))
		})

		t.Run("it can merge a parent value down when the right child is under the required minimum values", func(t *testing.T) {
			bTree := setupTree(g)
			// setup tree
			//                40,80
			//       /          |       \
			//    20,35        60      100
			//   /  |   \     /  \    /   \
			//  10, 30 ,38   50  70  90  110
			bTree.Delete(Key78, nil)
			bTree.Delete(Key75, nil)
			bTree.Delete(Key130, nil)
			bTree.Delete(Key120, nil)
			validateThreadSafeTree(g, bTree.root)

			// deleting anything on left or middle tree (in this case 110) to merge down to the left and look like so:
			//                40,80
			//       /          |       \
			//    20,35        60      [empty]
			//   /  |   \     /  \    /
			//  10, 30 ,38   50  70  90,100
			//
			//
			// with a final merge seting up the final tree:
			//            40
			//       /            \
			//    20,35        60, 80
			//   /  |   \     /   |   \
			//  10, 30 ,38   50  70  90,100
			bTree.Delete(Key110, nil)
			g.Expect(bTree.root.numberOfValues).To(Equal(1))
			g.Expect(bTree.root.numberOfChildren).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key40))

			// left child
			child1 := bTree.root.children[00]
			g.Expect(child1.numberOfChildren).To(Equal(3))
			g.Expect(child1.numberOfValues).To(Equal(2))
			g.Expect(child1.keyValues[0].key).To(Equal(Key20))
			g.Expect(child1.keyValues[1].key).To(Equal(Key35))

			gchild1_1 := child1.children[0]
			g.Expect(gchild1_1.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_1.numberOfValues).To(Equal(1))
			g.Expect(gchild1_1.keyValues[0].key).To(Equal(Key10))

			gchild1_2 := child1.children[1]
			g.Expect(gchild1_2.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_2.numberOfValues).To(Equal(1))
			g.Expect(gchild1_2.keyValues[0].key).To(Equal(Key30))

			gchild1_3 := child1.children[2]
			g.Expect(gchild1_3.numberOfChildren).To(Equal(0))
			g.Expect(gchild1_3.numberOfValues).To(Equal(1))
			g.Expect(gchild1_3.keyValues[0].key).To(Equal(Key38))

			// right child
			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfChildren).To(Equal(3))
			g.Expect(child2.numberOfValues).To(Equal(2))
			g.Expect(child2.keyValues[0].key).To(Equal(Key60))
			g.Expect(child2.keyValues[1].key).To(Equal(Key80))

			gchild2_1 := child2.children[0]
			g.Expect(gchild2_1.numberOfChildren).To(Equal(0))
			g.Expect(gchild2_1.numberOfValues).To(Equal(1))
			g.Expect(gchild2_1.keyValues[0].key).To(Equal(Key50))

			gchild2_2 := child2.children[1]
			g.Expect(gchild2_2.numberOfChildren).To(Equal(0))
			g.Expect(gchild2_2.numberOfValues).To(Equal(1))
			g.Expect(gchild2_2.keyValues[0].key).To(Equal(Key70))

			gchild2_3 := child2.children[2]
			g.Expect(gchild2_3.numberOfChildren).To(Equal(0))
			g.Expect(gchild2_3.numberOfValues).To(Equal(2))
			g.Expect(gchild2_3.keyValues[0].key).To(Equal(Key90))
			g.Expect(gchild2_3.keyValues[1].key).To(Equal(Key100))
		})
	})
}
