package btree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/DanLavine/willow/internal/datastructures/btree/testhelpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestBTree_CreateOrFind_SingleNode(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it does not save the value when onCreate returns an error", func(t *testing.T) {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem1, err := bTree.CreateOrFind(Key1, OnFindTest, NewBTreeTesterWithError)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("failure"))
		g.Expect(treeItem1).To(BeNil())

		g.Expect(bTree.root.numberOfValues).To(Equal(0))
	})

	t.Run("creates a new tree with proper size limits", func(t *testing.T) {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem1, err := bTree.CreateOrFind(Key1, OnFindTest, NewBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem2, err := bTree.CreateOrFind(Key2, OnFindTest, NewBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(Key1))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem1))
		g.Expect(bTree.root.values[0].item.(*BTreeTester).Value).To(Equal("1"))
		g.Expect(bTree.root.values[0].item.(*BTreeTester).OnFindCount).To(Equal(0))
		g.Expect(bTree.root.values[1].key).To(Equal(Key2))
		g.Expect(bTree.root.values[1].item).To(Equal(treeItem2))
		g.Expect(bTree.root.values[1].item.(*BTreeTester).Value).To(Equal("2"))
		g.Expect(bTree.root.values[1].item.(*BTreeTester).OnFindCount).To(Equal(0))
	})

	t.Run("adding the same item multiple times returns the original inserted item", func(t *testing.T) {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem1, err := bTree.CreateOrFind(Key1, OnFindTest, NewBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem2, err := bTree.CreateOrFind(Key1, OnFindTest, NewBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(treeItem1).To(Equal(treeItem2))
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].item.(*BTreeTester).Value).To(Equal("1"))
		g.Expect(bTree.root.values[0].item.(*BTreeTester).OnFindCount).To(Equal(1))
	})

	t.Run("inserts the items in the proper nodeSize, no matter how they were added to the tree", func(t *testing.T) {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem1, err := bTree.CreateOrFind(Key1, OnFindTest, NewBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem2, err := bTree.CreateOrFind(Key2, OnFindTest, NewBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(Key1))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem1))
		g.Expect(bTree.root.values[0].item.(*BTreeTester).Value).To(Equal("1"))
		g.Expect(bTree.root.values[0].item.(*BTreeTester).OnFindCount).To(Equal(0))
		g.Expect(bTree.root.values[1].key).To(Equal(Key2))
		g.Expect(bTree.root.values[1].item).To(Equal(treeItem2))
		g.Expect(bTree.root.values[1].item.(*BTreeTester).Value).To(Equal("2"))
		g.Expect(bTree.root.values[1].item.(*BTreeTester).OnFindCount).To(Equal(0))
	})

	t.Run("possible split returns an item that already exists", func(t *testing.T) {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		_, _ = bTree.CreateOrFind(Key2, OnFindTest, NewBTreeTester("2"))
		_, _ = bTree.CreateOrFind(Key1, OnFindTest, NewBTreeTester("1"))
		_, _ = bTree.CreateOrFind(Key1, OnFindTest, NewBTreeTester("3"))

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(Key1))
		g.Expect(bTree.root.values[0].item.(*BTreeTester).Value).To(Equal("1"))
		g.Expect(bTree.root.values[0].item.(*BTreeTester).OnFindCount).To(Equal(1))
		g.Expect(bTree.root.values[1].key).To(Equal(Key2))
		g.Expect(bTree.root.values[1].item.(*BTreeTester).Value).To(Equal("2"))
		g.Expect(bTree.root.values[1].item.(*BTreeTester).OnFindCount).To(Equal(0))
	})

	t.Run("it splits the node when adding a left pivot value", func(t *testing.T) {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem2, err := bTree.CreateOrFind(Key2, OnFindTest, NewBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem3, err := bTree.CreateOrFind(Key3, OnFindTest, NewBTreeTester("3"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem1, err := bTree.CreateOrFind(Key1, OnFindTest, NewBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(Key2))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem2))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(Key1))
		g.Expect(leftchild.values[0].item).To(Equal(treeItem1))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(Key3))
		g.Expect(rightchild.values[0].item).To(Equal(treeItem3))
	})

	t.Run("it splits the node when adding a pivot item value", func(t *testing.T) {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem3, err := bTree.CreateOrFind(Key3, OnFindTest, NewBTreeTester("3"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem2, err := bTree.CreateOrFind(Key2, OnFindTest, NewBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem1, err := bTree.CreateOrFind(Key1, OnFindTest, NewBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(Key2))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem2))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(Key1))
		g.Expect(leftchild.values[0].item).To(Equal(treeItem1))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(Key3))
		g.Expect(rightchild.values[0].item).To(Equal(treeItem3))
	})

	t.Run("it splits the node when adding a right pivot item value", func(t *testing.T) {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem2, err := bTree.CreateOrFind(Key2, OnFindTest, NewBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem1, err := bTree.CreateOrFind(Key1, OnFindTest, NewBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem3, err := bTree.CreateOrFind(Key3, OnFindTest, NewBTreeTester("3"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(Key2))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem2))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(Key1))
		g.Expect(leftchild.values[0].item).To(Equal(treeItem1))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(Key3))
		g.Expect(rightchild.values[0].item).To(Equal(treeItem3))
	})
}

