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
	keyOne    = NewIntTreeKey(1)
	itemOne   = &bTreeTester{value: "1"}
	keyTwo    = NewIntTreeKey(2)
	itemTwo   = &bTreeTester{value: "2"}
	keyThree  = NewIntTreeKey(3)
	itemThree = &bTreeTester{value: "3"}
)

func cleanup() {
	itemOne.onFindCount = 0
	itemTwo.onFindCount = 0
	itemThree.onFindCount = 0
}

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
		defer cleanup()
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem := bTree.FindOrCreate(keyOne, itemOne)
		g.Expect(treeItem).To(Equal(itemOne))

		treeItem = bTree.FindOrCreate(keyTwo, itemTwo)
		g.Expect(treeItem).To(Equal(itemTwo))

		g.Expect(len(bTree.root.values)).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(keyOne))
		g.Expect(bTree.root.values[0].item).To(Equal(itemOne))
		g.Expect(bTree.root.values[1].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[1].item).To(Equal(itemTwo))
		g.Expect(itemOne.onFindCount).To(Equal(1))
		g.Expect(itemTwo.onFindCount).To(Equal(1))
	})

	t.Run("adding the same item multiple times returns the original inserted item", func(t *testing.T) {
		defer cleanup()
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem := bTree.FindOrCreate(keyOne, itemOne)
		g.Expect(treeItem).To(Equal(itemOne))
		treeItem = bTree.FindOrCreate(keyOne, itemOne)
		g.Expect(treeItem).To(Equal(itemOne))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].item).To(Equal(itemOne))
		g.Expect(itemOne.onFindCount).To(Equal(2))
	})

	t.Run("inserts the items in the proper order, no matter how they were added to the tree", func(t *testing.T) {
		defer cleanup()
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		treeItem := bTree.FindOrCreate(keyTwo, itemTwo)
		g.Expect(treeItem).To(Equal(itemTwo))
		treeItem = bTree.FindOrCreate(keyOne, itemOne)
		g.Expect(treeItem).To(Equal(itemOne))

		g.Expect(len(bTree.root.values)).To(Equal(2))
		g.Expect(bTree.root.values[0].key).To(Equal(keyOne))
		g.Expect(bTree.root.values[0].item).To(Equal(itemOne))
		g.Expect(bTree.root.values[1].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[1].item).To(Equal(itemTwo))
		g.Expect(itemOne.onFindCount).To(Equal(1))
		g.Expect(itemTwo.onFindCount).To(Equal(1))
	})

	t.Run("possible split returns an item that already exists", func(t *testing.T) {
		defer cleanup()
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.FindOrCreate(keyTwo, itemTwo)).To(Equal(itemTwo))
		g.Expect(bTree.FindOrCreate(keyOne, itemOne)).To(Equal(itemOne))
		g.Expect(bTree.FindOrCreate(keyOne, itemOne)).To(Equal(itemOne))

		g.Expect(len(bTree.root.values)).To(Equal(2))
		g.Expect(itemOne.onFindCount).To(Equal(2))
		g.Expect(itemTwo.onFindCount).To(Equal(1))
	})

	t.Run("it splits the node when adding a left pivot value", func(t *testing.T) {
		defer cleanup()
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.FindOrCreate(keyTwo, itemTwo)).To(Equal(itemTwo))
		g.Expect(bTree.FindOrCreate(keyThree, itemThree)).To(Equal(itemThree))
		g.Expect(bTree.FindOrCreate(keyOne, itemOne)).To(Equal(itemOne))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[0].item).To(Equal(itemTwo))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(len(leftchild.values)).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(keyOne))
		g.Expect(leftchild.values[0].item).To(Equal(itemOne))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(len(rightchild.values)).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(keyThree))
		g.Expect(rightchild.values[0].item).To(Equal(itemThree))

		g.Expect(itemOne.onFindCount).To(Equal(1))
		g.Expect(itemTwo.onFindCount).To(Equal(1))
		g.Expect(itemThree.onFindCount).To(Equal(1))
	})

	t.Run("it splits the node when adding a pivot item value", func(t *testing.T) {
		defer cleanup()
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.FindOrCreate(keyThree, itemThree)).To(Equal(itemThree))
		g.Expect(bTree.FindOrCreate(keyOne, itemOne)).To(Equal(itemOne))
		g.Expect(bTree.FindOrCreate(keyTwo, itemTwo)).To(Equal(itemTwo))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[0].item).To(Equal(itemTwo))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(len(leftchild.values)).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(keyOne))
		g.Expect(leftchild.values[0].item).To(Equal(itemOne))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(len(rightchild.values)).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(keyThree))
		g.Expect(rightchild.values[0].item).To(Equal(itemThree))

		g.Expect(itemOne.onFindCount).To(Equal(1))
		g.Expect(itemTwo.onFindCount).To(Equal(1))
		g.Expect(itemThree.onFindCount).To(Equal(1))
	})

	t.Run("it splits the node when adding a right pivot item value", func(t *testing.T) {
		defer cleanup()
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.FindOrCreate(keyTwo, itemTwo)).To(Equal(itemTwo))
		g.Expect(bTree.FindOrCreate(keyOne, itemOne)).To(Equal(itemOne))
		g.Expect(bTree.FindOrCreate(keyThree, itemThree)).To(Equal(itemThree))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].key).To(Equal(keyTwo))
		g.Expect(bTree.root.values[0].item).To(Equal(itemTwo))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(len(leftchild.values)).To(Equal(1))
		g.Expect(leftchild.values[0].key).To(Equal(keyOne))
		g.Expect(leftchild.values[0].item).To(Equal(itemOne))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(len(rightchild.values)).To(Equal(1))
		g.Expect(rightchild.values[0].key).To(Equal(keyThree))
		g.Expect(rightchild.values[0].item).To(Equal(itemThree))

		g.Expect(itemOne.onFindCount).To(Equal(1))
		g.Expect(itemTwo.onFindCount).To(Equal(1))
		g.Expect(itemThree.onFindCount).To(Equal(1))
	})
}

