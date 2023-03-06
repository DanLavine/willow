package datastructures

import (
	"testing"

	. "github.com/onsi/gomega"
)

type twoThreeTester struct {
	num   int
	value string
}

func (tester *twoThreeTester) Less(compareItem TreeItem) bool {
	return tester.num < compareItem.(*twoThreeTester).num
}

func TestBPLusTtree_NewBTree(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("returns an error if the order is to small", func(t *testing.T) {
		bTree, err := NewBTree(1)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("order must be greater than 1"))
		g.Expect(bTree).To(BeNil())
	})
}

func TestBPLusTtree_FindOrCreate_SingleNode(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("creates a new tree with proper size limits", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		itemOne := &twoThreeTester{num: 1, value: "one"}
		itemTwo := &twoThreeTester{num: 2, value: "two"}

		treeItem := bTree.FindOrCreate(itemOne)
		g.Expect(treeItem).To(Equal(itemOne))

		treeItem = bTree.FindOrCreate(itemTwo)
		g.Expect(treeItem).To(Equal(itemTwo))

		g.Expect(bTree.root.count).To(Equal(uint(2)))
		g.Expect(bTree.root.values[0]).To(Equal(itemOne))
		g.Expect(bTree.root.values[1]).To(Equal(itemTwo))
	})

	t.Run("adding the same item multiple times returns the original inserted item", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		itemOne := &twoThreeTester{num: 1, value: "one"}
		itemTwo := &twoThreeTester{num: 1, value: "two"}

		treeItem := bTree.FindOrCreate(itemOne)
		g.Expect(treeItem).To(Equal(itemOne))

		treeItem = bTree.FindOrCreate(itemTwo)
		g.Expect(treeItem).To(Equal(itemOne))

		g.Expect(bTree.root.count).To(Equal(uint(1)))
		g.Expect(bTree.root.values[0]).To(Equal(itemOne))
	})

	t.Run("inserts the items in the proper order, no matter how they were added to the tree", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		itemOne := &twoThreeTester{num: 1, value: "one"}
		itemTwo := &twoThreeTester{num: 2, value: "two"}

		treeItem := bTree.FindOrCreate(itemTwo)
		g.Expect(treeItem).To(Equal(itemTwo))

		treeItem = bTree.FindOrCreate(itemOne)
		g.Expect(treeItem).To(Equal(itemOne))

		g.Expect(bTree.root.count).To(Equal(uint(2)))
		g.Expect(bTree.root.values[0]).To(Equal(itemOne))
		g.Expect(bTree.root.values[1]).To(Equal(itemTwo))
	})

	t.Run("it splits the node when adding an order+1 value", func(t *testing.T) {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		itemOne := &twoThreeTester{num: 1, value: "one"}
		itemTwo := &twoThreeTester{num: 2, value: "two"}
		itemThree := &twoThreeTester{num: 3, value: "three"}

		treeItem := bTree.FindOrCreate(itemOne)
		g.Expect(treeItem).To(Equal(itemOne))

		treeItem = bTree.FindOrCreate(itemTwo)
		g.Expect(treeItem).To(Equal(itemTwo))

		treeItem = bTree.FindOrCreate(itemThree)
		g.Expect(treeItem).To(Equal(itemThree))

		g.Expect(bTree.root.count).To(Equal(uint(1)))
		g.Expect(bTree.root.values[0]).To(Equal(itemTwo))

		// child 1
		child1 := bTree.root.children[0]
		g.Expect(child1.count).To(Equal(uint(1)))
		g.Expect(child1.values[0]).To(Equal(itemOne))
		// child 2
		child2 := bTree.root.children[1]
		g.Expect(child2.count).To(Equal(uint(1)))
		g.Expect(child2.values[0]).To(Equal(itemThree))
	})
}