func TestBTree_CreateOrFind_Tree_SimpleOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	// generate a tree of
	/*
	 *     20
	 *   /    \
	 *  10    30
	 */
	setupTree := func(g *GomegaWithT) *bTree {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		_, _ = bTree.CreateOrFind(Key10, OnFindTest, NewBTreeTester("10"))
		_, _ = bTree.CreateOrFind(Key20, OnFindTest, NewBTreeTester("20"))
		_, _ = bTree.CreateOrFind(Key30, OnFindTest, NewBTreeTester("30"))

		return bTree
	}

	t.Run("can return a entry on the leftChild if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem, err := bTree.CreateOrFind(Key10, OnFindTest, NewBTreeTester("10"))
		g.Expect(err).ToNot(HaveOccurred())

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(1))
		g.Expect(leftChild.values[0].item).To(Equal(treeItem))
	})

	t.Run("can return a entry on the rightChild if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem, err := bTree.CreateOrFind(Key30, OnFindTest, NewBTreeTester("30"))
		g.Expect(err).ToNot(HaveOccurred())

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(1))
		g.Expect(rightChild.values[0].item).To(Equal(treeItem))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  5,10   30
	 */
	t.Run("can add a new entry on the leftChild[0]", func(t *testing.T) {
		bTree := setupTree(g)
		_, _ = bTree.CreateOrFind(Key5, OnFindTest, NewBTreeTester("5"))

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(2))
		g.Expect(leftChild.values[0].key).To(Equal(Key5))
		g.Expect(leftChild.values[0].item.(*BTreeTester).Value).To(Equal("5"))
		g.Expect(leftChild.values[0].item.(*BTreeTester).OnFindCount).To(Equal(0))
		g.Expect(leftChild.values[1].key).To(Equal(Key10))
		g.Expect(leftChild.values[1].item.(*BTreeTester).Value).To(Equal("10"))
		g.Expect(leftChild.values[1].item.(*BTreeTester).OnFindCount).To(Equal(0))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10     25,30
	 */
	t.Run("can add a new entry on the rightChild[0]", func(t *testing.T) {
		bTree := setupTree(g)
		_, _ = bTree.CreateOrFind(Key25, OnFindTest, NewBTreeTester("25"))

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(2))
		g.Expect(rightChild.values[0].key).To(Equal(Key25))
		g.Expect(rightChild.values[0].item.(*BTreeTester).Value).To(Equal("25"))
		g.Expect(rightChild.values[0].item.(*BTreeTester).OnFindCount).To(Equal(0))
		g.Expect(rightChild.values[1].key).To(Equal(Key30))
		g.Expect(rightChild.values[1].item.(*BTreeTester).Value).To(Equal("30"))
		g.Expect(rightChild.values[1].item.(*BTreeTester).OnFindCount).To(Equal(0))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10,15   30
	 */
	t.Run("can add a new entry on the leftChild[1]", func(t *testing.T) {
		bTree := setupTree(g)
		_, _ = bTree.CreateOrFind(Key15, OnFindTest, NewBTreeTester("15"))

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(2))
		g.Expect(leftChild.values[0].key).To(Equal(Key10))
		g.Expect(leftChild.values[0].item.(*BTreeTester).Value).To(Equal("10"))
		g.Expect(leftChild.values[0].item.(*BTreeTester).OnFindCount).To(Equal(0))
		g.Expect(leftChild.values[1].key).To(Equal(Key15))
		g.Expect(leftChild.values[1].item.(*BTreeTester).Value).To(Equal("15"))
		g.Expect(leftChild.values[1].item.(*BTreeTester).OnFindCount).To(Equal(0))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10    30,35
	 */
	t.Run("can add a new entry on the rightChild[1]", func(t *testing.T) {
		bTree := setupTree(g)
		_, _ = bTree.CreateOrFind(Key35, OnFindTest, NewBTreeTester("35"))

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(2))
		g.Expect(rightChild.values[0].key).To(Equal(Key30))
		g.Expect(rightChild.values[0].item.(*BTreeTester).Value).To(Equal("30"))
		g.Expect(rightChild.values[0].item.(*BTreeTester).OnFindCount).To(Equal(0))
		g.Expect(rightChild.values[1].key).To(Equal(Key35))
		g.Expect(rightChild.values[1].item.(*BTreeTester).Value).To(Equal("35"))
		g.Expect(rightChild.values[1].item.(*BTreeTester).OnFindCount).To(Equal(0))
	})
}

func TestBTree_CreateOrFind_Tree_SimplePromoteOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("leftChild promotions", func(t *testing.T) {
		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 * 10,15   30
		 */
		setupTree := func(g *GomegaWithT) *bTree {
			bTree, err := New(2)
			g.Expect(err).ToNot(HaveOccurred())

			_, _ = bTree.CreateOrFind(Key30, OnFindTest, NewBTreeTester("30"))
			_, _ = bTree.CreateOrFind(Key10, OnFindTest, NewBTreeTester("10"))
			_, _ = bTree.CreateOrFind(Key20, OnFindTest, NewBTreeTester("20"))
			_, _ = bTree.CreateOrFind(Key15, OnFindTest, NewBTreeTester("15"))

			return bTree
		}

		// generate a tree of
		/*
		 *      10,20
		 *    /   |   \
		 *   8    15  30
		 */
		t.Run("adding the leftChild[0] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem, err := bTree.CreateOrFind(Key8, OnFindTest, NewBTreeTester("8"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(Key10))
			g.Expect(bTree.root.values[1].key).To(Equal(Key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(Key8))
			g.Expect(child1.values[0].item).To(Equal(treeItem))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(Key15))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(Key30))
		})

		// generate a tree of
		/*
		 *      12,20
		 *    /   |   \
		 *   10   15  30
		 */
		t.Run("adding the leftChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem, err := bTree.CreateOrFind(Key12, OnFindTest, NewBTreeTester("12"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(Key12))
			g.Expect(bTree.root.values[0].item).To(Equal(treeItem))
			g.Expect(bTree.root.values[1].key).To(Equal(Key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(Key15))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(Key30))
		})

		// generate a tree of
		/*
		 *      15,20
		 *    /   |   \
		 *   10   17  30
		 */
		t.Run("adding the leftChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem, err := bTree.CreateOrFind(Key17, OnFindTest, NewBTreeTester("17"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(Key15))
			g.Expect(bTree.root.values[1].key).To(Equal(Key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(Key17))
			g.Expect(child2.values[0].item).To(Equal(treeItem))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(Key30))
		})
	})

	t.Run("rightChild promotions", func(t *testing.T) {
		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 *   10   30,35
		 */
		setupTree := func(g *GomegaWithT) *bTree {
			bTree, err := New(2)
			g.Expect(err).ToNot(HaveOccurred())

			bTree.CreateOrFind(Key20, OnFindTest, NewBTreeTester("20"))
			bTree.CreateOrFind(Key30, OnFindTest, NewBTreeTester("30"))
			bTree.CreateOrFind(Key10, OnFindTest, NewBTreeTester("10"))
			bTree.CreateOrFind(Key35, OnFindTest, NewBTreeTester("35"))

			return bTree
		}

		// generate a tree of
		/*
		 *      20,30
		 *    /   |  \
		 *   10   25  35
		 */
		t.Run("adding the rightChild_0 node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem, err := bTree.CreateOrFind(Key25, OnFindTest, NewBTreeTester("25"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(Key20))
			g.Expect(bTree.root.values[1].key).To(Equal(Key30))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(Key25))
			g.Expect(child2.values[0].item).To(Equal(treeItem))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(Key35))
		})

		// generate a tree of
		/*
		 *      20,32
		 *    /   |   \
		 *   10   30  35
		 */
		t.Run("adding the rightChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem, err := bTree.CreateOrFind(Key32, OnFindTest, NewBTreeTester("32"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(Key20))
			g.Expect(bTree.root.values[1].key).To(Equal(Key32))
			g.Expect(bTree.root.values[1].item).To(Equal(treeItem))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(Key30))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(Key35))
		})

		// generate a tree of
		/*
		 *      20,35
		 *    /   |   \
		 *   10   30  37
		 */
		t.Run("adding the rightChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem, err := bTree.CreateOrFind(Key37, OnFindTest, NewBTreeTester("37"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(Key20))
			g.Expect(bTree.root.values[1].key).To(Equal(Key35))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(Key30))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(Key37))
			g.Expect(child3.values[0].item).To(Equal(treeItem))
		})
	})
}

