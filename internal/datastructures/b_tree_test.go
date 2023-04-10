package datastructures

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

type bTreeTester struct {
	onFindCount int
	value       string
}

func newBTreeTester(value string) func() (any, error) {
	return func() (any, error) {
		return &bTreeTester{
			onFindCount: 0,
			value:       value,
		}, nil
	}
}
func newBTreeTesterWithError() (any, error) {
	return nil, fmt.Errorf("failure")
}

func (btt *bTreeTester) OnFind() {
	btt.onFindCount++
}

func validateTree(g *GomegaWithT, bNode *bNode, parentKey TreeKey, less bool) {
	if bNode == nil {
		return
	}

	var index int
	for index = 0; index < bNode.numberOfValues-1; index++ {
		// check parent key
		if parentKey != nil {
			if less {
				g.Expect(bNode.values[index].key.Less(parentKey)).To(BeTrue())
			} else {
				g.Expect(bNode.values[index].key.Less(parentKey)).To(BeFalse())
			}
		}

		// check current vllue is less than the next index
		g.Expect(bNode.values[index].key.Less(bNode.values[index+1].key)).To(BeTrue())

		if bNode.numberOfChildren != 0 {
			validateTree(g, bNode.children[index], bNode.values[index].key, true)
		}
	}

	// if there are any children, we need to check the last 2 indexes
	if bNode.numberOfChildren != 0 {
		validateTree(g, bNode.children[index], bNode.values[index].key, true)
		validateTree(g, bNode.children[index+1], bNode.values[index].key, false)
	}
}

var (
	keyOne   = IntTreeKey(1)
	keyTwo   = IntTreeKey(2)
	keyThree = IntTreeKey(3)
	keyFour  = IntTreeKey(4)
	keyFive  = IntTreeKey(5)
)

func TestBTree_NewBTree(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns an error if the nodeSize is to small", func(t *testing.T) {
		bTree, err := NewBTree(1)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("nodeSize must be greater than 1 for BTree"))
		g.Expect(bTree).To(BeNil())
	})

	t.Run("returns an error if the nodeSize is to large", func(t *testing.T) {
		bTree, err := NewBTree(math.MaxInt)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(fmt.Sprintf("nodeSize must be 2 less than %d", math.MaxInt)))
		g.Expect(bTree).To(BeNil())
	})
}

