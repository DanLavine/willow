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

func newBTreeTester(value string) func() TreeItem {
	return func() TreeItem {
		return &bTreeTester{
			onFindCount: 0,
			value:       value,
		}
	}
}

func (btt *bTreeTester) OnFind() {
	btt.onFindCount++
}

func validateTree(g *GomegaWithT, bNode *bNode, parentKey TreeKey, less bool) {
	if bNode == nil {
		return
	}

	var index int
	for index = 0; index < len(bNode.values)-1; index++ {
		if parentKey != nil {
			if less {
				g.Expect(bNode.values[index].key.Less(parentKey)).To(BeTrue())
			} else {
				g.Expect(bNode.values[index].key.Less(parentKey)).To(BeFalse())
			}
		}

		g.Expect(bNode.values[index].key.Less(bNode.values[index+1].key)).To(BeTrue())

		if len(bNode.children) != 0 {
			validateTree(g, bNode.children[index], bNode.values[index].key, true)
		}
	}

	if len(bNode.children) != 0 {
		validateTree(g, bNode.children[index], bNode.values[index].key, true)
		validateTree(g, bNode.children[index+1], bNode.values[index].key, false)
	}
}

var (
	keyOne   = NewIntTreeKey(1)
	keyTwo   = NewIntTreeKey(2)
	keyThree = NewIntTreeKey(3)
)

func TestBTree_NewBTree(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns an error if the order is to small", func(t *testing.T) {
		bTree, err := NewBTree(1)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("order must be greater than 1 for BTree"))
		g.Expect(bTree).To(BeNil())
	})

	t.Run("returns an error if the order is to large", func(t *testing.T) {
		bTree, err := NewBTree(math.MaxInt)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal(fmt.Sprintf("order must be 2 less than %d", math.MaxInt)))
		g.Expect(bTree).To(BeNil())
	})
}

func TestBTree_FindOrCreate_SingleNode(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("creates a new tree with proper size limits", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem1 := bTree.FindOrCreate(keyOne, newBTreeTester("1"))
		treeItem2 := bTree.FindOrCreate(keyTwo, newBTreeTester("2"))

		g.Expect(len(bTree.root.values)).To(Equal(2))
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

		treeItem1 := bTree.FindOrCreate(keyOne, newBTreeTester("1"))
		treeItem2 := bTree.FindOrCreate(keyOne, newBTreeTester("2"))
		g.Expect(treeItem1).To(Equal(treeItem2))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].item.(*bTreeTester).value).To(Equal("1"))
		g.Expect(bTree.root.values[0].item.(*bTreeTester).onFindCount).To(Equal(1))
	})

	t.Run("inserts the items in the proper order, no matter how they were added to the tree", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem2 := bTree.FindOrCreate(keyTwo, newBTreeTester("2"))
		treeItem1 := bTree.FindOrCreate(keyOne, newBTreeTester("1"))

		g.Expect(len(bTree.root.values)).To(Equal(2))
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

		_ = bTree.FindOrCreate(keyTwo, newBTreeTester("2"))
		_ = bTree.FindOrCreate(keyOne, newBTreeTester("1"))
		_ = bTree.FindOrCreate(keyOne, newBTreeTester("3"))

		g.Expect(len(bTree.root.values)).To(Equal(2))
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

		treeItem2 := bTree.FindOrCreate(keyTwo, newBTreeTester("2"))
		treeItem3 := bTree.FindOrCreate(keyThree, newBTreeTester("3"))
		treeItem1 := bTree.FindOrCreate(keyOne, newBTreeTester("1"))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem2))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(len(leftchild.values)).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(keyOne))
		g.Expect(leftchild.values[0].item).To(Equal(treeItem1))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(len(rightchild.values)).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(keyThree))
		g.Expect(rightchild.values[0].item).To(Equal(treeItem3))
	})

	t.Run("it splits the node when adding a pivot item value", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem3 := bTree.FindOrCreate(keyThree, newBTreeTester("3"))
		treeItem2 := bTree.FindOrCreate(keyTwo, newBTreeTester("2"))
		treeItem1 := bTree.FindOrCreate(keyOne, newBTreeTester("1"))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem2))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(len(leftchild.values)).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(keyOne))
		g.Expect(leftchild.values[0].item).To(Equal(treeItem1))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(len(rightchild.values)).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(keyThree))
		g.Expect(rightchild.values[0].item).To(Equal(treeItem3))
	})

	t.Run("it splits the node when adding a right pivot item value", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem2 := bTree.FindOrCreate(keyTwo, newBTreeTester("2"))
		treeItem1 := bTree.FindOrCreate(keyOne, newBTreeTester("1"))
		treeItem3 := bTree.FindOrCreate(keyThree, newBTreeTester("3"))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[0].item).To(Equal(treeItem2))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(len(leftchild.values)).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(keyOne))
		g.Expect(leftchild.values[0].item).To(Equal(treeItem1))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(len(rightchild.values)).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(keyThree))
		g.Expect(rightchild.values[0].item).To(Equal(treeItem3))
	})
}