func TestBTree_FindOrCreate_Tree_SimpleOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	key10 := NewIntTreeKey(10)
	item10 := &bTreeTester{value: "10"}
	key20 := NewIntTreeKey(20)
	item20 := &bTreeTester{value: "20"}
	key30 := NewIntTreeKey(30)
	item30 := &bTreeTester{value: "30"}

	// generate a tree of
	/*
	 *     20
	 *   /    \
	 *  10    30
	 */
	setupTree := func(g *GomegaWithT) *BRoot {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		bTree.FindOrCreate(key10, item10)
		bTree.FindOrCreate(key20, item20)
		bTree.FindOrCreate(key30, item30)

		return bTree
	}

	t.Run("can return a entry on the leftChild if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem := bTree.FindOrCreate(key10, item10)
		g.Expect(treeItem).To(Equal(item10))

		leftChild := bTree.root.children[0]
		g.Expect(len(leftChild.values)).To(Equal(1))
		g.Expect(leftChild.values[0].item).To(Equal(item10))
	})

	t.Run("can return a entry on the rightChild if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem := bTree.FindOrCreate(key30, item30)
		g.Expect(treeItem).To(Equal(item30))

		rightChild := bTree.root.children[1]
		g.Expect(len(rightChild.values)).To(Equal(1))
		g.Expect(rightChild.values[0].item).To(Equal(item30))
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
		item5 := &bTreeTester{value: "5"}

		treeItem := bTree.FindOrCreate(key5, item5)
		g.Expect(treeItem).To(Equal(item5))

		leftChild := bTree.root.children[0]
		g.Expect(len(leftChild.values)).To(Equal(2))
		g.Expect(leftChild.values[0].item).To(Equal(item5))
		g.Expect(leftChild.values[1].item).To(Equal(item10))
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
		item25 := &bTreeTester{value: "25"}

		treeItem := bTree.FindOrCreate(key25, item25)
		g.Expect(treeItem).To(Equal(item25))

		rightChild := bTree.root.children[1]
		g.Expect(len(rightChild.values)).To(Equal(2))
		g.Expect(rightChild.values[0].item).To(Equal(item25))
		g.Expect(rightChild.values[1].item).To(Equal(item30))
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
		item15 := &bTreeTester{value: "15"}

		treeItem := bTree.FindOrCreate(key15, item15)
		g.Expect(treeItem).To(Equal(item15))

		leftChild := bTree.root.children[0]
		g.Expect(len(leftChild.values)).To(Equal(2))
		g.Expect(leftChild.values[0].item).To(Equal(item10))
		g.Expect(leftChild.values[1].item).To(Equal(item15))
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
		item35 := &bTreeTester{value: "35"}

		treeItem := bTree.FindOrCreate(key35, item35)
		g.Expect(treeItem).To(Equal(item35))

		rightChild := bTree.root.children[1]
		g.Expect(len(rightChild.values)).To(Equal(2))
		g.Expect(rightChild.values[0].item).To(Equal(item30))
		g.Expect(rightChild.values[1].item).To(Equal(item35))
	})
}