func TestBTree_FindOrCreate_SingleNode(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it does not save the value when onCreate returns an error", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem1, err := bTree.FindOrCreate(keyOne, "OnFind", newBTreeTesterWithError)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("failure"))
		g.Expect(treeItem1).To(BeNil())

		g.Expect(bTree.root.numberOfValues).To(Equal(0))
	})

	t.Run("creates a new tree with proper size limits", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem1, err := bTree.FindOrCreate(keyOne, "OnFind", newBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem2, err := bTree.FindOrCreate(keyTwo, "OnFind", newBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(keyOne))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem1))
		g.Expect(bTree.root.values[0].item.(*bTreeTester).value).To(Equal("1"))
		g.Expect(bTree.root.values[0].item.(*bTreeTester).onFindCount).To(Equal(0))
		g.Expect(bTree.root.values[1].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[1].item).To(Equal(treeItem2))
		g.Expect(bTree.root.values[1].item.(*bTreeTester).value).To(Equal("2"))
		g.Expect(bTree.root.values[1].item.(*bTreeTester).onFindCount).To(Equal(0))
	})

	t.Run("adding the same item multiple times returns the original inserted item", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem1, err := bTree.FindOrCreate(keyOne, "OnFind", newBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem2, err := bTree.FindOrCreate(keyOne, "OnFind", newBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(treeItem1).To(Equal(treeItem2))
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].item.(*bTreeTester).value).To(Equal("1"))
		g.Expect(bTree.root.values[0].item.(*bTreeTester).onFindCount).To(Equal(1))
	})

	t.Run("inserts the items in the proper nodeSize, no matter how they were added to the tree", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem1, err := bTree.FindOrCreate(keyOne, "OnFind", newBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem2, err := bTree.FindOrCreate(keyTwo, "OnFind", newBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(keyOne))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem1))
		g.Expect(bTree.root.values[0].item.(*bTreeTester).value).To(Equal("1"))
		g.Expect(bTree.root.values[0].item.(*bTreeTester).onFindCount).To(Equal(0))
		g.Expect(bTree.root.values[1].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[1].item).To(Equal(treeItem2))
		g.Expect(bTree.root.values[1].item.(*bTreeTester).value).To(Equal("2"))
		g.Expect(bTree.root.values[1].item.(*bTreeTester).onFindCount).To(Equal(0))
	})

	t.Run("possible split returns an item that already exists", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		_, _ = bTree.FindOrCreate(keyTwo, "OnFind", newBTreeTester("2"))
		_, _ = bTree.FindOrCreate(keyOne, "OnFind", newBTreeTester("1"))
		_, _ = bTree.FindOrCreate(keyOne, "OnFind", newBTreeTester("3"))

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(keyOne))
		g.Expect(bTree.root.values[0].item.(*bTreeTester).value).To(Equal("1"))
		g.Expect(bTree.root.values[0].item.(*bTreeTester).onFindCount).To(Equal(1))
		g.Expect(bTree.root.values[1].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[1].item.(*bTreeTester).value).To(Equal("2"))
		g.Expect(bTree.root.values[1].item.(*bTreeTester).onFindCount).To(Equal(0))
	})

	t.Run("it splits the node when adding a left pivot value", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem2, err := bTree.FindOrCreate(keyTwo, "OnFind", newBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem3, err := bTree.FindOrCreate(keyThree, "OnFind", newBTreeTester("3"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem1, err := bTree.FindOrCreate(keyOne, "OnFind", newBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem2))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(keyOne))
		g.Expect(leftchild.values[0].item).To(Equal(treeItem1))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(keyThree))
		g.Expect(rightchild.values[0].item).To(Equal(treeItem3))
	})

	t.Run("it splits the node when adding a pivot item value", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem3, err := bTree.FindOrCreate(keyThree, "OnFind", newBTreeTester("3"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem2, err := bTree.FindOrCreate(keyTwo, "OnFind", newBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem1, err := bTree.FindOrCreate(keyOne, "OnFind", newBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem2))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(keyOne))
		g.Expect(leftchild.values[0].item).To(Equal(treeItem1))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(keyThree))
		g.Expect(rightchild.values[0].item).To(Equal(treeItem3))
	})

	t.Run("it splits the node when adding a right pivot item value", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem2, err := bTree.FindOrCreate(keyTwo, "OnFind", newBTreeTester("2"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem1, err := bTree.FindOrCreate(keyOne, "OnFind", newBTreeTester("1"))
		g.Expect(err).ToNot(HaveOccurred())
		treeItem3, err := bTree.FindOrCreate(keyThree, "OnFind", newBTreeTester("3"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem2))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(keyOne))
		g.Expect(leftchild.values[0].item).To(Equal(treeItem1))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(keyThree))
		g.Expect(rightchild.values[0].item).To(Equal(treeItem3))
	})
}