func TestBTree_CreateOrFind_Tree_NewRootNode(t *testing.T) {
	g := NewGomegaWithT(t)

	// all tests in this section start with a base tree like so
	/*
	 *        20,40
	 *    /     |     \
	 *  5,10  25,30  45,50
	 */
	setupTree := func(g *GomegaWithT) *bTree {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		bTree.CreateOrFind(Key10, OnFindTest, NewBTreeTester("10"))
		bTree.CreateOrFind(Key20, OnFindTest, NewBTreeTester("20"))
		bTree.CreateOrFind(Key30, OnFindTest, NewBTreeTester("30"))
		bTree.CreateOrFind(Key40, OnFindTest, NewBTreeTester("40"))
		bTree.CreateOrFind(Key50, OnFindTest, NewBTreeTester("50"))
		bTree.CreateOrFind(Key5, OnFindTest, NewBTreeTester("5"))
		bTree.CreateOrFind(Key25, OnFindTest, NewBTreeTester("25"))
		bTree.CreateOrFind(Key45, OnFindTest, NewBTreeTester("45"))

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(Key20))
		g.Expect(bTree.root.values[1].key).To(Equal(Key40))

		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(2))
		g.Expect(child1.values[0].key).To(Equal(Key5))
		g.Expect(child1.values[1].key).To(Equal(Key10))

		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(2))
		g.Expect(child2.values[0].key).To(Equal(Key25))
		g.Expect(child2.values[1].key).To(Equal(Key30))

		child3 := bTree.root.children[2]
		g.Expect(child3.numberOfValues).To(Equal(2))
		g.Expect(child3.values[0].key).To(Equal(Key45))
		g.Expect(child3.values[1].key).To(Equal(Key50))

		return bTree
	}

	// generate a tree of
	/*
	 *             20
	 *         /         \
	 *        5          40
	 *      /  \      /      \
	 *     0   10   25,30  45,50
	 */
	t.Run("adding the leftChild promotes properly", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem, err := bTree.CreateOrFind(Key0, OnFindTest, NewBTreeTester("0"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(Key20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.values[0].key).To(Equal(Key5))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.values[0].key).To(Equal(Key0))
		g.Expect(gchild1.values[0].item).To(Equal(treeItem))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.values[0].key).To(Equal(Key10))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.values[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.values[0].key).To(Equal(Key25))
		g.Expect(gchild1.values[1].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.values[0].key).To(Equal(Key45))
		g.Expect(gchild2.values[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             25
	 *         /        \
	 *        20        40
	 *      /  \      /    \
	 *    5,10  22   30   45,50
	 */
	t.Run("adding the middleChild promotes properly", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem, err := bTree.CreateOrFind(Key22, OnFindTest, NewBTreeTester("22"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(Key25))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.values[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.values[0].key).To(Equal(Key5))
		g.Expect(gchild1.values[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.values[0].key).To(Equal(Key22))
		g.Expect(gchild2.values[0].item).To(Equal(treeItem))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.values[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.values[0].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.values[0].key).To(Equal(Key45))
		g.Expect(gchild2.values[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *              40
	 *         /         \
	 *        20         47
	 *      /    \      /   \
	 *    5,10  25,30  45   50
	 */
	t.Run("adding the right promotes properly", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem, err := bTree.CreateOrFind(Key47, OnFindTest, NewBTreeTester("47"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(Key40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.values[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.values[0].key).To(Equal(Key5))
		g.Expect(gchild1.values[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.values[0].key).To(Equal(Key25))
		g.Expect(gchild2.values[1].key).To(Equal(Key30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.values[0].key).To(Equal(Key47))
		g.Expect(child2.values[0].item).To(Equal(treeItem))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.values[0].key).To(Equal(Key45))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.values[0].key).To(Equal(Key50))
	})
}

func TestBTree_RandomAssertion(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("works for a tree nodeSize of 2", func(t *testing.T) {
		bTree, err := New(2)
		g.Expect(err).ToNot(HaveOccurred())

		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := datatypes.Int(num)
			_, _ = bTree.CreateOrFind(key, OnFindTest, NewBTreeTester(fmt.Sprintf("%d", num)))
		}

		validateTree(g, bTree.root, nil, true)
	})

	t.Run("works for a tree nodeSize of 3", func(t *testing.T) {
		bTree, err := New(3)
		g.Expect(err).ToNot(HaveOccurred())

		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := datatypes.Int(num)
			_, _ = bTree.CreateOrFind(key, OnFindTest, NewBTreeTester(fmt.Sprintf("%d", num)))
		}

		validateTree(g, bTree.root, nil, true)
	})

	t.Run("works for a tree nodeSize of 4", func(t *testing.T) {
		bTree, err := New(4)
		g.Expect(err).ToNot(HaveOccurred())

		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := datatypes.Int(num)
			_, _ = bTree.CreateOrFind(key, OnFindTest, NewBTreeTester(fmt.Sprintf("%d", num)))
		}

		validateTree(g, bTree.root, nil, true)
	})
}