func TestBTree_FindOrCreate_Tree_SimpleOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	key10 := NewIntTreeKey(10)
	key20 := NewIntTreeKey(20)
	key30 := NewIntTreeKey(30)

	// generate a tree of
	/*
	 *     20
	 *   /    \
	 *  10    30
	 */
	setupTree := func(g *GomegaWithT) *BRoot {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		bTree.FindOrCreate(key10, newBTreeTester("10"))
		bTree.FindOrCreate(key20, newBTreeTester("20"))
		bTree.FindOrCreate(key30, newBTreeTester("30"))

		return bTree
	}

	t.Run("can return a entry on the leftChild if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem := bTree.FindOrCreate(key10, newBTreeTester("10"))

		leftChild := bTree.root.children[0]
		g.Expect(len(leftChild.values)).To(Equal(1))
		g.Expect(leftChild.values[0].item).To(Equal(treeItem))
	})

	t.Run("can return a entry on the rightChild if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem := bTree.FindOrCreate(key30, newBTreeTester("30"))

		rightChild := bTree.root.children[1]
		g.Expect(len(rightChild.values)).To(Equal(1))
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
		key5 := NewIntTreeKey(5)

		_ = bTree.FindOrCreate(key5, newBTreeTester("5"))

		leftChild := bTree.root.children[0]
		g.Expect(len(leftChild.values)).To(Equal(2))
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
		key25 := NewIntTreeKey(25)

		_ = bTree.FindOrCreate(key25, newBTreeTester("25"))

		rightChild := bTree.root.children[1]
		g.Expect(len(rightChild.values)).To(Equal(2))
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
		key15 := NewIntTreeKey(15)

		_ = bTree.FindOrCreate(key15, newBTreeTester("15"))

		leftChild := bTree.root.children[0]
		g.Expect(len(leftChild.values)).To(Equal(2))
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
		key35 := NewIntTreeKey(35)

		_ = bTree.FindOrCreate(key35, newBTreeTester("35"))

		rightChild := bTree.root.children[1]
		g.Expect(len(rightChild.values)).To(Equal(2))
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
	key10 := NewIntTreeKey(10)
	key20 := NewIntTreeKey(20)
	key30 := NewIntTreeKey(30)

	t.Run("leftChild promotions", func(t *testing.T) {
		key15 := NewIntTreeKey(15)

		// values to add
		key8 := NewIntTreeKey(8)
		key12 := NewIntTreeKey(12)
		key17 := NewIntTreeKey(17)

		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 * 10,15   30
		 */
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			_ = bTree.FindOrCreate(key30, newBTreeTester("30"))
			_ = bTree.FindOrCreate(key10, newBTreeTester("10"))
			_ = bTree.FindOrCreate(key20, newBTreeTester("20"))
			_ = bTree.FindOrCreate(key15, newBTreeTester("15"))

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

			treeItem := bTree.FindOrCreate(key8, newBTreeTester("8"))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key10))
			g.Expect(bTree.root.values[1].key).To(Equal(key20))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key8))
			g.Expect(child1.values[0].item).To(Equal(treeItem))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key15))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
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

			treeItem := bTree.FindOrCreate(key12, newBTreeTester("12"))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key12))
			g.Expect(bTree.root.values[0].item).To(Equal(treeItem))
			g.Expect(bTree.root.values[1].key).To(Equal(key20))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key10))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key15))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
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

			treeItem := bTree.FindOrCreate(key17, newBTreeTester("17"))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key15))
			g.Expect(bTree.root.values[1].key).To(Equal(key20))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key10))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key17))
			g.Expect(child2.values[0].item).To(Equal(treeItem))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(key30))
		})
	})

	t.Run("rightChild promotions", func(t *testing.T) {
		key35 := NewIntTreeKey(35)

		// values to add
		key25 := NewIntTreeKey(25)
		key32 := NewIntTreeKey(32)
		key37 := NewIntTreeKey(37)

		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 *   10   30,35
		 */
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			bTree.FindOrCreate(key20, newBTreeTester("20"))
			bTree.FindOrCreate(key30, newBTreeTester("30"))
			bTree.FindOrCreate(key10, newBTreeTester("10"))
			bTree.FindOrCreate(key35, newBTreeTester("35"))

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

			treeItem := bTree.FindOrCreate(key25, newBTreeTester("25"))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key20))
			g.Expect(bTree.root.values[1].key).To(Equal(key30))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key10))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key25))
			g.Expect(child2.values[0].item).To(Equal(treeItem))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
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

			treeItem := bTree.FindOrCreate(key32, newBTreeTester("32"))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key20))
			g.Expect(bTree.root.values[1].key).To(Equal(key32))
			g.Expect(bTree.root.values[1].item).To(Equal(treeItem))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key10))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key30))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
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

			treeItem := bTree.FindOrCreate(key37, newBTreeTester("37"))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].key).To(Equal(key20))
			g.Expect(bTree.root.values[1].key).To(Equal(key35))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].key).To(Equal(key10))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].key).To(Equal(key30))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
			g.Expect(child3.values[0].key).To(Equal(key37))
			g.Expect(child3.values[0].item).To(Equal(treeItem))
		})
	})
}