func TestBTree_FindOrCreate_Tree_SimpleOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	key10 := IntTreeKey(10)
	key20 := IntTreeKey(20)
	key30 := IntTreeKey(30)

	// generate a tree of
	/*
	 *     20
	 *   /    \
	 *  10    30
	 */
	setupTree := func(g *GomegaWithT) *BRoot {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		_, _ = bTree.FindOrCreate(key10, "OnFind", newBTreeTester("10"))
		_, _ = bTree.FindOrCreate(key20, "OnFind", newBTreeTester("20"))
		_, _ = bTree.FindOrCreate(key30, "OnFind", newBTreeTester("30"))

		return bTree
	}

	t.Run("can return a entry on the leftChild if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem, err := bTree.FindOrCreate(key10, "OnFind", newBTreeTester("10"))
		g.Expect(err).ToNot(HaveOccurred())

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(1))
		g.Expect(leftChild.values[0].item).To(Equal(treeItem))
	})

	t.Run("can return a entry on the rightChild if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem, err := bTree.FindOrCreate(key30, "OnFind", newBTreeTester("30"))
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
		key5 := IntTreeKey(5)

		_, _ = bTree.FindOrCreate(key5, "OnFind", newBTreeTester("5"))

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(2))
		g.Expect(leftChild.values[0].key).To(Equal(key5))
		g.Expect(leftChild.values[0].item.(*bTreeTester).value).To(Equal("5"))
		g.Expect(leftChild.values[0].item.(*bTreeTester).onFindCount).To(Equal(0))
		g.Expect(leftChild.values[1].key).To(Equal(key10))
		g.Expect(leftChild.values[1].item.(*bTreeTester).value).To(Equal("10"))
		g.Expect(leftChild.values[1].item.(*bTreeTester).onFindCount).To(Equal(0))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10     25,30
	 */
	t.Run("can add a new entry on the rightChild[0]", func(t *testing.T) {
		bTree := setupTree(g)
		key25 := IntTreeKey(25)

		_, _ = bTree.FindOrCreate(key25, "OnFind", newBTreeTester("25"))

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(2))
		g.Expect(rightChild.values[0].key).To(Equal(key25))
		g.Expect(rightChild.values[0].item.(*bTreeTester).value).To(Equal("25"))
		g.Expect(rightChild.values[0].item.(*bTreeTester).onFindCount).To(Equal(0))
		g.Expect(rightChild.values[1].key).To(Equal(key30))
		g.Expect(rightChild.values[1].item.(*bTreeTester).value).To(Equal("30"))
		g.Expect(rightChild.values[1].item.(*bTreeTester).onFindCount).To(Equal(0))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10,15   30
	 */
	t.Run("can add a new entry on the leftChild[1]", func(t *testing.T) {
		bTree := setupTree(g)
		key15 := IntTreeKey(15)

		_, _ = bTree.FindOrCreate(key15, "OnFind", newBTreeTester("15"))

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(2))
		g.Expect(leftChild.values[0].key).To(Equal(key10))
		g.Expect(leftChild.values[0].item.(*bTreeTester).value).To(Equal("10"))
		g.Expect(leftChild.values[0].item.(*bTreeTester).onFindCount).To(Equal(0))
		g.Expect(leftChild.values[1].key).To(Equal(key15))
		g.Expect(leftChild.values[1].item.(*bTreeTester).value).To(Equal("15"))
		g.Expect(leftChild.values[1].item.(*bTreeTester).onFindCount).To(Equal(0))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10    30,35
	 */
	t.Run("can add a new entry on the rightChild[1]", func(t *testing.T) {
		bTree := setupTree(g)
		key35 := IntTreeKey(35)

		_, _ = bTree.FindOrCreate(key35, "OnFind", newBTreeTester("35"))

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(2))
		g.Expect(rightChild.values[0].key).To(Equal(key30))
		g.Expect(rightChild.values[0].item.(*bTreeTester).value).To(Equal("30"))
		g.Expect(rightChild.values[0].item.(*bTreeTester).onFindCount).To(Equal(0))
		g.Expect(rightChild.values[1].key).To(Equal(key35))
		g.Expect(rightChild.values[1].item.(*bTreeTester).value).To(Equal("35"))
		g.Expect(rightChild.values[1].item.(*bTreeTester).onFindCount).To(Equal(0))
	})
}