func TestBPLusTtree_FindOrCreate_Tree_SimpleOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	item10 := &twoThreeTester{num: 10, value: "10"}
	item20 := &twoThreeTester{num: 20, value: "20"}
	item30 := &twoThreeTester{num: 30, value: "30"}

	// generate a tree of
	/*
	 *     20
	 *   /    \
	 *  10    30
	 */
	setupTree := func(g *GomegaWithT) *TwoThreeRoot {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		bTree.FindOrCreate(item10)
		bTree.FindOrCreate(item20)
		bTree.FindOrCreate(item30)

		return bTree
	}

	t.Run("can return a entry on the leftChild if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem := bTree.FindOrCreate(item10)
		g.Expect(treeItem).To(Equal(item10))

		g.Expect(bTree.root.count).To(Equal(uint(1)))

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.count).To(Equal(uint(1)))
		g.Expect(leftChild.values[0]).To(Equal(item10))
	})

	t.Run("can return a entry on the rightChild if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		treeItem := bTree.FindOrCreate(item30)
		g.Expect(treeItem).To(Equal(item30))

		g.Expect(bTree.root.count).To(Equal(uint(1)))

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.count).To(Equal(uint(1)))
		g.Expect(rightChild.values[0]).To(Equal(item30))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  5,10   30
	 */
	t.Run("can add a new entry on the leftChild[0]", func(t *testing.T) {
		bTree := setupTree(g)
		item5 := &twoThreeTester{num: 5, value: "5"}

		treeItem := bTree.FindOrCreate(item5)
		g.Expect(treeItem).To(Equal(item5))

		g.Expect(bTree.root.count).To(Equal(uint(1)))

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.count).To(Equal(uint(2)))
		g.Expect(leftChild.values[0]).To(Equal(item5))
		g.Expect(leftChild.values[1]).To(Equal(item10))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10     25,30
	 */
	t.Run("can add a new entry on the rightChild[0]", func(t *testing.T) {
		bTree := setupTree(g)
		item25 := &twoThreeTester{num: 25, value: "25"}

		treeItem := bTree.FindOrCreate(item25)
		g.Expect(treeItem).To(Equal(item25))

		g.Expect(bTree.root.count).To(Equal(uint(1)))

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.count).To(Equal(uint(2)))
		g.Expect(rightChild.values[0]).To(Equal(item25))
		g.Expect(rightChild.values[1]).To(Equal(item30))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10,15   30
	 */
	t.Run("can add a new entry on the leftChild[1]", func(t *testing.T) {
		bTree := setupTree(g)
		item15 := &twoThreeTester{num: 15, value: "15"}

		treeItem := bTree.FindOrCreate(item15)
		g.Expect(treeItem).To(Equal(item15))

		g.Expect(bTree.root.count).To(Equal(uint(1)))

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.count).To(Equal(uint(2)))
		g.Expect(leftChild.values[0]).To(Equal(item10))
		g.Expect(leftChild.values[1]).To(Equal(item15))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10    30,35
	 */
	t.Run("can add a new entry on the rightChild[1]", func(t *testing.T) {
		bTree := setupTree(g)
		item35 := &twoThreeTester{num: 35, value: "35"}

		treeItem := bTree.FindOrCreate(item35)
		g.Expect(treeItem).To(Equal(item35))

		g.Expect(bTree.root.count).To(Equal(uint(1)))

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.count).To(Equal(uint(2)))
		g.Expect(rightChild.values[0]).To(Equal(item30))
		g.Expect(rightChild.values[1]).To(Equal(item35))
	})
}

func TestBPLusTtree_FindOrCreate_Tree_SimplePromoteOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	// setup values
	item10 := &twoThreeTester{num: 10, value: "10"} // both
	item20 := &twoThreeTester{num: 20, value: "20"} // both
	item30 := &twoThreeTester{num: 30, value: "30"} // both

	t.Run("leftChild promotions", func(t *testing.T) {
		item15 := &twoThreeTester{num: 15, value: "15"} // left

		// values to add
		item8 := &twoThreeTester{num: 8, value: "8"}
		item12 := &twoThreeTester{num: 12, value: "12"}
		item17 := &twoThreeTester{num: 17, value: "17"}

		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 * 10,15   30
		 */
		setupTree := func(g *GomegaWithT) *TwoThreeRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			bTree.FindOrCreate(item30)
			bTree.FindOrCreate(item10)
			bTree.FindOrCreate(item20)
			bTree.FindOrCreate(item15)

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

			treeItem := bTree.FindOrCreate(item8)
			g.Expect(treeItem).To(Equal(item8))

			g.Expect(bTree.root.count).To(Equal(uint(2)))
			g.Expect(bTree.root.values[0]).To(Equal(item10))
			g.Expect(bTree.root.values[1]).To(Equal(item20))

			child1 := bTree.root.children[0]
			g.Expect(child1.count).To(Equal(uint(1)))
			g.Expect(child1.values[0]).To(Equal(item8))

			child2 := bTree.root.children[1]
			g.Expect(child2.count).To(Equal(uint(1)))
			g.Expect(child2.values[0]).To(Equal(item15))

			child3 := bTree.root.children[2]
			g.Expect(child3.count).To(Equal(uint(1)))
			g.Expect(child3.values[0]).To(Equal(item30))
		})

		// generate a tree of
		/*
		 *      12,20
		 *    /   |   \
		 *   10   15  30
		 */
		t.Run("adding the leftChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem := bTree.FindOrCreate(item12)
			g.Expect(treeItem).To(Equal(item12))

			g.Expect(bTree.root.count).To(Equal(uint(2)))
			g.Expect(bTree.root.values[0]).To(Equal(item12))
			g.Expect(bTree.root.values[1]).To(Equal(item20))

			child1 := bTree.root.children[0]
			g.Expect(child1.count).To(Equal(uint(1)))
			g.Expect(child1.values[0]).To(Equal(item10))

			child2 := bTree.root.children[1]
			g.Expect(child2.count).To(Equal(uint(1)))
			g.Expect(child2.values[0]).To(Equal(item15))

			child3 := bTree.root.children[2]
			g.Expect(child3.count).To(Equal(uint(1)))
			g.Expect(child3.values[0]).To(Equal(item30))
		})

		// generate a tree of
		/*
		 *      15,20
		 *    /   |   \
		 *   10   17  30
		 */
		t.Run("adding the leftChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem := bTree.FindOrCreate(item17)
			g.Expect(treeItem).To(Equal(item17))

			g.Expect(bTree.root.count).To(Equal(uint(2)))
			g.Expect(bTree.root.values[0]).To(Equal(item15))
			g.Expect(bTree.root.values[1]).To(Equal(item20))

			child1 := bTree.root.children[0]
			g.Expect(child1.count).To(Equal(uint(1)))
			g.Expect(child1.values[0]).To(Equal(item10))

			child2 := bTree.root.children[1]
			g.Expect(child2.count).To(Equal(uint(1)))
			g.Expect(child2.values[0]).To(Equal(item17))

			child3 := bTree.root.children[2]
			g.Expect(child3.count).To(Equal(uint(1)))
			g.Expect(child3.values[0]).To(Equal(item30))
		})
	})

	t.Run("rightChild promotions", func(t *testing.T) {
		item35 := &twoThreeTester{num: 35, value: "35"}

		// values to add
		item25 := &twoThreeTester{num: 25, value: "25"}
		item32 := &twoThreeTester{num: 32, value: "32"}
		item37 := &twoThreeTester{num: 37, value: "37"}

		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 *   10   30,35
		 */
		setupTree := func(g *GomegaWithT) *TwoThreeRoot {
			bTree, err := NewBTree(2)
			g.Expect(err).ToNot(HaveOccurred())

			bTree.FindOrCreate(item20)
			bTree.FindOrCreate(item30)
			bTree.FindOrCreate(item10)
			bTree.FindOrCreate(item35)

			return bTree
		}

		// generate a tree of
		/*
		 *      20,30
		 *    /   |  \
		 *   10   25  35
		 */
		t.Run("adding the rightChild[0] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem := bTree.FindOrCreate(item25)
			g.Expect(treeItem).To(Equal(item25))

			g.Expect(bTree.root.count).To(Equal(uint(2)))
			g.Expect(bTree.root.values[0]).To(Equal(item20))
			g.Expect(bTree.root.values[1]).To(Equal(item30))

			child1 := bTree.root.children[0]
			g.Expect(child1.count).To(Equal(uint(1)))
			g.Expect(child1.values[0]).To(Equal(item10))

			child2 := bTree.root.children[1]
			g.Expect(child2.count).To(Equal(uint(1)))
			g.Expect(child2.values[0]).To(Equal(item25))

			child3 := bTree.root.children[2]
			g.Expect(child3.count).To(Equal(uint(1)))
			g.Expect(child3.values[0]).To(Equal(item35))
		})

		// generate a tree of
		/*
		 *      20,32
		 *    /   |   \
		 *   10   30  35
		 */
		t.Run("adding the rightChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem := bTree.FindOrCreate(item32)
			g.Expect(treeItem).To(Equal(item32))

			g.Expect(bTree.root.count).To(Equal(uint(2)))
			g.Expect(bTree.root.values[0]).To(Equal(item20))
			g.Expect(bTree.root.values[1]).To(Equal(item32))

			child1 := bTree.root.children[0]
			g.Expect(child1.count).To(Equal(uint(1)))
			g.Expect(child1.values[0]).To(Equal(item10))

			child2 := bTree.root.children[1]
			g.Expect(child2.count).To(Equal(uint(1)))
			g.Expect(child2.values[0]).To(Equal(item30))

			child3 := bTree.root.children[2]
			g.Expect(child3.count).To(Equal(uint(1)))
			g.Expect(child3.values[0]).To(Equal(item35))
		})

		// generate a tree of
		/*
		 *      20,35
		 *    /   |   \
		 *   10   30  37
		 */
		t.Run("adding the rightChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)

			treeItem := bTree.FindOrCreate(item37)
			g.Expect(treeItem).To(Equal(item37))

			g.Expect(bTree.root.count).To(Equal(uint(2)))
			g.Expect(bTree.root.values[0]).To(Equal(item20))
			g.Expect(bTree.root.values[1]).To(Equal(item35))

			child1 := bTree.root.children[0]
			g.Expect(child1.count).To(Equal(uint(1)))
			g.Expect(child1.values[0]).To(Equal(item10))

			child2 := bTree.root.children[1]
			g.Expect(child2.count).To(Equal(uint(1)))
			g.Expect(child2.values[0]).To(Equal(item30))

			child3 := bTree.root.children[2]
			g.Expect(child3.count).To(Equal(uint(1)))
			g.Expect(child3.values[0]).To(Equal(item37))
		})
	})
}