func TestBTree_FindOrCreate_Tree_SimplePromoteOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	// setup values
	key10 := NewIntTreeKey(10)
	item10 := &bTreeTester{value: "10"}
	key20 := NewIntTreeKey(20)
	item20 := &bTreeTester{value: "20"}
	key30 := NewIntTreeKey(30)
	item30 := &bTreeTester{value: "30"}

	t.Run("leftChild promotions", func(t *testing.T) {
		key15 := NewIntTreeKey(15)
		item15 := &bTreeTester{value: "15"}

		// values to add
		key8 := NewIntTreeKey(8)
		item8 := &bTreeTester{value: "8"}
		key12 := NewIntTreeKey(12)
		item12 := &bTreeTester{value: "12"}
		key17 := NewIntTreeKey(17)
		item17 := &bTreeTester{value: "17"}

		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 * 10,15   30
		 */
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			bTree.FindOrCreate(key30, item30)
			bTree.FindOrCreate(key10, item10)
			bTree.FindOrCreate(key20, item20)
			bTree.FindOrCreate(key15, item15)

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

			treeItem := bTree.FindOrCreate(key8, item8)
			g.Expect(treeItem).To(Equal(item8))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].item).To(Equal(item10))
			g.Expect(bTree.root.values[1].item).To(Equal(item20))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].item).To(Equal(item8))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].item).To(Equal(item15))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
			g.Expect(child3.values[0].item).To(Equal(item30))
		})

		// generate a tree of
		/*
		 *      12,20
		 *    /   |   \
		 *   10   15  30
		 */
		t.Run("adding the leftChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem := bTree.FindOrCreate(key12, item12)
			g.Expect(treeItem).To(Equal(item12))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].item).To(Equal(item12))
			g.Expect(bTree.root.values[1].item).To(Equal(item20))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].item).To(Equal(item10))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].item).To(Equal(item15))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
			g.Expect(child3.values[0].item).To(Equal(item30))
		})

		// generate a tree of
		/*
		 *      15,20
		 *    /   |   \
		 *   10   17  30
		 */
		t.Run("adding the leftChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem := bTree.FindOrCreate(key17, item17)
			g.Expect(treeItem).To(Equal(item17))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].item).To(Equal(item15))
			g.Expect(bTree.root.values[1].item).To(Equal(item20))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].item).To(Equal(item10))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].item).To(Equal(item17))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
			g.Expect(child3.values[0].item).To(Equal(item30))
		})
	})

	t.Run("rightChild promotions", func(t *testing.T) {
		key35 := NewIntTreeKey(35)
		item35 := &bTreeTester{value: "35"}

		// values to add
		key25 := NewIntTreeKey(25)
		item25 := &bTreeTester{value: "25"}
		key32 := NewIntTreeKey(32)
		item32 := &bTreeTester{value: "32"}
		key37 := NewIntTreeKey(37)
		item37 := &bTreeTester{value: "37"}

		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 *   10   30,35
		 */
		setupTree := func(g *GomegaWithT) *BRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			bTree.FindOrCreate(key20, item20)
			bTree.FindOrCreate(key30, item30)
			bTree.FindOrCreate(key10, item10)
			bTree.FindOrCreate(key35, item35)

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

			treeItem := bTree.FindOrCreate(key25, item25)
			g.Expect(treeItem).To(Equal(item25))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].item).To(Equal(item20))
			g.Expect(bTree.root.values[1].item).To(Equal(item30))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].item).To(Equal(item10))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].item).To(Equal(item25))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
			g.Expect(child3.values[0].item).To(Equal(item35))
		})

		// generate a tree of
		/*
		 *      20,32
		 *    /   |   \
		 *   10   30  35
		 */
		t.Run("adding the rightChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem := bTree.FindOrCreate(key32, item32)
			g.Expect(treeItem).To(Equal(item32))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].item).To(Equal(item20))
			g.Expect(bTree.root.values[1].item).To(Equal(item32))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].item).To(Equal(item10))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].item).To(Equal(item30))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
			g.Expect(child3.values[0].item).To(Equal(item35))
		})

		// generate a tree of
		/*
		 *      20,35
		 *    /   |   \
		 *   10   30  37
		 */
		t.Run("adding the rightChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem := bTree.FindOrCreate(key37, item37)
			g.Expect(treeItem).To(Equal(item37))

			g.Expect(len(bTree.root.values)).To(Equal(2))
			g.Expect(bTree.root.values[0].item).To(Equal(item20))
			g.Expect(bTree.root.values[1].item).To(Equal(item35))

			child1 := bTree.root.children[0]
			g.Expect(len(child1.values)).To(Equal(1))
			g.Expect(child1.values[0].item).To(Equal(item10))

			child2 := bTree.root.children[1]
			g.Expect(len(child2.values)).To(Equal(1))
			g.Expect(child2.values[0].item).To(Equal(item30))

			child3 := bTree.root.children[2]
			g.Expect(len(child3.values)).To(Equal(1))
			g.Expect(child3.values[0].item).To(Equal(item37))
		})
	})
}