func TestBTree_FindOrCreate_Tree_NewRootNode(t *testing.T) {
	g := NewGomegaWithT(t)

	// setup values
	key5 := NewIntTreeKey(5)
	key10 := NewIntTreeKey(10)
	key20 := NewIntTreeKey(20)
	key25 := NewIntTreeKey(25)
	key30 := NewIntTreeKey(30)
	key40 := NewIntTreeKey(40)
	key45 := NewIntTreeKey(45)
	key50 := NewIntTreeKey(50)

	// all tests in this section start with a base tree like so
	/*
	 *        20,40
	 *    /     |     \
	 *  5,10  25,30  45,50
	 */
	setupTree := func(g *GomegaWithT) *BRoot {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		bTree.FindOrCreate(key10, newBTreeTester("10"))
		bTree.FindOrCreate(key20, newBTreeTester("20"))
		bTree.FindOrCreate(key30, newBTreeTester("30"))
		bTree.FindOrCreate(key40, newBTreeTester("40"))
		bTree.FindOrCreate(key50, newBTreeTester("50"))
		bTree.FindOrCreate(key5, newBTreeTester("5"))
		bTree.FindOrCreate(key25, newBTreeTester("25"))
		bTree.FindOrCreate(key45, newBTreeTester("45"))

		g.Expect(len(bTree.root.values)).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(key20))
		g.Expect(bTree.root.values[1].key).To(Equal(key40))

		child1 := bTree.root.children[0]
		g.Expect(len(child1.values)).To(Equal(2))
		g.Expect(child1.values[0].key).To(Equal(key5))
		g.Expect(child1.values[1].key).To(Equal(key10))

		child2 := bTree.root.children[1]
		g.Expect(len(child2.values)).To(Equal(2))
		g.Expect(child2.values[0].key).To(Equal(key25))
		g.Expect(child2.values[1].key).To(Equal(key30))

		child3 := bTree.root.children[2]
		g.Expect(len(child3.values)).To(Equal(2))
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
		key0 := NewIntTreeKey(0)

		treeItem := bTree.FindOrCreate(key0, newBTreeTester("0"))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(key20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(len(child1.values)).To(Equal(1))
		g.Expect(child1.values[0].key).To(Equal(key5))

		gchild1 := child1.children[0]
		g.Expect(len(gchild1.values)).To(Equal(1))
		g.Expect(gchild1.values[0].key).To(Equal(key0))
		g.Expect(gchild1.values[0].item).To(Equal(treeItem))

		gchild2 := child1.children[1]
		g.Expect(len(gchild2.values)).To(Equal(1))
		g.Expect(gchild2.values[0].key).To(Equal(key10))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(len(child2.values)).To(Equal(1))
		g.Expect(child2.values[0].key).To(Equal(key40))

		gchild1 = child2.children[0]
		g.Expect(len(gchild1.values)).To(Equal(2))
		g.Expect(gchild1.values[0].key).To(Equal(key25))
		g.Expect(gchild1.values[1].key).To(Equal(key30))

		gchild2 = child2.children[1]
		g.Expect(len(gchild2.values)).To(Equal(2))
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
		key22 := NewIntTreeKey(22)

		treeItem := bTree.FindOrCreate(key22, newBTreeTester("22"))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(key25))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(len(child1.values)).To(Equal(1))
		g.Expect(child1.values[0].key).To(Equal(key20))

		gchild1 := child1.children[0]
		g.Expect(len(gchild1.values)).To(Equal(2))
		g.Expect(gchild1.values[0].key).To(Equal(key5))
		g.Expect(gchild1.values[1].key).To(Equal(key10))

		gchild2 := child1.children[1]
		g.Expect(len(gchild2.values)).To(Equal(1))
		g.Expect(gchild2.values[0].key).To(Equal(key22))
		g.Expect(gchild2.values[0].item).To(Equal(treeItem))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(len(child2.values)).To(Equal(1))
		g.Expect(child2.values[0].key).To(Equal(key40))

		gchild1 = child2.children[0]
		g.Expect(len(gchild1.values)).To(Equal(1))
		g.Expect(gchild1.values[0].key).To(Equal(key30))

		gchild2 = child2.children[1]
		g.Expect(len(gchild2.values)).To(Equal(2))
		g.Expect(gchild2.values[0].key).To(Equal(key45))
		g.Expect(gchild2.values[1].key).To(Equal(key50))
	})

	// generate a tree of
	/*
	 *              40
	 *         /         \
	 *        20         47
	 *      /    \        /   \
	 *    5,10  25,30  45   50
	 */
	t.Run("adding the right promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		key47 := NewIntTreeKey(47)

		treeItem := bTree.FindOrCreate(key47, newBTreeTester("47"))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(key40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(len(child1.values)).To(Equal(1))
		g.Expect(child1.values[0].key).To(Equal(key20))

		gchild1 := child1.children[0]
		g.Expect(len(gchild1.values)).To(Equal(2))
		g.Expect(gchild1.values[0].key).To(Equal(key5))
		g.Expect(gchild1.values[1].key).To(Equal(key10))

		gchild2 := child1.children[1]
		g.Expect(len(gchild2.values)).To(Equal(2))
		g.Expect(gchild2.values[0].key).To(Equal(key25))
		g.Expect(gchild2.values[1].key).To(Equal(key30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(len(child2.values)).To(Equal(1))
		g.Expect(child2.values[0].key).To(Equal(key47))
		g.Expect(child2.values[0].item).To(Equal(treeItem))

		gchild1 = child2.children[0]
		g.Expect(len(gchild1.values)).To(Equal(1))
		g.Expect(gchild1.values[0].key).To(Equal(key45))

		gchild2 = child2.children[1]
		g.Expect(len(gchild2.values)).To(Equal(1))
		g.Expect(gchild2.values[0].key).To(Equal(key50))
	})
}

func TestBTree_RandomAssertion(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("works for a tree order of 2", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := NewIntTreeKey(num)
			_ = bTree.FindOrCreate(key, newBTreeTester(fmt.Sprintf("%d", num)))
		}

		validateTree(g, bTree.root, nil, true)
	})

	t.Run("works for a tree order of 3", func(t *testing.T) {
		bTree, err := NewBTree(3)
		g.Expect(err).ToNot(HaveOccurred())

		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := NewIntTreeKey(num)
			_ = bTree.FindOrCreate(key, newBTreeTester(fmt.Sprintf("%d", num)))
		}

		validateTree(g, bTree.root, nil, true)
	})

	t.Run("works for a tree order of 4", func(t *testing.T) {
		bTree, err := NewBTree(4)
		g.Expect(err).ToNot(HaveOccurred())

		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := NewIntTreeKey(num)
			_ = bTree.FindOrCreate(key, newBTreeTester(fmt.Sprintf("%d", num)))
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
			_ = bTree.FindOrCreate(NewIntTreeKey(i), newBTreeTester(fmt.Sprintf("%d", i)))
		}

		return bTree
	}

	t.Run("returns nil if the item does not exist in the tree", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Find(NewIntTreeKey(1))).To(BeNil())
	})

	t.Run("returns the item in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem := bTree.Find(NewIntTreeKey(768))
		g.Expect(treeItem).ToNot(BeNil())
		g.Expect(treeItem.(*bTreeTester).value).To(Equal("768"))
		g.Expect(treeItem.(*bTreeTester).onFindCount).To(Equal(1))
	})
}
