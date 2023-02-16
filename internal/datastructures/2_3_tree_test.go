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
		item8 := &twoThreeTester{num: 8, value: "8"} // left
		//item12 := &twoThreeTester{num: 12, value: "12"} // left
		//item17 := &twoThreeTester{num: 17, value: "17"} // left

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
	})

	t.Run("rightChild promotions", func(t *testing.T) {
		//// generate a tree of
		///*
		// *      20
		// *    /    \
		// *  10    30,35
		// */
		//setupRightTree := func(g *GomegaWithT) *TwoThreeRoot {
		//	bTree, err := NewBTree(2)
		//	g.Expect(err).ToNot(HaveOccurred())

		//	bTree.FindOrCreate(item10)
		//	bTree.FindOrCreate(item20)
		//	bTree.FindOrCreate(item30)
		//	bTree.FindOrCreate(item35)

		//	return bTree
		//}
	})
}