func TestBTree_FindOrCreate_Tree_NewRootNode(t *testing.T) {
	g := NewGomegaWithT(t)

	// setup values
	key5 := NewIntTreeKey(5)
	item5 := &bTreeTester{value: "5"}
	key10 := NewIntTreeKey(10)
	item10 := &bTreeTester{value: "10"}
	key20 := NewIntTreeKey(20)
	item20 := &bTreeTester{value: "20"}
	key25 := NewIntTreeKey(25)
	item25 := &bTreeTester{value: "25"}
	key30 := NewIntTreeKey(30)
	item30 := &bTreeTester{value: "30"}
	key40 := NewIntTreeKey(40)
	item40 := &bTreeTester{value: "40"}
	key45 := NewIntTreeKey(45)
	item45 := &bTreeTester{value: "45"}
	key50 := NewIntTreeKey(50)
	item50 := &bTreeTester{value: "50"}

	// all tests in this section start with a base tree like so
	/*
	 *        20,40
	 *    /     |     \
	 *  5,10  25,30  45,50
	 */
	setupTree := func(g *GomegaWithT) *BRoot {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.FindOrCreate(key10, item10)).To(Equal(item10))
		g.Expect(bTree.FindOrCreate(key20, item20)).To(Equal(item20))
		g.Expect(bTree.FindOrCreate(key30, item30)).To(Equal(item30))
		g.Expect(bTree.FindOrCreate(key40, item40)).To(Equal(item40))
		g.Expect(bTree.FindOrCreate(key50, item50)).To(Equal(item50))
		g.Expect(bTree.FindOrCreate(key5, item5)).To(Equal(item5))
		g.Expect(bTree.FindOrCreate(key25, item25)).To(Equal(item25))
		g.Expect(bTree.FindOrCreate(key45, item45)).To(Equal(item45))

		g.Expect(len(bTree.root.values)).To(Equal(2))
		g.Expect(bTree.root.values[0].item).To(Equal(item20))
		g.Expect(bTree.root.values[1].item).To(Equal(item40))

		child1 := bTree.root.children[0]
		g.Expect(len(child1.values)).To(Equal(2))
		g.Expect(child1.values[0].item).To(Equal(item5))
		g.Expect(child1.values[1].item).To(Equal(item10))

		child2 := bTree.root.children[1]
		g.Expect(len(child2.values)).To(Equal(2))
		g.Expect(child2.values[0].item).To(Equal(item25))
		g.Expect(child2.values[1].item).To(Equal(item30))

		child3 := bTree.root.children[2]
		g.Expect(len(child3.values)).To(Equal(2))
		g.Expect(child3.values[0].item).To(Equal(item45))
		g.Expect(child3.values[1].item).To(Equal(item50))

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
		item0 := &bTreeTester{value: "0"}

		treeItem := bTree.FindOrCreate(key0, item0)
		g.Expect(treeItem).To(Equal(item0))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].item).To(Equal(item20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(len(child1.values)).To(Equal(1))
		g.Expect(child1.values[0].item).To(Equal(item5))

		gchild1 := child1.children[0]
		g.Expect(len(gchild1.values)).To(Equal(1))
		g.Expect(gchild1.values[0].item).To(Equal(item0))

		gchild2 := child1.children[1]
		g.Expect(len(gchild2.values)).To(Equal(1))
		g.Expect(gchild2.values[0].item).To(Equal(item10))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(len(child2.values)).To(Equal(1))
		g.Expect(child2.values[0].item).To(Equal(item40))

		gchild1 = child2.children[0]
		g.Expect(len(gchild1.values)).To(Equal(2))
		g.Expect(gchild1.values[0].item).To(Equal(item25))
		g.Expect(gchild1.values[1].item).To(Equal(item30))

		gchild2 = child2.children[1]
		g.Expect(len(gchild2.values)).To(Equal(2))
		g.Expect(gchild2.values[0].item).To(Equal(item45))
		g.Expect(gchild2.values[1].item).To(Equal(item50))
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
		item22 := &bTreeTester{value: "22"}

		treeItem := bTree.FindOrCreate(key22, item22)
		g.Expect(treeItem).To(Equal(item22))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].item).To(Equal(item25))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(len(child1.values)).To(Equal(1))
		g.Expect(child1.values[0].item).To(Equal(item20))

		gchild1 := child1.children[0]
		g.Expect(len(gchild1.values)).To(Equal(2))
		g.Expect(gchild1.values[0].item).To(Equal(item5))
		g.Expect(gchild1.values[1].item).To(Equal(item10))

		gchild2 := child1.children[1]
		g.Expect(len(gchild2.values)).To(Equal(1))
		g.Expect(gchild2.values[0].item).To(Equal(item22))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(len(child2.values)).To(Equal(1))
		g.Expect(child2.values[0].item).To(Equal(item40))

		gchild1 = child2.children[0]
		g.Expect(len(gchild1.values)).To(Equal(1))
		g.Expect(gchild1.values[0].item).To(Equal(item30))

		gchild2 = child2.children[1]
		g.Expect(len(gchild2.values)).To(Equal(2))
		g.Expect(gchild2.values[0].item).To(Equal(item45))
		g.Expect(gchild2.values[1].item).To(Equal(item50))
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
		item47 := &bTreeTester{value: "47"}

		treeItem := bTree.FindOrCreate(key47, item47)
		g.Expect(treeItem).To(Equal(item47))

		g.Expect(len(bTree.root.values)).To(Equal(1))
		g.Expect(bTree.root.values[0].item).To(Equal(item40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(len(child1.values)).To(Equal(1))
		g.Expect(child1.values[0].item).To(Equal(item20))

		gchild1 := child1.children[0]
		g.Expect(len(gchild1.values)).To(Equal(2))
		g.Expect(gchild1.values[0].item).To(Equal(item5))
		g.Expect(gchild1.values[1].item).To(Equal(item10))

		gchild2 := child1.children[1]
		g.Expect(len(gchild2.values)).To(Equal(2))
		g.Expect(gchild2.values[0].item).To(Equal(item25))
		g.Expect(gchild2.values[1].item).To(Equal(item30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(len(child2.values)).To(Equal(1))
		g.Expect(child2.values[0].item).To(Equal(item47))

		gchild1 = child2.children[0]
		g.Expect(len(gchild1.values)).To(Equal(1))
		g.Expect(gchild1.values[0].item).To(Equal(item45))

		gchild2 = child2.children[1]
		g.Expect(len(gchild2.values)).To(Equal(1))
		g.Expect(gchild2.values[0].item).To(Equal(item50))
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
			item := &bTreeTester{value: fmt.Sprintf("%d", num)}
			_ = bTree.FindOrCreate(key, item)
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
			item := &bTreeTester{value: fmt.Sprintf("%d", num)}
			_ = bTree.FindOrCreate(key, item)
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
			item := &bTreeTester{value: fmt.Sprintf("%d", num)}
			_ = bTree.FindOrCreate(key, item)
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
			item := &bTreeTester{value: fmt.Sprintf("%d", i)}
			_ = bTree.FindOrCreate(NewIntTreeKey(i), item)
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
		g.Expect(treeItem.(*bTreeTester).onFindCount).To(Equal(2))
	})
}