func TestBTree_FindOrCreate_Tree_SimplePromoteOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	// setup values
	key10 := IntTreeKey(10)
	key20 := IntTreeKey(20)
	key30 := IntTreeKey(30)

	t.Run("leftChild promotions", func(t *testing.T) {
		key15 := IntTreeKey(15)

		// values to add
		key8 := IntTreeKey(8)
		key12 := IntTreeKey(12)
		key17 := IntTreeKey(17)

		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 * 10,15   30
		 */
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			_, _ = bTree.FindOrCreate(key30, "OnFind", newBTreeTester("30"))
			_, _ = bTree.FindOrCreate(key10, "OnFind", newBTreeTester("10"))
			_, _ = bTree.FindOrCreate(key20, "OnFind", newBTreeTester("20"))
			_, _ = bTree.FindOrCreate(key15, "OnFind", newBTreeTester("15"))

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

			treeItem, err := bTree.FindOrCreate(key8, "OnFind", newBTreeTester("8"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key10))
			g.Expect(bTree.root.values[1].key).To(Equal(key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key8))
			g.Expect(child1.values[0].item).To(Equal(treeItem))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key15))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(key30))
		})

		// generate a tree of
		/*
		 *      12,20
		 *    /   |   \
		 *   10   15  30
		 */
		t.Run("adding the leftChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem, err := bTree.FindOrCreate(key12, "OnFind", newBTreeTester("12"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key12))
			g.Expect(bTree.root.values[0].item).To(Equal(treeItem))
			g.Expect(bTree.root.values[1].key).To(Equal(key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key15))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(key30))
		})

		// generate a tree of
		/*
		 *      15,20
		 *    /   |   \
		 *   10   17  30
		 */
		t.Run("adding the leftChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem, err := bTree.FindOrCreate(key17, "OnFind", newBTreeTester("17"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key15))
			g.Expect(bTree.root.values[1].key).To(Equal(key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key17))
			g.Expect(child2.values[0].item).To(Equal(treeItem))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(key30))
		})
	})

	t.Run("rightChild promotions", func(t *testing.T) {
		key35 := IntTreeKey(35)

		// values to add
		key25 := IntTreeKey(25)
		key32 := IntTreeKey(32)
		key37 := IntTreeKey(37)

		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 *   10   30,35
		 */
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			bTree.FindOrCreate(key20, "OnFind", newBTreeTester("20"))
			bTree.FindOrCreate(key30, "OnFind", newBTreeTester("30"))
			bTree.FindOrCreate(key10, "OnFind", newBTreeTester("10"))
			bTree.FindOrCreate(key35, "OnFind", newBTreeTester("35"))

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

			treeItem, err := bTree.FindOrCreate(key25, "OnFind", newBTreeTester("25"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key20))
			g.Expect(bTree.root.values[1].key).To(Equal(key30))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key25))
			g.Expect(child2.values[0].item).To(Equal(treeItem))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(key35))
		})

		// generate a tree of
		/*
		 *      20,32
		 *    /   |   \
		 *   10   30  35
		 */
		t.Run("adding the rightChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem, err := bTree.FindOrCreate(key32, "OnFind", newBTreeTester("32"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key20))
			g.Expect(bTree.root.values[1].key).To(Equal(key32))
			g.Expect(bTree.root.values[1].item).To(Equal(treeItem))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key30))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(key35))
		})

		// generate a tree of
		/*
		 *      20,35
		 *    /   |   \
		 *   10   30  37
		 */
		t.Run("adding the rightChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem, err := bTree.FindOrCreate(key37, "OnFind", newBTreeTester("37"))
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key20))
			g.Expect(bTree.root.values[1].key).To(Equal(key35))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key30))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(key37))
			g.Expect(child3.values[0].item).To(Equal(treeItem))
		})
	})
}