func TestBPLusTtree_FindOrCreate_Tree_NewRootNode(t *testing.T) {
	g := NewGomegaWithT(t)

	// setup values
	item5 := &twoThreeTester{num: 5, value: "5"}
	item10 := &twoThreeTester{num: 10, value: "10"}
	item20 := &twoThreeTester{num: 20, value: "20"}
	item25 := &twoThreeTester{num: 25, value: "25"}
	item30 := &twoThreeTester{num: 30, value: "30"}
	item40 := &twoThreeTester{num: 40, value: "40"}
	item45 := &twoThreeTester{num: 45, value: "45"}
	item50 := &twoThreeTester{num: 50, value: "50"}

	// all tests in this section start with a base tree like so
	/*
	 *        20,40
	 *    /     |     \
	 *  5,10  25,30  45,50
	 */
	setupTree := func(g *GomegaWithT) *TwoThreeRoot {
		bTree, err := NewBTree(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.FindOrCreate(item10)).To(Equal(item10))
		g.Expect(bTree.FindOrCreate(item20)).To(Equal(item20))
		g.Expect(bTree.FindOrCreate(item30)).To(Equal(item30))
		g.Expect(bTree.FindOrCreate(item40)).To(Equal(item40))
		g.Expect(bTree.FindOrCreate(item50)).To(Equal(item50))
		g.Expect(bTree.FindOrCreate(item5)).To(Equal(item5))
		g.Expect(bTree.FindOrCreate(item25)).To(Equal(item25))
		g.Expect(bTree.FindOrCreate(item45)).To(Equal(item45))

		g.Expect(bTree.root.count).To(Equal(uint(2)))
		g.Expect(bTree.root.values[0]).To(Equal(item20))
		g.Expect(bTree.root.values[1]).To(Equal(item40))

		child1 := bTree.root.children[0]
		g.Expect(child1.count).To(Equal(uint(2)))
		g.Expect(child1.values[0]).To(Equal(item5))
		g.Expect(child1.values[1]).To(Equal(item10))

		child2 := bTree.root.children[1]
		g.Expect(child2.count).To(Equal(uint(2)))
		g.Expect(child2.values[0]).To(Equal(item25))
		g.Expect(child2.values[1]).To(Equal(item30))

		child3 := bTree.root.children[2]
		g.Expect(child3.count).To(Equal(uint(2)))
		g.Expect(child3.values[0]).To(Equal(item45))
		g.Expect(child3.values[1]).To(Equal(item50))

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

		item0 := &twoThreeTester{num: 0, value: "0"}
		treeItem := bTree.FindOrCreate(item0)
		g.Expect(treeItem).To(Equal(item0))

		g.Expect(bTree.root.count).To(Equal(uint(1)))
		g.Expect(bTree.root.values[0]).To(Equal(item20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.count).To(Equal(uint(1)))
		g.Expect(child1.values[0]).To(Equal(item5))

		gchild1 := child1.children[0]
		g.Expect(gchild1.count).To(Equal(uint(1)))
		g.Expect(gchild1.values[0]).To(Equal(item0))

		gchild2 := child1.children[1]
		g.Expect(gchild2.count).To(Equal(uint(1)))
		g.Expect(gchild2.values[0]).To(Equal(item10))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.count).To(Equal(uint(1)))
		g.Expect(child2.values[0]).To(Equal(item40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.count).To(Equal(uint(2)))
		g.Expect(gchild1.values[0]).To(Equal(item25))
		g.Expect(gchild1.values[1]).To(Equal(item30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.count).To(Equal(uint(2)))
		g.Expect(gchild2.values[0]).To(Equal(item45))
		g.Expect(gchild2.values[1]).To(Equal(item50))
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

		item22 := &twoThreeTester{num: 22, value: "22"}
		treeItem := bTree.FindOrCreate(item22)
		g.Expect(treeItem).To(Equal(item22))

		g.Expect(bTree.root.count).To(Equal(uint(1)))
		g.Expect(bTree.root.values[0]).To(Equal(item25))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.count).To(Equal(uint(1)))
		g.Expect(child1.values[0]).To(Equal(item20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.count).To(Equal(uint(2)))
		g.Expect(gchild1.values[0]).To(Equal(item5))
		g.Expect(gchild1.values[1]).To(Equal(item10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.count).To(Equal(uint(1)))
		g.Expect(gchild2.values[0]).To(Equal(item22))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.count).To(Equal(uint(1)))
		g.Expect(child2.values[0]).To(Equal(item40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.count).To(Equal(uint(1)))
		g.Expect(gchild1.values[0]).To(Equal(item30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.count).To(Equal(uint(2)))
		g.Expect(gchild2.values[0]).To(Equal(item45))
		g.Expect(gchild2.values[1]).To(Equal(item50))
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

		item47 := &twoThreeTester{num: 47, value: "47"}
		treeItem := bTree.FindOrCreate(item47)
		g.Expect(treeItem).To(Equal(item47))

		bTree.root.print("")

		g.Expect(bTree.root.count).To(Equal(uint(1)))
		g.Expect(bTree.root.values[0]).To(Equal(item40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.count).To(Equal(uint(1)))
		g.Expect(child1.values[0]).To(Equal(item20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.count).To(Equal(uint(2)))
		g.Expect(gchild1.values[0]).To(Equal(item5))
		g.Expect(gchild1.values[1]).To(Equal(item10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.count).To(Equal(uint(2)))
		g.Expect(gchild2.values[0]).To(Equal(item25))
		g.Expect(gchild2.values[1]).To(Equal(item30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.count).To(Equal(uint(1)))
		g.Expect(child2.values[0]).To(Equal(item47))

		gchild1 = child2.children[0]
		g.Expect(gchild1.count).To(Equal(uint(1)))
		g.Expect(gchild1.values[0]).To(Equal(item45))

		gchild2 = child2.children[1]
		g.Expect(gchild2.count).To(Equal(uint(1)))
		g.Expect(gchild2.values[0]).To(Equal(item50))
	})
}