func TestBTree_FindOrCreate_Tree_NewRootNode(t *testing.T) {
	g := NewGomegaWithT(t)

	// setup values
	key5 := IntTreeKey(5)
	key10 := IntTreeKey(10)
	key20 := IntTreeKey(20)
	key25 := IntTreeKey(25)
	key30 := IntTreeKey(30)
	key40 := IntTreeKey(40)
	key45 := IntTreeKey(45)
	key50 := IntTreeKey(50)

	// all tests in this section start with a base tree like so
	/*
	 *        20,40
	 *    /     |     \
	 *  5,10  25,30  45,50
	 */
	setupTree := func(g *GomegaWithT) *BRoot {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		bTree.FindOrCreate(key10, "OnFind", newBTreeTester("10"))
		bTree.FindOrCreate(key20, "OnFind", newBTreeTester("20"))
		bTree.FindOrCreate(key30, "OnFind", newBTreeTester("30"))
		bTree.FindOrCreate(key40, "OnFind", newBTreeTester("40"))
		bTree.FindOrCreate(key50, "OnFind", newBTreeTester("50"))
		bTree.FindOrCreate(key5, "OnFind", newBTreeTester("5"))
		bTree.FindOrCreate(key25, "OnFind", newBTreeTester("25"))
		bTree.FindOrCreate(key45, "OnFind", newBTreeTester("45"))

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(key20))
		g.Expect(bTree.root.values[1].key).To(Equal(key40))

		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(2))
		g.Expect(child1.values[0].key).To(Equal(key5))
		g.Expect(child1.values[1].key).To(Equal(key10))

		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(2))
		g.Expect(child2.values[0].key).To(Equal(key25))
		g.Expect(child2.values[1].key).To(Equal(key30))

		child3 := bTree.root.children[2]
		g.Expect(child3.numberOfValues).To(Equal(2))
		g.Expect(child3.values[0].key).To(Equal(key45))
		g.Expect(child3.values[1].key).To(Equal(key50))

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
		key0 := IntTreeKey(0)

		treeItem, err := bTree.FindOrCreate(key0, "OnFind", newBTreeTester("0"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(key20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.values[0].key).To(Equal(key5))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.values[0].key).To(Equal(key0))
		g.Expect(gchild1.values[0].item).To(Equal(treeItem))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.values[0].key).To(Equal(key10))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.values[0].key).To(Equal(key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.values[0].key).To(Equal(key25))
		g.Expect(gchild1.values[1].key).To(Equal(key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.values[0].key).To(Equal(key45))
		g.Expect(gchild2.values[1].key).To(Equal(key50))
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
		key22 := IntTreeKey(22)

		treeItem, err := bTree.FindOrCreate(key22, "OnFind", newBTreeTester("22"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(key25))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.values[0].key).To(Equal(key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.values[0].key).To(Equal(key5))
		g.Expect(gchild1.values[1].key).To(Equal(key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.values[0].key).To(Equal(key22))
		g.Expect(gchild2.values[0].item).To(Equal(treeItem))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.values[0].key).To(Equal(key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.values[0].key).To(Equal(key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.values[0].key).To(Equal(key45))
		g.Expect(gchild2.values[1].key).To(Equal(key50))
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
		key47 := IntTreeKey(47)

		treeItem, err := bTree.FindOrCreate(key47, "OnFind", newBTreeTester("47"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(key40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.values[0].key).To(Equal(key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.values[0].key).To(Equal(key5))
		g.Expect(gchild1.values[1].key).To(Equal(key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.values[0].key).To(Equal(key25))
		g.Expect(gchild2.values[1].key).To(Equal(key30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.values[0].key).To(Equal(key47))
		g.Expect(child2.values[0].item).To(Equal(treeItem))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.values[0].key).To(Equal(key45))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.values[0].key).To(Equal(key50))
	})
}

func TestBTree_RandomAssertion(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("works for a tree nodeSize of 2", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := IntTreeKey(num)
			_, _ = bTree.FindOrCreate(key, "OnFind", newBTreeTester(fmt.Sprintf("%d", num)))
		}

		validateTree(g, bTree.root, nil, true)
	})

	t.Run("works for a tree nodeSize of 3", func(t *testing.T) {
		bTree, err := NewBTree(3)
		g.Expect(err).ToNot(HaveOccurred())

		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := IntTreeKey(num)
			_, _ = bTree.FindOrCreate(key, "OnFind", newBTreeTester(fmt.Sprintf("%d", num)))
		}

		validateTree(g, bTree.root, nil, true)
	})

	t.Run("works for a tree nodeSize of 4", func(t *testing.T) {
		bTree, err := NewBTree(4)
		g.Expect(err).ToNot(HaveOccurred())

		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := IntTreeKey(num)
			_, _ = bTree.FindOrCreate(key, "OnFind", newBTreeTester(fmt.Sprintf("%d", num)))
		}

		validateTree(g, bTree.root, nil, true)
	})
}

func TestBTree_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *BRoot {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			_, _ = bTree.FindOrCreate(IntTreeKey(i), "OnFind", newBTreeTester(fmt.Sprintf("%d", i)))
		}

		return bTree
	}

	t.Run("returns nil if the item does not exist in the tree", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Find(IntTreeKey(1), "OnFind")).To(BeNil())
	})

	t.Run("returns the item in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem := bTree.Find(IntTreeKey(768), "OnFind")
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*bTreeTester).value).To(Equal("768"))
		g.Expect(treeItem.(*bTreeTester).onFindCount).To(Equal(1))
	})
}

func TestBTree_Iterate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it does not run the iterative function if there are no values", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		iterate := func(val any) {
			panic("should not call")
		}

		g.Expect(func() { bTree.Iterate(iterate) }).ToNot(Panic())
	})

	t.Run("it calls the iterative function on each tree item with a value", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			_, _ = bTree.FindOrCreate(IntTreeKey(i), "OnFind", newBTreeTester(fmt.Sprintf("%d", i)))
		}

		seenValues := map[string]struct{}{}
		count := 0
		iterate := func(val any) {
			bTreeTester := val.(*bTreeTester)

			// check that each value is unique
			g.Expect(seenValues).ToNot(HaveKey(bTreeTester.value))
			seenValues[bTreeTester.value] = struct{}{}

			count++
		}

		bTree.Iterate(iterate)
		g.Expect(count).To(Equal(1_024))
	})
}
